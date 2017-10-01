// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"bytes"
	"errors"
	"log"
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
	observed    map[xproto.Window]bool
}

type Window struct {
	ID    int
	Class string
	Name  string
}

type Tracker interface {
	Seen(Window)
	Idle(time.Duration)
}

var (
	ErrNoValue = errors.New("empty value")
	ErrNoClass = errors.New("empty class")
)

func (x Xorg) atom(aname string) (*xproto.InternAtomReply, error) {
	return xproto.InternAtom(x.conn, true, uint16(len(aname)), aname).Reply()
}

func (x Xorg) property(w xproto.Window, a *xproto.InternAtomReply) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(x.conn, false, w, a.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
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
	if name.ValueLen > 0 {
		return string(name.Value), nil
	}
	name, err = x.property(w, x.nameAtom)
	if err != nil {
		return "", err
	}
	if name.ValueLen > 0 {
		return string(name.Value), nil
	}
	return "", ErrNoValue
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
	return "", ErrNoClass
}

func (x Xorg) window() (Window, bool) {
	w := x.active()
	/* skip invalid window id */
	if w == 0 {
		return Window{}, false
	}
	class, err := x.class(w)
	if err != nil {
		return Window{}, false
	}
	name, err := x.name(w)
	if err != nil {
		return Window{}, false
	}
	x.spy(w)
	return Window{ID: int(w), Class: class, Name: name}, true
}

func (x Xorg) spy(w xproto.Window) {
	if !x.observed[w] {
		xproto.ChangeWindowAttributes(x.conn, w, xproto.CwEventMask,
			[]uint32{xproto.EventMaskPropertyChange})
		x.observed[w] = true
	}
}

func (x Xorg) Close() {
	x.conn.Close()
}

func Connect(display string) (Xorg, error) {
	var x Xorg
	var err error

	x.conn, err = xgb.NewConnDisplay(display)
	if err != nil {
		return Xorg{}, err
	}

	err = screensaver.Init(x.conn)
	if err != nil {
		return Xorg{}, err
	}

	setup := xproto.Setup(x.conn)
	x.root = setup.DefaultScreen(x.conn).Root

	drw := xproto.Drawable(x.root)
	screensaver.SelectInput(x.conn, drw, screensaver.EventNotifyMask)

	x.activeAtom, err = x.atom("_NET_ACTIVE_WINDOW")
	if err != nil {
		return Xorg{}, err
	}
	x.netNameAtom, err = x.atom("_NET_WM_NAME")
	if err != nil {
		return Xorg{}, err
	}
	x.nameAtom, err = x.atom("WM_NAME")
	if err != nil {
		return Xorg{}, err
	}
	x.classAtom, err = x.atom("WM_CLASS")
	if err != nil {
		return Xorg{}, err
	}
	x.observed = make(map[xproto.Window]bool)
	x.spy(x.root)

	return x, nil
}

func (x Xorg) waitForEvent(events chan<- xgb.Event) {
	for {
		ev, err := x.conn.WaitForEvent()
		if err != nil {
			log.Println("wait for event:", err)
			continue
		}
		events <- ev
	}
}

func (x Xorg) queryIdle() (time.Duration, error) {
	info, err := screensaver.QueryInfo(x.conn, xproto.Drawable(x.root)).Reply()
	if err != nil {
		return 0, err
	}
	return time.Duration(info.MsSinceUserInput) * time.Millisecond, nil
}

func (x Xorg) Collect(t Tracker, timeout time.Duration) {
	if win, ok := x.window(); ok {
		t.Seen(win)
	}
	events := make(chan xgb.Event, 1)
	go x.waitForEvent(events)
	for {
		select {
		case event := <-events:
			switch e := event.(type) {
			case xproto.PropertyNotifyEvent:
				if win, ok := x.window(); ok {
					//t.Idle(0)
					t.Seen(win)
				}
			case screensaver.NotifyEvent:
				switch e.State {
				case screensaver.StateOn:
					idle, err := x.queryIdle()
					if err != nil {
						log.Println(err)
					}
					t.Idle(idle)
				case screensaver.StateOff:
					t.Idle(0)
				}
			}
		case <-time.After(timeout):
			idle, err := x.queryIdle()
			if err != nil {
				log.Println(err)
			}
			t.Idle(idle)
		}
	}
}
