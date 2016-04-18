Gone Time Tracker
=================

Where has my time gone? X11 automatic work activity tracker in pure Go.

_Disclaimer_: may contain traces of bugs or be sometimes fubar. :)
_Gone_ is work in progress. Be patient, update often, file bugs and
send pull-requests.  After all, I'm just a hobbyist gopher.

Synopsis
--------

_Gone_ performs automatic time accounting on EWMH capable Window Managers
by looking at _NET_ACTIVE_WINDOW and storing time spent on a particular
application.

_Gone_ is aware of ScreenSaver and suspends accounting if ScreenSaver
triggers.

As fallback (see caveats) _gone_ also stops at user inactivity.
The inactive time is counted separated.

Results are presented at http://localhost:8001/

Installation
------------

    go get github.com/dim13/gone

Caveats
-------

* For _xmonad_ Window Manager _EwmhDesktop_ extension is required.

* _xscreensaver_ seems to ignore MIT-SCREEN-SAVER extension.
Use xidle/xlock instead and/or activate X11 ScreenSaver:

    xset s default

Alternatives
------------

http://arbtt.nomeata.de/ - automatic, rule-based time tracker

Dockerize
---------

    docker run -ti --rm -e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix dim13/gone
