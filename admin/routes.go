package admin

import (
	"log/slog"
	"net/http"

	t "github.com/datasektionen/pls4/admin/templates"
)

func Mount(admin *Admin) {
	http.Handle("/", page(admin, index))
	http.Handle("/role/", http.StripPrefix("/role/", page(admin, role)))
	http.Handle("/role/subrole", partial(admin, roleSubrole))
	http.Handle("/role/name/", http.StripPrefix("/role/name/", partial(admin, roleName)))
	http.Handle("/role/description/", http.StripPrefix("/role/description/", partial(admin, roleDescription)))

	http.Handle("/login", route(admin, login))
	http.Handle("/logout", route(admin, logout))
}

func route(admin *Admin, handler func(s *Admin, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(admin, w, r)
	}
}

func page(admin *Admin, handler func(s *Admin, w http.ResponseWriter, r *http.Request) t.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := handler(admin, w, r)
		var err error
		if r.Header.Get("HX-Boosted") == "true" {
			err = admin.t.Render(w, t)
		} else {
			if t.Code != 0 {
				w.WriteHeader(t.Code)
			}
			session, _ := admin.GetSession(r)
			if err == nil {
				err = admin.t.RenderWithLayout(w, t, session.DisplayName)
			}
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func partial(admin *Admin, handler func(s *Admin, w http.ResponseWriter, r *http.Request) t.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := handler(admin, w, r)
		if err := admin.t.Render(w, t); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func index(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	roles, err := admin.ListRoles(r.Context())
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Roles(roles)
}

func role(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	ctx := r.Context()

	id := r.URL.Path
	role, err := admin.GetRole(ctx, id)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return admin.t.Error(http.StatusNotFound, "No role with id "+id)
	}
	subroles, err := admin.GetSubroles(ctx, id)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	members, err := admin.GetRoleMembers(ctx, id, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	canUpdate, err := admin.CanUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Role(*role, subroles, members, canUpdate)
}

func roleName(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	id := r.URL.Path

	if r.Method == http.MethodPost {
		displayName := r.FormValue("display-name")
		session, err := admin.GetSession(r)
		if err != nil {
			// TODO: redirect to login?
			return admin.t.Error(http.StatusUnauthorized)
		}
		if err := admin.UpdateRole(r.Context(), session.KTHID, id, displayName, ""); err != nil {
			slog.Error("Could not update role name", "error", err)
			return admin.t.Error(http.StatusInternalServerError)
		}
		return admin.t.RoleName(id, displayName, true)
	} else {
		role, err := admin.GetRole(r.Context(), id)
		if err != nil {
			slog.Error("Could not get role", "error", err, "role_id", id)
			return admin.t.Error(http.StatusInternalServerError)
		}
		if role == nil {
			return admin.t.Error(http.StatusNotFound, "No role with id "+id)
		}
		return admin.t.RoleEditName(role.ID, role.DisplayName)
	}
}

func roleDescription(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	id := r.URL.Path

	if r.Method == http.MethodPost {
		description := r.FormValue("description")
		session, err := admin.GetSession(r)
		if err != nil {
			// TODO: redirect to login?
			return admin.t.Error(http.StatusUnauthorized)
		}
		if err := admin.UpdateRole(r.Context(), session.KTHID, id, "", description); err != nil {
			slog.Error("Could not update role description", "error", err)
			return admin.t.Error(http.StatusInternalServerError)
		}
		return admin.t.RoleDescription(id, description, true)
	} else {
		role, err := admin.GetRole(r.Context(), id)
		if err != nil {
			slog.Error("Could not get role", "error", err, "role_id", id)
			return admin.t.Error(http.StatusInternalServerError)
		}
		if role == nil {
			return admin.t.Error(http.StatusNotFound, "No role with id "+id)
		}
		return admin.t.RoleEditDescription(role.ID, role.Description)
	}
}

func roleSubrole(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	ctx := r.Context()
	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("id")

		options, err := admin.ListRoles(ctx)
		if err != nil {
			slog.Error("Could not list roles", "error", err)
			return admin.t.Error(http.StatusInternalServerError)
		}
		return admin.t.RoleAddSubroleForm(id, options)
	}

	if r.Method != http.MethodPost {
		return admin.t.Error(http.StatusMethodNotAllowed)
	}

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}

	id := r.FormValue("id")
	subrole := r.FormValue("subrole")
	action := r.FormValue("action")

	slog.Info("roleSubrole", "id", id, "subrole", subrole)
	if action == "Add" {
		if err := admin.AddSubrole(ctx, session.KTHID, id, subrole); err != nil {
			slog.Error("Could not add subrole", "error", err, "role_id", id, "subrole_id", subrole)
			return admin.t.Error(http.StatusInternalServerError)
		}
	} else if action == "Remove" {
		if err := admin.RemoveSubrole(ctx, session.KTHID, id, subrole); err != nil {
			slog.Error("Could not remove subrole", "error", err, "role_id", id, "subrole_id", subrole)
			return admin.t.Error(http.StatusInternalServerError)
		}
	}

	subroles, err := admin.GetSubroles(ctx, id)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	canUpdate, err := admin.CanUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Subroles(id, subroles, canUpdate)
}

func login(admin *Admin, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	sessionToken, err := admin.Login(code)
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
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(admin *Admin, w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session")
	if cookie != nil {
		admin.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
