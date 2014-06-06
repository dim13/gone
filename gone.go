package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type Tracker map[Window]*Track

type Track struct {
	Start time.Time
	Spent time.Duration
}

type Window struct {
	Class string
	Name  string
}

type Xorg struct {
	X           *xgb.Conn
	root        xproto.Window
	activeAtom  *xproto.InternAtomReply
	netNameAtom *xproto.InternAtomReply
	nameAtom    *xproto.InternAtomReply
	classAtom   *xproto.InternAtomReply
}

func (t Track) String() string {
	return fmt.Sprint(t.Spent)
}

func (w Window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (x Xorg) atom(aname string) *xproto.InternAtomReply {
	a, err := xproto.InternAtom(x.X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal("atom: ", err)
	}
	return a
}

func (x Xorg) property(w xproto.Window, a *xproto.InternAtomReply) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(x.X, false, w, a.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
}

func (x Xorg) active() xproto.Window {
	p, err := x.property(x.root, x.activeAtom)
	if err != nil {
		log.Fatal("active: ", err)
	}
	return xproto.Window(xgb.Get32(p.Value))
}

func (x Xorg) name(w xproto.Window) string {
	name, err := x.property(w, x.netNameAtom)
	if err != nil {
		log.Fatal("net name: ", err)
	}
	if string(name.Value) != "" {
		return string(name.Value)
	}
	name, err = x.property(w, x.nameAtom)
	if err != nil {
		log.Fatal("wm name: ", err)
	}
	return string(name.Value)
}

func (x Xorg) class(w xproto.Window) string {
	class, err := x.property(w, x.classAtom)
	if err != nil {
		log.Fatal("class: ", err)
	}
	i := bytes.IndexByte(class.Value, 0)
	if i == -1 || string(class.Value[:i]) == "" {
		return "unknown"
	}
	return string(class.Value[:i])
}

func (x Xorg) winName() (Window, bool) {
	windowId := x.active()
	/* skip invalid window id */
	if windowId == 0 {
		return Window{}, false
	}
	x.spy(windowId)
	return Window{
		Class: x.class(windowId),
		Name:  x.name(windowId),
	}, true
}

func (x Xorg) spy(w xproto.Window) {
	xproto.ChangeWindowAttributes(x.X, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
}

func (x Xorg) Update(t Tracker) (prev *Track) {
	if win, ok := x.winName(); ok {
		if _, ok := t[win]; !ok {
			t[win] = new(Track)
		}
		t[win].Start = time.Now()
		prev = t[win]
	}
	return
}

func connect() Xorg {
	var x Xorg
	var err error

	x.X, err = xgb.NewConn()
	if err != nil {
		log.Fatal("xgb: ", err)
	}

	err = screensaver.Init(x.X)
	if err != nil {
		log.Fatal("screensaver: ", err)
	}

	setup := xproto.Setup(x.X)
	x.root = setup.DefaultScreen(x.X).Root

	drw := xproto.Drawable(x.root)
	screensaver.SelectInput(x.X, drw, screensaver.EventNotifyMask)

	x.activeAtom = x.atom("_NET_ACTIVE_WINDOW")
	x.netNameAtom = x.atom("_NET_WM_NAME")
	x.nameAtom = x.atom("WM_NAME")
	x.classAtom = x.atom("WM_CLASS")

	x.spy(x.root)

	return x
}

func (t Tracker) collect() {
	x := connect()
	defer x.X.Close()

	prev := x.Update(t)
	for {
		ev, everr := x.X.WaitForEvent()
		if everr != nil {
			log.Fatal("wait for event: ", everr)
		}
		switch event := ev.(type) {
		case xproto.PropertyNotifyEvent:
			if prev != nil {
				prev.Spent += time.Since(prev.Start)
			}
			prev = x.Update(t)
			zzz = false
		case screensaver.NotifyEvent:
			switch event.State {
			case screensaver.StateOn:
				fmt.Println("away from keyboard")
				prev = nil
				zzz = true
			}
		}
	}
}

func (t Tracker) cleanup(d time.Duration) {
	for k, v := range t {
		if time.Since(v.Start) > d {
			log.Println("removing", k)
			delete(t, k)
		}
	}
}

func (t Tracker) load(fname string) {
	dump, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	err = dec.Decode(&t)
	if err != nil {
		log.Println(err)
	}
}

func (t Tracker) store(fname string) {
	dump, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	enc := gob.NewEncoder(dump)
	err = enc.Encode(t)
	if err != nil {
		log.Println(err)
	}
}

type Index struct {
	Title  string
	Tracks Tracks
	Class  Tracks
	Total  time.Duration
	Zzz    bool
}

type Tracks []track

type track struct {
	Class string
	Name  string
	Time  time.Duration
}

func (t Tracks) Len() int           { return len(t) }
func (t Tracks) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Tracks) Less(i, j int) bool { return t[i].Time < t[j].Time }

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var i Index
	i.Title = "Time Tracker"
	i.Zzz = zzz
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		i.Total += v.Spent
		if class != "" && class != k.Class {
			continue
		}
		i.Tracks = append(i.Tracks, track{
			Class: k.Class,
			Name:  k.Name,
			Time:  v.Spent})
	}
	for k, v := range classtotal {
		i.Class = append(i.Class, track{Class: k, Time: v})
	}
	sort.Sort(sort.Reverse(i.Class))
	sort.Sort(sort.Reverse(i.Tracks))
	err := tmpl.Execute(w, i)
	if err != nil {
		log.Println(err)
	}
}

const (
	port = ":8001"
	file = "dump.gob"
)

var (
	tracks = make(Tracker)
	tmpl   = template.Must(template.ParseFiles("index.html"))
	zzz    bool
)

func main() {
	tracks.load(file)
	go tracks.collect()
	go func() {
		for {
			tracks.cleanup(12 * time.Hour)
			tracks.store(file)
			time.Sleep(time.Minute)
		}
	}()
	log.Println("listen on", port)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(port, nil)
}
