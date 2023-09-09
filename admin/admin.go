package admin

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"io"
	"net/http"
	"time"

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
	Roles(roles []models.Role) Template
	Role(role models.Role, subroles []models.Role, members []models.Member) Template

	ListRoles(ctx context.Context) ([]models.Role, error)
	GetRole(ctx context.Context, id string) (models.Role, error)
	GetSubroles(ctx context.Context, id string) ([]models.Role, error)
	GetRoleMembers(ctx context.Context, id string, onlyCurrent bool) ([]models.Member, error)
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
	s.t, err = template.New("").Funcs(map[string]any{
		"date": func(date time.Time) string {
			return date.Format(time.DateOnly)
		},
	}).ParseFS(templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s.db = db
	s.sessions = make(map[string]session)

	go s.deleteOldSessionsForever()

	return s, nil
}
