// Gone Time Tracker -or- Where has my time gone?
package main

//go:generate esc -o static.go static/

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type Tracks struct {
	tracks  map[Window]Track
	current Window
	logger  *log.Logger
	zzz     bool
}

type Track struct {
	Seen  time.Time
	Spent time.Duration
	Idle  time.Duration
}

func (t Track) String() string {
	return fmt.Sprintf("%s %s", t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
}

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (t *Tracks) Snooze(idle time.Duration) {
	if !t.zzz {
		t.logger.Println("away from keyboard, idle for", idle)
		if c, ok := t.tracks[t.current]; ok {
			c.Idle += idle
			t.tracks[t.current] = c
		}
		t.zzz = true
	}
}

func (t *Tracks) Wakeup() {
	if t.zzz {
		t.logger.Println("back to keyboard")
		if c, ok := t.tracks[t.current]; ok {
			c.Seen = time.Now()
			t.tracks[t.current] = c
		}
		t.zzz = false
	}
}

func (t *Tracks) Update(w Window) {
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
			t.logger.Println(v, k)
			delete(t.tracks, k)
		}
	}
}

func Load(fname string) *Tracks {
	t := &Tracks{tracks: make(map[Window]Track)}
	dump, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return t
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	err = dec.Decode(&t)
	if err != nil {
		log.Println(err)
	}
	return t
}

func (t Tracks) Store(fname string) {
	tmp := fname + ".tmp"
	dump, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	enc := gob.NewEncoder(dump)
	err = enc.Encode(t)
	if err != nil {
		log.Println(err)
		os.Remove(tmp)
		return
	}
	os.Rename(tmp, fname)
}

func (t Tracks) Cleanup(every, since time.Duration, dump string) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	for range tick.C {
		t.RemoveSince(since)
		t.Store(dump)
	}
}

func main() {
	var (
		display  = flag.String("display", os.Getenv("DISPLAY"), "X11 display")
		listen   = flag.String("listen", "127.0.0.1:8001", "web reporter")
		timeout  = flag.Duration("timeout", time.Minute*5, "idle timeout")
		expire   = flag.Duration("expire", time.Hour*8, "expire timeout")
		refresh  = flag.Duration("refresh", time.Minute, "refresh interval")
		logFile  = flag.String("logfile", path.Join(CachePath(), "gone.log"), "log file")
		dumpFile = flag.String("dumpfile", path.Join(CachePath(), "gone.gob"), "dump file")
	)
	flag.Parse()

	X := Connect(*display)
	defer X.Close()

	logfile, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	tracks := Load(*dumpFile)
	tracks.logger = log.New(logfile, "", log.LstdFlags)
	defer tracks.Store(*dumpFile)

	go X.Collect(tracks, *timeout)
	go tracks.Cleanup(*refresh, *expire, *dumpFile)

	if err := webReporter(tracks, *listen); err != nil {
		log.Fatal(err)
	}
}
