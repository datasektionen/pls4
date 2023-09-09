package admin

import (
	"context"

	"github.com/datasektionen/pls4/models"
)

func (s *service) ListRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		select
			r.id, r.display_name, r.description,
			count(rr.subrole_id), count(ru.id)
		from roles r
		left join roles_roles rr on rr.superrole_id = r.id
		left join roles_users ru on ru.role_id = r.id
		group by r.id
	`)
	if err != nil {
		return nil, err
	}
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID, &role.DisplayName, &role.Description,
			&role.SubroleCount, &role.MemberCount,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *service) GetRole(ctx context.Context, id string) (*models.Role, error) {
	rows := s.db.QueryRowContext(ctx, `
		select
			r.id, r.display_name, r.description,
			count(rr.subrole_id), count(ru.id)
		from roles r
		left join roles_roles rr on rr.superrole_id = r.id
		left join roles_users ru on ru.role_id = r.id
		where r.id = $1
		group by r.id
	`, id)
	var r models.Role
	err := rows.Scan(&r.ID, &r.DisplayName, &r.Description, &r.SubroleCount, &r.MemberCount)
	if r.ID == "" {
		return nil, nil
	}
	return &r, err
}

func (s *service) GetSubroles(ctx context.Context, id string) ([]models.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		select
			r.id, r.display_name, r.description,
			count(sub.subrole_id), count(ru.id)
		from roles_roles rr
		inner join roles r on r.id = rr.subrole_id
		left join roles_roles sub on sub.superrole_id = r.id
		left join roles_users ru on ru.role_id = r.id
		where rr.superrole_id = $1
		group by r.id
	`, id)
	if err != nil {
		return nil, err
	}
	var roles []models.Role
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(
			&r.ID, &r.DisplayName, &r.Description,
			&r.SubroleCount, &r.MemberCount,
		); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, nil
}

func (s *service) GetRoleMembers(ctx context.Context, id string, onlyCurrent bool, includeIndirect bool) ([]models.Member, error) {
	query := `
		select
			kth_id, comment, modified_by,
			modified_at, start_date, end_date
		from roles_users
		where role_id = $1
		and ($2 or now() between start_date and end_date)
	`
	if includeIndirect {
		query = `
			with recursive all_subroles (role_id, indirect) as (
				select $1, false as role_id
				union all
				select subrole_id, true from all_subroles
				inner join roles_roles
				on superrole_id = role_id
			)
			select
				kth_id, comment, modified_by,
				modified_at, start_date, end_date,
				indirect
			from all_subroles
			inner join roles_users using (role_id)
			where ($2 or now() between start_date and end_date)
		`
	}

	rows, err := s.db.QueryContext(ctx, query, id, !onlyCurrent)
	if err != nil {
		return nil, err
	}
	var members []models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(
			&m.KTHID, &m.Comment, &m.ModifiedBy,
			&m.ModifiedAt, &m.StartDate, &m.EndDate,
			&m.Indirect,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
