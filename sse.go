package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Event struct {
	Type string
	Data string
}

type Broker struct {
	clients map[chan Event]bool
	sync.Mutex
}

func NewBroker() Broker {
	return Broker{clients: make(map[chan Event]bool)}
}

func (b Broker) Send(ev Event) {
	for c := range b.clients {
		c <- ev
	}
}

func (b Broker) Register(c chan Event) {
	b.Lock()
	defer b.Unlock()
	b.clients[c] = true
}

func (b Broker) Deregister(c chan Event) {
	b.Lock()
	defer b.Unlock()
	delete(b.clients, c)
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

	c := make(chan Event)
	defer close(c)

	b.Register(c)
	defer b.Deregister(c)

	for ev := range c {
		select {
		case <-r.Context().Done():
			return
		default:
			if ev.Type != "" {
				fmt.Fprintf(w, "event: %s\n", ev.Type)
			}
			fmt.Fprintf(w, "data: %s\n\n", ev.Data)
			flusher.Flush()
		}
	}
}
