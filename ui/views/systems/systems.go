package systems

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/models"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
)

func ListSystems(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systems, err := ui.GetAllSystems(ctx)
	if err != nil {
		slog.Error("Could not get systems", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	mayDelete, err := ui.MayDeleteSystems(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not see which systems user may delete", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return listSystems(systems, mayDelete)
}

func GetSystem(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	permissions, err := ui.GetPermissions(ctx, systemID)
	if err != nil {
		slog.Error("Could not get permissions for system", "error", err, "system", systemID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdatePermissionsInSystem(ctx, session.KTHID, systemID)
	if err != nil {
		slog.Error("Could not get permissions", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return permissionsForSystem(systemID, permissions, mayUpdate)
}

func CreateSystem(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.FormValue("system-id")
	if err := ui.CreateSystem(ctx, systemID, session.KTHID); err != nil {
		slog.Error("Could not create system", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayDelete, err := ui.MayDeleteSystems(ctx, session.KTHID)
	if err != nil {
		slog.Error("Could not see which systems user may delete", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return system(systemID, mayDelete)
}

func DeleteSystem(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	if err := ui.DeleteSystem(ctx, systemID, session.KTHID); err != nil {
		slog.Error("Could not delete system", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return nil
}

func CreatePermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	permissionID := r.FormValue("permission-id")
	hasScope := r.Form.Has("has-scope")
	if err := ui.CreatePermission(ctx, systemID, permissionID, hasScope, session.KTHID); err != nil {
		slog.Error("Could not delete system", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdatePermissionsInSystem(ctx, session.KTHID, systemID)
	if err != nil {
		slog.Error("Could not get permissions", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return permission(systemID, models.Permission{ID: permissionID, HasScope: hasScope}, mayUpdate)
}

func DeletePermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	permissionID := r.PathValue("permissionID")
	if err := ui.DeletePermission(ctx, systemID, permissionID, session.KTHID); err != nil {
		slog.Error("Could not delete permission", "error", err, "system", systemID, "permission", permissionID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return nil
}

func AddScopeToPermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	permissionID := r.PathValue("permissionID")
	defaultScope := r.Header.Get("hx-prompt")
	if err := ui.AddScopeToPermission(ctx, systemID, permissionID, defaultScope, session.KTHID); err != nil {
		slog.Error("Could not add scope to permission", "error", err, "system", systemID, "permission", permissionID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdatePermissionsInSystem(ctx, session.KTHID, systemID)
	if err != nil {
		slog.Error("Could not get permissions", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return permission(systemID, models.Permission{ID: permissionID, HasScope: true}, mayUpdate)
}

func RemoveScopeFromPermission(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systemID := r.PathValue("id")
	permissionID := r.PathValue("permissionID")
	if err := ui.RemoveScopeFromPermission(ctx, systemID, permissionID, session.KTHID); err != nil {
		slog.Error("Could not remove scope from permission", "error", err, "system", systemID, "permission", permissionID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdatePermissionsInSystem(ctx, session.KTHID, systemID)
	if err != nil {
		slog.Error("Could not get permissions", "error", err, "system", systemID, "kth_id", session.KTHID)
		return errors.Error(http.StatusInternalServerError)
	}
	return permission(systemID, models.Permission{ID: permissionID, HasScope: false}, mayUpdate)
}
