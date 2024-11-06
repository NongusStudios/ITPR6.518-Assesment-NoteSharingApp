package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/icza/session"
	"golang.org/x/crypto/bcrypt"
)

func isAuthenticated(w http.ResponseWriter, r *http.Request) bool {
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
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}

	return authenticated
}

func setupAuth() {
	// Initialize the session manager - this is a global
	// For testing purposes, we want cookies to be sent over HTTP too (not just HTTPS)
	// refer to the auth.go for the authentication handlers using the sessions
	session.Global.Close()
	session.Global = session.NewCookieManagerOptions(session.NewInMemStore(), &session.CookieMngrOptions{AllowHTTP: true})

}

type AuthData struct {
	LogErrMsg string
	RegErrMsg string
}

var authData AuthData = AuthData{LogErrMsg: "", RegErrMsg: ""}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if method != "POST" {
		executeTemplate(w, "login.html", "web/login.html",
			template.FuncMap{}, authData)
		authData.LogErrMsg = ""
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := a.db.QueryRow("SELECT user_id, username, pass FROM users WHERE username=$1",
		username).Scan(&user.Id, &user.Username, &user.Password)

	if err == sql.ErrNoRows {
		authData.LogErrMsg = "Incorrect Username"
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	checkInternalServerError(err, w)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		authData.LogErrMsg = "Incorrect Password"
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	authData.LogErrMsg = ""

	// Successful Login
	createUserSession(w, user)
	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}

func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if method != "POST" {
		executeTemplate(w, "register.html", "web/register.html",
			template.FuncMap{}, authData)
		authData.RegErrMsg = ""
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// User name can't contain spaces. My reasoning is that sql statements require spaces so sql injection would be impossible
	if !ValidateString(username, []rune{' '}, []ValidateRequire{}) {
		authData.RegErrMsg = "Username can't contain spaces"
		http.Redirect(w, r, "/register", http.StatusMovedPermanently)
		return
	}

	if !ValidateString(password, []rune{' '} /* No spaces */, []ValidateRequire{
		{amount: 2, requiredChar: []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}}, // Atleast 2 numbers
		{amount: 1, requiredChar: []rune{'`', '~', '!', '@', '#', '$', '%', '^', '&', '*', // Atleast 1 special character
			'(', ')', '-', '_', '+', '=', ':', ';', '"', '\'',
			',', '<', '.', '>', '?', '/', '{', '}', '[', ']'}},
	}) {
		authData.RegErrMsg = "Password must contain no spaces, atleast two numbers, and atleast 1 special character (e.g. '@')"
		http.Redirect(w, r, "/register", http.StatusMovedPermanently)
		return
	}

	var user User
	err := a.db.QueryRow("SELECT user_id, username, pass FROM users WHERE username=$1", username).Scan(&user.Id, &user.Username, &user.Password)

	switch {
	case err == sql.ErrNoRows:
		authData.RegErrMsg = ""

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		checkInternalServerError(err, w)

		_, err = a.db.Exec("INSERT INTO users(username, pass) VALUES($1, $2)", username, hashedPassword)
		checkInternalServerError(err, w)

		err = a.db.QueryRow("SELECT user_id FROM users WHERE username=$1", username).Scan(&user.Id)
		checkInternalServerError(err, w)
		_, err = a.db.Exec("INSERT INTO user_settings(user_id, colleagues) VALUES($1, ARRAY[]::INTEGER[])", user.Id)
		checkInternalServerError(err, w)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		authData.RegErrMsg = "User Already Exists."
		http.Redirect(w, r, "/register", http.StatusMovedPermanently)
	}
}

func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	log.Printf("User %s", s.CAttr("username").(string))
	session.Remove(s, w)
	s = nil

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}
