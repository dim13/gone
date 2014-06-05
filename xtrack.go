package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/xgb"
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
		log.Fatal(err)
	}
	return a
}

func prop(X *xgb.Conn, w xproto.Window, a *xproto.InternAtomReply) *xproto.GetPropertyReply {
	p, err := xproto.GetProperty(X, false, w, a.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		log.Fatal(err)
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

func collect(tracks tracker) {
	var prev *track

	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	root := rootWin(X)
	spy(X, root)

	for {
		if _, everr := X.WaitForEvent(); everr != nil {
			log.Fatal(err)
		}
		if prev != nil {
			prev.Spent += time.Since(prev.Start)
		}
		if win, ok := winName(X, root); ok {
			if _, ok := tracks[win]; !ok {
				tracks[win] = new(track)
			}
			tracks[win].Start = time.Now()
			prev = tracks[win]
		}
	}
}

func display(tracks tracker) {
	for {
		var total time.Duration
		classtotal := make(map[string]time.Duration)
		for n, t := range tracks {
			log.Println(n, t)
			total += t.Spent
			classtotal[n.Class] += t.Spent
		}
		fmt.Println("")
		for k, v := range classtotal {
			fmt.Println(k, v)
		}
		fmt.Println("Total:", total)
		fmt.Println("")
		time.Sleep(5 * time.Second)
	}
}

func main() {
	tracks := make(tracker)
	go collect(tracks)
	display(tracks)
}
