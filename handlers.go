package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/icza/session"
)

type DashboardData struct {
	Username string
	Notes    []Note
}

type Note struct {
	Owner   int
	Share   []int
	Name    string
	Date    time.Time
	Content string
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

func (a *App) fetchNotes() ([]Note, error) {
	noteCount := 0

	rows, err := a.db.Query("SELECT COUNT(note_id) FROM notes")
	if err != nil {
		return make([]Note, 0), err
	}
	defer rows.Close()

	for rows.Next() {
		if e := rows.Scan(&noteCount); e != nil {
			return make([]Note, 0), e
		}
	}

	notes := make([]Note, 0, noteCount)

	rows, err = a.db.Query("SELECT note_name, note_date, note_content FROM notes")

	for rows.Next() {
		note := Note{}
		if e := rows.Scan(&note.Name, &note.Date, &note.Content); e != nil {
			return notes, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}

func (a *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	sess := session.Get(r)
	user := "[guest]"

	if sess != nil {
		user = sess.CAttr("username").(string)
	}

	notes, err := a.fetchNotes()
	checkInternalServerError(err, w)

	tmplData := DashboardData{
		Username: user,
		Notes:    notes,
	}

	executeTemplate(w, "dashboard.html", "web/dashboard.html",
		template.FuncMap{
			"addOne": func(n int) int {
				return n + 1
			},
		},
		tmplData)
}

func (a *App) getUserIDfromUsername(name string) (int, error) {
	id := 0
	err := a.db.QueryRow("SELECT user_id, username FROM users WHERE username=$1", name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a *App) createNoteHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	sess := session.Get(r)
	user := "[guest]"

	if sess != nil {
		user = sess.CAttr("username").(string)
	}

	// TODO
	userID, err := a.getUserIDfromUsername(user)
	if err != nil {
		return
	}

	log.Print(userID)

	noteName := r.FormValue("create-note-name")
	noteContent := r.FormValue("create-note-content")

	var note Note
	err = a.db.QueryRow("SELECT note_name, FROM notes WHERE note_name=$1", noteName).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		_, err = a.db.Exec("INSERT INTO notes(note_owner, note_name, note_date, note_content) VALUES($1, $2, $3)", noteName, time.Now(), noteContent)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}
