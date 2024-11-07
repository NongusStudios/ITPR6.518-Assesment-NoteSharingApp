package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var (
	searchByKeyword = ""
	searchByUser    = -1
	searchByDate    = ""
	searchByFlag    = -1
)

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}

func (a *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	settings, err := a.fetchUserSettings(user)
	checkInternalServerError(err, w)

	notes, err := a.fetchNotes()

	checkInternalServerError(err, w)

	notes = getAccessibleNotes(user, notes)
	notes = searchNotes(notes, searchByKeyword, searchByUser, searchByDate, searchByFlag)

	otherUsers, err := a.fetchUsersExclude(user)
	clearUserPasswordHash(otherUsers)

	checkInternalServerError(err, w)

	tmplData := DashboardData{
		CurrentUser:         user,
		CurrentUserSettings: settings,
		Users:               otherUsers,
		Notes:               notes,
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
			"isColleague": func(settings UserSettings, id int32) bool {
				for _, colleague := range settings.Colleagues {
					if colleague == id {
						return true
					}
				}
				return false
			},
			"shortDate": func(date time.Time) string {
				return date.Format("02/01/2006")
			},
			"completedDate": func(note Note) string {
				if note.Flag == NoteFlagCompleted {
					return note.CompletionDate.Format("02/01/2006")
				}
				return "N/A"
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

func (a *App) searchHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}

	searchByKeyword = r.FormValue("search-by-keyword")
	searchByUser, _ = strconv.Atoi(r.FormValue("search-by-user"))
	searchByDate = r.FormValue("search-by-date")
	searchByFlag, _ = strconv.Atoi(r.FormValue("search-by-flags"))

	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}

func (a *App) createNoteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	noteNameRaw := r.FormValue("create-note-name")
	noteContent := r.FormValue("create-note-content")
	noteFlag, err := strconv.Atoi(r.FormValue("create-note-flags"))

	if err != nil || noteFlag >= NoteFlagMax {
		checkInternalServerError(errors.New("invalid note flag passed from create form"), w)
		return
	}

	noteName := noteNameRaw[:minInt(len(noteNameRaw), NoteNameMaxLength)]

	otherUsers, err := a.fetchUsersExclude(user)
	checkInternalServerError(err, w)

	share := getShareDetails("create", otherUsers, w, r)

	var note Note
	err = a.db.QueryRow("SELECT note_name FROM notes WHERE note_name=$1", noteName).Scan(&note.Name)

	switch {
	case err == sql.ErrNoRows:
		_, err = a.db.Exec("INSERT INTO notes(note_owner, note_share, note_name, note_date, note_completion_date, note_flag, note_content) VALUES($1, $2, $3, $4, $5, $6, $7)",
			user.Id, share, noteName, time.Now(), time.Now(), noteFlag, noteContent)
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
	if !isAuthenticated(w, r) {
		return
	}

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	noteToEdit := r.FormValue("edit-select-note")
	editedNameRaw := r.FormValue("edit-note-name")
	editedContent := r.FormValue("edit-note-content")
	editedFlag, err := strconv.Atoi(r.FormValue("edit-note-flags"))

	if err != nil || editedFlag >= NoteFlagMax {
		checkInternalServerError(errors.New("invalid note flag passed from edit form"), w)
		return
	}

	editedName := editedNameRaw[:minInt(len(editedNameRaw), NoteNameMaxLength)]

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
		_, err = a.db.Exec("UPDATE notes SET note_share=$1, note_name=$2, note_completion_date=$3, note_flag=$4, note_content=$5 WHERE note_name=$6",
			editedShare, editedName, time.Now(), editedFlag, editedContent, noteToEdit)
		checkInternalServerError(err, w)
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
	}
}

func (a *App) deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}

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

func (a *App) editSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(w, r) {
		return
	}

	user, err := a.fetchCurrentUser(r)
	checkInternalServerError(err, w)

	otherUsers, err := a.fetchUsersExclude(user)
	checkInternalServerError(err, w)

	colleagues := getShareDetails("settings", otherUsers, w, r)
	if colleagues[0] == -1 {
		return
	}

	settings, err := a.fetchUserSettings(user)
	checkInternalServerError(err, w)

	_, err = a.db.Exec("UPDATE user_settings SET colleagues=$1 WHERE setting_id=$2", colleagues, settings.Id)
	checkInternalServerError(err, w)

	http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
}
