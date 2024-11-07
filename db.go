package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

/* - Entry from 'users' table - */
type User struct {
	Id       int32
	Username string
	Password string
}

/* - Entry from 'user_settings' table - */
type UserSettings struct {
	Id         int32
	UserId     int32
	Colleagues pq.Int32Array
}

/* - Entry from 'notes' table - */
type Note struct {
	Owner          int32
	Share          pq.Int32Array
	Name           string
	Date           time.Time
	CompletionDate time.Time
	Flag           int
	Content        string
}

/*
- Reads a sql script from a file and executes it on the database
Args:

	db: database script will be executed on
	scriptPath: path to the script

return: nil or an error
*/
func execSqlScript(db *sql.DB, scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	sql := string(bytes)
	_, err = db.Exec(sql)

	return err
}

/*
- setup the database by creating required tables
*/
func (a *App) importData() {
	log.Printf("Creating Tables")

	err := execSqlScript(a.db, "sqlScripts/createTables.sql")
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(fmt.Sprintf("./%s", dbFileLock))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}
