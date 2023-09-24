package api

import (
	"database/sql"
)

type API struct {
	db *sql.DB
}

func New(db *sql.DB) *API {
	return &API{db}
}
