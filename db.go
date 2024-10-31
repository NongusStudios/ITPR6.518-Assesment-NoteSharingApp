package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

type User struct {
	Id       int32
	Username string
	Password string
}

type UserSettings struct {
	Id         int32
	UserId     int32
	Colleagues pq.Int32Array
}

type Note struct {
	Owner          int32
	Share          pq.Int32Array
	Name           string
	Date           time.Time
	CompletionDate time.Time
	Flag           int
	Content        string
}

func execSqlScript(db *sql.DB, scriptPath string) error {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	sql := string(bytes)
	_, err = db.Exec(sql)

	return err
}

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
