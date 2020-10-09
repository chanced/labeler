// Package labeler marshals and unmarshals map[string]string utilizing struct tags.
package labeler

// Labeler Marshals and Unmarshals map[string]string based on struct tags and options
type Labeler struct {
	options Options
}

// NewLabeler returns a new Labeler instance based upon Options (if any) provided.
func NewLabeler(opts ...Option) Labeler {
	o := newOptions(opts)
	lbl := Labeler{
		options: o,
	}
	return lbl
}

// ValidateOptions checks the options provided
func (lbl Labeler) ValidateOptions() error {
	return lbl.options.Validate()
}

//Unmarshal input into v using the Options provided to Labeler
func (lbl *Labeler) Unmarshal(input interface{}, v interface{}) error {
	o := lbl.options
	sub, err := newSubject(v, o)
	if err != nil {
		return err
	}
	kvs := newKeyvalues()
	in, err := newInput(input, o)
	if err != nil {
		return err
	}
	err = in.Marshal(&kvs, o)
	if err != nil {
		return err
	}
	return sub.Unmarshal(&kvs, o)
}

// Marshal v into map[string]string using the Options provided to Labeler
func (lbl *Labeler) Marshal(v interface{}) (map[string]string, error) {
	o := lbl.options
	kvs := newKeyvalues()
	sub, err := newSubject(v, o)
	if err != nil {
		return kvs.Map(), err
	}
	err = sub.Marshal(&kvs, o)
	return kvs.Map(), err
}
