package admin

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/datasektionen/pls4/models"
)

type Template struct {
	name string
	code int
	data any
}

func (s *Admin) RenderWithLayout(wr io.Writer, t Template, userID string) error {
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

func (s *Admin) Render(wr io.Writer, t Template) error {
	return s.t.ExecuteTemplate(wr, t.name, t.data)
}

func (s *Admin) Roles(roles []models.Role) Template {
	return Template{"roles.html", http.StatusOK, roles}
}

func (s *Admin) Role(role models.Role, subroles []models.Role, members []models.Member, canUpdate bool) Template {
	return Template{"role.html", http.StatusOK, map[string]any{
		"ID": role.ID,
		"DisplayName": role.DisplayName,
		"Description": role.Description,
		"Subroles":    subroles,
		"Members":     members,
		"CanUpdate":   canUpdate,
	}}
}

func (s *Admin) RoleName(id, displayName string, canUpdate bool) Template {
	return Template{"role-name", http.StatusOK, map[string]any{
		"ID": id,
		"DisplayName": displayName,
		"CanUpdate": canUpdate,
	}}
}

func (s *Admin) RoleEditName(id, displayName string) Template {
	return Template{"role-edit-name", http.StatusOK, map[string]any{"ID": id, "DisplayName": displayName}}
}

func (s *Admin) Error(code int, messages ...string) Template {
	return Template{"error.html", code, map[string]any{
		"StatusCode": code,
		"StatusText": http.StatusText(code),
		"Messages":   messages,
	}}
}
