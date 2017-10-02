package main

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	EventSeen = "seen"
	EventIdle = "idle"
)

type App struct {
	broker   Broker
	current  Window
	lastSeen time.Time
	idle     time.Duration
}

func NewApp(b Broker) *App {
	return &App{
		broker:   b,
		lastSeen: time.Now(),
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
		Active: time.Since(a.lastSeen) - a.idle,
	}
	b, _ := json.Marshal(data)
	a.broker.Send(Event{
		Type: EventSeen,
		Data: string(b),
	})
	a.current = w
	a.lastSeen = time.Now()
	a.idle = 0
}

func (a *App) Idle(idle time.Duration) {
	a.idle = idle
	data := struct {
		Idle time.Duration
	}{
		Idle: idle,
	}
	b, _ := json.Marshal(data)
	a.broker.Send(Event{
		Type: EventIdle,
		Data: string(b),
	})
}

func (a *App) Serve(addr string) {
	http.Handle("/", http.FileServer(Dir(true, "/static")))
	http.Handle("/events", a.broker)
	http.ListenAndServe(addr, nil)
}
