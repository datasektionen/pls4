package service

import (
	"context"

	"github.com/datasektionen/pls4/models"
)

func (ui *UI) GetSubroles(ctx context.Context, id string) ([]models.Role, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
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

func (ui *UI) AddSubrole(ctx context.Context, kthID, roleID, subroleID string) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	_, err := ui.db.ExecContext(ctx, `--sql
		insert into roles_roles (superrole_id, subrole_id)
		values ($1, $2)
	`, roleID, subroleID)
	if err != nil {
		return err
	}
	return nil
}

func (ui *UI) RemoveSubrole(ctx context.Context, kthID, roleID, subroleID string) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := ui.db.ExecContext(ctx, `--sql
		delete from roles_roles
		where superrole_id = $1 and subrole_id = $2
	`, roleID, subroleID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		// TODO: invalid roleID didn't have subroleID as a subrole
	}
	return nil
}

