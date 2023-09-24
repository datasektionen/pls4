package admin

import (
	"database/sql"
	"embed"
	"html/template"
	"time"

	"github.com/datasektionen/pls4/api"
)

//go:embed templates/*.html
var templates embed.FS

type Admin struct {
	db          *sql.DB
	api         *api.API
	t           *template.Template
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

	var err error
	s.t, err = template.New("").Funcs(map[string]any{
		"date": func(date time.Time) string {
			return date.Format(time.DateOnly)
		},
	}).ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s.db = db

	go s.deleteOldSessionsForever()

	return s, nil
}
