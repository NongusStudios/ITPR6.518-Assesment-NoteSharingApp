package main

import "log"

var (
	Version string
	Build   string
)

// Initialise app and run
func main() {
	app, err := InitApp()

	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
