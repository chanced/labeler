package labeler

import (
	"errors"
	"reflect"
)

type input struct {
	meta
}

func newInput(v interface{}, o Options) (input, error) {
	rv := reflect.ValueOf(v)
	in := input{
		meta: newMeta(rv),
	}
	in.marshal = getMarshal(&in, o)
	if in.marshal == nil {
		return in, ErrInvalidInput
	}
	return in, nil
}

func (in *input) Unmarshal(kvs *keyvalues, o Options) error {
	return errors.New("cannot unmarshal input")
}

func (in *input) Marshal(kvs *keyvalues, o Options) error {
	if in.marshal == nil {
		return ErrInvalidInput
	}
	return in.marshal(in, kvs, o)
}

func (in *input) IsContainer(o Options) bool {
	return false
}

func (in *input) Path() string {
	return ""
}

func (in *input) Topic() topic {
	return inputTopic
}

func (in *input) Save() {}
