package labeler

import "reflect"

type isetter func(r reflected, o Options) setter
type setter func(kvs keyvalueSet, o Options) error
type fieldSetter func(f *field, kvs keyvalueSet, o Options) error
type fieldStrSetter func(f *field, s string, o Options) error
type pkgFieldStrSetterMap map[string]map[string]fieldStrSetter

type isetterSet []isetter

func (iss isetterSet) getSetter(r reflected, o Options) setter {
	for _, fs := range containerSetters {
		set := fs(r, o)
		if set != nil {
			return set
		}
	}
	return nil
}

var fieldSetters isetterSet = isetterSet{
	unmarshalerWithOptsSetter,
	unmarshalerSetter,
	pkgFieldStrSetters,
	stringeeFieldSetter,
	textUnmarshalerFieldSetter,
	basicFieldStrSetters,
}

var containerSetters isetterSet = isetterSet{
	unmarshalerWithOptsSetter,
	unmarshalerSetter,
	genericLabeleeSetter,
	strictLabeleeSetter,
	labeleeSetter,
	mapFieldSetter,
}

var subjectSetters isetterSet = isetterSet{
	unmarshalerWithOptsSetter,
	unmarshalerSetter,
	genericLabeleeSetter,
	strictLabeleeSetter,
	labeleeSetter,
}

func getSetter(r reflected, o Options) setter {
	switch r.topic() {
	case fieldTopic:
		return getFieldSetter(r, o)
	case subjectTopic:
		return subjectSetters.getSetter(r, o)

	}
	return nil
}

var getFieldSetter isetter = func(r reflected, o Options) setter {
	if r.topic() != fieldTopic {
		return nil
	}
	if r.(*field).IsContainer {
		return containerSetters.getSetter(r, o)
	}
	return fieldSetters.getSetter(r, o)
}

func (fs fieldSetter) setter(r reflected, o Options) setter {
	f, ok := r.(*field)
	if !ok {
		return nil
	}
	var set setter = func(kvs keyvalueSet, o Options) error {
		return fs(f, kvs, o)
	}
	return set
}

func (fss fieldStrSetter) setter(r reflected, o Options) setter {
	var set fieldSetter = func(f *field, kvs keyvalueSet, o Options) error {
		kv, ok := kvs.Get(f.Key, f.ignoreCase(o))
		if o.OmitEmpty && !ok {
			return nil
		}
		return fss(f, kv.Value, o)
	}
	return set.setter(r, o)
}

var mapFieldSetter isetter = func(r reflected, o Options) setter {
	if r.topic() != fieldTopic || !r.Assignable(mapType) {
		return nil
	}
	var set fieldSetter = func(f *field, kvs keyvalueSet, o Options) error {
		return f.setMap(kvs.Map(), o)
	}
	return set.setter(r, o)
}

var unmarshalerSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(unmarshalerType) {
		return nil
	}
	u := m.Interface.(Unmarshaler)
	var set setter = func(kvs keyvalueSet, o Options) error {
		return u.UnmarshalLabels(kvs.Map())
	}
	return set
}

var labeleeSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(labeleeType) {
		return nil
	}
	u := m.Interface.(Labelee)
	var set setter = func(kvs keyvalueSet, o Options) error {
		u.SetLabels(kvs.Map())
		return nil
	}
	return set
}

var strictLabeleeSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(strictLabeleeType) {
		return nil
	}
	u := m.Interface.(StrictLabelee)
	var set setter = func(kvs keyvalueSet, o Options) error {
		u.SetLabels(kvs.Map())
		return nil
	}
	return set
}

var genericLabeleeSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(genericLabeleeType) {
		return nil
	}
	u := m.Interface.(Labelee)
	var set setter = func(kvs keyvalueSet, o Options) error {
		u.SetLabels(kvs.Map())
		return nil
	}
	return set
}

var unmarshalerWithOptsSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(unmarshalerWithOptsType) {
		return nil
	}
	u := m.Interface.(UnmarshalerWithOpts)
	var set setter = func(kvs keyvalueSet, o Options) error {
		return u.UnmarshalLabels(kvs.Map(), o)
	}
	return set
}

var stringeeFieldSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(stringeeType) {
		return nil
	}
	var fs fieldStrSetter = func(f *field, s string, o Options) error {
		u := m.Interface.(Stringee)
		u.FromString(s)
		return nil
	}
	return fs.setter(r, o)
}

var textUnmarshalerFieldSetter isetter = func(r reflected, o Options) setter {
	m := r.Meta()
	if !m.CanInterface() || !r.Implements(textUnmarshalerType) {
		return nil
	}
	var fs fieldStrSetter = func(f *field, s string, o Options) error {
		u := m.Interface.(TextUnmarshaler)
		return u.UnmarshalText([]byte(s))

	}
	return fs.setter(r, o)
}

var pkgFieldStrSetters isetter = func(r reflected, o Options) setter {
	if r.topic() != fieldTopic {
		return nil
	}
	m := r.Meta()
	set, ok := pkgFieldStrSetterLookup[m.PkgPath][m.TypeName]
	if ok {
		return set.setter(r, o)
	}
	return nil
}

var pkgFieldStrSetterLookup pkgFieldStrSetterMap = pkgFieldStrSetterMap{
	"time": {
		"Time": func(f *field, s string, o Options) error {
			return f.setTime(s, o)
		},
		"Duration": func(f *field, s string, o Options) error {
			return f.setDuration(s, o)
		},
	},
}

var basicFieldStrSetters isetter = func(r reflected, o Options) setter {
	if !r.CanSet() {
		return nil
	}
	if fss, ok := basicFieldStrSettersMap[r.Meta().Kind]; ok {
		return fss.setter(r, o)
	}
	return nil
}

var basicFieldStrSettersMap = map[reflect.Kind]fieldStrSetter{
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
