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
	broker Broker
}

func NewApp(b Broker) *App {
	return &App{broker: b}
}

func (a *App) Seen(w Window) {
	data := struct {
		ID    int
		Class string
		Name  string
		Date  time.Time
	}{
		ID:    w.ID,
		Class: w.Class,
		Name:  w.Name,
		Date:  time.Now(),
	}
	b, _ := json.Marshal(data)
	a.broker.Send(Event{
		Type: EventSeen,
		Data: string(b),
	})
}

func (a *App) Idle(d time.Duration) {
	data := struct {
		Idle time.Duration
	}{
		Idle: d,
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
