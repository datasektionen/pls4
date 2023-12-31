package admin

import (
	"log/slog"
	"net/http"
	"time"

	t "github.com/datasektionen/pls4/admin/templates"
	"github.com/google/uuid"
)

func Mount(admin *Admin) {
	http.Handle("/", page(admin, index))
	http.Handle("/roles", partial(admin, roles))
	http.Handle("/role/", http.StripPrefix("/role/", page(admin, role)))
	http.Handle("/role/name", partial(admin, roleName))
	http.Handle("/role/description", partial(admin, roleDescription))
	http.Handle("/role/subrole", partial(admin, roleSubrole))
	http.Handle("/role/member", partial(admin, roleMember))

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
	if r.URL.Path != "/" {
		return admin.t.Error(http.StatusNotFound)
	}
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}
	ctx := r.Context()
	mayCreate, err := admin.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return admin.t.Error(http.StatusInternalServerError)
	}
	mayDelete, err := admin.MayDeleteRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return admin.t.Error(http.StatusInternalServerError)
	}
	roles, err := admin.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Roles(roles, mayCreate, mayDelete)
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
	permissions, err := admin.GetRolePermissions(ctx, id, session.KTHID)
	if err != nil {
		slog.Error("Could not get role persmissions", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Role(*role, subroles, members, permissions, mayUpdate)
}

func roles(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	if r.Method == http.MethodGet {
		return admin.t.CreateRoleForm()
	}
	ctx := r.Context()
	action := r.FormValue("action")
	id := r.FormValue("id")

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}

	if action == "Create" {
		displayName := r.FormValue("display-name")
		description := r.FormValue("description")
		if err := admin.CreateRole(ctx, session.KTHID, id, displayName, description); err != nil {
			slog.Error("Could not create role", "error", err, "role_id", id)
			return admin.t.Error(http.StatusInternalServerError)
		}
	} else if action == "Delete" {
		if err := admin.DeleteRole(ctx, session.KTHID, id); err != nil {
			slog.Error("Could not delete role", "error", err, "role_id", id)
			return admin.t.Error(http.StatusInternalServerError)
		}
	}

	mayCreate, err := admin.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return admin.t.Error(http.StatusInternalServerError)
	}
	mayDelete, err := admin.MayDeleteRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return admin.t.Error(http.StatusInternalServerError)
	}
	roles, err := admin.ListRoles(r.Context())
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Roles(roles, mayCreate, mayDelete)
}

func roleName(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	id := r.FormValue("id")

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
	id := r.FormValue("id")

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
	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Subroles(id, subroles, mayUpdate)
}

func roleMember(admin *Admin, w http.ResponseWriter, r *http.Request) t.Template {
	ctx := r.Context()
	id := r.FormValue("id")
	member, _ := uuid.Parse(r.FormValue("member"))
	addNew := r.Form.Has("new")

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return admin.t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}

	if r.Method == http.MethodGet {
		members, err := admin.GetRoleMembers(ctx, id, true, true)
		if err != nil {
			slog.Error("Could not list roles", "error", err)
			return admin.t.Error(http.StatusInternalServerError)
		}
		return admin.t.Members(id, members, mayUpdate, member, addNew)
	}

	if r.Method != http.MethodPost {
		return admin.t.Error(http.StatusMethodNotAllowed)
	}

	action := r.FormValue("action")

	if action == "Remove" {
		if err := admin.RemoveMember(ctx, session.KTHID, id, member); err != nil {
			slog.Error("Could not edit member", "error", err, "member", member)
			return admin.t.Error(http.StatusInternalServerError)
		}
	}

	kthID := r.FormValue("kth-id")
	startDate, err := time.Parse(time.DateOnly, r.FormValue("start-date"))
	if err != nil && r.Form.Has("start-date") {
		return admin.t.Error(http.StatusBadRequest)
	}
	endDate, err := time.Parse(time.DateOnly, r.FormValue("end-date"))
	if err != nil && r.Form.Has("end-date") {
		return admin.t.Error(http.StatusBadRequest)
	}
	comment := r.FormValue("comment")

	if action == "Save" {
		if err := admin.UpdateMember(ctx, session.KTHID, id, member, startDate, endDate, comment); err != nil {
			slog.Error("Could not edit member", "error", err, "member", member)
			return admin.t.Error(http.StatusInternalServerError)
		}
	} else if action == "Add" {
		if err := admin.AddMember(ctx, session.KTHID, id, kthID, comment, startDate, endDate); err != nil {
			slog.Error("Could not add member", "error", err, "role_id", id, "kth_id", kthID)
			return admin.t.Error(http.StatusInternalServerError)
		}
	}

	members, err := admin.GetRoleMembers(ctx, id, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", id)
		return admin.t.Error(http.StatusInternalServerError)
	}
	return admin.t.Members(id, members, mayUpdate, uuid.Nil, false)
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
