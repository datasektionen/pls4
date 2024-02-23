package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
)

func renderRoles(ui *service.UI, ctx context.Context, session service.Session) templ.Component {
	roles, err := ui.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return Error(http.StatusInternalServerError)
	}
	mayCreate, err := ui.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return Error(http.StatusInternalServerError)
	}
	var roleIDs []string
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	deletable, err := ui.MayDeleteRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return Error(http.StatusInternalServerError)
	}
	return Roles(roles, mayCreate, deletable)
}

func index(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}
	return renderRoles(ui, ctx, session)
}

func getRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	role, err := ui.GetRole(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	if role == nil {
		return Error(http.StatusNotFound, "No role with id "+roleID)
	}
	subroles, err := ui.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	members, err := ui.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	permissions, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}

	mayAddPermissions, err := ui.MayAddPermissions(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)

	return Role(*role, subroles, members, permissions, mayUpdate, mayAddPermissions, mayDeleteInSystems)
}

func createRoleForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roles, err := ui.GetUserRoles(r.Context(), session.KTHID)
	if err != nil {
		slog.Error("Could not get roles for user", "error", err, "kth_id", session.KTHID)
		return Error(http.StatusInternalServerError)
	}
	return CreateRoleForm(roles)
}

func createRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roleID := r.FormValue("id")
	displayName := r.FormValue("display-name")
	description := r.FormValue("description")
	owner := r.FormValue("owner")
	if err := ui.CreateRole(ctx, session.KTHID, roleID, displayName, description, owner); err != nil {
		slog.Error("Could not create role", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func deleteRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	if err := ui.DeleteRole(ctx, session.KTHID, roleID); err != nil {
		slog.Error("Could not delete role", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func roleNameForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	if role == nil {
		return Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return RoleNameForm(*role)
}

func updateRoleName(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	displayName := r.FormValue("display-name")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, displayName, ""); err != nil {
		slog.Error("Could not update role name", "error", err)
		return Error(http.StatusInternalServerError)
	}
	return RoleNameDisplay(roleID, displayName, true)
}

func roleDescriptionForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	if role == nil {
		return Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return RoleDescriptionForm(*role)
}

func updateRoleDescription(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	description := r.FormValue("description")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, "", description); err != nil {
		slog.Error("Could not update role description", "error", err)
		return Error(http.StatusInternalServerError)
	}
	return RoleDescriptionDisplay(roleID, description, true)
}

