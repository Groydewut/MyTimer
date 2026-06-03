package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	timer          *Timer
	listenerCtx    context.Context
	listenerCancel context.CancelFunc
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.timer = NewTimer(300)
	// НЕ запускаем слушателя здесь!
}

// startListener — запускает горутину для прослушивания
func (a *App) startListener() {
	// Отменяем предыдущего слушателя
	if a.listenerCancel != nil {
		a.listenerCancel()
	}

	a.listenerCtx, a.listenerCancel = context.WithCancel(a.ctx)

	go func() {
		for left := range a.timer.GetTickChannel() {
			select {
			case <-a.listenerCtx.Done():
				return
			default:
				runtime.EventsEmit(a.ctx, "tick", left)
				if left == 0 {
					runtime.EventsEmit(a.ctx, "finished")
				}
			}
		}
	}()
}
func (s Status) String() string {
	switch s {
	case Stopped:
		return "Stopped"
	case Running:
		return "Running"
	case Finished:
		return "Finished"
	default:
		return "Unknown"
	}
}

func (a *App) Start() error {
	err := a.timer.Start()
	if err == nil {
		a.startListener() // ← Перезапускаем!
	}
	return err
}

func (a *App) Resume() error {
	err := a.timer.Resume()
	if err == nil {
		a.startListener() // ← Перезапускаем!
	}
	return err
}

func (a *App) SetDuration(seconds int) {
	a.timer.SetDuration(seconds)
	a.startListener() // ← Перезапускаем!
}

func (a *App) Stop() error {
	return a.timer.Stop()
}

func (a *App) Reset() {
	a.timer.Reset()
}

func (a *App) GetLeft() int {
	return a.timer.GetLeft()
}

func (a *App) GetStatus() string {
	return a.timer.GetStatus().String()
}
