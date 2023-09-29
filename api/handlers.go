package api

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

func (s *API) CheckUser(ctx context.Context, kthID, system string, permission string) (bool, error) {
	rawPermissions, err := s.GetUserRaw(ctx, kthID, system)
	if err != nil {
		return false, err
	}
	return hasPermission(rawPermissions, permission), nil
}

func (s *API) FilterForUser(ctx context.Context, kthID, system string, queries []string) ([]string, error) {
	rawPermissions, err := s.GetUserRaw(ctx, kthID, system)
	if err != nil {
		return []string{}, err
	}
	// A note for the future: this can probably be optimized for large cases. I don't however think
	// that we will have very large cases, so I can't be bothered.
	granted := make([]string, 0)
	for _, query := range queries {
		if hasPermission(rawPermissions, query) {
			granted = append(granted, query)
		}
	}
	return granted, nil
}

func hasPermission(rawPermissions []string, permission string) bool {
	for _, rawPerm := range rawPermissions {
		if rawPerm == permission {
			return true
		}
		if rawPerm[len(rawPerm)-1] == '*' && strings.HasPrefix(permission, rawPerm[0:len(rawPerm)-1]) {
			return true
		}
	}
	return false
}

func (s *API) CheckToken(ctx context.Context, secret uuid.UUID, system, permission string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	row := tx.QueryRowContext(ctx, `
		select t.id from api_tokens t
		inner join api_tokens_permissions tp
			on tp.api_token_id = t.id
		where t.secret = $1
		and tp.system = $2
		and $3 like replace(tp.permission, '*', '%')
	`, secret, system, permission)
	var id uuid.UUID
	if err := row.Scan(&id); err == sql.ErrNoRows {
		return false, tx.Commit()
	} else if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `
		update api_tokens
		set last_used_at = now()
		where id = $1
	`, id); err != nil {
		slog.ErrorContext(ctx, "Could not update last_used_at", "id", id)
	}
	return true, tx.Commit()
}

func (s *API) GetUserRaw(ctx context.Context, kthID, system string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		with recursive all_roles (role_id) as (
			select role_id from roles_users
			where kth_id = $1 and now() between start_date and end_date
			union
			select superrole_id from all_roles
			inner join roles_roles
				on subrole_id = role_id
		)
		select permission from roles_permissions p
		inner join all_roles a
			on a.role_id = p.role_id
		where system = $2
	`, kthID, system)
	if err != nil {
		return nil, err
	}
	perms := make([]string, 0)
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}
