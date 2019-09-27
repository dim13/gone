package ui

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/zserge/webview"
)

type Row struct {
	Class string
	Title string
	Spent string
}

type UI struct {
	Records  []Row
	Classes  []Row
	webview  webview.WebView
	listener net.Listener
	sync     func()
}

func New() (*UI, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	settings := webview.Settings{
		Title:  "gone",
		URL:    "http://" + l.Addr().String(),
		Width:  800,
		Height: 600,
		Debug:  true,
	}
	w := webview.New(settings)
	ui := &UI{webview: w, listener: l}
	w.Dispatch(func() {
		ui.sync, _ = w.Bind("ui", ui)
	})
	return ui, nil
}

func (ui *UI) Serve() error {
	log.Println("listen at", ui.listener.Addr())
	http.Handle("/", http.FileServer(http.Dir("public")))
	go http.Serve(ui.listener, nil)
	ui.webview.Run()
	return nil
}

func (ui *UI) OnEvent(class, name string, seen time.Time, active time.Duration) error {
	if class == "" {
		return nil
	}
	ui.Records = append(ui.Records, Row{Class: class, Title: name, Spent: active.String()})
	ui.webview.Dispatch(func() {
		if ui.sync != nil {
			ui.sync()
		}
		ui.webview.Eval(`updateTable();`)
	})
	log.Println(class, name, seen, active)
	return nil
}

func (ui *UI) Close() {
	ui.webview.Exit()
	ui.listener.Close()
}
