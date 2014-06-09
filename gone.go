package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
)

const (
	port    = ":8001"
	dump    = "gone.gob"
	logf    = "gone.log"
	unknown = "unknown"
)

var (
	tracks = make(Tracker)
	tmpl   = template.Must(template.ParseFiles("index.html"))
	zzz    bool
	m      sync.Mutex
)

type Tracker map[Window]*Track

type Track struct {
	Seen  time.Time
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
		return x.root
	}
	return xproto.Window(xgb.Get32(p.Value))
}

func (x Xorg) name(w xproto.Window) string {
	name, err := x.property(w, x.netNameAtom)
	if err != nil {
		return unknown
	}
	if string(name.Value) == "" {
		name, err = x.property(w, x.nameAtom)
		if err != nil || string(name.Value) == "" {
			return unknown
		}
	}
	return string(name.Value)
}

func (x Xorg) class(w xproto.Window) string {
	class, err := x.property(w, x.classAtom)
	if err != nil {
		return unknown
	}
	i := bytes.IndexByte(class.Value, 0)
	if i == -1 || string(class.Value[:i]) == "" {
		return unknown
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

func (x Xorg) update(t Tracker) (current *Track) {
	if win, ok := x.winName(); ok {
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

	current := x.update(t)
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
			current = x.update(t)
		case screensaver.NotifyEvent:
			switch event.State {
			case screensaver.StateOn:
				fmt.Println("away from keyboard")
				current = nil
				zzz = true
			default:
				fmt.Println("back to keyboard")
				zzz = false
			}
		}
	}
}

type logger struct {
	*os.File
}

func openLog(fname string) logger {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return logger{f}
}

func (l logger) logForget(w Window, t Track) {
	log.Println("removing", w.Name)
	log.SetOutput(l)
	log.Println(t.Seen.Format("2006/01/02 15:04:05"),
		t.Spent, w.Class, w.Name)
	log.SetOutput(os.Stderr)
}

func (t Tracker) cleanup(d time.Duration) {
	f := openLog(logf)
	defer f.Close()
	m.Lock()
	for k, v := range t {
		if time.Since(v.Seen) > d {
			f.logForget(k, *v)
			delete(t, k)
		}
	}
	m.Unlock()
}

func (t Tracker) reset() {
	f := openLog(logf)
	defer f.Close()
	m.Lock()
	for k, v := range t {
		f.logForget(k, *v)
		delete(t, k)
	}
	m.Unlock()
}

func (t Tracker) load(fname string) {
	dump, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()
	dec := gob.NewDecoder(dump)
	m.Lock()
	err = dec.Decode(&t)
	m.Unlock()
	if err != nil {
		log.Println(err)
	}
}

func (t Tracker) store(fname string) {
	dump, err := os.Create(fname + ".tmp")
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
	}
	os.Rename(fname+".tmp", fname)
}

type Index struct {
	Title   string
	Records Records
	Classes Records
	Total   time.Duration
	Zzz     bool
}

type Records []Record

type Record struct {
	Class string
	Name  string
	Spent time.Duration
	Seen  time.Time
	Odd   bool `json:"-"`
}

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Title = "Gone Time Tracker"
	idx.Zzz = zzz
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	m.Lock()
	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		idx.Total += v.Spent
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: v.Spent})
	}
	m.Unlock()
	for k, v := range classtotal {
		idx.Classes = append(idx.Classes, Record{Class: k, Spent: v})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	for j := range idx.Records {
		idx.Records[j].Odd = j%2 == 0
	}
	err := tmpl.Execute(w, idx)
	if err != nil {
		log.Println(err)
	}
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	var rec Records

	m.Lock()
	for k, v := range tracks {
		rec = append(rec, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: v.Spent,
			Seen:  v.Seen})
	}
	m.Unlock()

	data, err := json.MarshalIndent(rec, "", "\t")
	if err != nil {
		log.Println("dump:", err)
	}
	w.Write(data)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	tracks.reset()
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	tracks.load(dump)
	go tracks.collect()
	go func() {
		for {
			tracks.cleanup(8 * time.Hour)
			tracks.store(dump)
			time.Sleep(time.Minute)
		}
	}()
	log.Println("listen on", port)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gone.json", dumpHandler)
	http.HandleFunc("/reset", resetHandler)
	http.ListenAndServe(port, nil)
}
