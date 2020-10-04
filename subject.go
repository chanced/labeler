package labeler

import (
	"errors"
	"reflect"
)

type subject struct {
	meta
	fieldset
	options Options
}

func newSubject(v interface{}, o Options) subject {
	rv := reflect.ValueOf(v)
	sub := subject{
		meta:     newMeta(rv),
		fieldset: newFieldset(),
	}

	return sub
}

func (sub *subject) topic() topic {
	return subjectTopic
}

func (sub *subject) init(o Options) error {
	if ok := sub.deref(); !ok {
		return ErrInvalidValue
	}

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
	o.SetFromTag(sub.containerTag())

	sub.options = o
	return nil
}
