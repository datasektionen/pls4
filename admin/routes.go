package admin

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	t "github.com/datasektionen/pls4/admin/templates"
	"github.com/google/uuid"
)

func Mount(admin *Admin) {
	http.Handle("/{$}", page(admin, index))
	http.Handle("POST /role", partial(admin, createRole))
	http.Handle("DELETE /role/{id}", partial(admin, deleteRole))
	http.Handle("GET /role", partial(admin, createRoleForm))
	http.Handle("GET /role/{id}", page(admin, role))
	http.Handle("POST /role/name", partial(admin, updateRoleName))
	http.Handle("GET /role/name", partial(admin, roleNameForm))
	http.Handle("POST /role/description", partial(admin, updateRoleDescription))
	http.Handle("GET /role/description", partial(admin, roleDescriptionForm))
	http.Handle("POST /role/{id}/subrole", partial(admin, roleAddSubrole))
	http.Handle("DELETE /role/{id}/subrole/{subroleID}", partial(admin, roleRemoveSubrole))
	http.Handle("GET /role/subrole", partial(admin, roleSubroleForm))

	http.Handle("GET /role/{id}/member", partial(admin, getRoleMembers))
	http.Handle("POST /role/{id}/member", partial(admin, roleAddMember))
	http.Handle("POST /role/{id}/member/{memberID}", partial(admin, roleUpdateMember))
	http.Handle("DELETE /role/{id}/member/{memberID}", partial(admin, roleRemoveMember))
	http.Handle("POST /role/{id}/member/{memberID}/end", partial(admin, roleEndMember))

	http.Handle("POST /role/{id}/permission", partial(admin, roleAddPermission))
	http.Handle("DELETE /role/{id}/permission/{sysperm}", partial(admin, roleRemovePermission))

	http.Handle("/login", route(admin, login))
	http.Handle("/logout", route(admin, logout))
}

func route(admin *Admin, handler func(s *Admin, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(admin, w, r)
	}
}

