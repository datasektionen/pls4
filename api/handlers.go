package api

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
)

func (s *API) CheckUser(ctx context.Context, kthID, system, permission string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `
		with recursive all_roles (role_id) as (
			select role_id from roles_users
			where kth_id = $1 and now() between start_date and end_date
			union all
			select superrole_id from all_roles
			inner join roles_roles
				on subrole_id = role_id
		),
		found as (
			select from permissions p
			inner join roles_permissions gp
				on gp.permission_id = p.id
			inner join all_roles ag
				on ag.role_id = gp.role_id
			where system = $2
			and name = $3
		)
		select exists(select 1 from found)
	`, kthID, system, permission)
	var found bool
	if err := row.Scan(&found); err != nil {
		return false, err
	}
	return found, nil
}

func (s *API) ListForUser(ctx context.Context, kthID, system string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		with recursive all_roles (role_id) as (
			select role_id from roles_users
			where kth_id = $1 and now() between start_date and end_date
			union all
			select superrole_id from all_roles
			inner join roles_roles
				on subrole_id = role_id
		)
		select name from permissions p
		inner join roles_permissions gp
			on gp.permission_id = p.id
		inner join all_roles ag
			on ag.role_id = gp.role_id
		where system = $2
	`, kthID, system)
	if err != nil {
		return nil, err
	}
	var perms []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
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
		inner join permissions p
			on p.id = tp.permission_id
		where t.secret = $1
		and system = $2
		and name = $3
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

