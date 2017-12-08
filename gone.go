// Gone Time Tracker -or- Where has my time gone?
package main

//go:generate go get github.com/mjibson/esc
//go:generate esc -ignore '^\..*' -o public.go public/

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dim13/gone/internal/sse"
	"github.com/dim13/gone/internal/xorg"
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
func (a *App) ListenAndServe(addr string) error {
	http.Handle("/", http.FileServer(Dir(true, "/public")))
	http.Handle("/events", a.broker)
	return http.ListenAndServe(addr, nil)
}

func main() {
	var (
		display = flag.String("display", os.Getenv("DISPLAY"), "X11 display")
		listen  = flag.String("listen", "127.0.0.1:8001", "web reporter")
	)
	flag.Parse()

	X, err := xorg.Connect(*display)
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	app := NewApp(sse.NewBroker())

	go X.Collect(app)

	if err := app.ListenAndServe(*listen); err != nil {
		log.Fatal(err)
	}
}
