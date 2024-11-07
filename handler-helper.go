package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/icza/session"
)

type DashboardData struct {
	CurrentUser         User
	CurrentUserSettings UserSettings
	Users               []User
	Notes               []Note
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

	rows, err = a.db.Query(
		"SELECT note_owner, note_share, note_name, note_date, note_completion_date, note_flag, note_content FROM notes ORDER BY note_id DESC")
	if err != nil {
		return make([]Note, 0), err
	}

	for rows.Next() {
		note := Note{}

		if e := rows.Scan(&note.Owner, &note.Share, &note.Name, &note.Date, &note.CompletionDate, &note.Flag, &note.Content); e != nil {
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

func searchNotes(notes []Note, keyword string, user int, date string, flag int) []Note {
	filteredNotes := make([]Note, 0, len(notes))

	for _, note := range notes {
		hasKeyword, hasUser, hasDate, hasFlag := false, false, false, false

		// Keyword search
		if keyword == "" ||
			strings.Contains(strings.ToLower(note.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(note.Content), strings.ToLower(keyword)) {
			hasKeyword = true
		}

		// User search
		if user == -1 || note.Owner == int32(user) {
			hasUser = true
		}

		// Date search
		if date == "" || note.Date.Format("2006-01-02") == date || (note.Flag == NoteFlagCompleted && note.CompletionDate.Format("2006-01-02") == date) {
			hasDate = true
		}

		// Flag search
		if flag == -1 || note.Flag == flag {
			hasFlag = true
		}

		if hasKeyword && hasUser && hasDate && hasFlag {
			filteredNotes = append(filteredNotes, note)
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

func (a *App) fetchUserSettings(user User) (UserSettings, error) {
	var settings UserSettings
	err := a.db.QueryRow("SELECT setting_id, user_id, colleagues FROM user_settings WHERE user_id=$1", user.Id).Scan(&settings.Id, &settings.UserId, &settings.Colleagues)
	if err != nil {
		return UserSettings{}, err
	}
	return settings, nil
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

// Run this before sending user data to the website
func clearUserPasswordHash(users []User) {
	for i := range users {
		users[i].Password = ""
	}
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
