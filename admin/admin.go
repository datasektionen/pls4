package admin

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/datasektionen/pls4/api"
	"github.com/datasektionen/pls4/models"
)

//go:embed templates/*.html
var templates embed.FS

type Admin interface {
	GetSession(r *http.Request) (Session, error)
	Login(code string) (string, error)
	DeleteSession(sessionID string) error

	RenderWithLayout(wr io.Writer, t Template, userID string) error
	Render(wr io.Writer, t Template) error
	Roles(roles []models.Role) Template
	Role(role models.Role, subroles []models.Role, members []models.Member, canUpdate bool) Template
	RoleName(id, displayName string, canUpdate bool) Template
	RoleEditName(id, displayName string) Template
	Error(code int, messages ...string) Template

	ListRoles(ctx context.Context) ([]models.Role, error)
	GetRole(ctx context.Context, id string) (*models.Role, error)
	GetSubroles(ctx context.Context, id string) ([]models.Role, error)
	GetRoleMembers(ctx context.Context, id string, onlyCurrent bool, includeIndirect bool) ([]models.Member, error)

	UpdateRole(ctx context.Context, kthID, roleID, displayName string) error

	CanUpdateRole(ctx context.Context, kthID, roleID string) (bool, error)
}

type service struct {
	db          *sql.DB
	api         api.API
	t           *template.Template
	loginURL    string
	loginAPIKey string
	hodisURL    string
}

func New(db *sql.DB, api api.API, loginURL, loginAPIKey, hodisURL string) (Admin, error) {
	s := &service{}

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