func page(admin *Admin, handler func(s *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		component := handler(admin, ctx, w, r)
		var err error
		if e, ok := component.(t.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		layout := t.Body()
		if r.Header.Get("HX-Boosted") != "true" {
			session, _ := admin.GetSession(r)
			layout = t.Layout(session.DisplayName, admin.loginFrontendURL)
		}
		err = layout.Render(templ.WithChildren(ctx, component), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func partial(admin *Admin, handler func(s *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c := handler(admin, ctx, w, r)
		if e, ok := c.(t.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		if err := c.Render(ctx, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func renderRoles(admin *Admin, ctx context.Context, session Session) templ.Component {
	roles, err := admin.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	mayCreate, err := admin.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	var roleIDs []string
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	deletable, err := admin.MayDeleteRoles(ctx, session.KTHID, roleIDs)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.Roles(roles, mayCreate, deletable)
}

func index(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return renderRoles(admin, ctx, session)
}

func createRole(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	id := r.FormValue("id")
	displayName := r.FormValue("display-name")
	description := r.FormValue("description")
	owner := r.FormValue("owner")
	if err := admin.CreateRole(ctx, session.KTHID, id, displayName, description, owner); err != nil {
		slog.Error("Could not create role", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}

	return renderRoles(admin, ctx, session)
}

func deleteRole(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	id := r.PathValue("id")
	if err := admin.DeleteRole(ctx, session.KTHID, id); err != nil {
		slog.Error("Could not delete role", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}

	return renderRoles(admin, ctx, session)
}

func createRoleForm(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roles, err := admin.GetUserRoles(r.Context(), session.KTHID)
	if err != nil {
		slog.Error("Could not get roles for user", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.CreateRoleForm(roles)
}

func role(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.PathValue("id")
	role, err := admin.GetRole(ctx, id)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+id)
	}
	subroles, err := admin.GetSubroles(ctx, id)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	members, err := admin.GetRoleMembers(ctx, id, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	permissions, err := admin.GetRolePermissions(ctx, id, session.KTHID)
	if err != nil {
		slog.Error("Could not get role persmissions", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}

	allSystems, err := admin.GetAllSystems(ctx)
	if err != nil {
		slog.Error("Could not get systems", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	mayAddInSystems, err := admin.MayUpdatePermissionsInSystems(ctx, session.KTHID, allSystems)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := admin.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)

	return t.Role(*role, subroles, members, permissions, mayUpdate, mayAddInSystems, mayDeleteInSystems)
}

func updateRoleName(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.FormValue("id")

	displayName := r.FormValue("display-name")
	session, err := admin.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return t.Error(http.StatusUnauthorized)
	}
	if err := admin.UpdateRole(r.Context(), session.KTHID, id, displayName, ""); err != nil {
		slog.Error("Could not update role name", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.RoleNameDisplay(id, displayName, true)
}

func roleNameForm(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.FormValue("id")

	role, err := admin.GetRole(r.Context(), id)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+id)
	}
	return t.RoleNameForm(*role)
}

func updateRoleDescription(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.FormValue("id")

	description := r.FormValue("description")
	session, err := admin.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return t.Error(http.StatusUnauthorized)
	}
	if err := admin.UpdateRole(r.Context(), session.KTHID, id, "", description); err != nil {
		slog.Error("Could not update role description", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.RoleDescriptionDisplay(id, description, true)
}

func roleDescriptionForm(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.FormValue("id")

	role, err := admin.GetRole(r.Context(), id)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+id)
	}
	return t.RoleDescriptionForm(*role)
}

func roleAddSubrole(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subrole := r.FormValue("subrole")

	if err := admin.AddSubrole(ctx, session.KTHID, roleID, subrole); err != nil {
		slog.Error("Could not add subrole", "error", err, "role_id", roleID, "subrole_id", subrole)
		return t.Error(http.StatusInternalServerError)
	}

	return renderSubroles(admin, ctx, session, roleID)
}

func roleRemoveSubrole(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subroleID := r.PathValue("subroleID")

	if err := admin.RemoveSubrole(ctx, session.KTHID, roleID, subroleID); err != nil {
		slog.Error("Could not remove subrole", "error", err, "role_id", roleID, "subrole_id", subroleID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderSubroles(admin, ctx, session, roleID)
}

func renderSubroles(admin *Admin, ctx context.Context, session Session, roleID string) templ.Component {
	subroles, err := admin.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.Subroles(roleID, subroles, mayUpdate)
}

func roleSubroleForm(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.URL.Query().Get("id")

	options, err := admin.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not list roles", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.AddSubroleForm(id, options)
}

func roleAddMember(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	kthID := r.FormValue("kth-id")

	startDate, err := time.Parse(time.DateOnly, r.FormValue("start-date"))
	if err != nil && r.Form.Has("start-date") {
		return t.Error(http.StatusBadRequest)
	}
	endDate, err := time.Parse(time.DateOnly, r.FormValue("end-date"))
	if err != nil && r.Form.Has("end-date") {
		return t.Error(http.StatusBadRequest)
	}

	if err := admin.AddMember(ctx, session.KTHID, roleID, kthID, startDate, endDate); err != nil {
		slog.Error("Could not add member", "error", err, "role_id", roleID, "kth_id", kthID)
		return t.Error(http.StatusInternalServerError)
	}

	members, err := admin.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Members(roleID, members, mayUpdate, uuid.Nil, false)
}

func roleUpdateMember(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	memberID, err := uuid.Parse(r.PathValue("memberID"))
	if err != nil {
		return t.Error(http.StatusBadRequest, "Invalid syntax for member uuid")
	}

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	startDate, err := time.Parse(time.DateOnly, r.FormValue("start-date"))
	if err != nil && r.Form.Has("start-date") {
		return t.Error(http.StatusBadRequest, "Invalid syntax for start date")
	}
	endDate, err := time.Parse(time.DateOnly, r.FormValue("end-date"))
	if err != nil && r.Form.Has("end-date") {
		return t.Error(http.StatusBadRequest, "Invalid syntax for start date")
	}

	if err := admin.UpdateMember(ctx, session.KTHID, roleID, memberID, startDate, endDate); err != nil {
		slog.Error("Could not edit member", "error", err, "member", memberID)
		return t.Error(http.StatusInternalServerError)
	}

	members, err := admin.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Members(roleID, members, mayUpdate, uuid.Nil, false)
}

func roleRemoveMember(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	if err := admin.RemoveMember(ctx, session.KTHID, roleID, member); err != nil {
		slog.Error("Could not remove member", "error", err, "member", member)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(admin, ctx, session, roleID)
}

func roleEndMember(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	if err := admin.UpdateMember(ctx, session.KTHID, roleID, member, time.Time{}, time.Now().AddDate(0, 0, -1)); err != nil {
		slog.Error("Could not edit member", "error", err, "role_id", roleID, "member", member)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(admin, ctx, session, roleID)
}

func renderMembers(admin *Admin, ctx context.Context, session Session, roleID string) templ.Component {
	members, err := admin.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Members(roleID, members, mayUpdate, uuid.Nil, false)
}

func getRoleMembers(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	id := r.PathValue("id")
	toUpdateMember, _ := uuid.Parse(r.FormValue("updateMemberID"))
	addNew := r.Form.Has("new")

	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := admin.MayUpdateRole(ctx, session.KTHID, id)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", id)
		return t.Error(http.StatusInternalServerError)
	}

	members, err := admin.GetRoleMembers(ctx, id, true, true)
	if err != nil {
		slog.Error("Could not list roles", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.Members(id, members, mayUpdate, toUpdateMember, addNew)
}

func roleAddPermission(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	system := r.FormValue("system")
	permission := r.FormValue("permission")

	if err := admin.AddPermission(ctx, session.KTHID, roleID, system, permission); err != nil {
		slog.Error("Could add permission to role", "error", err, "role_id", roleID, "system", system, "permission", permission)
		return t.Error(http.StatusInternalServerError)
	}

	return renderPermissions(admin, ctx, session, roleID)
}

func roleRemovePermission(admin *Admin, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := admin.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	roleID := r.PathValue("id")
	sysperm := strings.SplitN(r.PathValue("sysperm"), ":", 2)
	if len(sysperm) != 2 {
		return t.Error(http.StatusBadRequest, "Invalid system:permission")
	}
	system, permission := sysperm[0], sysperm[1]

	if err := admin.RemovePermission(ctx, session.KTHID, roleID, system, permission); err != nil {
		slog.Error("Could remove permission from role", "error", err, "role_id", roleID, "system", system, "permission", permission)
		return t.Error(http.StatusInternalServerError)
	}

	return renderPermissions(admin, ctx, session, roleID)
}

func renderPermissions(admin *Admin, ctx context.Context, session Session, roleID string) templ.Component {
	permissions, err := admin.GetRolePermissions(ctx, roleID, session.KTHID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	allSystems, err := admin.GetAllSystems(ctx)
	if err != nil {
		slog.Error("Could not get systems", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	mayAddInSystems, err := admin.MayUpdatePermissionsInSystems(ctx, session.KTHID, allSystems)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := admin.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Permissions(roleID, permissions, mayAddInSystems, mayDeleteInSystems)
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
		SameSite: http.SameSiteLaxMode,
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
