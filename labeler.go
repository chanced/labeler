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

// UnmarshalerWithOptions is implemented by any type that has the method
// UnmarshalLabels, providing a means of unmarshaling map[string]string
// that also accepts ...Option themselves.
type UnmarshalerWithOptions interface {
	UnmarshalLabels(v map[string]string, opts Options) error
}

// Marshaler is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type Marshaler interface {
	MarshalLabels() (map[string]string, error)
}

// MarshalerWithOptions is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type MarshalerWithOptions interface {
	MarshalLabels(o Options) (map[string]string, error)
}

// Unmarshal parses labels and unmarshals them into v. The input can
// be either implement Labeled with GetLabels() map[string]string or
// GetLabels(tag string) map[string]string or simply be a map[string]string.
// There must be a way of setting labels in the form of map[string]string.
// v can implement Labeler by having a SetLabels(map[string]string),
// implement GenericLabeler by having a SetLabels(map[string], tag string)
// method, have a tag `labels:"*"` (note: "*" and "labels" are both
// configurable) on a field somewhere in within v, or by setting the option
// ContainerField either with UseContainerField(f string) or a custom Option
func Unmarshal(input interface{}, v interface{}, opts ...Option) error {

	o, optErr := newOptions(opts)
	if optErr != nil {
		return optErr
	}
	lbl, err := newLabeler(v, o)
	if err != nil {
		return newInvalidValueErrorForUnmarshaling(o)

	}
	err = lbl.initFields()
	if err != nil {
		return err
	}

	err = lbl.unmarshal(input)
	if err != nil {
		return err
	}
	return nil
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
	Labels  map[string]string
	Fields  fields
	RValue  reflect.Value
	RType   reflect.Type
	RKind   reflect.Kind
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

	return lbl.getRefType().NumField()
}

func (lbl labeler) isStruct() bool {
	return true
}

type fields struct {
	Tagged    []field
	Container *field
}

func (f *fields) setContainerLabels(v interface{}, l map[string]string, o Options) error {
	var errSettingLabels error
	if f.Container != nil {
		return f.Container.set(l, o)
	}
	switch t := v.(type) {
	case GenericLabeler:
		errSettingLabels = t.SetLabels(l, o.Tag)
	case StrictLabeler:
		errSettingLabels = t.SetLabels(l)
	case Labeler:
		t.SetLabels(l)
	default:
		errSettingLabels = ErrInvalidValue
	}
	if errSettingLabels != nil {
		return ErrSettingLabels
	}
	return nil
}

func newFields() fields {
	return fields{
		Tagged: []field{},
	}
}

func newLabeler(v interface{}, o Options) (labeler, error) {
	lbl := labeler{
		Value:   v,
		Options: o,
		Labels:  map[string]string{},
	}

	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if rv.IsNil() || kind != reflect.Ptr {
		return lbl, ErrInvalidInput
	}
	rv = rv.Elem()
	kind = rv.Kind()
	lbl.RValue = rv
	lbl.RKind = kind
	lbl.RType = rv.Type()
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
	processing := true
	for processing {
		select {
		case f := <-ch.fieldCh:
			switch {
			case f.IsContainer:
				if containerField != nil {
					// TODO: Sure this up.
					if containerField.Name != f.Name {
						return ErrMultipleContainers
					}
				} else if containerField == nil {
					containerField = &f
					lbl.Options.SetFromTag(f.Tag)
				}

			case f.IsTagged:
				tagged = append(tagged, f)
			}
		case err := <-ch.errCh:
			var fieldErr *FieldError
			if errors.As(err, &fieldErr) {
				errs = append(errs, fieldErr)
			} else {
				return err
			}
		case <-ch.doneCh:
			processing = false
		}
	}
	if len(errs) > 0 {
		return NewParsingError(errs)
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

func (lbl *labeler) setLabels(input interface{}) error {
	var in map[string]string
	switch t := input.(type) {
	case GenericallyLabeled:
		in = t.GetLabels(lbl.Options.Tag)
	case Labeled:
		in = t.GetLabels()
	case map[string]string:
		in = input.(map[string]string)
	default:
		return ErrInvalidInput
	}
	if in == nil {
		in = map[string]string{}
	}
	for key, val := range in {
		lbl.Labels[key] = val
	}

	return nil
}

func (lbl *labeler) unmarshal(input interface{}) error {
	err := lbl.setLabels(input)
	if err != nil {
		return err
	}
	return nil
}
