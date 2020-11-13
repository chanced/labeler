package labeler

import (
	"errors"
	"reflect"
)

type subject struct {
	meta
	fieldset
}

func newSubject(v interface{}, o Options) (subject, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return subject{}, ErrInvalidValue
	}
	sub := subject{
		meta:     newMeta(rv),
		fieldset: newFieldset(),
	}

	sub.marshal = getMarshal(&sub, o)
	sub.unmarshal = getUnmarshal(&sub, o)
	err := sub.init(o)
	return sub, err
}

func (sub *subject) init(o Options) error {
	ch := newChannels(sub, o)
	go ch.processFields()
	errs := []*FieldError{}
	fieldCh := ch.fieldCh
	errCh := ch.errCh
	for fieldCh != nil || errCh != nil {
		select {
		case f, ok := <-fieldCh:
			if !ok {
				fieldCh = nil
				break
			}
			err := sub.processField(f, o)
			if err != nil {
				return err
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
	o = o.FromTag(sub.containerTag())
	return nil
}

func (sub *subject) Save() {
	sub.save()
}

func (sub *subject) Unmarshal(kvs *keyValues, o Options) error {
	if sub.unmarshal == nil && (sub.container == nil || sub.container.unmarshal == nil) {
		return ErrMissingContainer
	}
	fieldErrs := []*FieldError{}
	for _, f := range sub.tagged {
		err := f.Unmarshal(kvs, o)
		if err != nil {
			fieldErrs = append(fieldErrs, f.err(err))
		}
		if f.ShouldDiscard(o) {
			kvs.Delete(f.key)
		}
		if f.wasSet {
			f.Save()
		}
	}
	if len(fieldErrs) > 0 {
		return NewParsingError(fieldErrs)
	}
	if sub.unmarshal != nil {
		return sub.unmarshal(sub, kvs, o)
	}
	return sub.container.Unmarshal(kvs, o)
}

func (sub *subject) Marshal(kvs *keyValues, o Options) error {
	if sub.marshal == nil && (sub.container == nil || sub.container.marshal == nil) {
		return ErrMissingContainer
	}
	fieldErrs := []*FieldError{}
	for _, f := range sub.tagged {
		err := f.Marshal(kvs, o)
		if err != nil {
			fieldErrs = append(fieldErrs, f.err(err))
		}
	}
	if len(fieldErrs) > 0 {
		return NewParsingError(fieldErrs)
	}
	if sub.marshal != nil {
		return sub.marshal(sub, kvs, o)
	}
	return sub.container.Marshal(kvs, o)
}

func (sub *subject) Path() string {
	return ""
}

func (sub *subject) Topic() topic {
	return subjectTopic
}

func (sub *subject) IsContainer(o Options) bool {
	return false
}
