// Gone Time Tracker -or- Where has my time gone?
package main

//go:generate go get github.com/mjibson/esc
//go:generate esc -ignore '^\..*' -o public.go public/

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/dim13/gone/internal/sse"
	"github.com/dim13/gone/internal/xorg"
	"github.com/zserge/webview"
)

// App holds application context
type App struct {
	broker  sse.Broker
	current xorg.Window
	seen    time.Time
}

type seenEvent struct {
	Class  string
	Name   string
	Seen   time.Time
	Active time.Duration
}

// NewApp creates a new application insctance
func NewApp(b sse.Broker) *App {
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
func (a *App) Seen(w xorg.Window) error {
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
func (a *App) Serve(l net.Listener) error {
	http.Handle("/", http.FileServer(Dir(true, "/public")))
	http.Handle("/events", a.broker)
	return http.Serve(l, nil)
}

func main() {
	X, err := xorg.Connect(os.Getenv("DISPLAY"))
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	app := NewApp(sse.NewBroker())
	go X.Collect(app)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	go app.Serve(l)

	addr := "http://" + l.Addr().String()
	webview.Open("Gone", addr, 800, 600, false)
}
