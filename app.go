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
	idle    time.Duration
}

func NewApp(b Broker) *App {
	return &App{
		broker: b,
		seen:   time.Now(),
	}
}

func (a *App) Seen(w Window) {
	data := struct {
		Class  string
		Name   string
		Seen   time.Time
		Active time.Duration
	}{
		Class:  a.current.Class,
		Name:   a.current.Name,
		Seen:   time.Now(),
		Active: time.Since(a.seen) - a.idle,
	}
	b, _ := json.Marshal(data)
	a.broker.Send(Event{
		Type: "seen",
		Data: string(b),
	})
	a.current = w
	a.seen = time.Now()
	a.idle = 0
}

func (a *App) Idle(idle time.Duration) {
	a.idle = idle
}

func (a *App) Serve(addr string) {
	http.Handle("/", http.FileServer(Dir(true, "/static")))
	http.Handle("/events", a.broker)
	http.ListenAndServe(addr, nil)
}
