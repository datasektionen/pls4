package api

import (
	"database/sql"

	"github.com/datasektionen/pls4/database"
)

type service struct {
	db *sql.DB
}

var s service

func init() {
	s.db = database.DB
}
