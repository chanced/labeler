package labeler

import (
	"errors"
	"reflect"
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
	SetLabels(labels map[string]string, token string) error
}

// Labeled is implemented by types with a method GetLabels, which returns
// a map[string]string of labels and values
type Labeled interface {
	GetLabels() map[string]string
}

// GenericallyLabeled is implemented by types with a method GetLabels, which
// accepts a string and returns a map[string]string of labels and values
type GenericallyLabeled interface {
	GetLabels(t string) map[string]string
}

// Unmarshaler is implemented by any type that has the method UnmarshalLabels,
// providing a means of unmarshaling map[string]string themselves.
type Unmarshaler interface {
	UnmarshalLabels(v map[string]string) error
}

// UnmarshalerWithOpts is implemented by any type that has the method
// UnmarshalLabels, providing a means of unmarshaling map[string]string
// that also accepts Options.
type UnmarshalerWithOpts interface {
	UnmarshalLabels(v map[string]string, opts Options) error
}

// UnmarshalerWithTagAndOpts providing a means of unmarshaling map[string]string
// that also accepts Tag and Options.
type UnmarshalerWithTagAndOpts interface {
	UnmarshalLabels(v map[string]string, t Tag, opts Options) error
}

// Marshaler is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type Marshaler interface {
	MarshalLabels() (map[string]string, error)
}

// MarshalerWithOpts is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type MarshalerWithOpts interface {
	MarshalLabels(o Options) (map[string]string, error)
}

// MarshalerWithTagAndOptions is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type MarshalerWithTagAndOptions interface {
	MarshalLabels(t Tag, o Options) (map[string]string, error)
}

// Stringee is implemented by any value that has a FromString method,
// which parses the “native” format for that value from a string and
// returns a bool value to indicate success (true) or failure (false)
// of parsing.
// Use StringeeStrict if returning an error is preferred.
type Stringee interface {
	FromString(s string) error
}

// Unmarshal parses labels and unmarshals them into v. See README.md for
// available options for input and v.
func Unmarshal(input interface{}, v interface{}, opts ...Option) error {
	o, err := newOptions(opts)
	if err != nil {
		return err
	}
	lbl, err := newLabeler(v, o)
	if err != nil {
		return newInvalidValueErrorForUnmarshaling(o)
	}
	return lbl.unmarshal(input)
}

// Marshal parses v, pulling the values from tagged fields (default: "label") as
// well as any labels returned from GetLabels(), GetLabels(tag string), or the
// value of the ContainerField, set either with a tag `labels:"*"` (note: both "*"
// and labels are configurable). Tagged field values take precedent over these
// values as they are just present to ensure that all labels are stored, regardless
// of unmarshaling.
func Marshal(v interface{}, opts ...Option) (l map[string]string, err error) {
	return nil, nil

}

type fields struct {
	Tagged    []field
	Container *field
}

type keyValue struct {
	Key   string
	Value string
}

type reflected interface {
	getRefKind() reflect.Kind
	getRefType() reflect.Type
	getRefValue() reflect.Value
	isStruct() bool
	getRefNumField() int
}

type labeler struct {
	Value   interface{}
	Options Options
	// Labels  map[string]string
	Fields fields
	RValue reflect.Value
	RType  reflect.Type
	RKind  reflect.Kind
}

func (lbl *labeler) unmarshal(input interface{}) error {
	err := lbl.initFields()
	if err != nil {
		return err
	}

	labels, err := lbl.getLabels(input)
	if err != nil {
		return err
	}
	errs := []*FieldError{}
	if err != nil {
		return err
	}
	for _, f := range lbl.Fields.Tagged {
		err = f.set(labels, lbl.Options)
		var fieldErr *FieldError
		if err != nil {
			if errors.As(err, &fieldErr) {
				errs = append(errs, fieldErr)
			} else {
				errs = append(errs, f.err(err))
			}
		} else if !f.Keep {
			delete(labels, f.Key)
		}
	}
	if len(errs) > 0 {
		return NewParsingError(errs)
	}
	return nil
}

