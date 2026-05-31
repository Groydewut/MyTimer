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
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.status = Running

	go t.run(ctx)
	return nil
}

func (t *Timer) Pause() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == Paused || t.status == Stopped {
		return errors.New("Таймер уже остановлен.Ошибка.")
	}

	if t.cancel != nil {
		t.cancel()
	}

	t.status = Paused
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
	t.cancel()
	return nil
}

func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == Running || t.status == Paused {
		t.status = Stopped
	}
	t.left = t.duration
	t.status = Stopped

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
			currentLeft := t.left
			t.mu.Unlock()

			fmt.Println("Осталось:", currentLeft)

			if currentLeft <= 0 {
				t.mu.Lock()
				t.status = Finished
				t.mu.Unlock()

				fmt.Println("Время вышло.")
				return
			}

		}
	}
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

	time.Sleep(10 * time.Second)

}
