package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
)

//go:generate templ generate

func Mount(ui *service.UI) {
	http.Handle("/{$}", page(ui, index))

	http.Handle("GET /role/{id}", page(ui, getRole))
	http.Handle("GET /role", partial(ui, createRoleForm))
	http.Handle("POST /role", partial(ui, createRole))
	http.Handle("DELETE /role/{id}", partial(ui, deleteRole))

	http.Handle("GET /role/{id}/name", partial(ui, roleNameForm))
	http.Handle("POST /role/{id}/name", partial(ui, updateRoleName))

	http.Handle("GET /role/{id}/description", partial(ui, roleDescriptionForm))
	http.Handle("POST /role/{id}/description", partial(ui, updateRoleDescription))

	http.Handle("GET /role/{id}/subrole", partial(ui, roleSubroleForm))
	http.Handle("POST /role/{id}/subrole", partial(ui, roleAddSubrole))
	http.Handle("DELETE /role/{id}/subrole/{subroleID}", partial(ui, roleRemoveSubrole))

	http.Handle("GET /role/{id}/member", partial(ui, getRoleMembers))
	http.Handle("POST /role/{id}/member", partial(ui, roleAddMember))
	http.Handle("POST /role/{id}/member/{memberID}", partial(ui, roleUpdateMember))
	http.Handle("POST /role/{id}/member/{memberID}/end", partial(ui, roleEndMember))
	http.Handle("DELETE /role/{id}/member/{memberID}", partial(ui, roleRemoveMember))

	http.Handle("POST /role/{id}/permission", partial(ui, roleAddPermission))
	http.Handle("DELETE /role/{id}/permission/{instanceID}", partial(ui, roleRemovePermission))
	http.Handle("GET /role/{id}/add-permission-form", partial(ui, addPermissionForm))
	http.Handle("GET /permission-select", partial(ui, permissionSelect))
	http.Handle("GET /scope-input", partial(ui, scopeInput))

	http.Handle("/login", route(ui, login))
	http.Handle("/logout", route(ui, logout))
}

func route(ui *service.UI, handler func(s *service.UI, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(ui, w, r)
	}
}

func page(ui *service.UI, handler func(s *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		ctx := r.Context()
		component := handler(ui, ctx, w, r)
		var err error
		if e, ok := component.(ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		layout := Body()
		if r.Header.Get("HX-Boosted") != "true" {
			session, _ := ui.GetSession(r)
			layout = Layout(session.DisplayName, ui.LoginFrontendURL())
		}
		err = layout.Render(templ.WithChildren(ctx, component), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func partial(ui *service.UI, handler func(s *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		ctx := r.Context()
		c := handler(ui, ctx, w, r)
		if e, ok := c.(ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		if err := c.Render(ctx, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func login(ui *service.UI, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	sessionToken, err := ui.Login(code)
	if err != nil {
		// TODO: this could also be bad/stale request
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not verify login code", "error", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		MaxAge:   60 * 60,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(ui *service.UI, w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session")
	if cookie != nil {
		ui.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func ternary[T any](condition bool, then T, elze T) T {
	if condition {
		return then
	} else {
		return elze
	}
}
