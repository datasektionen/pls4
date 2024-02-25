package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
	"github.com/datasektionen/pls4/ui/views/members"
	"github.com/datasektionen/pls4/ui/views/permissions"
	"github.com/datasektionen/pls4/ui/views/roles"
	"github.com/datasektionen/pls4/ui/views/subroles"
)

//go:generate templ generate

func Mount(ui *service.UI) {
	http.Handle("/{$}", page(ui, roles.Index))

	http.Handle("GET /role/{id}", page(ui, roles.GetRole))
	http.Handle("GET /role", partial(ui, roles.CreateRoleForm))
	http.Handle("POST /role", partial(ui, roles.CreateRole))
	http.Handle("DELETE /role/{id}", partial(ui, roles.DeleteRole))

	http.Handle("GET /role/{id}/name", partial(ui, roles.RoleNameForm))
	http.Handle("POST /role/{id}/name", partial(ui, roles.UpdateRoleName))

	http.Handle("GET /role/{id}/description", partial(ui, roles.RoleDescriptionForm))
	http.Handle("POST /role/{id}/description", partial(ui, roles.UpdateRoleDescription))

	http.Handle("GET /role/{id}/subrole", partial(ui, subroles.RoleSubroleForm))
	http.Handle("POST /role/{id}/subrole", partial(ui, subroles.RoleAddSubrole))
	http.Handle("DELETE /role/{id}/subrole/{subroleID}", partial(ui, subroles.RoleRemoveSubrole))

	http.Handle("GET /role/{id}/member", partial(ui, members.GetRoleMembers))
	http.Handle("POST /role/{id}/member", partial(ui, members.RoleAddMember))
	http.Handle("POST /role/{id}/member/{memberID}", partial(ui, members.RoleUpdateMember))
	http.Handle("POST /role/{id}/member/{memberID}/end", partial(ui, members.RoleEndMember))
	http.Handle("DELETE /role/{id}/member/{memberID}", partial(ui, members.RoleRemoveMember))

	http.Handle("POST /role/{id}/permission", partial(ui, permissions.RoleAddPermission))
	http.Handle("DELETE /role/{id}/permission/{instanceID}", partial(ui, permissions.RoleRemovePermission))
	http.Handle("GET /role/{id}/add-permission-form", partial(ui, permissions.AddPermissionForm))
	http.Handle("GET /permission-select", partial(ui, permissions.PermissionSelect))
	http.Handle("GET /scope-input", partial(ui, permissions.ScopeInput))

	http.Handle("/login", route(ui, login))
	http.Handle("/logout", route(ui, logout))
}

func route(ui *service.UI, handler func(s *service.UI, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(ui, w, r)
	}
}

func getCtxAndSession(ui *service.UI, w http.ResponseWriter, r *http.Request) (context.Context, service.Session) {
	ctx := r.Context()

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		errors.Error(http.StatusInternalServerError).Render(ctx, w)
		return nil, service.Session{}
	}

	return ctx, session
}

func page(ui *service.UI, handler func(s *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		ctx, session := getCtxAndSession(ui, w, r)
		if ctx == nil {
			return
		}

		component := handler(ui, ctx, session, w, r)
		if e, ok := component.(errors.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		layout := body()
		if r.Header.Get("HX-Boosted") != "true" {
			layout = document(session.DisplayName, ui.LoginFrontendURL())
		}
		if err := layout.Render(templ.WithChildren(ctx, component), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func partial(ui *service.UI, handler func(s *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		ctx, session := getCtxAndSession(ui, w, r)
		if ctx == nil {
			return
		}

		component := handler(ui, ctx, session, w, r)
		if e, ok := component.(errors.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		if err := component.Render(ctx, w); err != nil {
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
