package sql_conn

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"

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

func DetermineURL(filePath string, defaultURL string) string {
	if _, err := os.Stat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return defaultURL
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			return scanner.Text()
		} else if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
			return defaultURL
		}
	}
	return defaultURL
}

func Prepare(url string) (db *sql.DB, err error) {
	db, err = createDatabase("tick_test", url)
	if err != nil {
		return
	}
	return
}
