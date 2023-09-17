package admin

import (
	"log/slog"
	"net/http"
)

func Mount(admin Admin) {
	http.Handle("/", page(admin, index))
	http.Handle("/role/", http.StripPrefix("/role/", page(admin, role)))
	http.Handle("/role/name/", http.StripPrefix("/role/name/", partial(admin, roleName)))

	http.Handle("/login", route(admin, login))
	http.Handle("/logout", route(admin, logout))
}

func route(admin Admin, handler func(t Admin, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(admin, w, r)
	}
}

func page(admin Admin, handler func(t Admin, w http.ResponseWriter, r *http.Request) Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := handler(admin, w, r)
		var err error
		if r.Header.Get("HX-Boosted") == "true" {
			err = admin.Render(w, t)
		} else {
			if t.code != 0 {
				w.WriteHeader(t.code)
			}
			err = admin.RenderWithLayout(w, t, admin.LoggedInKTHID(r))
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func partial(admin Admin, handler func(t Admin, w http.ResponseWriter, r *http.Request) Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := handler(admin, w, r)
		if err := admin.Render(w, t); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func index(admin Admin, w http.ResponseWriter, r *http.Request) Template {
	roles, err := admin.ListRoles(r.Context())
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return admin.Error(http.StatusInternalServerError)
	}
	return admin.Roles(roles)
}

func role(admin Admin, w http.ResponseWriter, r *http.Request) Template {
	ctx := r.Context()

	id := r.URL.Path
	role, err := admin.GetRole(ctx, id)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", id)
		return admin.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return admin.Error(http.StatusNotFound, "No role with id "+id)
	}
	subroles, err := admin.GetSubroles(ctx, id)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", id)
		return admin.Error(http.StatusInternalServerError)
	}
	members, err := admin.GetRoleMembers(ctx, id, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", id)
		return admin.Error(http.StatusInternalServerError)
	}
	kthID := admin.LoggedInKTHID(r)
	canUpdate, err := admin.CanUpdateRole(ctx, kthID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.Error(http.StatusInternalServerError)
	}
	return admin.Role(*role, subroles, members, canUpdate)
}

func roleName(admin Admin, w http.ResponseWriter, r *http.Request) Template {
	id := r.URL.Path

	if r.Method == http.MethodPost {
		displayName := r.FormValue("display-name")
		kthID := admin.LoggedInKTHID(r)
		if err := admin.UpdateRole(r.Context(), kthID, id, displayName); err != nil {
			return admin.Error(http.StatusInternalServerError)
		}
		return admin.RoleName(id, displayName, true)
	} else {
		role, err := admin.GetRole(r.Context(), id)
		if err != nil {
			slog.Error("Could not get role", "error", err, "role_id", id)
			return admin.Error(http.StatusInternalServerError)
		}
		if role == nil {
			return admin.Error(http.StatusNotFound, "No role with id "+id)
		}
		return admin.RoleEditName(role.ID, role.DisplayName)
	}
}

func login(admin Admin, w http.ResponseWriter, r *http.Request) {
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

func logout(admin Admin, w http.ResponseWriter, r *http.Request) {
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
