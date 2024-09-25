package main

import (
	"html/template"
	"net/http"

	"github.com/icza/session"
)

type DashboardData struct {
	Username string
}

func executeTemplate(w http.ResponseWriter, name, path string, funcMap template.FuncMap, data any) {
	t, err := template.New(name).Funcs(funcMap).ParseFiles(path)
	checkInternalServerError(err, w)
	err = t.Execute(w, data)
	checkInternalServerError(err, w)
}

func createUserSession(w http.ResponseWriter, user User) {
	s := session.NewSessionOptions(&session.SessOptions{
		CAttrs: map[string]interface{}{"username": user.username, "userid": user.id},
		Attrs:  map[string]interface{}{"count": 1},
	})
	session.Add(s, w)
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}

func (a *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	sess := session.Get(r)
	user := "[guest]"

	if sess != nil {
		user = sess.CAttr("username").(string)
	}

	tmplData := DashboardData{
		Username: user,
	}

	executeTemplate(w, "dashboard.html", "web/dashboard.html",
		template.FuncMap{},
		tmplData)
}
