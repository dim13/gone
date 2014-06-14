// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/mewkiz/pkg/goutil"
)

const (
	port = "127.0.0.1:8001"
	dump = "gone.gob"
	logf = "gone.log"
)

var (
	goneDir string
	tracks  Tracker
	tmpl    *template.Template
	zzz     bool
	m       sync.Mutex
	logger  *log.Logger
)

func init() {
	var err error
	goneDir, err = goutil.SrcDir("github.com/dim13/gone")
	if err != nil {
		log.Fatal("init: ", err)
	}
	tmpl = template.Must(template.ParseFiles(filepath.Join(goneDir, "index.html")))
}

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
	return fmt.Sprintf("%s %s", t.Seen.Format("2006/01/02 15:04:05"), t.Spent)
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

func (x Xorg) name(w xproto.Window) (string, error) {
	name, err := x.property(w, x.netNameAtom)
	if err != nil {
		return "", err
	}
	if string(name.Value) == "" {
		name, err = x.property(w, x.nameAtom)
		if err != nil {
			return "", err
		}
		if string(name.Value) == "" {
			return "", errors.New("empty value")
		}
	}
	return string(name.Value), nil
}

func (x Xorg) class(w xproto.Window) (string, error) {
	class, err := x.property(w, x.classAtom)
	if err != nil {
		return "", err
	}
	zero := []byte{0}
	s := bytes.Split(bytes.TrimSuffix(class.Value, zero), zero)
	if l := len(s); l > 0 && len(s[l-1]) != 0 {
		return string(s[l-1]), nil
	}
	return "", errors.New("empty class")
}

func (x Xorg) window() (Window, bool) {
	id := x.active()
	/* skip invalid window id */
	if id == 0 {
		return Window{}, false
	}
	class, err := x.class(id)
	if err != nil {
		return Window{}, false
	}
	name, err := x.name(id)
	if err != nil {
		return Window{}, false
	}
	x.spy(id)
	return Window{
		Class: class,
		Name:  name,
	}, true
}

func (x Xorg) spy(w xproto.Window) {
	xproto.ChangeWindowAttributes(x.X, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
}

func (x Xorg) update(t Tracker) (current *Track) {
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

func connect() Xorg {
	var x Xorg
	var err error

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	x.X, err = xgb.NewConnDisplay(display)
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

func (t Tracker) cleanup(d time.Duration) {
	m.Lock()
	for k, v := range t {
		if time.Since(v.Seen) > d {
			logger.Println(v, k)
			delete(t, k)
		}
	}
	m.Unlock()
}

func load(fname string) Tracker {
	t := make(Tracker)
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

func (t Tracker) store(fname string) {
	tmp := fname+".tmp"
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

type Index struct {
	Title   string
	Records Records
	Classes Records
	Total   Duration
	Zzz     bool
}

type Records []Record
type Duration time.Duration

type Record struct {
	Class string
	Name  string
	Spent Duration
	Seen  time.Time
	Odd   bool `json:"-"`
}

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (d Duration) String() string {
	h := int(time.Duration(d).Hours())
	m := int(time.Duration(d).Minutes()) % 60
	s := int(time.Duration(d).Seconds()) % 60
	var ret string
	if h > 0 {
		ret += fmt.Sprintf("%dh", h)
	}
	if m > 0 {
		ret += fmt.Sprintf("%dm", m)
	}
	return ret + fmt.Sprintf("%ds", s)
}

func (d Duration) Seconds() int {
	return int(time.Duration(d).Seconds())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Title = "Gone Time Tracker"
	idx.Zzz = zzz
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	m.Lock()
	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent)})
	}
	m.Unlock()
	for k, v := range classtotal {
		idx.Classes = append(idx.Classes, Record{Class: k, Spent: Duration(v)})
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
			Spent: Duration(v.Spent),
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
	tracks.cleanup(0)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	logfile, err := os.OpenFile(filepath.Join(goneDir, logf), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.LstdFlags)

	dumpPath := filepath.Join(goneDir, dump)
	tracks = load(dumpPath)

	go tracks.collect()
	go func() {
		for {
			tracks.cleanup(8 * time.Hour)
			tracks.store(dumpPath)
			time.Sleep(time.Minute)
		}
	}()
	log.Println("listen on", port)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gone.json", dumpHandler)
	http.HandleFunc("/reset", resetHandler)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
