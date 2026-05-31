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
		tickCh:   make(chan int, 100),
	}
}

func (t *Timer) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == Running {
		return errors.New("Ошибка.Таймер уже работает.")
	}

	if t.status == Finished || t.left <= 0 {
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

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.status = Running

	go t.run(ctx)
	return nil

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

func main() {
	fmt.Println("Timer backend initialized")
	t := NewTimer(15)
	go func() {
		for left := range t.GetTickChannel() {
			fmt.Printf("⏱️  Осталось: %d сек | Статус: %v\n", left, t.GetStatus())
			if left == 0 {
				fmt.Println("🔔 Таймер завершил работу!")
			}
		}

	}()

	if err := t.Start(); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Работаю...")

	if err := t.Stop(); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Пауза")
	time.Sleep(3 * time.Second)

	if err := t.Resume(); err != nil {
		fmt.Println("Ошибка:", err)
	}
	fmt.Println("Продолжаем")

	time.Sleep(20 * time.Second)
	fmt.Println("\n✅ Демонстрация завершена. Backend готов к подключению Wails.")
}
