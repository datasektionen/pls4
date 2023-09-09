package admin

import (
	"log/slog"
	"net/http"
)

func Mount(admin Admin) {
	http.Handle("/", route(admin, index))
	http.Handle("/role/", http.StripPrefix("/role/", route(admin, role)))

	http.Handle("/login", route(admin, login))
	http.Handle("/logout", route(admin, logout))
}

func route(admin Admin, handler func(t Admin, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(admin, w, r)
	}
}

func index(admin Admin, w http.ResponseWriter, r *http.Request) {
	roles, err := admin.ListRoles(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not get roles", "error", err)
	}
	t := admin.Roles(roles)

	if r.Header.Get("HX-Boosted") == "true" {
		err = admin.Render(w, t)
	} else {
		err = admin.RenderWithLayout(w, t, admin.LoggedInKTHID(r))
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not render template", "error", err)
	}
}

func role(admin Admin, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Path
	role, err := admin.GetRole(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not get role", "error", err, "role_id", id)
	}
	subroles, err := admin.GetSubroles(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not get subroles", "error", err, "role_id", id)
	}
	members, err := admin.GetRoleMembers(ctx, id, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not get role members", "error", err, "role_id", id)
	}
	t := admin.Role(role, subroles, members)

	if r.Header.Get("HX-Boosted") == "true" {
		err = admin.Render(w, t)
	} else {
		err = admin.RenderWithLayout(w, t, admin.LoggedInKTHID(r))
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not render template", "error", err)
	}
}

func login(admin Admin, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	sessionToken, err := admin.Login(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Could not verify login code", "error", err)
		return
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
