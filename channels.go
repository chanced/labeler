package labeler

import (
	"errors"
	"reflect"
	"sync"
)

// should rename this
type channels struct {
	fieldCh chan field
	errCh   chan error
	doneCh  chan int
	field   *field
	labeler *labeler
	sub     reflected
	opts    Options
	wg      *sync.WaitGroup
}

func newChannels(r reflected, o Options) channels {
	i := r.getRefNumField()
	ch := channels{
		sub:     r,
		opts:    o,
		fieldCh: make(chan field, i),
		errCh:   make(chan error, i),
		doneCh:  make(chan int, 2),
		wg:      &sync.WaitGroup{},
	}
	switch t := r.(type) {
	case *field:
		ch.field = t
	case *labeler:
		ch.labeler = t
	}
	return ch
}

func (ch channels) pipe(w channels) {
	for {
		select {
		case err := <-w.errCh:
			ch.handleErr(err)
		case f := <-w.fieldCh:
			ch.fieldCh <- f
		case <-w.doneCh:
			ch.doneCh <- 1
			return
		}
	}
}

func (ch channels) finished() {
	ch.doneCh <- 1
	close(ch.doneCh)
	close(ch.errCh)
	close(ch.fieldCh)
}

func (ch channels) processFields() {
	defer ch.finished()
	r := ch.sub
	numField := r.getRefNumField()
	rt := r.getRefType()
	rv := r.getRefValue()

	for i := 0; i < numField; i++ {
		ch.wg.Add(1)
		structField := rt.Field(i)
		valueField := rv.Field(i)
		go ch.processField(structField, valueField)
	}
	ch.wg.Wait()
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

	// defer func() {
	// 	var err error
	// 	if r := recover(); r != nil {
	// 		switch e := r.(type) {
	// 		case error:
	// 			err = e
	// 		case string:
	// 			err = errors.New(e)
	// 		default:
	// 			err = errors.New("unkown error")
	// 		}
	// 		ch.handleErr(err)
	// 	}
	// }()
	defer ch.wg.Done()
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
		go wch.pipe(ch)
		go wch.processFields()

	}
}
