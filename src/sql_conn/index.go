package sql_conn

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func createDatabase(name string, dbPath string) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", dbPath)
	if err != nil {
		return
	}

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS " + name)
	if err != nil {
		return
	}

	return
}

func Prepare(url string) (db *sql.DB, err error) {
	db, err = createDatabase("tick_test", url)
	if err != nil {
		return
	}
	return
}
