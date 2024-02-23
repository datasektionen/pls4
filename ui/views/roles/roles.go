package roles

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
)

func Index(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return renderRoles(ui, ctx, session)
}

func GetRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	role, err := ui.GetRole(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return errors.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	subroles, err := ui.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	members, err := ui.GetRoleMembers(ctx, roleID, true, true)
	if err != nil {
		slog.Error("Could not get role members", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	permissions, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	mayAddPermissions, err := ui.MayAddPermissions(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range permissions {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)

	return roleComponent(*role, subroles, members, permissions, mayUpdate, mayAddPermissions, mayDeleteInSystems)
}

func CreateRoleForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	roles, err := ui.GetUserRoles(r.Context(), session.KTHID)
	if err != nil {
		slog.Error("Could not get roles for user", "error", err, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return createRoleForm(roles)
}

func CreateRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	roleID := r.FormValue("id")
	displayName := r.FormValue("display-name")
	description := r.FormValue("description")
	owner := r.FormValue("owner")
	if err := ui.CreateRole(ctx, session.KTHID, roleID, displayName, description, owner); err != nil {
		slog.Error("Could not create role", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func DeleteRole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	if err := ui.DeleteRole(ctx, session.KTHID, roleID); err != nil {
		slog.Error("Could not delete role", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderRoles(ui, ctx, session)
}

func RoleNameForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return errors.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return roleNameForm(*role)
}

func UpdateRoleName(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	displayName := r.FormValue("display-name")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return errors.Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, displayName, ""); err != nil {
		slog.Error("Could not update role name", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return roleNameDisplay(roleID, displayName, true)
}

func RoleDescriptionForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	role, err := ui.GetRole(r.Context(), roleID)
	if err != nil {
		slog.Error("Could not get role", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}
	if role == nil {
		return errors.Error(http.StatusNotFound, "No role with id "+roleID)
	}
	return roleDescriptionForm(*role)
}

func UpdateRoleDescription(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	description := r.FormValue("description")
	session, err := ui.GetSession(r)
	if err != nil {
		// TODO: redirect to login?
		return errors.Error(http.StatusUnauthorized)
	}
	if err := ui.UpdateRole(r.Context(), session.KTHID, roleID, "", description); err != nil {
		slog.Error("Could not update role description", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return roleDescriptionDisplay(roleID, description, true)
}

func renderRoles(ui *service.UI, ctx context.Context, session service.Session) templ.Component {
	roles, err := ui.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not get roles", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	mayCreate, err := ui.MayCreateRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may create roles", "error", err, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	var roleIDs []string
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	deletable, err := ui.MayDeleteRoles(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not check if user may delete roles", "error", err, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return roleList(roles, mayCreate, deletable)
}
