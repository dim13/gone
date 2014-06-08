Gone Time Tracker
=================

Name
----

Where has my time gone? X11 automatic work activity tracker in pure Go.

Synopsis
--------

_Gone_ performs automatic time accounting on EWMH capable Window Managers by
looking at _NET_ACTIVE_WINDOW and storing spent time on a particular window.

_Gone_ is aware of ScreenSaver and suspends accounting if ScreenSaver triggers.

Results are presented at http://localhost:8001/

Installation
------------

    go get github.com/dim13/gone

Caveats
-------

For _xmonad_ Window Manager _EwmhDesktop_ extension is required.

_xscreensaver_ seems to ignore MIT-SCREEN-SAVER extention.
So you can either use xidle / xlock instead and/or setup X11 ScreenSaver with:

    xset s 600
