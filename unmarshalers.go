package labeler

import (
	"reflect"
	"strings"
)

type unmarshalerFunc func(r reflected, o Options) unmarshalFunc
type unmarshalFunc func(r reflected, kvs *keyValues, o Options) error
type unmarshalFieldFunc func(f *field, kvs *keyValues, o Options) error
type unmarshalerFuncs []unmarshalerFunc

func getUnmarshal(r reflected, o Options) unmarshalFunc {
	switch r.Topic() {
	case fieldTopic:
		if r.IsContainer(o) {
			return containerUnmarshalers.Unmarshaler(r, o)
		}
		return fieldUnmarshalers.Unmarshaler(r, o)
	case subjectTopic:
		return subjectUnmarshalers.Unmarshaler(r, o)
	}
	return nil
}

var fieldUnmarshalers = unmarshalerFuncs{
	unmarshalSlice,
	unmarshalArray,
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalFieldPkgString,
	unmarshalFieldStringee,
	unmarshalFieldTextUnmarshaler,
	unmarshalFieldString,
}

var collectionUnmarshalers = unmarshalerFuncs{
	unmarshalFieldPkgString,
	unmarshalFieldStringee,
	unmarshalFieldTextUnmarshaler,
	unmarshalFieldString,
}

var containerUnmarshalers = unmarshalerFuncs{
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalGenericLabelee,
	unmarshalStrictLabelee,
	unmarshalLabelee,
	unmarshalMap,
}

var subjectUnmarshalers = unmarshalerFuncs{
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalGenericLabelee,
	unmarshalStrictLabelee,
	unmarshalLabelee,
}

func (list unmarshalerFuncs) Unmarshaler(r reflected, o Options) unmarshalFunc {
	for _, loader := range list {
		unmarsh := loader(r, o)
		if unmarsh != nil {
			return unmarsh
		}
	}
	return nil
}

func (set unmarshalFieldFunc) Unmarshaler(r reflected, o Options) unmarshalFunc {
	if r.Topic() != fieldTopic {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		return set(r.(*field), kvs, o)
	}
}

var unmarshalMap = func(r reflected, o Options) unmarshalFunc {
	if r.Topic() != fieldTopic || !r.CanSet() || !r.Assignable(mapType) {
		return nil
	}
	var set unmarshalFieldFunc = func(f *field, kvs *keyValues, o Options) error {
		return f.setMap(kvs.Map(), o)
	}
	return set.Unmarshaler(r, o)
}

var unmarshalUnmarshaler = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(unmarshalerType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(Unmarshaler)
		return u.UnmarshalLabels(kvs.Map())
	}
}

var unmarshalUnmarshalerWithOpts = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(unmarshalerWithOptsType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(UnmarshalerWithOpts)
		return u.UnmarshalLabels(kvs.Map(), o)
	}
}

