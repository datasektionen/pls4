package members

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/datasektionen/pls4/ui/service"
	"github.com/datasektionen/pls4/ui/views/errors"
	"github.com/google/uuid"
)

func GetRoleMembers(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	toUpdateMember, _ := uuid.Parse(r.FormValue("update-member-id"))
	addNew := r.Form.Has("new")
	includeExpired := r.Form.Has("include-expired")

	members, err := ui.GetRoleMembers(ctx, roleID, includeExpired, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	return Members(roleID, members, toUpdateMember, mayUpdate, addNew, includeExpired)
}

func RoleAddMember(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")

	kthID := r.FormValue("kth-id")

	startDate, err := time.Parse(time.DateOnly, r.FormValue("start-date"))
	if err != nil && r.Form.Has("start-date") {
		return errors.Error(http.StatusBadRequest)
	}
	endDate, err := time.Parse(time.DateOnly, r.FormValue("end-date"))
	if err != nil && r.Form.Has("end-date") {
		return errors.Error(http.StatusBadRequest)
	}

	if err := ui.AddMember(ctx, session.KTHID, roleID, kthID, startDate, endDate); err != nil {
		slog.Error("Could not add member", "error", err, "role_id", roleID, "kth_id", kthID)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func RoleUpdateMember(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	memberID, err := uuid.Parse(r.PathValue("memberID"))
	if err != nil {
		return errors.Error(http.StatusBadRequest, "Invalid syntax for member uuid")
	}

	startDate, err := time.Parse(time.DateOnly, r.FormValue("start-date"))
	if err != nil && r.Form.Has("start-date") {
		return errors.Error(http.StatusBadRequest, "Invalid syntax for start date")
	}
	endDate, err := time.Parse(time.DateOnly, r.FormValue("end-date"))
	if err != nil && r.Form.Has("end-date") {
		return errors.Error(http.StatusBadRequest, "Invalid syntax for start date")
	}

	if err := ui.UpdateMember(ctx, session.KTHID, roleID, memberID, startDate, endDate); err != nil {
		slog.Error("Could not edit member", "error", err, "member", memberID)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func RoleEndMember(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	if err := ui.UpdateMember(ctx, session.KTHID, roleID, member, time.Time{}, time.Now().AddDate(0, 0, -1)); err != nil {
		slog.Error("Could not edit member", "error", err, "role_id", roleID, "member", member)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func RoleRemoveMember(ui *service.UI, ctx context.Context, session service.Session, w http.ResponseWriter, r *http.Request) templ.Component {
	roleID := r.PathValue("id")
	member, _ := uuid.Parse(r.PathValue("memberID"))

	if err := ui.RemoveMember(ctx, session.KTHID, roleID, member); err != nil {
		slog.Error("Could not remove member", "error", err, "member", member)
		return errors.Error(http.StatusInternalServerError)
	}

	return renderMembers(ui, ctx, session, roleID)
}

func renderMembers(ui *service.UI, ctx context.Context, session service.Session, roleID string) templ.Component {
	members, err := ui.GetRoleMembers(ctx, roleID, false, true)
	if err != nil {
		slog.Error("Could not get members", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	mayUpdate, err := ui.MayUpdateRole(ctx, session.KTHID, roleID)
	if err != nil {
		slog.Error("Could not check if role may be updated", "error", err, "role_id", roleID)
		return errors.Error(http.StatusInternalServerError)
	}

	return Members(roleID, members, uuid.Nil, mayUpdate, false, false)
}
