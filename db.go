package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func execSqlScript(db *sql.DB, scriptPath string) (error) {
	bytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	sql := string(bytes)
	_, err = db.Exec(sql)

	return err
}

func (a *App) importData(){
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