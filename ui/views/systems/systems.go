package systems

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
)

func ListSystems(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	systems, err := ui.GetAllSystems(ctx)
	if err != nil {
		slog.Error("Could not get systems", "error", err)
		return errors.Error(http.StatusInternalServerError)
	}
	return listSystems(systems)
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
	return system(systemID, permissions, mayUpdate)
}
