package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/icza/session"
)

type DashboardData struct {
	CurrentUser User
	Users       []User
	Notes       []Note
}

func executeTemplate(w http.ResponseWriter, name, path string, funcMap template.FuncMap, data any) {
	t, err := template.New(name).Funcs(funcMap).ParseFiles(path)
	checkInternalServerError(err, w)
	err = t.Execute(w, data)
	checkInternalServerError(err, w)
}

func createUserSession(w http.ResponseWriter, user User) {
	s := session.NewSessionOptions(&session.SessOptions{
		CAttrs: map[string]interface{}{"username": user.Username, "userid": user.Id},
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

	rows, err = a.db.Query("SELECT note_owner, note_share, note_name, note_date, note_flag, note_content FROM notes ORDER BY note_id DESC")
	if err != nil {
		return make([]Note, 0), err
	}

	for rows.Next() {
		note := Note{}

		if e := rows.Scan(&note.Owner, &note.Share, &note.Name, &note.Date, &note.Flag, &note.Content); e != nil {
			return notes, e
		}

		notes = append(notes, note)
	}

	return notes, nil
}

// Returns slice of notes that user has access to
func getAccessibleNotes(user User, notes []Note) []Note {
	filteredNotes := make([]Note, 0, len(notes))
	for _, note := range notes {
		if len(note.Share) == 0 {
			filteredNotes = append(filteredNotes, note)
		}

		for _, share_id := range note.Share {
			if share_id == user.Id || note.Owner == user.Id {
				filteredNotes = append(filteredNotes, note)
				break
			}
		}
	}
	return filteredNotes
}

func (a *App) fetchCurrentUser(r *http.Request) (User, error) {
	sess := session.Get(r)
	name := "[guest]"

	if sess != nil {
		name = sess.CAttr("username").(string)
	}

	var user User
	err := a.db.QueryRow("SELECT user_id, username, pass FROM users WHERE username=$1", name).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// exclude - exclude this user from list
func (a *App) fetchUsersExclude(exclude User) ([]User, error) {
	rows, err := a.db.Query("SELECT user_id, username, pass FROM users WHERE username!=$1 AND username!='__placeholder__user__'", exclude.Username)
	if err != nil {
		return make([]User, 0), err
	}

	users := []User{}
	for rows.Next() {
		var user User
		if e := rows.Scan(&user.Id, &user.Username, &user.Password); e != nil {
			return make([]User, 0), err
		}

		users = append(users, user)
	}

	return users, nil
}

func (a *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	notes, err := a.fetchNotes()
	checkInternalServerError(err, w)

	notes = getAccessibleNotes(user, notes)

	otherUsers, err := a.fetchUsersExclude(user)

	checkInternalServerError(err, w)

	tmplData := DashboardData{
		CurrentUser: user,
		Users:       otherUsers,
		Notes:       notes,
	}

	executeTemplate(w, "dashboard.html", "web/dashboard.html",
		template.FuncMap{
			"addOne": func(n int) int {
				return n + 1
			},
			"getUserName": func(id int32) string {
				name := ""
				err := a.db.QueryRow("SELECT username FROM users WHERE user_id=$1", id).Scan(&name)
				checkInternalServerError(err, w)

				if name == "__placeholder__user__" {
					return ""
				}

				return name
			},
			"shortDate": func(date time.Time) string {
				return date.Format("02/01/2006")
			},
			"isNoteOwned": func(note Note) bool {
				return note.Owner == user.Id
			},
			"noteFlagToString": func(noteFlag int) string {
				return []string{
					"Note",
					"In Progress",
					"Completed",
					"Cancelled",
					"Delegated",
				}[noteFlag]
			},
			"json": func(s interface{}) string {
				jsonBytes, err := json.Marshal(s)
				if err != nil {
					return ""
				}
				return string(jsonBytes)
			},
		},
		tmplData)
}

func getShareDetails(formIdPrefix string, otherUsers []User, w http.ResponseWriter, r *http.Request) []int {
	var share []int

	for _, u := range otherUsers {
		shareFormValueStr := r.FormValue(formIdPrefix + "-" + u.Username)
		if shareFormValueStr != "" {
			shareFormValue, err := strconv.Atoi(shareFormValueStr)
			checkInternalServerError(err, w)

			share = append(share, shareFormValue)
		}
	}

	if len(share) == 0 {
		share = append(share, -1)
	}

	return share
}

func (a *App) createNoteHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	noteName := r.FormValue("create-note-name")
	noteContent := r.FormValue("create-note-content")
	noteFlag, _ := strconv.Atoi(r.FormValue("create-note-flags"))

	otherUsers, err := a.fetchUsersExclude(user)
	checkInternalServerError(err, w)

	share := getShareDetails("create", otherUsers, w, r)

	var note Note
	err = a.db.QueryRow("SELECT note_name FROM notes WHERE note_name=$1", noteName).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		_, err = a.db.Exec("INSERT INTO notes(note_owner, note_share, note_name, note_date, note_flag, note_content) VALUES($1, $2, $3, $4, $5, $6)",
			user.Id, share, noteName, time.Now(), noteFlag, noteContent)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}

func (a *App) editNoteHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	noteToEdit := r.FormValue("edit-select-note")
	editedName := r.FormValue("edit-note-name")
	editedContent := r.FormValue("edit-note-content")
	editedFlag, _ := strconv.Atoi(r.FormValue("edit-note-flag"))

	otherUsers, err := a.fetchUsersExclude(user)
	checkInternalServerError(err, w)

	editedShare := getShareDetails("edit", otherUsers, w, r)

	var note Note
	err = a.db.QueryRow("SELECT note_name FROM notes WHERE note_name=$1", noteToEdit).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		_, err = a.db.Exec("UPDATE notes SET note_share=$1, note_name=$2, note_flag=$3 note_content=$4 WHERE note_name=$5",
			editedShare, editedName, editedFlag, editedContent, noteToEdit)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}

func (a *App) deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	noteToDelete := r.FormValue("delete-select-note")

	var note Note
	err = a.db.QueryRow("SELECT note_name FROM notes WHERE note_name=$1", noteToDelete).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		_, err = a.db.Exec("DELETE FROM notes WHERE note_name=$1 AND note_owner=$2", noteToDelete, user.Id)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}
