package ui

import (
	"database/sql"

	"github.com/datasektionen/pls4/api"
)

type UI struct {
	db               *sql.DB
	api              *api.API
	loginFrontendURL string
	loginAPIURL      string
	loginAPIKey      string
	hodisURL         string
}

func New(db *sql.DB, api *api.API, loginFrontendURL, loginAPIURL, loginAPIKey, hodisURL string) (*UI, error) {
	s := &UI{}

	s.api = api
	s.loginFrontendURL = loginFrontendURL
	s.loginAPIURL = loginAPIURL
	s.loginAPIKey = loginAPIKey
	s.hodisURL = hodisURL

	s.db = db

	go s.deleteOldSessionsForever()

	return s, nil
}
