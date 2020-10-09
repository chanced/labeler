package labeler

import (
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
	numField := r.NumField()
	wg := &sync.WaitGroup{}
	wg.Add(numField)
	ch := channels{
		reflected: r,
		options:   o,
		fieldCh:   make(chan *field, numField),
		errCh:     make(chan error, numField),
		waitGroup: wg,
	}
	return ch
}

func (ch channels) pipe(w channels, total int) {
	errCh := ch.errCh
	fieldCh := ch.fieldCh
	for errCh != nil || fieldCh != nil {
		select {
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				break
			}
			w.errCh <- err
		case f, ok := <-fieldCh:
			if !ok {
				fieldCh = nil
				break
			}
			w.fieldCh <- f
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
	numField := ch.reflected.NumField()
	for i := 0; i < numField; i++ {
		go ch.processField(i)
	}
	ch.waitGroup.Wait()
}

func (ch channels) processField(i int) {
	defer ch.waitGroup.Done()
	f, err := newField(ch.reflected, i, ch.options)
	if err != nil {
		ch.errCh <- err
		return
	}
	switch {
	case f.IsTagged:
		ch.fieldCh <- f
	case f.IsContainer:
		ch.fieldCh <- f
	case f.IsStruct():
		wch := newChannels(f, ch.options)
		go wch.processFields()
		wch.pipe(ch, f.NumField())
		//wch.waitGroup.Wait()
	default:
		ch.fieldCh <- f
	}
}
