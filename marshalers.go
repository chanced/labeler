package labeler

import (
	"fmt"
	"reflect"
	"strings"
)

type marshalerFunc func(r reflected, o Options) marshalFunc
type marshalerFuncs []marshalerFunc
type marshalFunc func(r reflected, kvs *keyValues, o Options) error
type marshalFieldFunc func(f *field, kvs *keyValues, o Options) error

func getMarshal(r reflected, o Options) marshalFunc {
	switch r.Topic() {
	case fieldTopic:
		if r.IsContainer(o) {
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

var fieldMarshalers = marshalerFuncs{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalFieldPkgString,
	marshalFieldStringer,
	marshalArrayOrSlice,
	marshalFieldTextMarshaler,
	marshalFieldString,
}

var collectionMarshalers = marshalerFuncs{
	marshalFieldPkgString,
	marshalFieldStringer,
	marshalFieldTextMarshaler,
	marshalFieldString,
}

var containerMarshalers = marshalerFuncs{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalGenericallyLabeled,
	marshalLabeled,
	marshalMap,
}

var subjectMarshalers = marshalerFuncs{
	marshalMarshalerWithOpts,
	marshalMarshaler,
	marshalGenericallyLabeled,
	marshalLabeled,
}

var inputMarshalers = marshalerFuncs{
	marshalGenericallyLabeled,
	marshalLabeled,
	marshalMap,
}

func (list marshalerFuncs) Marshaler(r reflected, o Options) marshalFunc {
	for _, loader := range list {
		marsh := loader(r, o)
		if marsh != nil {
			return marsh
		}
	}
	return nil
}

var marshalMarshalerWithOpts = func(r reflected, o Options) marshalFunc {
	if !r.CanInterface() || !r.Implements(marshalerWithOptsType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(MarshalerWithOpts)
		m, err := u.MarshalLabels(o)
		if err != nil {
			return err
		}
		kvs.Add(m)
		return nil
	}
}

var marshalMarshaler = func(r reflected, o Options) marshalFunc {
	if !r.CanInterface() || !r.Implements(marshalerType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(Marshaler)
		m, err := u.MarshalLabels()
		if err != nil {
			return err
		}
		kvs.Add(m)
		return nil
	}
}

var marshalFieldPkgString = func(r reflected, o Options) marshalFunc {
	if r.Topic() != fieldTopic {
		return nil
	}
	pkg, ok := fieldStringerPkgs[r.PkgPath()][r.TypeName()]
	if !ok {
		return nil
	}
	return pkg.Marshaler(r, o)
}

var marshalFieldStringer = func(r reflected, o Options) marshalFunc {
	if !r.CanInterface() || !r.Implements(stringerType) {
		return nil
	}
	var fstr fieldStringer = func(f *field, o Options) (string, error) {
		u := r.Interface().(fmt.Stringer)
		return u.String(), nil
	}
	return fstr.Marshaler(r, o)
}

var marshalFieldTextMarshaler = func(r reflected, o Options) marshalFunc {
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

var marshalFieldString = func(r reflected, o Options) marshalFunc {
	strGetter := fieldStringerBasic[r.Kind()]
	if strGetter == nil {
		return nil
	}
	return strGetter.Marshaler(r, o)
}

var marshalArrayOrSlice = func(r reflected, o Options) marshalFunc {
	if (!r.IsArray() && !r.IsSlice()) || r.ColType().Implements(stringeeType) || r.IsElem() {
		return nil
	}
	r.SetIsElem(true)
	defer r.SetIsElem(false)

	fn := collectionMarshalers.Marshaler(r, o)

	if fn == nil {
		return nil
	}

	return func(r reflected, kvs *keyValues, o Options) error {
		defer r.SetValue(r.ColValue())

		f := r.(*field)
		strs := []string{}

		for i := 0; i < r.Len(); i++ {
			r.SetValue(r.ColValue().Index(i))
			if r.deref() {
				r.PtrValue().Set(r.Value().Addr())
			}
			nkvs := newKeyValues()

			if err := fn(r, &nkvs, o); err != nil {
				return err
			}

			if v, ok := nkvs.Get(f.key, f.ignoreCase(o)); ok {
				strs = append(strs, v.Value)
			}
		}

		kvs.Set(f.key, strings.Join(strs, f.split(o)))

		return nil
	}
}

var marshalGenericallyLabeled = func(r reflected, o Options) marshalFunc {
	if !r.CanInterface() || !r.Implements(genericallyLabeledType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(GenericallyLabeled)
		m := u.GetLabels(o.Tag)
		kvs.Add(m)
		return nil
	}
}
var marshalLabeled = func(r reflected, o Options) marshalFunc {
	if !r.CanInterface() || !r.Implements(labeledType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(Labeled)
		m := u.GetLabels()
		kvs.Add(m)
		return nil
	}
}

var marshalMap = func(r reflected, o Options) marshalFunc {
	if !r.Assignable(mapType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
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

type fieldStringer func(f *field, o Options) (string, error)

func (get fieldStringer) Marshaler(r reflected, o Options) marshalFunc {
	if r.Topic() != fieldTopic {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		f := r.(*field)
		s, err := get(f, o)
		if err != nil {
			return err
		}
		if s == "" {
			s = f.Default(o)
		}
		if o.OmitEmpty && s == "" {
			return nil
		}
		kvs.Set(f.key, s)
		return nil
	}
}

var fieldStringerPkgs = map[string]map[string]fieldStringer{
	"time": {
		"Time": func(f *field, o Options) (string, error) {
			return f.formatTime(o)
		},
		"Duration": func(f *field, o Options) (string, error) {
			return f.formatDuration(o)
		},
	},
}

var fieldStringerBasic = map[reflect.Kind]fieldStringer{
	reflect.Bool: func(f *field, o Options) (string, error) {
		return f.formatBool(o)
	},
	reflect.Float64: func(f *field, o Options) (string, error) {
		return f.formatFloat(o)
	},
	reflect.Float32: func(f *field, o Options) (string, error) {
		return f.formatFloat(o)
	},
	reflect.Int: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int8: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int16: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int32: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int64: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.String: func(f *field, o Options) (string, error) {
		return f.formatString(o)
	},
	reflect.Uint: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint8: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint16: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint32: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint64: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Complex64: func(f *field, o Options) (string, error) {
		return f.formatComplex(o)
	},
	reflect.Complex128: func(f *field, o Options) (string, error) {
		return f.formatComplex(o)
	},
}
