package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Status int

const (
	Stopped Status = iota
	Running
	Paused
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
		tickCh:   make(chan int, 10),
	}
}

func (t *Timer) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tickCh = make(chan int, 10)

	if t.status == Running {
		return errors.New("Ошибка.Таймер уже работает.")
	}

	if t.left <= 0 {
		t.left = t.duration

	}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.status = Running

	go t.run(ctx)
	return nil
}

func (t *Timer) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == Stopped || t.status == Finished {
		return errors.New("Ошибка, таймер уже остановлен или завершён.")
	}

	t.status = Stopped
	t.left = t.duration

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

	t.left = t.duration
	t.status = Stopped

}

func (t *Timer) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	defer close(t.tickCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			t.left--

			t.mu.Unlock()

			t.tickCh <- t.left

			if t.left <= 0 {
				t.mu.Lock()
				t.status = Finished
				t.mu.Unlock()
				return
			}

		}
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

func main() {
	fmt.Println("Timer backend initialized")
	timer := NewTimer(5)

	fmt.Println("не заблокированно")

	if err := timer.Start(); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Start success!")
	}

	go func() {
		for left := range timer.tickCh {
			fmt.Println(left)
		}
	}()

	time.Sleep(10 * time.Second)

}
