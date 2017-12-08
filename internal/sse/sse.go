// Package sse implements server-sent events (HTML5)
package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type msg struct {
	event string
	data  string
}

// Broker for SSE connections
type Broker struct {
	clients map[chan msg]bool
}

// NewBroker allocates a new broker
func NewBroker() Broker {
	return Broker{clients: make(map[chan msg]bool)}
}

// Send event
func (b Broker) Send(event, data string) error {
	for c := range b.clients {
		c <- msg{event: event, data: data}
	}
	return nil
}

// SendJSON event
func (b Broker) SendJSON(event string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.Send(event, string(data))
}

func (b Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "not a flusher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	c := make(chan msg)
	defer close(c)

	b.clients[c] = true
	defer delete(b.clients, c)

	for ev := range c {
		select {
		case <-r.Context().Done():
			return
		default:
			if ev.event != "" {
				fmt.Fprintln(w, "event:", ev.event)
			}
			for _, data := range strings.Split(ev.data, "\n") {
				fmt.Fprintln(w, "data:", data)
			}
			fmt.Fprintln(w, "")
			flusher.Flush()
		}
	}
}
