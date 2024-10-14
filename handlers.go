package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/icza/session"
)

type DashboardData struct {
	Username string
	Users    []string
	Notes    []Note
}

type Note struct {
	Owner   string
	Share   []string
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

	rows, err = a.db.Query("SELECT note_owner, note_share, note_name, note_date, note_content FROM notes")
	if err != nil {
		return make([]Note, 0), err
	}

	for rows.Next() {
		note := Note{}
		noteShare := ""
		if e := rows.Scan(&note.Owner, &noteShare, &note.Name, &note.Date, &note.Content); e != nil {
			return notes, err
		}

		note.Share = strings.Split(noteShare, ",")
		notes = append(notes, note)
	}

	return notes, nil
}

// Returns slice of notes that user has access to
func filterNotesByUser(user string, notes []Note) []Note {
	filteredNotes := make([]Note, 0, len(notes))
	for _, note := range notes {
		if slices.IndexFunc(note.Share, func(u string) bool { return u == user || u == "global" }) != -1 || note.Owner == user {
			filteredNotes = append(filteredNotes, note)
		}
	}
	return filteredNotes
}

func (a *App) fetchUsernamesExclude(exclude string) ([]string, error) {
	rows, err := a.db.Query("SELECT username FROM users WHERE username!=$1 AND username!='__placeholder__user__'", exclude)
	if err != nil {
		return make([]string, 0), err
	}

	usernames := []string{}
	for rows.Next() {
		name := ""
		if e := rows.Scan(&name); e != nil {
			return make([]string, 0), err
		}

		usernames = append(usernames, name)
	}

	return usernames, nil
}

func (a *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	sess := session.Get(r)
	user := "[guest]"

	if sess != nil {
		user = sess.CAttr("username").(string)
	}

	notes, err := a.fetchNotes()
	notes = filterNotesByUser(user, notes)

	checkInternalServerError(err, w)

	otherUsers, err := a.fetchUsernamesExclude(user)

	checkInternalServerError(err, w)

	tmplData := DashboardData{
		Username: user,
		Users:    otherUsers,
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

	noteName := r.FormValue("create-note-name")
	noteContent := r.FormValue("create-note-content")

	otherUsers, err := a.fetchUsernamesExclude(user)
	checkInternalServerError(err, w)

	var shareSb strings.Builder

	for _, u := range otherUsers {
		shareToUser := r.FormValue(u)
		if shareToUser == u {
			shareSb.WriteString(",")
			shareSb.WriteString(u)
		}
	}

	var note Note
	err = a.db.QueryRow("SELECT note_name FROM notes WHERE note_name=$1", noteName).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		_, err = a.db.Exec("INSERT INTO notes(note_owner, note_share, note_name, note_date, note_content) VALUES($1, $2, $3, $4, $5)",
			user, shareSb.String(), noteName, time.Now(), noteContent)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}
