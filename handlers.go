package main

import (
	"html/template"
	"net/http"
)

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	//isAuthenticated(w, r)
	
	// Server index.html
	t, err := template.New("index.html").ParseFiles("web/index.html")
	checkInternalServerError(err, w)
	err = t.Execute(w, nil)
	checkInternalServerError(err, w)
}
