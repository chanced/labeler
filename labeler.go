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

type process struct {
	containerField *fieldRef
	fields         []fieldRef
	errs           []*FieldError
	opts           *Options
	labels         map[string]string
	numField       int
	mutex          sync.Mutex
}

// Unmarshal parses the labels returned from method GetLabels of labeled.
// There must be a means of setting the labels. The top level type must either
// implement Labeler (by having a SetLabels(map[string]string) method), have a
// tag `labels:"*"` on a field at the top level of v, or by having set LabelsField
// in the options (use the Field option func)
func Unmarshal(input interface{}, v interface{}, opts ...Option) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case error:
				err = e
			case string:
				err = errors.New(e)
			default:
				err = ErrParsing
			}
		}
	}()
	o, optErr := getOptions(opts)
	if optErr != nil {
		return optErr
	}

	var in map[string]string

	switch t := input.(type) {
	case Labeled:
		in = t.GetLabels()
	case map[string]string:
		in = input.(map[string]string)
	case *map[string]string:
		in = *input.(*map[string]string)
	default:
		return ErrInvalidInput
	}

	l := map[string]string{}

	for key, val := range in {
		l[key] = val
	}

	mutex := &sync.Mutex{}

	p, uErr := unmarshalLabels(v, l, mutex, o)
	if uErr != nil {
		return uErr
	}

	var errSettingLabels error
	if p.containerField != nil {
		return p.containerField.set(l, o)
	}
	switch t := v.(type) {
	case GenericLabeler:
		errSettingLabels = t.SetLabels(l, o.Tag)
	case StrictLabeler:
		errSettingLabels = t.SetLabels(l)
	case Labeler:
		t.SetLabels(l)
	default:
		return ErrInvalidValue
	}
	if errSettingLabels != nil {
		return ErrSettingLabels
	}

	return nil
}

func unmarshalLabels(v interface{}, l map[string]string, mutex *sync.Mutex, o *Options) (*process, error) {
	switch t := v.(type) {
	// this should probably occur after fields have been parsed. Leaving it as-is for now.
	case UnmarshalerWithOptions:
		return nil, t.UnmarshalLabels(l, *o)
	case Unmarshaler:
		return nil, t.UnmarshalLabels(l)
	}
	rv := reflect.ValueOf(v)

	kind := rv.Kind()
	if kind != reflect.Ptr || rv.IsNil() {
		return nil, ErrInvalidValue
	}

	rv = rv.Elem()
	kind = rv.Kind()
	if !rv.CanAddr() {
		return nil, ErrInvalidValue
	}

	if kind != reflect.Struct {
		return nil, ErrInvalidValue
	}
	rt := rv.Type()
	numField := rv.NumField()

	fieldCh := make(chan fieldRef, numField)
	errCh := make(chan error, numField)

	for i := 0; i < numField; i++ {
		structField := rt.Field(i)
		valueField := rv.Field(i)
		go getFieldMeta(structField, valueField, fieldCh, errCh, mutex, o)
	}
	p, err := getProcess(numField, fieldCh, errCh)
	close(fieldCh)
	close(errCh)
	if err != nil {
		return nil, err
	}
	if len(p.errs) > 0 {
		return nil, NewParsingError(p.errs)
	}

	fieldErrCh := make(chan *FieldError, len(p.fields))

	wg := &sync.WaitGroup{}
	for _, f := range p.fields {
		wg.Add(1)
		go unmarshalField(f, l, p, wg, fieldErrCh, o)
	}
	go func() {
		wg.Wait()
		close(fieldErrCh)
	}()
	errs := getErrors(fieldErrCh)

	if len(errs) > 0 {
		return nil, NewParsingError(errs)
	}
	if p.containerField != nil {
		err = p.containerField.set(l, o)
		if err != nil {
			var fieldErr *FieldError
			if errors.As(err, &fieldErr) {
				return nil, fieldErr
			}
			return nil, p.containerField.err(err)
		}
	}
	return p, nil
}

func getErrors(errCh <-chan *FieldError) []*FieldError {
	errs := []*FieldError{}
	for err := range errCh {
		errs = append(errs, err)
	}
	return errs
}

func getProcess(numField int, fieldCh <-chan fieldRef, errCh <-chan error) (*process, error) {

	p := &process{
		numField: numField,
		errs:     []*FieldError{},
		fields:   []fieldRef{},
	}
	for i := 0; i < numField; i++ {
		select {
		case f := <-fieldCh:
			if f.IsContainer && p.containerField != nil {
				return p, ErrMultipleContainers
			} else if f.IsContainer {
				p.containerField = &f
			} else if f.IsTagged || f.IsStruct {
				p.fields = append(p.fields, f)
			}
		case err := <-errCh:
			var fe *FieldError
			if errors.As(err, &fe) {
				p.errs = append(p.errs, fe)
			} else {
				return p, err
			}
		}
	}
	return p, nil

}

func handleUnmarshalFieldPanic(f fieldRef, errCh chan<- *FieldError) {
	if err := recover(); err != nil {
		// TODO: add details to the err
		errCh <- f.err(ErrParsing)
	}

}

func unmarshalField(f fieldRef, l map[string]string, pr *process, wg *sync.WaitGroup, errCh chan<- *FieldError, o *Options) {
	defer wg.Done()
	defer handleUnmarshalFieldPanic(f, errCh)
	if f.IsStruct {
		var e *ParsingError
		p, err := unmarshalLabels(f.Interface, l, f.Mutex, o)
		if err != nil {
			if errors.As(err, &e) {
				for _, fErr := range e.Errs {
					errCh <- newFieldErrorFromNested(f, fErr)
				}
			} else {
				errCh <- newFieldError(&f, err)
			}
		} else if p != nil && len(p.errs) != 0 {
			var e *ParsingError
			for _, err := range p.errs {
				if errors.As(err, &e) {
					for _, fErr := range e.Errs {
						errCh <- newFieldErrorFromNested(f, fErr)
					}
				} else {
					errCh <- newFieldError(&f, err)
				}
			}
		} else if p != nil {
			pr.mutex.Lock()
			if p.containerField != nil && pr.containerField != nil {
				errCh <- newFieldError(&f, ErrMultipleContainers)
			} else if p.containerField != nil {
				pr.containerField = p.containerField
			}
			pr.mutex.Unlock()
		}
	} else if f.IsTagged {
		err := f.set(l, o)

		if err != nil {
			var e *FieldError
			if errors.As(err, &e) {
				errCh <- e
			} else {
				errCh <- newFieldError(&f, e)
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
