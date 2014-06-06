package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"log"
	"strings"
	"time"
)

type tracker map[window]*track

type track struct {
	Start time.Time
	Spent time.Duration
}

type window struct {
	Class string
	Name  string
}

type xorg struct {
	X           *xgb.Conn
	root        xproto.Window
	activeAtom  *xproto.InternAtomReply
	netNameAtom *xproto.InternAtomReply
	nameAtom    *xproto.InternAtomReply
	classAtom   *xproto.InternAtomReply
}

func (t track) String() string {
	return fmt.Sprint(t.Spent)
}

func (w window) String() string {
	return fmt.Sprintf("%s %s", w.Class, w.Name)
}

func (x xorg) atom(aname string) *xproto.InternAtomReply {
	a, err := xproto.InternAtom(x.X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal("atom: ", err)
	}
	return a
}

func (x xorg) property(w xproto.Window, a *xproto.InternAtomReply) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(x.X, false, w, a.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
}

func (x xorg) active() xproto.Window {
	p, err := x.property(x.root, x.activeAtom)
	if err != nil {
		log.Fatal("active: ", err)
	}
	return xproto.Window(xgb.Get32(p.Value))
}

func (x xorg) name(w xproto.Window) string {
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

func (x xorg) class(w xproto.Window) string {
	class, err := x.property(w, x.classAtom)
	if err != nil {
		log.Fatal("class: ", err)
	}
	i := bytes.IndexByte(class.Value, 0)
	if i == -1 {
		return ""
	}
	return string(class.Value[:i])
}

func (x xorg) winName() (window, bool) {
	windowId := x.active()
	/* skip invalid window id */
	if windowId == 0 {
		return window{}, false
	}
	x.spy(windowId)
	return window{
		Class: x.class(windowId),
		Name:  x.name(windowId),
	}, true
}

func (x xorg) spy(w xproto.Window) {
	xproto.ChangeWindowAttributes(x.X, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
}

func (x xorg) Update(t tracker) (prev *track) {
	if win, ok := x.winName(); ok {
		if _, ok := t[win]; !ok {
			t[win] = new(track)
		}
		t[win].Start = time.Now()
		prev = t[win]
	}
	return
}

func connect() xorg {
	var x xorg
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

func (t tracker) collect() {
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
		case screensaver.NotifyEvent:
			switch event.State {
			case screensaver.StateOn:
				fmt.Println("away from keyboard")
				prev = nil
			}
		}
	}
}

func (t tracker) String() string {
	var ret []string
	var total time.Duration
	classtotal := make(map[string]time.Duration)
	for k, v := range t {
		ret = append(ret, fmt.Sprintf("%s %s", k, v))
		total += v.Spent
		classtotal[k.Class] += v.Spent
	}
	ret = append(ret, "")
	for k, v := range classtotal {
		ret = append(ret, fmt.Sprintf("%s %s", k, v))
	}
	ret = append(ret, fmt.Sprintf("Total %s", total))
	ret = append(ret, "")
	return strings.Join(ret, "\n")
}

func (t tracker) cleanup(d time.Duration) {
	for k, v := range t {
		if time.Since(v.Start) > d {
			log.Println("removing", k)
			delete(t, k)
		}
	}
}

func main() {
	tracks := make(tracker)
	go tracks.collect()
	go func() {
		for {
			tracks.cleanup(12 * time.Hour)
			time.Sleep(5 * time.Minute)
		}
	}()
	for {
		fmt.Println(tracks)
		time.Sleep(3 * time.Second)
	}
}
