Gone Time Tracker
=================

Where has my time gone? X11 automatic work activity tracker in pure Go.

Synopsis
--------

_Gone_ performs automatic time accounting on EWMH capable Window Managers by
looking at _NET_ACTIVE_WINDOW and storing spent time on a particular application.

_Gone_ is aware of ScreenSaver and suspends accounting if ScreenSaver triggers.
As fallback (see caveats) it also observes user activity and stops after 5 minutes
of incativity. The inactive time is not counted.

Results are presented at http://localhost:8001/

Installation
------------

    go get github.com/dim13/gone

Caveats
-------

For _xmonad_ Window Manager _EwmhDesktop_ extension is required.

_xscreensaver_ ignores MIT-SCREEN-SAVER extension.
Use xidle/xlock instead and/or activate X11 ScreenSaver

    xset s 600
