Gone Time Tracker
=================

Name
----

Where has my time gone? X11 automatic work ativity tracker in pure Go.

Synopsis
--------

_Gone_ performes automatic time accounting on EWMH capable Window Managers by
looking at _NET_ACTVE_WINDOW and storing spent time on a particular window.

_Gone_ is aware of ScreenSaver and suspends accounting if ScreenSaver triggers.

Results are presented at http://localhost:8001/

Installation
------------

    go get github.com/dim13/gone
