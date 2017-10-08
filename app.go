package main

import (
	"net/http"
	"time"
)

// App holds application context
type App struct {
	broker  Broker
	current Window
	seen    time.Time
}

type seenEvent struct {
	Class  string
	Name   string
	Seen   time.Time
	Active time.Duration
}

// NewApp creates a new application insctance
func NewApp(b Broker) *App {
	return &App{broker: b, seen: time.Now()}
}

func (a *App) sendEvent(idle time.Duration) error {
	ev := seenEvent{
		Class:  a.current.Class,
		Name:   a.current.Name,
		Seen:   a.seen,
		Active: time.Since(a.seen) - idle,
	}
	return a.broker.SendJSON("seen", ev)
}

// Seen Window event handler
func (a *App) Seen(w Window) error {
	defer func() {
		a.seen = time.Now()
		a.current = w
	}()
	return a.sendEvent(0)
}

// Idle event handler
func (a *App) Idle(idle time.Duration) error {
	defer func() {
		a.seen = time.Now()
	}()
	if idle == 0 {
		return nil
	}
	return a.sendEvent(idle)
}

// ListenAndServe launches http server
func (a *App) ListenAndServe(addr string) error {
	http.Handle("/", http.FileServer(Dir(true, "/public")))
	http.Handle("/events", a.broker)
	return http.ListenAndServe(addr, nil)
}
