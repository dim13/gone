// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mewkiz/pkg/goutil"
)

var (
	goneDir       string
	dumpFileName  string
	logFileName   string
	indexFileName string
	tracks        Tracks
	zzz           bool
	logger        *log.Logger
	current       Window
)

func init() {
	var err error
	goneDir, err = goutil.SrcDir("github.com/dim13/gone")
	if err != nil {
		log.Fatal("init: ", err)
	}
	dumpFileName = filepath.Join(goneDir, "gone.gob")
	logFileName = filepath.Join(goneDir, "gone.log")
	indexFileName = filepath.Join(goneDir, "index.html")
}

type Tracker interface {
	Update(Window)
	Snooze(time.Duration)
	Wakeup()
}

type Tracks map[Window]Track

type Track struct {
	Seen  time.Time
	Spent time.Duration
}

type Window struct {
	Class string
	Name  string
}

func (t Track) String() string {
	return fmt.Sprintf("%s %s", t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
}

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (t Tracks) Snooze(idle time.Duration) {
	if zzz == false {
		log.Println("away from keyboard, idle for", idle)
		if c, ok := t[current]; ok {
			if idle > c.Spent {
				c.Spent -= idle
				t[current] = c
			}
		}
		zzz = true
	}
}

func (t Tracks) Wakeup() {
	if zzz == true {
		log.Println("back to keyboard")
		zzz = false
	}
}

func (t Tracks) Update(w Window) {
	if zzz == false {
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
		if time.Since(v.Seen) > d {
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
		tracks.Remove(8 * time.Hour)
		tracks.Store(dumpFileName)
		time.Sleep(time.Minute)
	}
}

func main() {
	X := Connect()
	defer X.Close()

	logfile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.LstdFlags)

	tracks = Load(dumpFileName)

	go X.Collect(tracks)
	go tracks.Cleanup()

	webReporter("127.0.0.1:8001")
}
