// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mewkiz/pkg/goutil"
)

type Tracks map[Window]Track

type Track struct {
	Seen  time.Time
	Spent time.Duration
	Idle  time.Duration
}

var (
	goneDir       string
	dumpFileName  string
	logFileName   string
	indexFileName string
	tracks        Tracks
	zzz           bool
	logger        *log.Logger
	current       Window
	display       string
	listen        string
	timeout       time.Duration
	expire        time.Duration
	refresh       time.Duration
)

func init() {
	var err error
	goneDir, err = goutil.SrcDir("github.com/dim13/gone")
	if err != nil {
		log.Fatal("init: ", err)
	}

	dumpFileName = filepath.Join(goneDir, "gone.gob")
	logFileName = filepath.Join(goneDir, "gone.log")
	indexFileName = filepath.Join(goneDir, "root.tmpl")
	initTemplate(indexFileName)

	flag.StringVar(&display, "display", ":0", "X11 display")
	flag.StringVar(&listen, "listen", "127.0.0.1:8001", "web reporter")
	flag.DurationVar(&timeout, "timeout", time.Minute*5, "idle timeout")
	flag.DurationVar(&expire, "expire", time.Hour*8, "expire timeout")
	flag.DurationVar(&refresh, "refresh", time.Minute, "refresh interval")
	flag.Parse()
}

func (t Track) String() string {
	return fmt.Sprintf("%s %s",
		t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
}

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (t Tracks) Snooze(idle time.Duration) {
	if !zzz {
		logger.Println("away from keyboard, idle for", idle)
		if c, ok := t[current]; ok {
			if c.Idle < idle {
				c.Idle = idle
			}
			t[current] = c
		}
		zzz = true
	}
}

func (t Tracks) Wakeup() {
	if zzz {
		logger.Println("back to keyboard")
		if c, ok := t[current]; ok {
			c.Seen = time.Now()
			t[current] = c
		}
		zzz = false
	}
}

func (t Tracks) Update(w Window) {
	if !zzz {
		if c, ok := t[current]; ok {
			c.Spent += time.Since(c.Seen)
			t[current] = c
		}
	}

	if _, ok := t[w]; !ok {
		t[w] = Track{}
	}

	s := t[w]
	s.Seen = time.Now()
	t[w] = s

	current = w
}

func (t Tracks) Remove(d time.Duration) {
	for k, v := range t {
		if time.Since(v.Seen) > d || v.Idle > d {
			logger.Println(v, k)
			delete(t, k)
		}
	}
}

func Load(fname string) Tracks {
	t := make(Tracks)
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

func (t Tracks) Cleanup() {
	for {
		tracks.Remove(expire)
		tracks.Store(dumpFileName)
		time.Sleep(time.Minute)
	}
}

func main() {
	X := Connect()
	defer X.Close()

	logfile, err := os.OpenFile(logFileName,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.LstdFlags)

	tracks = Load(dumpFileName)

	go X.Collect(tracks, timeout)
	go tracks.Cleanup()

	webReporter(listen)
}
