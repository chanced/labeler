package labeler

import (
	"errors"
	"reflect"
	"sync"
)

// should rename this
type channels struct {
	fieldCh   chan field
	errCh     chan error
	field     *field
	labeler   *labeler
	sub       reflected
	opts      Options
	waitGroup *sync.WaitGroup
}

func newChannels(r reflected, o Options) channels {
	i := r.refNumField()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ch := channels{
		sub:       r,
		opts:      o,
		fieldCh:   make(chan field, i),
		errCh:     make(chan error, i),
		waitGroup: wg,
	}

	switch t := r.(type) {
	case *field:
		ch.field = t
	case *labeler:
		ch.labeler = t
	}
	return ch
}

func (ch channels) pipe(w channels, total int) {
	errCh := ch.errCh
	fieldCh := ch.fieldCh
	wg := ch.waitGroup
	wg.Add(total) // sures up a race condition
	for errCh != nil || fieldCh != nil {
		select {
		case err, ok := <-errCh:
			if ok {
				ch.handleErr(err)
				wg.Done()
			} else {
				errCh = nil
			}
		case f, ok := <-fieldCh:
			if ok {
				w.fieldCh <- f
				wg.Done()
			} else {
				fieldCh = nil
			}
		}
	}
}

func (ch channels) finished() {
	ch.waitGroup.Wait()
	close(ch.errCh)
	close(ch.fieldCh)
}

func (ch channels) processFields() {
	defer ch.finished()
	r := ch.sub
	numField := r.refNumField()
	rt := r.refType()
	rv := r.refValue()
	ch.waitGroup.Add(numField)
	for i := 0; i < numField; i++ {
		structField := rt.Field(i)
		valueField := rv.Field(i)
		go ch.processField(structField, valueField)
	}
	ch.waitGroup.Done()
	ch.waitGroup.Wait()
}

func (ch channels) handleErr(err error) {
	var fieldErr *FieldError
	if errors.As(err, &fieldErr) {
		switch {
		case ch.field != nil:
			ch.errCh <- newFieldErrorFromNested(ch.field, fieldErr)
		case ch.labeler != nil:
			ch.errCh <- fieldErr
		}
	} else {
		ch.errCh <- err
	}
}

func (ch channels) processField(structField reflect.StructField, valueField reflect.Value) {
	defer ch.waitGroup.Done()
	f, err := newField(structField, valueField, ch.opts)

	if err != nil {
		ch.handleErr(err)
		return
	}
	switch {
	case f.IsTagged:
		ch.fieldCh <- f
	case f.IsContainer:
		ch.fieldCh <- f
	case f.IsStruct:
		wch := newChannels(f, ch.opts)
		go wch.pipe(ch, f.refNumField())
		go wch.processFields()
		wch.waitGroup.Wait()
	}
}
