APP = gone

XDG_DATA_HOME   ?= ${HOME}/.local/share/
XDG_CONFIG_HOME ?= ${HOME}/.config/
XDG_CACHE_HOME  ?= ${HOME}/.cache/

all: install xdg tmpl

install:
	go install -v

xdg:
	install -d ${XDG_DATA_HOME}${APP}
	install -d ${XDG_CONFIG_HOME}${APP}
	install -d ${XDG_CACHE_HOME}${APP}

tmpl:
	install gone.tmpl ${XDG_DATA_HOME}${APP}
