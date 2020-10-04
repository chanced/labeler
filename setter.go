package labeler

import "reflect"

type isetter = func(r reflected, o Options) setter
type setter func(r reflected, kvs keyvalueSet, o Options) error
type fieldSetter func(f *field, kvs keyvalueSet, o Options) error
type fieldStrSetter func(f *field, s string, o Options) error

func (fs fieldSetter) setter(r reflected, o Options) setter {
	if r.topic() != fieldTopic {
		return nil
	}
	var f setter = func(r reflected, kvs keyvalueSet, o Options) error {
		f := r.(*field)
		return fs(f, kvs, o)
	}
	return f
}

func (fss fieldStrSetter) setter(r reflected, o Options) setter {
	var fs fieldSetter = func(f *field, kvs keyvalueSet, o Options) error {
		kv, ok := kvs.Get(f.Key, f.ignoreCase(o))
		if o.OmitEmpty && !ok {
			return nil
		}
		return fss(f, kv.Value, o)
	}
	return fs.setter(r, o)
}

var getFieldSetter isetter = func(r reflected, o Options) setter {

	pkgsetter := getPkgFieldSetter(r, o)
}

var getPkgieldSetter isetter = func(r reflected, o Options) setter {
	if r.topic() != fieldTopic {
		return nil
	}
	m := r.Meta()
	if pset, ok := pkgFieldSetters[m.PkgPath]; ok {
		if fss, ok := pset[m.TypeName]; ok {
			return fss.setter(r, o)
		}
	}
	return nil
}

type fieldStrSetters = map[string]fieldStrSetter
type pkgSetters = map[string]fieldStrSetters

var pkgFieldSetters pkgSetters = pkgSetters{
	"time": fieldStrSetters{
		"Time": func(f *field, s string, o Options) error {
			return f.setTime(s, o)
		},
		"Duration": func(f *field, s string, o Options) error {
			return f.setDuration(s, o)
		},
	},
}

var basicFieldSetters = map[reflect.Kind]fieldStrSetter{
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
