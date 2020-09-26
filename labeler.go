package labeler

import (
	"errors"
	"reflect"
	"sync"
)

// Labeler is implemented by any type with a SetLabels method, which
// accepts map[string]string and handles assignment of those values.
type Labeler interface {
	SetLabels(labels map[string]string)
}

// StrictLabeler is implemented by types with a SetLabels method, which accepts
// map[string]string and handles assignment of those values, returning error if
// there was an issue assigning the value.
type StrictLabeler interface {
	SetLabels(labels map[string]string) error
}

// GenericLabeler is implemented by any type with a SetLabels method, which
// accepts map[string]string and handles assignment of those values.
type GenericLabeler interface {
	SetLabels(labels map[string]string, tag string) error
}

// Stringee is implemented by any value that has a FromString method,
// which parses the “native” format for that value from a string and
// returns a bool value to indicate success (true) or failure (false)
// of parsing.
// Use StringeeStrict if returning an error is preferred.
type Stringee interface {
	FromString(s string) error
}

// Labeled is the interface implemented by types with a method GetLabels,
// which returns a map[string]string of labels and values
type Labeled interface {
	GetLabels() map[string]string
}

// Unmarshaler is implemented by any type that has the method UnmarshalLabels,
// providing a means of unmarshaling map[string]string themselves.
type Unmarshaler interface {
	UnmarshalLabels(v map[string]string) error
}

// UnmarshalerWithOptions is implemented by any type that has the method
// UnmarshalLabels, providing a means of unmarshaling map[string]string
// that also accepts ...Option themselves.
type UnmarshalerWithOptions interface {
	UnmarshalLabels(v map[string]string, opts Options) error
}

// Marshaler is the interface implemented by types that can marshal a value
// into map[string]string
type Marshaler interface {
	MarshalLabels() (map[string]string, error)
}

type fieldsToProcess struct {
	containerField *fieldRef
	fields         []fieldRef
	errs           []*FieldError
	numField       int
}

// Unmarshal parses the labels returned from method GetLabels of labeled.
// There must be a means of setting the labels. The top level type must either
// implement Labeler (by having a SetLabels(map[string]string) method), have a
// tag `labels:"*"` on a field at the top level of v, or by having set LabelsField
// in the options (use the Field option func)
func Unmarshal(input interface{}, v interface{}, opts ...Option) error {

	o, err := getOptions(opts)
	if err != nil {
		return err
	}

	var in map[string]string

	switch t := input.(type) {
	case Labeled:
		in = t.GetLabels()
	case map[string]string:
		in = input.(map[string]string)
	default:
		return ErrInvalidInput
	}

	l := map[string]string{}

	for key, val := range in {
		l[key] = val
	}
	switch t := v.(type) {
	// this should probably occur after fields have been parsed. Leaving it as-is for now.
	case UnmarshalerWithOptions:
		return t.UnmarshalLabels(l, *o)
	case Unmarshaler:
		return t.UnmarshalLabels(l)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidValue
	}
	rv = rv.Elem()
	if !rv.CanAddr() {
		return ErrInvalidValue
	}
	rvi := rv.Addr().Interface()

	if rv.Kind() != reflect.Struct {
		return ErrInvalidValue
	}
	t := rv.Type()

	numField := rv.NumField()

	fieldCh := make(chan fieldRef, numField)
	errCh := make(chan error, numField)
	mutex := &sync.Mutex{}

	for i := 0; i < numField; i++ {
		structField := t.Field(i)
		valueField := rv.Field(i)
		go getFieldMeta(structField, valueField, fieldCh, errCh, mutex, o)
	}
	p, err := getFieldsToProcess(numField, fieldCh, errCh)
	close(fieldCh)
	close(errCh)
	if err != nil {
		return err
	}
	if len(p.errs) > 0 {
		return NewParsingError(p.errs)
	}

	fieldErrCh := make(chan *FieldError, len(p.fields))

	wg := &sync.WaitGroup{}
	for _, f := range p.fields {
		wg.Add(1)
		go unmarshalField(f, l, wg, fieldErrCh, o)
	}
	go func() {
		wg.Wait()
		close(fieldErrCh)
	}()
	errs := getErrors(fieldErrCh)

	if len(errs) > 0 {
		return NewParsingError(errs)
	}
	if p.containerField != nil {
		err = p.containerField.set(l, *o)
		if err != nil {
			var fe *FieldError
			if errors.As(err, &fe) {
				return fe
			}
			return p.containerField.err(err)
		}
		return nil
	}

	var errSettingLabels error
	switch t := rvi.(type) {
	case GenericLabeler:
		errSettingLabels = t.SetLabels(l, o.Tag)
	case StrictLabeler:
		errSettingLabels = t.SetLabels(l)
	case Labeler:
		t.SetLabels(l)
	default:
		if !o.isNested {
			errSettingLabels = ErrSettingLabels
		}
	}
	if errSettingLabels != nil {
		return ErrSettingLabels
	}

	return nil
}

func getErrors(errCh <-chan *FieldError) []*FieldError {
	errs := []*FieldError{}
	for err := range errCh {
		errs = append(errs, err)
	}
	return errs
}

func getFieldsToProcess(numField int, fieldCh <-chan fieldRef, errCh <-chan error) (fieldsToProcess, error) {

	pf := fieldsToProcess{
		numField: numField,
		errs:     []*FieldError{},
		fields:   []fieldRef{},
	}
	for i := 0; i < numField; i++ {
		select {
		case f := <-fieldCh:
			if f.IsContainer && pf.containerField != nil {
				return pf, ErrMultipleContainers
			} else if f.IsContainer {
				pf.containerField = &f
			} else if f.IsTagged || f.IsStruct {
				pf.fields = append(pf.fields, f)
			}
		case err := <-errCh:
			var fe *FieldError
			if errors.As(err, &fe) {
				pf.errs = append(pf.errs, fe)
			} else {
				return pf, err
			}
		}
	}
	return pf, nil

}

func handleUnmarshalFieldPanic(f fieldRef, errCh chan<- *FieldError) {
	if err := recover(); err != nil {
		// TODO: add details to the err
		errCh <- f.err(ErrParsing)
	}
}

func unmarshalField(f fieldRef, l map[string]string, wg *sync.WaitGroup, errCh chan<- *FieldError, o *Options) {
	defer wg.Done()
	defer handleUnmarshalFieldPanic(f, errCh)
	if f.IsStruct {
		err := Unmarshal(l, f.Interface, isNested(), CopyOptions(o))
		var e *ParsingError
		if err != nil {
			if errors.As(err, &e) {
				for _, fErr := range e.Errs {
					errCh <- newFieldErrorFromNested(f, fErr)
				}
			} else {
				errCh <- newFieldError(&f, err)
			}
		}
	}
}

func getFieldMeta(structField reflect.StructField, valueField reflect.Value, ch chan<- fieldRef, errs chan<- error, mutex *sync.Mutex, o *Options) {
	f, err := newFieldRef(structField, valueField, mutex, o)
	if err != nil {
		errs <- err
		return
	}
	ch <- f
}
