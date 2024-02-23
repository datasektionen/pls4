package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
)

func roleSubroleForm(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	options, err := ui.ListRoles(ctx)
	if err != nil {
		slog.Error("Could not list roles", "error", err)
		return Error(http.StatusInternalServerError)
	}
	return AddSubroleForm(roleID, options)
}

func roleAddSubrole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subrole := r.FormValue("subrole")

	if err := ui.AddSubrole(ctx, session.KTHID, roleID, subrole); err != nil {
		slog.Error("Could not add subrole", "error", err, "role_id", roleID, "subrole_id", subrole)
		return Error(http.StatusInternalServerError)
	}

	return renderSubroles(ui, ctx, session, roleID)
}

func roleRemoveSubrole(ui *service.UI, ctx context.Context, w http.ResponseWriter, r *http.Request) templ.Component {
	session, err := ui.GetSession(r)
	if err != nil {
		slog.Error("Could not get current session", "error", err)
		return Error(http.StatusInternalServerError)
	}

	roleID := r.PathValue("id")
	subroleID := r.PathValue("subroleID")

	if err := ui.RemoveSubrole(ctx, session.KTHID, roleID, subroleID); err != nil {
		slog.Error("Could not remove subrole", "error", err, "role_id", roleID, "subrole_id", subroleID)
		return Error(http.StatusInternalServerError)
	}

	return renderSubroles(ui, ctx, session, roleID)
}

func renderSubroles(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	subroles, err := ui.GetSubroles(ctx, roleID)
	if err != nil {
		slog.Error("Could not get subroles", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return Error(http.StatusInternalServerError)
	}
	return Subroles(roleID, subroles, mayUpdate)
}

