// Gone Time Tracker -or- Where has my time gone?
package main

//go:generate go get github.com/mjibson/esc
//go:generate esc -ignore '.*\.swp' -o static.go static/

import (
	"flag"
	"log"
	"os"
	"time"
)

type Tracks struct {
	tracks   map[Window]Track
	current  Window
	zzz      bool
	interval time.Duration
}

type Track struct {
	Seen  time.Time
	Spent time.Duration
	Idle  time.Duration
}

func (t *Tracks) Idle(idle time.Duration) {
	if idle > 0 && !t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Idle += idle
			t.tracks[t.current] = c
		}
		t.zzz = true
	}
	if idle == 0 && t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Seen = time.Now()
			t.tracks[t.current] = c
		}
		t.zzz = false
	}
}

func (t *Tracks) Seen(w Window) {
	if !t.zzz {
		if c, ok := t.tracks[t.current]; ok {
			c.Spent += time.Since(c.Seen)
			t.tracks[t.current] = c
		}
	}

	if _, ok := t.tracks[w]; !ok {
		t.tracks[w] = Track{}
	}

	s := t.tracks[w]
	s.Seen = time.Now()
	t.tracks[w] = s

	t.current = w
}

func (t Tracks) RemoveSince(d time.Duration) {
	for k, v := range t.tracks {
		if time.Since(v.Seen) > d || v.Idle > d {
			delete(t.tracks, k)
		}
	}
}

func (t Tracks) Cleanup(since time.Duration) {
	tick := time.NewTicker(t.interval)
	defer tick.Stop()
	for range tick.C {
		t.RemoveSince(since)
	}
}

func main() {
	var (
		display = flag.String("display", os.Getenv("DISPLAY"), "X11 display")
		listen  = flag.String("listen", "127.0.0.1:8001", "web reporter")
		timeout = flag.Duration("timeout", time.Minute*5, "idle timeout")
		//expire  = flag.Duration("expire", time.Hour*8, "expire timeout")
		//refresh = flag.Duration("refresh", time.Minute, "refresh interval")
	)
	flag.Parse()

	X, err := Connect(*display)
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	/* old
	tracks := &Tracks{
		tracks:   make(map[Window]Track),
		interval: *refresh,
	}

	go X.Collect(tracks, *timeout)
	go tracks.Cleanup(*expire)
	if err := webReporter(tracks, *listen); err != nil {
		log.Fatal(err)
	}
	*/

	b := NewBroker()
	a := NewApp(b)
	go X.Collect(a, *timeout)
	a.Serve(*listen)
}
