package admin

import (
	"database/sql"
	"embed"
	"html/template"
	"io"
	"net/http"
)

//go:embed templates/*.html
var templates embed.FS

type Admin interface {
	LoggedInName(r *http.Request) string
	Login(code string) (string, error)

	RenderIndex(wr io.Writer, p IndexParameters) error
}

type service struct {
	db          *sql.DB
	t           *template.Template
	loginURL    string
	loginAPIKey string
	hodisURL    string
	sessions    map[string]session
}

func New(db *sql.DB, loginURL, loginAPIKey, hodisURL string) (Admin, error) {
	s := &service{}

	s.loginURL = loginURL
	s.loginAPIKey = loginAPIKey
	s.hodisURL = hodisURL

	var err error
	s.t, err = template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s.db = db
	s.sessions = make(map[string]session)

	go s.deleteOldSessionsForever()

	return s, nil
}