var unmarshalLabelee = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(labeleeType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(Labelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalStrictLabelee = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(strictLabeleeType) {
		return nil
	}

	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(StrictLabelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalGenericLabelee = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(genericLabeleeType) {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		u := r.Interface().(Labelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalFieldStringee = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(stringeeType) {
		return nil
	}
	var fstr fieldStrUnmarshalFunc = func(f *field, s string, o Options) error {
		u := f.Interface().(Stringee)
		u.FromString(s)
		return nil
	}
	return fstr.Unmarshaler(r, o)
}

var unmarshalFieldTextUnmarshaler = func(r reflected, o Options) unmarshalFunc {
	if !r.CanInterface() || !r.Implements(textUnmarshalerType) {
		return nil
	}
	u := r.Interface().(TextUnmarshaler)
	var set fieldStrUnmarshalFunc = func(f *field, s string, o Options) error {
		return u.UnmarshalText([]byte(s))
	}
	return set.Unmarshaler(r, o)
}

var unmarshalFieldPkgString = func(r reflected, o Options) unmarshalFunc {
	if r.Topic() != fieldTopic {
		return nil
	}
	set, ok := fieldStringeePkgs[r.PkgPath()][r.TypeName()]
	if !ok {
		return nil
	}
	return set.Unmarshaler(r, o)
}

var unmarshalFieldString = func(r reflected, o Options) unmarshalFunc {
	if !r.CanSet() {
		return nil
	}
	if fss, ok := fieldStringeeBasic[r.Kind()]; ok {
		return fss.Unmarshaler(r, o)
	}
	return nil
}

func splitFieldValue(f *field, kvs *keyValues, o Options) ([]string, bool) {
	kv, ok := kvs.Get(f.key, f.ignoreCase(o))
	var s string
	switch {
	case ok:
		s = kv.Value
	case f.HasDefault(o):
		s = f.Default(o)
	case f.OmitEmpty(o):
		return nil, false
	}
	strs := strings.Split(s, f.split(o))
	return strs, true
}

func unmarshalElem(f *field, rv reflect.Value, s string, fn unmarshalFunc, o Options) error {
	f.SetValue(rv)
	isPtr := f.deref()
	nkvs := newKeyValues()

	nkvs.Set(f.key, s)

	if err := fn(f, &nkvs, o); err != nil {
		return err
	}

	if isPtr {
		f.PtrValue().Set(f.Value().Addr())
	}

	return nil
}

var unmarshalArray = func(r reflected, o Options) unmarshalFunc {
	if !r.IsArray() || r.ColType().Implements(stringeeType) || r.IsElem() {
		return nil
	}

	r.SetIsElem(true)
	defer r.SetIsElem(false)

	r.PrepCollection()
	defer r.ResetCollection()

	fn := collectionUnmarshalers.Unmarshaler(r, o)
	if fn == nil {
		return nil
	}

	return func(r reflected, kvs *keyValues, o Options) error {
		r.PrepCollection()
		defer r.ResetCollection()

		f := r.(*field)
		strs, hasVal := splitFieldValue(f, kvs, o)
		if !hasVal {
			return nil
		}
		for i, s := range strs {
			if i >= f.len {
				break
			}
			rv := f.ColValue().Index(i)
			if err := unmarshalElem(f, rv, s, fn, o); err != nil {
				return err
			}
		}

		return nil
	}
}

var unmarshalSlice = func(r reflected, o Options) unmarshalFunc {
	if !r.IsSlice() || r.ColType().Implements(stringeeType) || r.IsElem() {
		return nil
	}
	r.SetIsElem(true)
	defer r.SetIsElem(false)

	fn := collectionUnmarshalers.Unmarshaler(r, o)
	if fn == nil {
		return nil
	}

	return func(r reflected, kvs *keyValues, o Options) error {
		defer r.ResetCollection()

		f := r.(*field)
		strs, hasVal := splitFieldValue(f, kvs, o)
		if !hasVal {
			return nil
		}
		for _, s := range strs {
			rv := reflect.New(r.Type()).Elem()
			if err := unmarshalElem(f, rv, s, fn, o); err != nil {
				return err
			}
			r.ColValue().Set(reflect.Append(r.ColValue(), r.Value()))
		}

		return nil
	}
}

type fieldStrUnmarshalFunc func(f *field, s string, o Options) error

func (setStr fieldStrUnmarshalFunc) Unmarshaler(r reflected, o Options) unmarshalFunc {
	if r.Topic() != fieldTopic {
		return nil
	}
	return func(r reflected, kvs *keyValues, o Options) error {
		f := r.(*field)
		kv, ok := kvs.Get(f.key, f.ignoreCase(o))
		var s string
		switch {
		case ok:
			s = kv.Value
		case f.HasDefault(o):
			s = f.Default(o)
		case f.OmitEmpty(o):
			return nil
		}
		f.wasSet = true
		return setStr(f, s, o)
	}
}

var fieldStringeePkgs = map[string]map[string]fieldStrUnmarshalFunc{
	"time": {
		"Time": func(f *field, s string, o Options) error {
			return f.setTime(s, o)
		},
		"Duration": func(f *field, s string, o Options) error {
			return f.setDuration(s, o)
		},
	},
}

var fieldStringeeBasic = map[reflect.Kind]fieldStrUnmarshalFunc{
	reflect.Bool: func(f *field, s string, o Options) error {
		return f.setBool(s, o)
	},
	reflect.Float64: func(f *field, s string, o Options) error {
		return f.setFloat(s, 64, o)
	},
	reflect.Float32: func(f *field, s string, o Options) error {
		return f.setFloat(s, 32, o)
	},
	reflect.Int: func(f *field, s string, o Options) error {
		return f.setInt(s, 0, o)
	},
	reflect.Int8: func(f *field, s string, o Options) error {
		return f.setInt(s, 8, o)
	},
	reflect.Int16: func(f *field, s string, o Options) error {
		return f.setInt(s, 16, o)
	},
	reflect.Int32: func(f *field, s string, o Options) error {
		return f.setInt(s, 32, o)
	},
	reflect.Int64: func(f *field, s string, o Options) error {
		return f.setInt(s, 64, o)
	},
	reflect.String: func(f *field, s string, o Options) error {
		return f.setString(s, o)
	},
	reflect.Uint: func(f *field, s string, o Options) error {
		return f.setUint(s, 0, o)
	},
	reflect.Uint8: func(f *field, s string, o Options) error {
		return f.setUint(s, 8, o)
	},
	reflect.Uint16: func(f *field, s string, o Options) error {
		return f.setUint(s, 16, o)
	},
	reflect.Uint32: func(f *field, s string, o Options) error {
		return f.setUint(s, 32, o)
	},
	reflect.Uint64: func(f *field, s string, o Options) error {
		return f.setUint(s, 64, o)
	},
	reflect.Complex64: func(f *field, s string, o Options) error {
		return f.setComplex(s, 64, o)
	},
	reflect.Complex128: func(f *field, s string, o Options) error {
		return f.setComplex(s, 128, o)
	},
}
