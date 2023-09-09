package admin

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"io"
	"net/http"

	"github.com/datasektionen/pls4/models"
)

//go:embed templates/*.html
var templates embed.FS

type Admin interface {
	LoggedInKTHID(r *http.Request) string
	Login(code string) (string, error)
	DeleteSession(sessionID string)

	RenderWithLayout(wr io.Writer, t Template, userID string) error
	Render(wr io.Writer, t Template) error
	Index(roles []models.Role) Template
	Role() Template

	ListRoles(ctx context.Context) ([]models.Role, error)
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
