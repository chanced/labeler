package labeler

import (
	"reflect"
	"sync"
)

// should rename this
type channels struct {
	fieldCh   chan *field
	errCh     chan error
	reflected reflected
	waitGroup *sync.WaitGroup
	options   Options
}

func newChannels(r reflected, o Options) channels {
	i := r.numField()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ch := channels{
		reflected: r,
		fieldCh:   make(chan *field, i),
		errCh:     make(chan error, i),
		waitGroup: wg,
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
			if !ok {
				errCh = nil
				break
			}
			w.errCh <- err
			wg.Done()
		case f, ok := <-fieldCh:
			if ok {
				fieldCh = nil
				break
			}
			w.fieldCh <- f
			wg.Done()
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
	r := ch.reflected
	m := r.Meta()

	ch.waitGroup.Add(m.NumField)
	for i := 0; i < m.NumField; i++ {
		sf := m.Type.Field(i)
		vf := m.Value.Field(i)
		go ch.processField(sf, vf)
	}
	ch.waitGroup.Done()
	ch.waitGroup.Wait()
}

func (ch channels) processField(structField reflect.StructField, valueField reflect.Value) {
	defer ch.waitGroup.Done()
	f, err := newField(structField, valueField, ch.reflected, ch.options)
	if err != nil {
		ch.errCh <- err
	}
	switch {
	case f.IsTagged:
		ch.fieldCh <- f
	case f.IsContainer:
		ch.fieldCh <- f
	case f.IsStruct():
		wch := newChannels(f, ch.options)
		go wch.pipe(ch, f.NumField)
		go wch.processFields()
		wch.waitGroup.Wait()
	}
}
