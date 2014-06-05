package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"log"
	"time"
)

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

func winName(X *xgb.Conn) string {
	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	activeAtom := atom(X, "_NET_ACTIVE_WINDOW")
	nameAtom := atom(X, "WM_NAME")

	reply := prop(X, root, activeAtom)
	windowId := xproto.Window(xgb.Get32(reply.Value))

	reply = prop(X, windowId, nameAtom)

	return string(reply.Value)
}

func main() {
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	for {
		//X.WaitForEvent()
		log.Println(winName(X))
		time.Sleep(time.Second)

	}
}
