//go:build !darwin
// +build !darwin

package main

import (
	"os"
	"path"
)

func CachePath() string {
	if p := os.Getenv("XDG_CACHE_HOME"); p != "" {
		return p
	}
	return os.ExpandEnv(path.Join("$HOME", ".cache"))
}

func ConfigPath() string {
	if p := os.Getenv("XDG_CONFIG_HOME"); p != "" {
		return p
	}
	return os.ExpandEnv(path.Join("$HOME", ".config"))
}

func DataPath() string {
	if p := os.Getenv("XDG_DATA_HOME"); p != "" {
		return p
	}
	return os.ExpandEnv(path.Join("$HOME", ".local", "share"))
}
