package templates

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
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
		"nilUUID": func() uuid.UUID {
			return uuid.Nil
		},
		"now": func() time.Time {
			return time.Now()
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

func (s *Templates) Error(code int, messages ...string) Template {
	return Template{code, "error.html", map[string]any{
		"StatusCode": code,
		"StatusText": http.StatusText(code),
		"Messages":   messages,
	}}
}

func (s *Templates) Roles(roles []models.Role, mayCreate, mayDelete bool) Template {
	return Template{http.StatusOK, "roles.html", map[string]any{
		"Roles":     roles,
		"MayCreate": mayCreate,
		"MayDelete": mayDelete,
	}}
}

func (s *Templates) Role(role models.Role, subroles []models.Role, members []models.Member, permissions []models.SystemPermissions, mayUpdate bool) Template {
	return Template{http.StatusOK, "role.html", map[string]any{
		"ID":          role.ID,
		"DisplayName": role.DisplayName,
		"Description": role.Description,
		"Subroles":    subroles,
		"Members":     members,
		"Permissions": permissions,
		"MayUpdate":   mayUpdate,
	}}
}

func (s *Templates) RoleName(id, displayName string, mayUpdate bool) Template {
	return Template{http.StatusOK, "role-name", map[string]any{
		"ID":          id,
		"DisplayName": displayName,
		"MayUpdate":   mayUpdate,
	}}
}

func (s *Templates) RoleEditName(id, displayName string) Template {
	return Template{http.StatusOK, "role-edit-name", map[string]any{"ID": id, "DisplayName": displayName}}
}

func (s *Templates) RoleDescription(id, description string, mayUpdate bool) Template {
	return Template{http.StatusOK, "role-description", map[string]any{
		"ID":          id,
		"Description": description,
		"MayUpdate":   mayUpdate,
	}}
}

func (s *Templates) RoleEditDescription(id, description string) Template {
	return Template{http.StatusOK, "role-edit-description", map[string]any{"ID": id, "Description": description}}
}

func (s *Templates) Subroles(id string, subroles []models.Role, mayUpdate bool) Template {
	return Template{http.StatusOK, "subroles.html", map[string]any{
		"ID":        id,
		"Subroles":  subroles,
		"MayUpdate": mayUpdate,
	}}
}

func (s *Templates) Members(id string, members []models.Member, mayUpdate bool, toUpdateMemberID uuid.UUID, addNew bool) Template {
	return Template{http.StatusOK, "members.html", map[string]any{
		"ID":               id,
		"Members":          members,
		"MayUpdate":        mayUpdate,
		"ToUpdateMemberID": toUpdateMemberID,
		"AddNew":           addNew,
	}}
}

func (s *Templates) RoleAddSubroleForm(id string, options []models.Role) Template {
	return Template{http.StatusOK, "add-subrole", map[string]any{"ID": id, "Options": options}}
}

func (s *Templates) CreateRoleForm() Template {
	return Template{http.StatusOK, "create-role", nil}
}
