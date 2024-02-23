package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/google/uuid"
)

func roleAddPermission(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	system := r.FormValue("system")
	permission := r.FormValue("permission")
	scope := r.FormValue("scope")

	if err := ui.AddPermissionToRole(ctx, session.KTHID, roleID, system, permission, scope); err != nil {
		slog.Error("Could add permission to role", "error", err, "role_id", roleID, "system", system, "permission", permission)
		return Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func roleRemovePermission(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}
	roleID := r.PathValue("id")
	instanceID, err := uuid.Parse(r.PathValue("instanceID"))
	if err != nil {
		return Error(http.StatusBadRequest, "Invalid uuid syntax")
	}
	// NOTE: we don't check that these match which is only fine as long as we
	// don't rely on that for authorization

	if err := ui.RemovePermission(ctx, session.KTHID, instanceID); err != nil {
		slog.Error("Could remove permission from role", "error", err, "role_id", roleID, "instance_id", instanceID)
		return Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func renderPermissions(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	permissions, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
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
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return Error(http.StatusInternalServerError)
	}

	return Permissions(roleID, permissions, mayAddPermissions, mayDeleteInSystems)
}

func addPermissionForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	systemSet, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not systems in which user may update permissions", "error", err)
	}
	var systems []string
	for system := range systemSet {
		systems = append(systems, system)
	}

	return RoleAddPermissionForm(roleID, systems)
}

func permissionSelect(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")

	permissions, err := ui.GetPermissions(r.Context(), system)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return Error(http.StatusInternalServerError)
	}
	return PermissionSelect(permissions)
}

func scopeInput(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")
	permission := r.FormValue("permission")

	hasScope, err := ui.PermissionHasScope(r.Context(), system, permission)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return Error(http.StatusInternalServerError)
	}
	return ScopeInput(hasScope)
}

