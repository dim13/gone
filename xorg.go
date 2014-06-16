// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
)

type Xorg struct {
	conn        *xgb.Conn
	root        xproto.Window
	activeAtom  *xproto.InternAtomReply
	netNameAtom *xproto.InternAtomReply
	nameAtom    *xproto.InternAtomReply
	classAtom   *xproto.InternAtomReply
	event       chan xgb.Event
}

type Window struct {
	Class string
	Name  string
}

type Tracker interface {
	Update(Window)
	Snooze(time.Duration)
	Wakeup()
}

func (x Xorg) atom(aname string) *xproto.InternAtomReply {
	a, err := xproto.InternAtom(x.conn, true,
		uint16(len(aname)), aname).Reply()
	if err != nil {
		log.Fatal("atom: ", err)
	}
	return a
}

func (x Xorg) property(w xproto.Window,
	a *xproto.InternAtomReply) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(x.conn, false, w, a.Atom,
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
	xproto.ChangeWindowAttributes(x.conn, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
}

func (x Xorg) Close() {
	x.conn.Close()
}

func Connect() Xorg {
	var x Xorg
	var err error

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	x.conn, err = xgb.NewConnDisplay(display)
	if err != nil {
		log.Fatal("xgb: ", err)
	}

	err = screensaver.Init(x.conn)
	if err != nil {
		log.Fatal("screensaver: ", err)
	}

	setup := xproto.Setup(x.conn)
	x.root = setup.DefaultScreen(x.conn).Root

	drw := xproto.Drawable(x.root)
	screensaver.SelectInput(x.conn, drw, screensaver.EventNotifyMask)

	x.activeAtom = x.atom("_NET_ACTIVE_WINDOW")
	x.netNameAtom = x.atom("_NET_WM_NAME")
	x.nameAtom = x.atom("WM_NAME")
	x.classAtom = x.atom("WM_CLASS")
	x.event = make(chan xgb.Event, 1)

	x.spy(x.root)

	return x
}

func (x Xorg) waitForEvent() <-chan xgb.Event {
	go func() {
		ev, err := x.conn.WaitForEvent()
		if err != nil {
			log.Println("wait for event:", err)
		}
		x.event <- ev
	}()
	return x.event
}

func (x Xorg) queryIdle() time.Duration {
	info, err := screensaver.QueryInfo(x.conn,
		xproto.Drawable(x.root)).Reply()
	if err != nil {
		log.Println("query idle:", err)
		return 0
	}
	return time.Duration(info.MsSinceUserInput) * time.Millisecond
}

func (x Xorg) Collect(t Tracker, timeout time.Duration) {
	if win, ok := x.window(); ok {
		t.Update(win)
	}

	for {
		select {
		case event := <-x.waitForEvent():
			switch e := event.(type) {
			case xproto.PropertyNotifyEvent:
				if win, ok := x.window(); ok {
					t.Wakeup()
					t.Update(win)
				}
			case screensaver.NotifyEvent:
				switch e.State {
				case screensaver.StateOn:
					t.Snooze(x.queryIdle())
				default:
					t.Wakeup()
				}
			}
		case <-time.After(timeout):
			t.Snooze(x.queryIdle())
		}
	}
}
