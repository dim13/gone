package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
	"log"
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

func (t track) String() string {
	return fmt.Sprint(t.Spent)
}

func (w window) String() string {
	return fmt.Sprintf("%s: %s", w.Class, w.Name)
}

func getClass(b []byte) string {
	i := bytes.IndexByte(b, 0)
	if i == -1 {
		return ""
	}
	return string(b[:i])
}

func asciizToString(b []byte) (s []string) {
	for _, x := range bytes.Split(b, []byte{0}) {
		s = append(s, string(x))
	}
	if len(s) > 0 && s[len(s)-1] == "" {
		s = s[:len(s)-1]
	}
	return
}

func atom(X *xgb.Conn, aname string) *xproto.InternAtomReply {
	a, err := xproto.InternAtom(X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal("atom: ", err)
	}
	return a
}

func prop(X *xgb.Conn, w xproto.Window, a *xproto.InternAtomReply) *xproto.GetPropertyReply {
	p, err := xproto.GetProperty(X, false, w, a.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		log.Fatal("property: ", err)
	}
	return p
}

func winName(X *xgb.Conn, root xproto.Window) (window, bool) {
	activeAtom := atom(X, "_NET_ACTIVE_WINDOW")
	netNameAtom := atom(X, "_NET_WM_NAME")
	nameAtom := atom(X, "WM_NAME")
	classAtom := atom(X, "WM_CLASS")

	active := prop(X, root, activeAtom)
	windowId := xproto.Window(xgb.Get32(active.Value))

	/* skip root window */
	if windowId == 0 {
		return window{}, false
	}

	spy(X, windowId)

	name := prop(X, windowId, netNameAtom)
	if string(name.Value) == "" {
		name = prop(X, windowId, nameAtom)
	}
	class := prop(X, windowId, classAtom)

	w := window{
		Class: getClass(class.Value),
		Name:  string(name.Value),
	}

	return w, true
}

func rootWin(X *xgb.Conn) xproto.Window {
	setup := xproto.Setup(X)
	return setup.DefaultScreen(X).Root
}

func spy(X *xgb.Conn, w xproto.Window) {
	xproto.ChangeWindowAttributes(X, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
}

func (t tracker) Update(X *xgb.Conn, w xproto.Window) (prev *track) {
	if win, ok := winName(X, w); ok {
		if _, ok := t[win]; !ok {
			t[win] = new(track)
		}
		t[win].Start = time.Now()
		prev = t[win]
	}
	return
}

func collect(tracks tracker) {
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal("xgb: ", err)
	}
	defer X.Close()

	err = screensaver.Init(X)
	if err != nil {
		log.Fatal("screensaver: ", err)
	}

	root := rootWin(X)

	drw := xproto.Drawable(root)
	screensaver.SelectInput(X, drw, screensaver.EventNotifyMask)

	spy(X, root)
	prev := tracks.Update(X, root)

	for {
		ev, everr := X.WaitForEvent()
		if everr != nil {
			log.Fatal("wait for event: ", everr)
		}
		switch event := ev.(type) {
		case xproto.PropertyNotifyEvent:
			if prev != nil {
				prev.Spent += time.Since(prev.Start)
			}
			prev = tracks.Update(X, root)
		case screensaver.NotifyEvent:
			switch event.State {
			case screensaver.StateOn:
				fmt.Println("away from keyboard")
				prev = nil
			}
		}
	}
}

func display(tracks tracker) {
	for {
		var total time.Duration
		classtotal := make(map[string]time.Duration)
		for n, t := range tracks {
			fmt.Println(n, t)
			total += t.Spent
			classtotal[n.Class] += t.Spent
		}
		fmt.Println("")
		for k, v := range classtotal {
			fmt.Println(k, v)
		}
		fmt.Println("Total:", total)
		fmt.Println("")
		time.Sleep(3 * time.Second)
	}
}

func cleanup(tracks tracker) {
	for {
		for k, v := range tracks {
			if time.Since(v.Start).Hours() > 12.0 {
				log.Println("removing", k)
				delete(tracks, k)
			}
		}
		time.Sleep(time.Minute)
	}
}

func main() {
	tracks := make(tracker)
	go collect(tracks)
	go cleanup(tracks)
	display(tracks)
}