// func (f *labeler) setContainerLabels(v interface{}, l map[string]string, o Options) error {
// var errSettingLabels error
// if f.Container != nil {
// 	return f.Container.set(l, o)
// }
// switch t := v.(type) {
// case GenericLabeler:
// 	errSettingLabels = t.SetLabels(l, o.Tag)
// case StrictLabeler:
// 	errSettingLabels = t.SetLabels(l)
// case Labeler:
// 	t.SetLabels(l)
// default:
// 	errSettingLabels = ErrInvalidValue
// }
// if errSettingLabels != nil {
// 	return ErrSettingLabels
// }
// return nil
// }

func newFields() fields {
	return fields{
		Tagged: []field{},
	}
}

func newLabeler(v interface{}, o Options) (labeler, error) {
	lbl := labeler{
		Value:   v,
		Options: o,
		// Labels:  map[string]string{},
	}

	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if rv.IsNil() || kind != reflect.Ptr {
		return lbl, ErrInvalidInput
	}
	rv = rv.Elem()
	kind = rv.Kind()
	rt := rv.Type()
	lbl.RValue = rv
	lbl.RKind = kind
	lbl.RType = rt
	if !rv.CanAddr() || kind != reflect.Struct {
		return lbl, ErrInvalidValue
	}

	return lbl, nil
}

func (lbl *labeler) initFields() error {
	ch := newChannels(lbl, lbl.Options)
	go ch.processFields()
	errs := []*FieldError{}
	tagged := []field{}
	var containerField *field
	fieldCh := ch.fieldCh
	errCh := ch.errCh
	for fieldCh != nil || errCh != nil {
		select {
		case f, ok := <-fieldCh:
			if !ok {
				fieldCh = nil
				break
			}
			switch {
			case f.IsContainer && containerField == nil:
				containerField = &f
			case f.IsContainer && containerField.Name != "" && containerField.Name == f.Name:
				return ErrMultipleContainers
			case f.IsTagged:
				tagged = append(tagged, f)
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				break
			}
			var fieldErr *FieldError
			if errors.As(err, &fieldErr) {
				errs = append(errs, fieldErr)
			} else {
				return err
			}
		}
	}
	if len(errs) > 0 {
		return NewParsingError(errs)
	}
	if containerField != nil {
		lbl.Options.SetFromTag(containerField.Tag)
	}
	lbl.Fields.Container = containerField
	lbl.Fields.Tagged = tagged
	return nil
}

func handleFieldPanic(f field, errCh chan<- *FieldError) {
	if err := recover(); err != nil {
		errCh <- f.err(ErrParsing)
	}
}

func (lbl *labeler) getLabels(input interface{}) (map[string]string, error) {
	var l map[string]string
	container := lbl.Fields.Container
	var target interface{}
	if container != nil {
		target = container.Interface
	} else {
		target = input
	}
	switch t := target.(type) {
	case GenericallyLabeled:
		l = t.GetLabels(lbl.Options.Tag)
	case Labeled:
		l = t.GetLabels()
	case map[string]string:
		l = input.(map[string]string)
	default:
		return nil, ErrInvalidInput
	}
	if l == nil {
		l = map[string]string{}
	}
	return l, nil
}

func (lbl *labeler) setLabels(l map[string]string) error {
	o := lbl.Options
	container := lbl.Fields.Container
	if container != nil {
		err := container.set(l, o)
		if err != nil {
			return ErrSettingLabels
		}
		return nil
	}
	var err error
	switch t := lbl.Value.(type) {
	case GenericLabeler:
		err = t.SetLabels(l, o.Tag)
	case StrictLabeler:
		err = t.SetLabels(l)
	case Labeler:
		t.SetLabels(l)
	default:
		err = ErrSettingLabels
	}
	if err != nil {
		return ErrSettingLabels
	}

	return nil
}

func (lbl labeler) getRefKind() reflect.Kind {
	return lbl.RKind
}
func (lbl labeler) getRefType() reflect.Type {
	return lbl.RType
}
func (lbl labeler) getRefValue() reflect.Value {
	return lbl.RValue
}

func (lbl labeler) getRefNumField() int {

	return lbl.RType.NumField()
}

func (lbl labeler) isStruct() bool {
	return true
}
