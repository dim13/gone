// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/mewkiz/pkg/goutil"
)

var (
	goneDir       string
	dumpFileName  string
	logFileName   string
	indexFileName string
	tracks        Tracks
	zzz           bool
	m             sync.Mutex
	logger        *log.Logger
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

type Tracks map[Window]*Track

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

func (t Tracks) update(x Xorg) (current *Track) {
	if win, ok := x.window(); ok {
		m.Lock()
		if _, ok := t[win]; !ok {
			t[win] = new(Track)
		}
		t[win].Seen = time.Now()
		current = t[win]
		m.Unlock()
	}
	return
}

func (t Tracks) collect() {
	x := connect()
	defer x.X.Close()

	current := t.update(x)
	for {
		ev, everr := x.X.WaitForEvent()
		if everr != nil {
			log.Println("wait for event:", everr)
			continue
		}
		switch event := ev.(type) {
		case xproto.PropertyNotifyEvent:
			if current != nil {
				m.Lock()
				current.Spent += time.Since(current.Seen)
				m.Unlock()
			}
			current = t.update(x)
		case screensaver.NotifyEvent:
			switch event.State {
			case screensaver.StateOn:
				log.Println("away from keyboard")
				current = nil
				zzz = true
			default:
				log.Println("back to keyboard")
				zzz = false
			}
		}
	}
}

func (t Tracks) cleanup(d time.Duration) {
	m.Lock()
	for k, v := range t {
		if time.Since(v.Seen) > d {
			logger.Println(v, k)
			delete(t, k)
		}
	}
	m.Unlock()
}

func load(fname string) Tracks {
	t := make(Tracks)
	dump, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return t
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	m.Lock()
	err = dec.Decode(&t)
	m.Unlock()
	if err != nil {
		log.Println(err)
	}
	return t
}

func (t Tracks) store(fname string) {
	tmp := fname + ".tmp"
	dump, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	enc := gob.NewEncoder(dump)
	m.Lock()
	err = enc.Encode(t)
	m.Unlock()
	if err != nil {
		log.Println(err)
		os.Remove(tmp)
		return
	}
	os.Rename(tmp, fname)
}

func main() {
	logfile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.LstdFlags)

	tracks = load(dumpFileName)

	go tracks.collect()
	go func() {
		for {
			tracks.cleanup(8 * time.Hour)
			tracks.store(dumpFileName)
			time.Sleep(time.Minute)
		}
	}()

	webReporter("127.0.0.1:8001")
}
