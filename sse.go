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

func (b Broker) Send(ev Event) error {
	for c := range b.clients {
		c <- ev
	}
	return nil
}

func (b Broker) register(c chan Event) {
	b.Lock()
	b.clients[c] = true
	b.Unlock()
}

func (b Broker) deregister(c chan Event) {
	b.Lock()
	delete(b.clients, c)
	b.Unlock()
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

	b.register(c)
	defer b.deregister(c)

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
