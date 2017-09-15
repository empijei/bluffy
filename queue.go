package bluffy

import "errors"

var QueueFullError = errors.New("Queue is full")

var QueueClosedError = errors.New("Queue is closed")

type Queue struct {
	q chan *Account

	done chan struct{}
}

func NewQueue(maxsize int) *Queue {
	return &Queue{
		q:    make(chan *Account, maxsize),
		done: make(chan struct{}, 2),
	}
}

func (q *Queue) Run() {
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

func (q *Queue) Close() {
	//TODO can channel operations be reordered by the compiler?
	close(q.done)
	close(q.q)
}

func (q *Queue) Enqueue(a *Account) error {
	select {
	case <-q.done:
		return QueueClosedError
	default:
		select {
		case q.q <- a:
			//account enqueued successfully
		default:
			return QueueFullError
		}
	}
	return nil
}
