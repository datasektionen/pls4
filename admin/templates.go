package admin

import (
	"bytes"
	"html/template"
	"io"

	"github.com/datasektionen/pls4/models"
)

type Template struct {
	name string
	data any
}

func (s *service) RenderWithLayout(wr io.Writer, t Template, userID string) error {
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

func (s *service) Render(wr io.Writer, t Template) error {
	return s.t.ExecuteTemplate(wr, t.name, t.data)
}

func (s *service) Index(roles []models.Role) Template {
	return Template{"index.html", struct{ Roles []models.Role }{Roles: roles}}
}

func (s *service) Role() Template {
	return Template{"role.html", nil}
}
