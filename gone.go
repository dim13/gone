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

	"github.com/dim13/gone/internal/ui"
	"github.com/dim13/gone/internal/xorg"
)

// App holds application context
type App struct {
	ui       *ui.UI
	current  xorg.Window
	lastSeen time.Time
}

type seenEvent struct {
	Class  string
	Name   string
	Seen   time.Time
	Active time.Duration
}

func (a *App) sendEvent(idle time.Duration) error {
	return a.ui.OnEvent(a.current.Class, a.current.Name, a.lastSeen, time.Since(a.lastSeen)-idle)
}

// Seen Window event handler
func (a *App) Seen(w xorg.Window) error {
	defer func() {
		a.lastSeen = time.Now()
		a.current = w
	}()
	return a.sendEvent(0)
}

// Idle event handler
func (a *App) Idle(idle time.Duration) error {
	defer func() {
		a.lastSeen = time.Now()
	}()
	if idle == 0 {
		return nil
	}
	return a.sendEvent(idle)
}

// Serve launches http server
func (a *App) Serve(l net.Listener) error {
	log.Println("listen on", l.Addr())
	http.Handle("/", http.FileServer(Dir(true, "/public")))
	return http.Serve(l, nil)
}

func main() {
	X, err := xorg.Connect(os.Getenv("DISPLAY"))
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	ui, err := ui.New()
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	app := &App{
		lastSeen: time.Now(),
		ui:       ui,
	}
	go X.Collect(app)
	ui.Serve()
}
