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

// Xorg holds X11 relavant properties
type Xorg struct {
	conn        *xgb.Conn
	root        xproto.Window
	activeAtom  *xproto.InternAtomReply
	netNameAtom *xproto.InternAtomReply
	nameAtom    *xproto.InternAtomReply
	classAtom   *xproto.InternAtomReply
	observed    map[xproto.Window]bool
}

// Window description
type Window struct {
	ID    int
	Class string
	Name  string
}

// Tracker interface
type Tracker interface {
	Seen(Window) error
	Idle(time.Duration) error
}

func (x Xorg) atom(aname string) (*xproto.InternAtomReply, error) {
	return xproto.InternAtom(x.conn, true, uint16(len(aname)), aname).Reply()
}

func (x Xorg) atomMust(aname string) *xproto.InternAtomReply {
	a, err := x.atom(aname)
	if err != nil {
		panic(err)
	}
	return a
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
	return "", errors.New("no value")
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
	return "", errors.New("no class")
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
	x.observe(w)
	return Window{ID: int(w), Class: class, Name: name}, true
}

func (x Xorg) observe(w xproto.Window) {
	if x.observed[w] {
		return
	}
	xproto.ChangeWindowAttributes(x.conn, w, xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})
	x.observed[w] = true
}

// Close X11 connection
func (x Xorg) Close() {
	x.conn.Close()
}

// Connect to X11 server
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

	x.root = xproto.Setup(x.conn).DefaultScreen(x.conn).Root
	screensaver.SelectInput(x.conn, xproto.Drawable(x.root), screensaver.EventNotifyMask)

	x.activeAtom = x.atomMust("_NET_ACTIVE_WINDOW")
	x.netNameAtom = x.atomMust("_NET_WM_NAME")
	x.nameAtom = x.atomMust("WM_NAME")
	x.classAtom = x.atomMust("WM_CLASS")

	x.observed = make(map[xproto.Window]bool)
	x.observe(x.root)

	return x, nil
}

func (x Xorg) queryIdle() (time.Duration, error) {
	info, err := screensaver.QueryInfo(x.conn, xproto.Drawable(x.root)).Reply()
	if err != nil {
		return 0, err
	}
	return time.Duration(info.MsSinceUserInput) * time.Millisecond, nil
}

// Collect active window data
func (x Xorg) Collect(t Tracker) {
	if win, ok := x.window(); ok {
		err := t.Seen(win)
		if err != nil {
			log.Println("seen", err)
		}
	}
	for {
		event, err := x.conn.WaitForEvent()
		if err != nil {
			log.Println("wait for event", err)
			continue
		}
		switch ev := event.(type) {
		case xproto.PropertyNotifyEvent:
			if win, ok := x.window(); ok {
				if err := t.Seen(win); err != nil {
					log.Println("seen", err)
				}
			}
		case screensaver.NotifyEvent:
			switch ev.State {
			case screensaver.StateOn:
				idle, err := x.queryIdle()
				if err != nil {
					log.Println("query idle", err)
				}
				if err := t.Idle(idle); err != nil {
					log.Println("idle on", err)
				}
			case screensaver.StateOff:
				if err := t.Idle(0); err != nil {
					log.Println("idle off", err)
				}
			}
		}
	}
}
