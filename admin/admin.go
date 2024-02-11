package admin

import (
	"database/sql"

	"github.com/datasektionen/pls4/admin/templates"
	"github.com/datasektionen/pls4/api"
)

type Admin struct {
	db          *sql.DB
	api         *api.API
	t           *templates.Templates
	loginURL    string
	loginAPIKey string
	hodisURL    string
}

func New(db *sql.DB, api *api.API, loginURL, loginAPIKey, hodisURL string) (*Admin, error) {
	s := &Admin{}

	s.api = api
	s.loginURL = loginURL
	s.loginAPIKey = loginAPIKey
	s.hodisURL = hodisURL

	s.db = db

	go s.deleteOldSessionsForever()

	return s, nil
}
