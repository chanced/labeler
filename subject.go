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

func (sub *subject) Unmarshal(kvs *keyvalues, o Options) error {
	var unmarshalLabels unmarshal
	if sub.unmarshal != nil {
		unmarshalLabels = sub.unmarshal
	} else if sub.container != nil {
		unmarshalLabels = sub.container.unmarshal
	} else {
		return ErrMissingContainer
	}
	fieldErrs := []*FieldError{}
	for _, f := range sub.tagged {
		err := f.Unmarshal(kvs, o)
		if err != nil {
			fieldErrs = append(fieldErrs, f.err(err))
		}
		if f.ShouldDiscard(o) {
			kvs.Delete(f.Key)
		}
		if f.WasSet {
			f.Save()
		}
	}
	if len(fieldErrs) > 0 {
		return NewParsingError(fieldErrs)
	}
	return unmarshalLabels(sub, kvs, o)
}

func (sub *subject) Marshal(kvs *keyvalues, o Options) error {
	return nil
}

func (sub *subject) Path() string {
	return ""
}

func (sub *subject) Topic() topic {
	return subjectTopic
}

func (sub *subject) IsFieldContainer() bool {
	return false
}
