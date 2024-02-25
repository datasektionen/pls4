package permissions

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
	"github.com/google/uuid"
)

func RoleAddPermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	system := r.FormValue("system")
	permission := r.FormValue("permission")
	scope := r.FormValue("scope")

	if err := ui.AddPermissionToRole(ctx, session.KTHID, roleID, system, permission, scope); err != nil {
		slog.Error("Could add permission to role", "error", err, "role_id", roleID, "system", system, "permission", permission)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func RoleRemovePermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	instanceID, err := uuid.Parse(r.PathValue("instanceID"))
	if err != nil {
		return errors.Error(http.StatusBadRequest, "Invalid uuid syntax")
	}
	// NOTE: we don't check that these match which is only fine as long as we
	// don't rely on that for authorization

	if err := ui.RemovePermission(ctx, session.KTHID, instanceID); err != nil {
		slog.Error("Could remove permission from role", "error", err, "role_id", roleID, "instance_id", instanceID)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderPermissions(ui, ctx, session, roleID)
}

func AddPermissionForm(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	systemSet, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not systems in which user may update permissions", "error", err)
	}
	var systems []string
	for system := range systemSet {
		systems = append(systems, system)
	}

	return roleAddPermissionForm(roleID, systems)
}

func PermissionSelect(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")

	permissions, err := ui.GetPermissions(r.Context(), system)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return errors.Error(http.StatusInternalServerError)
	}
	permissionIDs := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionIDs[i] = perm.ID
	}
	return permissionSelect(permissionIDs)
}

func ScopeInput(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	system := r.FormValue("system")
	permission := r.FormValue("permission")

	hasScope, err := ui.PermissionHasScope(r.Context(), system, permission)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", system)
		return errors.Error(http.StatusInternalServerError)
	}
	return scopeInput(hasScope)
}

func renderPermissions(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	perms, err := ui.GetRolePermissions(ctx, roleID)
	if err != nil {
		slog.Error("Could not get role permissions", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	mayAddPermissions, err := ui.MayAddPermissions(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	var systems []string
	for _, perm := range perms {
		systems = append(systems, perm.System)
	}
	mayDeleteInSystems, err := ui.MayUpdatePermissionsInSystems(ctx, session.KTHID, systems)
	if err != nil {
		slog.Error("Could not filter systems for permissions", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}

	return Permissions(roleID, perms, mayAddPermissions, mayDeleteInSystems)
}
