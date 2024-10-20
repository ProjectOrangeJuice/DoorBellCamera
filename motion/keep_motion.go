package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB_NAME string = "./motions.db"

func main() {

	if DBExists(DB_NAME) {
		readyAndListen()
	} else {
		createDatabase()
	}

}

func readyAndListen() {

}

func createDatabase() {
	db, err := sql.Open("sqlite3", DB_NAME)
	failOnError(err, "Error on database creation")
	defer db.Close()

	sqlStmt := `CREATE TABLE 'motion' (
		'motionId'	INTEGER PRIMARY KEY AUTOINCREMENT,
		'motionCode'	TEXT,
		'location'	TEXT,
		'time'	TEXT
	);`

	_, err = db.Exec(sqlStmt)
	failOnError(err, "Error creating table")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func DBExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
