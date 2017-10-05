// Gone Time Tracker -or- Where has my time gone?
package main

//go:generate go get github.com/mjibson/esc
//go:generate esc -ignore '^\..*' -o static.go static/

import (
	"flag"
	"log"
	"os"
)

func main() {
	var (
		display = flag.String("display", os.Getenv("DISPLAY"), "X11 display")
		listen  = flag.String("listen", "127.0.0.1:8001", "web reporter")
	)
	flag.Parse()

	X, err := Connect(*display)
	if err != nil {
		log.Fatal(err)
	}
	defer X.Close()

	app := NewApp(NewBroker())

	go X.Collect(app)

	if err := app.ListenAndServe(*listen); err != nil {
		log.Fatal(err)
	}
}
