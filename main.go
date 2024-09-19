package main

import "log"

func main() {
	app, err := InitApp()

	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
