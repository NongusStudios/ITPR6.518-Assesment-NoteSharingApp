package main

import "log"

var (
	Version string
	Build   string
)

func main() {
	app, err := InitApp()

	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
