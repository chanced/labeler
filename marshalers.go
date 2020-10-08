package labeler

import (
	"fmt"
)

// need a better name for this.
// marshaler is a func that returns a marshal func
type marshaler func(r reflected, o Options) marshal
type marshalers []marshaler
type marshal func(r reflected, kvs *keyvalues, o Options) error
type marshalField func(f *field, kvs *keyvalues, o Options) error

func getMarshal(r reflected, o Options) marshal {
	switch r.Topic() {
	case fieldTopic:
		if r.IsFieldContainer() {
			return containerMarshalers.Marshaler(r, o)
		}
		return fieldMarshalers.Marshaler(r, o)
	case subjectTopic:
		return subjectMarshalers.Marshaler(r, o)
	case inputTopic:
		return inputMarshalers.Marshaler(r, o)
	}
	return nil
}

var fieldMarshalers marshalers = marshalers{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalFieldPkgString,
	marshalFieldStringer,
	marshalFieldTextMarshaler,
	marshalFieldString,
}

var containerMarshalers marshalers = marshalers{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalGenericallyLabeled,
	marshalLabeled,
	marshalMap,
}

var subjectMarshalers marshalers = marshalers{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalGenericallyLabeled,
	marshalLabeled,
}

var inputMarshalers marshalers = marshalers{
	marshalGenericallyLabeled,
	marshalLabeled,
	marshalMap,
}

func (list marshalers) Marshaler(r reflected, o Options) marshal {
	for _, loader := range list {
		marsh := loader(r, o)
		if marsh != nil {
			return marsh
		}
	}
	return nil
}

var marshalMarshalerWithOpts = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(marshalerWithOptsType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(MarshalerWithOpts)
		m, err := u.MarshalLabels(o)
		if err != nil {
			return err
		}
		kvs.Add(m)
		return nil
	}
}

var marshalMarshaler = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(marshalerType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(Marshaler)
		m, err := u.MarshalLabels()
		if err != nil {
			return err
		}
		kvs.Add(m)
		return nil
	}
}

var marshalFieldPkgString marshaler = func(r reflected, o Options) marshal {
	if r.Topic() != fieldTopic {
		return nil
	}
	pkg, ok := fieldStringerPkgs[r.PkgPath()][r.TypeName()]
	if !ok {
		return nil
	}
	return pkg.Marshaler(r, o)
}

var marshalFieldStringer marshaler = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(stringerType) {
		return nil
	}
	var fstr fieldStringer = func(f *field, o Options) (string, error) {
		u := r.Interface().(fmt.Stringer)
		return u.String(), nil
	}
	return fstr.Marshaler(r, o)
}

var marshalFieldTextMarshaler marshaler = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(textMarshalerType) {
		return nil
	}
	var fstr fieldStringer = func(f *field, o Options) (string, error) {
		u := r.Interface().(TextMarshaler)
		t, err := u.MarshalText()
		return string(t), err
	}
	return fstr.Marshaler(r, o)
}

var marshalFieldString marshaler = func(r reflected, o Options) marshal {
	strGetter := fieldStringerBasic[r.Kind()]
	if strGetter == nil {
		return nil
	}
	return strGetter.Marshaler(r, o)
}

var marshalGenericallyLabeled marshaler = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(genericallyLabeledType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(GenericallyLabeled)
		m := u.GetLabels(o.Tag)
		kvs.Add(m)
		return nil
	}
}
var marshalLabeled marshaler = func(r reflected, o Options) marshal {
	if !r.CanInterface() || !r.Implements(labeledType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(Labeled)
		m := u.GetLabels()
		kvs.Add(m)
		return nil
	}
}

var marshalMap marshaler = func(r reflected, o Options) marshal {
	if !r.Assignable(mapType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		if r.Value().IsNil() {
			return ErrInvalidInput
		}
		iter := r.Value().MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v := iter.Value().String()
			if o.OmitEmpty && v == "" {
				continue
			}
			kvs.Set(k, v)
		}
		return nil
	}
}
