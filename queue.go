package bluffy

import "errors"

var (
	e_queueFull   = errors.New("Queue is full")
	e_queueClosed = errors.New("Queue is closed")
	e_nilAccount  = errors.New("Cannot add nil account to Queue")
)

type queue struct {
	q chan *account

	done chan struct{}
}

func newQueue(maxsize int) *queue {
	return &queue{
		q:    make(chan *account, maxsize),
		done: make(chan struct{}, 2),
	}
}

func (q *queue) run() {
	for {
		a1, a2 := <-q.q, <-q.q
		if a1 != nil && a2 != nil {
			_ = newMatch(a1, a2)
			continue
		}
		//TODO tell the non-nil account it has been disconnected
		return
	}
}

func (q *queue) close() {
	//TODO can channel operations be reordered by the compiler?
	close(q.done)
	close(q.q)
}

func (q *queue) enqueue(a *account) error {
	if a == nil {
		return e_nilAccount
	}
	select {
	case <-q.done:
		return e_queueClosed
	default:
	}
	select {
	case q.q <- a:
		//TODO account enqueued successfully
	default:
		return e_queueFull
	}
	return nil
}
