package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// isWriteable checks if dir can be written to by opening gone.log for writing.
// If the file did not exist and was created, isWriteable does not remove it.
func isWriteable(dir string) bool {
	f, err := os.OpenFile(filepath.Join(dir, "gone.log"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// getGoneDir tries to find the gone executable and returns a directory where it was found.
func getGoneDir() (string, error) {
	goneDir := filepath.Dir(os.Args[0])
	if goneDir[0] != '/' {
		path, err := exec.LookPath(os.Args[0])
		if err != nil {
			return "", err
		}
		goneDir = filepath.Dir(path)
		if goneDir[0] != '/' {
			cwd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			goneDir = filepath.Join(cwd, goneDir)
		}
	}
	return goneDir, nil
}

// findWriteableDir returns the first writeable directory.
func findWriteableDir(dirs ...string) (string, error) {
	for _, dir := range dirs {
		if isWriteable(dir) {
			if dir[0] != '/' {
				cwd, err := os.Getwd()
				if err != nil {
					return "", err
				}
				dir = filepath.Join(cwd, dir)
			}
			return dir, nil
		}
	}
	return "", errors.New("couldn't find a writeable directory")
}
