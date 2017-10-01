package main

import "time"

const (
	EventSeen = "seen"
	EventIdle = "idle"
)

type App struct{}

func (a App) Seen(w Window)      {}
func (a App) Idle(time.Duration) {}
