package observer

import (
	"log/slog"
	"sync"
	"time"
)

type Observer[T any] interface {
	Subscribe(time.Duration) <-chan T
	Notify(T)
	Unsubscribe(chan T)
}

type Impl[T any] struct {
	sync.Mutex
	subscribers map[chan T]time.Duration
}

func New[T any]() *Impl[T] {
	return &Impl[T]{
		subscribers: make(map[chan T]time.Duration),
	}
}

func (impl *Impl[T]) Subscribe(timeout time.Duration) chan T {
	impl.Lock()
	defer impl.Unlock()
	ch := make(chan T, 1)
	impl.subscribers[ch] = timeout
	return ch
}

func (impl *Impl[T]) Unsubscribe(ch chan T) {
	impl.Lock()
	defer impl.Unlock()
	if _, ok := impl.subscribers[ch]; ok {
		close(ch)
		delete(impl.subscribers, ch)
	}
}

func (impl *Impl[T]) Notify(event T) {
	impl.Lock()
	defer impl.Unlock()
	for ch, duration := range impl.subscribers {
		go func(ch chan T, duration time.Duration) {
			select {
			case ch <- event:
			case <-time.After(duration):
				slog.Warn("timeout writing to channel")
			}
		}(ch, duration)
	}
}
