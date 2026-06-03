package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Status int

const (
	Stopped Status = iota
	Running
	Finished
)

type Timer struct {
	mu       sync.Mutex
	duration int
	left     int
	status   Status
	cancel   context.CancelFunc
	tickCh   chan int
}

func NewTimer(second int) *Timer {
	return &Timer{
		duration: second,
		left:     second,
		status:   Stopped,
	}
}

func (t *Timer) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == Running {
		return errors.New("Ошибка.Таймер уже работает.")
	}

	if t.left <= 0 {
		t.left = t.duration
	}

	// ← ДОБАВЬ ЭТУ СТРОКУ:
	t.tickCh = make(chan int, 10)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.status = Running

	go t.run(ctx)
	return nil
}

func (t *Timer) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != Running {
		return errors.New("Нельзя остановить, таймер не запуще!")
	}

	t.status = Stopped
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	return nil
}

func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.status = Stopped
	t.left = t.duration

}

func (t *Timer) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			t.left--

			if t.left <= 0 {
				t.left = 0
				t.status = Finished
				t.mu.Unlock()

				select {
				case t.tickCh <- 0:
				default:
				}
				return
			}

			currentLeft := t.left
			t.mu.Unlock()

			select {
			case t.tickCh <- currentLeft:
			default:
			}
		}
	}
}

func (t *Timer) Resume() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != Stopped {
		return errors.New("нельзя продолжить таймер не на паузе")

	}

	if t.left <= 0 {
		return errors.New("Таймер уже завершился")
	}

	t.tickCh = make(chan int, 10)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.status = Running

	go t.run(ctx)
	return nil

}

func (t *Timer) SetDuration(seconds int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if seconds > 0 {
		t.duration = seconds
		t.left = seconds
		t.status = Stopped
		t.tickCh = make(chan int, 10)
	}
}
func (t *Timer) GetStatus() Status {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.status
}

func (t *Timer) GetLeft() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.left
}

func (t *Timer) GetTickChannel() <-chan int {
	return t.tickCh
}

func (t *Timer) GetDuration() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.duration
}
