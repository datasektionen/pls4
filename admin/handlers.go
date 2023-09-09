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
