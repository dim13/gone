package main

import (
	"encoding/json"
	"net/http"
	"time"
)

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

func NewApp(b Broker) *App {
	return &App{broker: b, seen: time.Now()}
}

func (a *App) sendEvent(idle time.Duration) error {
	defer func() { a.seen = time.Now() }()
	b, err := json.Marshal(seenEvent{
		Class:  a.current.Class,
		Name:   a.current.Name,
		Seen:   a.seen,
		Active: time.Since(a.seen) - idle,
	})
	if err != nil {
		return err
	}
	return a.broker.Send(Event{Type: "seen", Data: string(b)})
}

func (a *App) Seen(w Window) error {
	defer func() { a.current = w }()
	return a.sendEvent(0)
}

func (a *App) Idle(idle time.Duration) error {
	return a.sendEvent(idle)
}

func (a *App) ListenAndServe(addr string) error {
	http.Handle("/", http.FileServer(Dir(true, "/static")))
	http.Handle("/events", a.broker)
	return http.ListenAndServe(addr, nil)
}
