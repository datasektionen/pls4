package templates

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/datasektionen/pls4/models"
)

type Template struct {
	Code int
	name string
	data any
}

type Templates struct {
	t        *template.Template
	loginURL string
}

//go:embed *.html
var templateFS embed.FS

func New(loginURL string) (*Templates, error) {
	funcs := map[string]any{
		"date": func(date time.Time) string {
			return date.Format(time.DateOnly)
		},
	}

	t, err := template.New("").Funcs(funcs).ParseFS(templateFS, "*.html")
	if err != nil {
		return nil, err
	}

	return &Templates{
		t:        t,
		loginURL: loginURL,
	}, nil
}

func (s *Templates) RenderWithLayout(wr io.Writer, t Template, userID string) error {
	var buffer bytes.Buffer
	if err := s.t.ExecuteTemplate(&buffer, t.name, t.data); err != nil {
		return err
	}
	contents := template.HTML(buffer.String())
	return s.t.ExecuteTemplate(wr, "layout.html", map[string]any{
		"UserID":   userID,
		"LoginURL": s.loginURL,
		"Contents": contents,
	})
}

func (s *Templates) Render(wr io.Writer, t Template) error {
	return s.t.ExecuteTemplate(wr, t.name, t.data)
}

func (s *Templates) Roles(roles []models.Role) Template {
	return Template{http.StatusOK, "roles.html", roles}
}

func (s *Templates) Role(role models.Role, subroles []models.Role, members []models.Member, canUpdate bool) Template {
	return Template{http.StatusOK, "role.html", map[string]any{
		"ID":          role.ID,
		"DisplayName": role.DisplayName,
		"Description": role.Description,
		"Subroles":    subroles,
		"Members":     members,
		"CanUpdate":   canUpdate,
	}}
}

func (s *Templates) RoleName(id, displayName string, canUpdate bool) Template {
	return Template{http.StatusOK, "role-name", map[string]any{
		"ID":          id,
		"DisplayName": displayName,
		"CanUpdate":   canUpdate,
	}}
}

func (s *Templates) RoleEditName(id, displayName string) Template {
	return Template{http.StatusOK, "role-edit-name", map[string]any{"ID": id, "DisplayName": displayName}}
}

func (s *Templates) Error(code int, messages ...string) Template {
	return Template{code, "error.html", map[string]any{
		"StatusCode": code,
		"StatusText": http.StatusText(code),
		"Messages":   messages,
	}}
}
