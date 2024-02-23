package ui

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	t "github.com/datasektionen/pls4/ui/templates"
	"github.com/google/uuid"
)

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
		if e, ok := component.(t.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		layout := t.Body()
		if r.Header.Get("HX-Boosted") != "true" {
			session, _ := ui.GetSession(r)
			layout = t.Layout(session.DisplayName, ui.LoginFrontendURL())
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
		if e, ok := c.(t.ErrorComponent); ok {
			w.WriteHeader(e.Code)
		}
		if err := c.Render(ctx, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Could not render template", "error", err)
		}
	}
}

func renderRoles(ui *service.UI, ctx context.Context, session service.Session) templ.Component {
	roles, err := ui.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	mayCreate, err := ui.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	var roleIDs []string
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	deletable, err := ui.MayDeleteRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.Roles(roles, mayCreate, deletable)
}

func index(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return renderRoles(ui, ctx, session)
}

func getRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	role, err := ui.GetRole(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	subroles, err := ui.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	members, err := ui.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	permissions, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayAddPermissions, err := ui.MayAddPermissions(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)

	return t.Role(*role, subroles, members, permissions, mayUpdate, mayAddPermissions, mayDeleteInSystems)
}

func createRoleForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roles, err := ui.GetUserRoles(r.Context(), session.KTHID)
	if err != nil {
		slog.Error("Could not get roles for user", "error", err, "kth_id", session.KTHID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.CreateRoleForm(roles)
}

func createRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.FormValue("id")
	displayName := r.FormValue("display-name")
	description := r.FormValue("description")
	owner := r.FormValue("owner")
	if err := ui.CreateRole(ctx, session.KTHID, roleID, displayName, description, owner); err != nil {
		slog.Error("Could not create role", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func deleteRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	if err := ui.DeleteRole(ctx, session.KTHID, roleID); err != nil {
		slog.Error("Could not delete role", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func roleNameForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return t.RoleNameForm(*role)
}

func updateRoleName(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	displayName := r.FormValue("display-name")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return t.Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, displayName, ""); err != nil {
		slog.Error("Could not update role name", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.RoleNameDisplay(roleID, displayName, true)
}

func roleDescriptionForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return t.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return t.RoleDescriptionForm(*role)
}

func updateRoleDescription(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	description := r.FormValue("description")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return t.Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, "", description); err != nil {
		slog.Error("Could not update role description", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.RoleDescriptionDisplay(roleID, description, true)
}

func roleSubroleForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	options, err := ui.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not list roles", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	return t.AddSubroleForm(roleID, options)
}

func roleAddSubrole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subrole := r.FormValue("subrole")

	if err := ui.AddSubrole(ctx, session.KTHID, roleID, subrole); err != nil {
		slog.Error("Could not add subrole", "error", err, "role_id", roleID, "subrole_id", subrole)
		return t.Error(http.StatusInternalServerError)
	}

	return renderSubroles(ui, ctx, session, roleID)
}

func roleRemoveSubrole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subroleID := r.PathValue("subroleID")

	if err := ui.RemoveSubrole(ctx, session.KTHID, roleID, subroleID); err != nil {
		slog.Error("Could not remove subrole", "error", err, "role_id", roleID, "subrole_id", subroleID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderSubroles(ui, ctx, session, roleID)
}

func renderSubroles(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	subroles, err := ui.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}
	return t.Subroles(roleID, subroles, mayUpdate)
}

func getRoleMembers(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	toUpdateMember, _ := uuid.Parse(r.FormValue("updateMemberID"))
	addNew := r.Form.Has("new")

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	members, err := ui.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Members(roleID, members, mayUpdate, toUpdateMember, addNew)
}

func roleAddMember(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	session, err := ui.GetSession(r)
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

	if err := ui.AddMember(ctx, session.KTHID, roleID, kthID, startDate, endDate); err != nil {
		slog.Error("Could not add member", "error", err, "role_id", roleID, "kth_id", kthID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func roleUpdateMember(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	memberID, err := uuid.Parse(r.PathValue("memberID"))
	if err != nil {
		return t.Error(http.StatusBadRequest, "Invalid syntax for member uuid")
	}

	session, err := ui.GetSession(r)
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

	if err := ui.UpdateMember(ctx, session.KTHID, roleID, memberID, startDate, endDate); err != nil {
		slog.Error("Could not edit member", "error", err, "member", memberID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func roleEndMember(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	if err := ui.UpdateMember(ctx, session.KTHID, roleID, member, time.Time{}, time.Now().AddDate(0, 0, -1)); err != nil {
		slog.Error("Could not edit member", "error", err, "role_id", roleID, "member", member)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func roleRemoveMember(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	if err := ui.RemoveMember(ctx, session.KTHID, roleID, member); err != nil {
		slog.Error("Could not remove member", "error", err, "member", member)
		return t.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func renderMembers(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	members, err := ui.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Members(roleID, members, mayUpdate, uuid.Nil, false)
}

func roleAddPermission(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	system := r.FormValue("system")
	permission := r.FormValue("permission")
	scope := r.FormValue("scope")

	if err := ui.AddPermissionToRole(ctx, session.KTHID, roleID, system, permission, scope); err != nil {
		slog.Error("Could add permission to role", "error", err, "role_id", roleID, "system", system, "permission", permission)
		return t.Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func roleRemovePermission(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}
	roleID := r.PathValue("id")
	instanceID, err := uuid.Parse(r.PathValue("instanceID"))
	if err != nil {
		return t.Error(http.StatusBadRequest, "Invalid uuid syntax")
	}
	// NOTE: we don't check that these match which is only fine as long as we
	// don't rely on that for authorization

	if err := ui.RemovePermission(ctx, session.KTHID, instanceID); err != nil {
		slog.Error("Could remove permission from role", "error", err, "role_id", roleID, "instance_id", instanceID)
		return t.Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func renderPermissions(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	permissions, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return t.Error(http.StatusInternalServerError)
	}

	mayAddPermissions, err := ui.MayAddPermissions(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	return t.Permissions(roleID, permissions, mayAddPermissions, mayDeleteInSystems)
}

func addPermissionForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return t.Error(http.StatusInternalServerError)
	}

	systemSet, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not systems in which user may update permissions", "error", err)
	}
	var systems []string
	for system := range systemSet {
		systems = append(systems, system)
	}

	return t.RoleAddPermissionForm(roleID, systems)
}

func permissionSelect(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")

	permissions, err := ui.GetPermissions(r.Context(), system)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return t.Error(http.StatusInternalServerError)
	}
	return t.PermissionSelect(permissions)
}

func scopeInput(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")
	permission := r.FormValue("permission")

	hasScope, err := ui.PermissionHasScope(r.Context(), system, permission)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return t.Error(http.StatusInternalServerError)
	}
	return t.ScopeInput(hasScope)
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
