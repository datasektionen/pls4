package api

import (
	"context"
	"github.com/google/uuid"
)

func CheckUser(ctx context.Context, kthID, system, permission string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `
		with recursive all_groups (group_id) as (
			select group_id from groups_users
			where kth_id = $1 and now() between start_date and end_date
			union all
			select supergroup_id from all_groups
			inner join groups_groups
			on subgroup_id = group_id
		),
		found as (
			select from permissions p
			inner join groups_permissions gp
			on gp.permission_id = p.id
			inner join all_groups ag
			on ag.group_id = gp.group_id
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

func ListForUser(ctx context.Context, kthID, system string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		with recursive all_groups (group_id) as (
			select group_id from groups_users
			where kth_id = $1 and now() between start_date and end_date
			union all
			select supergroup_id from all_groups
			inner join groups_groups
			on subgroup_id = group_id
		)
		select name from permissions p
		inner join groups_permissions gp
		on gp.permission_id = p.id
		inner join all_groups ag
		on ag.group_id = gp.group_id
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

func CheckToken(ctx context.Context, secret uuid.UUID, system, permission string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `
		select exists(
			select 1 from api_tokens t
			inner join api_tokens_permissions tp
			on tp.api_token_id = t.id
			inner join permissions p
			on p.id = tp.permission_id
			where t.secret = $1
			and system = $2
			and name = $3
		)
	`, secret, system, permission)
	var found bool
	if err := row.Scan(&found); err != nil {
		return false, err
	}
	return found, nil
}

