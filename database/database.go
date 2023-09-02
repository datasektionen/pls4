package database

import (
	"database/sql"
	"flag"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	var databaseURL string
	flag.StringVar(&databaseURL, "database-url",
		os.Getenv("DATABASE_URL"),
		"URL to the postgresql database to use",
	)

	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		panic(err)
	}
}
