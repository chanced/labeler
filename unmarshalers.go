package labeler

// need a better name for this.
// unmarshaler is a func that returns an unmarshal func
type unmarshaler func(r reflected, o Options) unmarshal
type unmarshal func(r reflected, kvs *keyvalues, o Options) error
type unmarshalField func(f *field, kvs *keyvalues, o Options) error
type unmarshalers []unmarshaler

func getUnmarshal(r reflected, o Options) unmarshal {
	switch r.Topic() {
	case fieldTopic:
		if r.IsFieldContainer() {
			return containerUnmarshalers.Unmarshaler(r, o)
		}
		return fieldUnmarshalers.Unmarshaler(r, o)
	case subjectTopic:
		return subjectUnmarshalers.Unmarshaler(r, o)
	}
	return nil
}

var fieldUnmarshalers unmarshalers = unmarshalers{
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalFieldPkgString,
	unmarshalFieldStringee,
	unmarshalFieldTextUnmarshaler,
	unmarhsalFieldString,
}

var containerUnmarshalers unmarshalers = unmarshalers{
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalGenericLabelee,
	unmarshalStrictLabelee,
	unmarshalLabelee,
	unmarshalMap,
}

var subjectUnmarshalers unmarshalers = unmarshalers{
	unmarshalUnmarshalerWithOpts,
	unmarshalUnmarshaler,
	unmarshalGenericLabelee,
	unmarshalStrictLabelee,
	unmarshalLabelee,
}

func (list unmarshalers) Unmarshaler(r reflected, o Options) unmarshal {
	for _, loader := range list {
		unmarsh := loader(r, o)
		if unmarsh != nil {
			return unmarsh
		}
	}
	return nil
}

func (set unmarshalField) Unmarshaler(r reflected, o Options) unmarshal {
	if r.Topic() != fieldTopic {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		return set(r.(*field), kvs, o)
	}
}

var unmarshalMap unmarshaler = func(r reflected, o Options) unmarshal {
	if r.Topic() != fieldTopic || !r.Assignable(mapType) {
		return nil
	}
	var set unmarshalField = func(f *field, kvs *keyvalues, o Options) error {
		return f.setMap(kvs.Map(), o)
	}
	return set.Unmarshaler(r, o)
}

var unmarshalUnmarshaler unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(unmarshalerType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(Unmarshaler)
		return u.UnmarshalLabels(kvs.Map())
	}
}

var unmarshalUnmarshalerWithOpts unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(unmarshalerWithOptsType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(UnmarshalerWithOpts)
		return u.UnmarshalLabels(kvs.Map(), o)
	}
}

var unmarshalLabelee unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(labeleeType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(Labelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalStrictLabelee unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(strictLabeleeType) {
		return nil
	}

	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(StrictLabelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalGenericLabelee unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(genericLabeleeType) {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		u := r.Interface().(Labelee)
		u.SetLabels(kvs.Map())
		return nil
	}
}

var unmarshalFieldStringee unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(stringeeType) {
		return nil
	}
	var fstr fieldStringee = func(f *field, s string, o Options) error {
		u := r.Interface().(Stringee)
		u.FromString(s)
		return nil
	}
	return fstr.Unmarshaler(r, o)
}

var unmarshalFieldTextUnmarshaler unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanInterface() || !r.Implements(textUnmarshalerType) {
		return nil
	}
	u := r.Interface().(TextUnmarshaler)
	var set fieldStringee = func(f *field, s string, o Options) error {
		return u.UnmarshalText([]byte(s))
	}
	return set.Unmarshaler(r, o)
}

var unmarshalFieldPkgString unmarshaler = func(r reflected, o Options) unmarshal {
	if r.Topic() != fieldTopic {
		return nil
	}
	set, ok := fieldStringeePkgs[r.PkgPath()][r.TypeName()]
	if !ok {
		return nil
	}
	return set.Unmarshaler(r, o)
}

var unmarhsalFieldString unmarshaler = func(r reflected, o Options) unmarshal {
	if !r.CanSet() {
		return nil
	}
	if fss, ok := fieldStringeeBasic[r.Kind()]; ok {
		return fss.Unmarshaler(r, o)
	}
	return nil
}
