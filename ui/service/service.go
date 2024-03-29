package service

import (
	"context"
	"database/sql"

	"github.com/datasektionen/pls4/api"
	"github.com/datasektionen/pls4/models"
)

type UI struct {
	db               *sql.DB
	api              *api.API
	loginFrontendURL string
	loginAPIURL      string
	loginAPIKey      string
}

func New(ctx context.Context, db *sql.DB, api *api.API, loginFrontendURL, loginAPIURL, loginAPIKey string) *UI {
	s := &UI{}

	s.api = api
	s.loginFrontendURL = loginFrontendURL
	s.loginAPIURL = loginAPIURL
	s.loginAPIKey = loginAPIKey
	s.db = db

	go s.deleteOldSessionsForever(ctx)

	return s
}

func (ui *UI) LoginFrontendURL() string {
	return ui.loginFrontendURL
}

func (ui *UI) GetUserRoles(
	ctx context.Context,
	kthID string,
) ([]models.Role, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
		with recursive all_roles (role_id) as (
			select role_id from roles_users
			where kth_id = $1 and now() between start_date and end_date
			union
			select superrole_id from all_roles
			inner join roles_roles
				on subrole_id = role_id
		)
		select r.id, r.display_name
		from all_roles a
		inner join roles r on a.role_id = r.id
	`, kthID)
	if err != nil {
		return nil, err
	}
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID, &role.DisplayName,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
