package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"log"
)

type window struct {
	Class []string
	Name  string
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

func (w window) String() string {
	return fmt.Sprintf("%s (%s) %s", w.Class[0], w.Class[1], w.Name)
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
		Class: asciizToString(class.Value),
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

func main() {
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
		if name, ok := winName(X, root); ok {
			log.Println(name)
		}
	}
}
