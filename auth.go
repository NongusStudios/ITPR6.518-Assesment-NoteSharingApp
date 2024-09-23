package main

import (
	"database/sql"
	"net/http"

	"github.com/icza/session"
	"golang.org/x/crypto/bcrypt"
)

func isAuthenticated(w http.ResponseWriter, r *http.Request) {
	authenticated := false

	//m := map[string]interface{}{}

	// get the current session variables
	sess := session.Get(r)

	if sess != nil {
		u := sess.CAttr("username").(string)
		c := sess.Attr("count").(int)

		//just a simple authentication check for the current user
		if c > 0 && len(u) > 0 {
			authenticated = true
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/login", 301)
	}
}

func setupAuth() {
	// Initialize the session manager - this is a global
	// For testing purposes, we want cookies to be sent over HTTP too (not just HTTPS)
	// refer to the auth.go for the authentication handlers using the sessions
	session.Global.Close()
	session.Global = session.NewCookieManagerOptions(session.NewInMemStore(), &session.CookieMngrOptions{AllowHTTP: true})

}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if method != "POST" {
		http.ServeFile(w, r, "web/login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := a.db.QueryRow("SELECT id, username, password FROM user WHERE username=$1",
		                 username).Scan(&user.id, &user.username, &user.password);
	
	if err == sql.ErrNoRows {
		http.Redirect(w, r, "/register", http.StatusMovedPermanently)
		return
	}
	
	checkInternalServerError(err, w)

	err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	// Successful Login
	createUserSession(w, user)
	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}


func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if method != "POST" {
		http.ServeFile(w, r, "web/register.html")
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := a.db.QueryRow("SELECT username, password FROM user WHERE username=$1", username).Scan(&user.username, &user.password)

	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		checkInternalServerError(err, w)

		_, err = a.db.Exec("INSERT INTO user(username, password) VALUES($1, $2)", username, hashedPassword)
		checkInternalServerError(err, w)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
}
