package service

import (
	"context"
	"time"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
)

func (ui *UI) GetRoleMembers(ctx context.Context, id string, includeExpired bool, includeIndirect bool) ([]models.Member, error) {
	query := `--sql
		select
			id, kth_id, modified_by,
			modified_at, start_date, end_date
		from roles_users
		where role_id = $1
		and ($2 or now() between start_date and end_date)
		order by kth_id
	`
	if includeIndirect {
		query = `/*sql*/ with recursive all_subroles (role_id) as (
				select subrole_id
				from roles_roles
				where superrole_id = $1
				union
				select subrole_id from all_subroles
				inner join roles_roles
				on superrole_id = role_id
		) (` + query + `/*sql*/) union all
		(select
			null as id, kth_id, '' as modified_by,
			max(modified_at), min(start_date), max(end_date)
		from all_subroles
		inner join roles_users using (role_id)
		where ($2 or now() between start_date and end_date)
		group by kth_id
		order by kth_id)`
	}

	rows, err := ui.db.QueryContext(ctx, query, id, includeExpired)
	if err != nil {
		return nil, err
	}
	var members []models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(
			&m.MemberID, &m.KTHID, &m.ModifiedBy,
			&m.ModifiedAt, &m.StartDate, &m.EndDate,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (ui *UI) UpdateMember(
	ctx context.Context,
	kthID, roleID string,
	memberID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := ui.db.ExecContext(ctx, `--sql
		update roles_users
		set
			start_date = case when $3 then $4 else start_date end,
			end_date = case when $5 then $6 else end_date end
		where id = $1 and role_id = $2
	`, memberID, roleID, startDate != time.Time{}, startDate, endDate != time.Time{}, endDate)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		// TODO: invalid id
	}
	return nil
}

func (ui *UI) AddMember(
	ctx context.Context,
	kthID, roleID, memberKTHID string,
	startDate time.Time,
	endDate time.Time,
) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := ui.db.ExecContext(ctx, `--sql
		insert into roles_users (role_id, kth_id, modified_by, start_date, end_date)
		values ($1, $2, $3, $4, $5)
	`, roleID, memberKTHID, kthID, startDate, endDate)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		// TODO: invalid id
	}
	return nil
}

func (ui *UI) RemoveMember(
	ctx context.Context,
	kthID, roleID string,
	memberID uuid.UUID,
) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := ui.db.ExecContext(ctx, `--sql
		delete from roles_users
		where role_id = $1 and id = $2
	`, roleID, memberID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		// TODO: invalid id
	}
	return nil
}

