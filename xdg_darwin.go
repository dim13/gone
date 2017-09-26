package main

import (
	"os"
	"path"
)

func CachePath() string {
	return os.ExpandEnv(path.Join("$HOME", "Library", "Caches"))
}

func ConfigPath() string {
	return os.ExpandEnv(path.Join("$HOME", "Library", "Preferences"))
}

func DataPath() string {
	return os.ExpandEnv(path.Join("$HOME", "Library"))
}
