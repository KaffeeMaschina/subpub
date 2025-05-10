package mysubpub

import (
	"context"
	"errors"
	"sync"
)

// MassageHandler is a callback function that processes messages delivered to subscriber.
type MassageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given subject.
	Subscribe(subject string, cb MassageHandler) (Subscription, error)
	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg interface{}) error
	// Close will shut down the sub-pub system.
	// May be blocked by data delivery until the context is closed.
	Close(ctx context.Context) error
}

type subscriber struct {
	cb    MassageHandler
	queue chan interface{}
	ctx   context.Context
	stop  context.CancelFunc
}
type eventBus struct {
	subscribers map[string][]*subscriber
	mu          sync.RWMutex
	closed      bool
	closeOnce   sync.Once
	ctx         context.Context
	cancel      context.CancelFunc
}
type subscription struct {
	eb      *eventBus
	subject string
	sub     *subscriber
}

func (s *subscription) Unsubscribe() {
	s.eb.mu.Lock()
	defer s.eb.mu.Unlock()
	if s.eb.closed {
		return
	}
	s.sub.stop()
	close(s.sub.queue)
	subscribers := s.eb.subscribers[s.subject]
	for i, sub := range subscribers {
		if sub == s.sub {
			s.eb.subscribers[s.subject] = append(subscribers[:i], subscribers[i+1:]...)
			return
		}
	}

}
func (eb *eventBus) Subscribe(subject string, cb MassageHandler) (Subscription, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.closed {
		return nil, errors.New("already closed")
	}

	ctx, cancel := context.WithCancel(eb.ctx)
	sub := &subscriber{
		cb:    cb,
		queue: make(chan interface{}, 100),
		ctx:   ctx,
		stop:  cancel,
	}
	eb.subscribers[subject] = append(eb.subscribers[subject], sub)

	go func() {
		for {
			select {
			case msg, ok := <-sub.queue:
				if !ok {
					return
				}

				sub.cb(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	return &subscription{
		eb:      eb,
		subject: subject,
		sub:     sub,
	}, nil
}
func (eb *eventBus) Publish(subject string, msg interface{}) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return errors.New("already closed")
	}
	if subscribers, ok := eb.subscribers[subject]; ok {
		for _, sub := range subscribers {
			select {
			case sub.queue <- msg:
			case <-sub.ctx.Done():
				continue
			}
		}
	} else {
		return errors.New("not subscribed")
	}

	return nil
}
func (eb *eventBus) Close(ctx context.Context) error {

	done := make(chan struct{})

	eb.closeOnce.Do(func() {
		eb.mu.Lock()
		if eb.closed {
			eb.mu.Unlock()
			close(done)
			return
		}
		eb.closed = true
		eb.cancel()
		eb.mu.Unlock()

		go func() {
			eb.mu.Lock()
			defer eb.mu.Unlock()
			for _, subs := range eb.subscribers {
				for _, sub := range subs {
					sub.stop()
					close(sub.queue)
				}
			}
			eb.subscribers = nil
			close(done)
		}()
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func NewSubPub() SubPub {
	ctx, cancel := context.WithCancel(context.Background())
	return &eventBus{
		subscribers: make(map[string][]*subscriber),
		mu:          sync.RWMutex{},
		ctx:         ctx,
		cancel:      cancel,
	}
}
