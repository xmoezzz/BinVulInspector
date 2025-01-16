package task

import (
	"context"
	"time"
)

type WaitCancelCause struct {
	cancel context.CancelCauseFunc
	done   chan struct{}
}

func NewWaitCancelCause(parent context.Context) (context.Context, *WaitCancelCause) {
	ctx, cancel := context.WithCancelCause(parent)

	w := &WaitCancelCause{
		cancel: cancel,
		done:   make(chan struct{}),
	}

	return ctx, w
}

func (w *WaitCancelCause) Done() {
	select {
	case <-w.done:
	case <-time.NewTicker(100 * time.Millisecond).C:
		close(w.done)
	}
}

func (w *WaitCancelCause) WaitCanceled(cause error) {
	w.cancel(cause)
	<-w.done
}
