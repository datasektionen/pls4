package admin

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
)

func (s *Admin) ListRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		select
			r.id, r.display_name, r.description,
			count(rr.subrole_id), count(ru.id)
		from roles r
		left join roles_roles rr on rr.superrole_id = r.id
		left join roles_users ru on ru.role_id = r.id
		where (ru.id is null or now() between ru.start_date and ru.end_date)
		group by r.id
		order by r.display_name
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

func (s *Admin) GetRole(ctx context.Context, id string) (*models.Role, error) {
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

func (s *Admin) GetSubroles(ctx context.Context, id string) ([]models.Role, error) {
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

func (s *Admin) GetRoleMembers(ctx context.Context, id string, onlyCurrent bool, includeIndirect bool) ([]models.Member, error) {
	query := `select
		id, kth_id, modified_by,
		modified_at, start_date, end_date
	from roles_users
	where role_id = $1
	and ($2 or now() between start_date and end_date)
	order by kth_id`
	if includeIndirect {
		query = `with recursive all_subroles (role_id) as (
				select subrole_id
				from roles_roles
				where superrole_id = $1
				union
				select subrole_id from all_subroles
				inner join roles_roles
				on superrole_id = role_id
		) (` + query + `) union all
		(select
			null as id, kth_id, '' as modified_by,
			max(modified_at), min(start_date), max(end_date)
		from all_subroles
		inner join roles_users using (role_id)
		where ($2 or now() between start_date and end_date)
		group by kth_id
		order by kth_id)`
	}

	rows, err := s.db.QueryContext(ctx, query, id, !onlyCurrent)
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

func (s *Admin) GetRolePermissions(ctx context.Context, id, kthID string) ([]models.SystemPermissions, error) {
	rows, err := s.db.QueryContext(ctx, `
		select system, permission
		from roles_permissions
		where role_id = $1
		order by system
	`, id)
	if err != nil {
		return nil, err
	}
	perms := make([]models.SystemPermissions, 1)
	p := &perms[0]
	for rows.Next() {
		var system, permission string
		if err := rows.Scan(&system, &permission); err != nil {
			return nil, err
		}
		if p.System != system {
			mayEdit, err := s.MayUpdatePermissions(ctx, kthID, system)
			if err != nil {
				return nil, err
			}
			perms = append(perms, models.SystemPermissions{
				System:      system,
				Permissions: []string{},
				MayEdit:     mayEdit,
			})
			p = &perms[len(perms)-1]
		}
		p.Permissions = append(p.Permissions, permission)
	}
	return perms[1:], nil
}

func (s *Admin) UpdateRole(ctx context.Context, kthID, roleID, displayName, description string) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
		update roles
		set
			display_name = coalesce(nullif($2, ''), display_name),
			description = coalesce(nullif($3, ''), description)
		where id = $1
	`, roleID, displayName, description)
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

func (s *Admin) AddSubrole(ctx context.Context, kthID, roleID, subroleID string) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		insert into roles_roles (superrole_id, subrole_id)
		values ($1, $2)
	`, roleID, subroleID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Admin) RemoveSubrole(ctx context.Context, kthID, roleID, subroleID string) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
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

func (s *Admin) UpdateMember(
	ctx context.Context,
	kthID, roleID string,
	memberID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
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

func (s *Admin) AddMember(
	ctx context.Context,
	kthID, roleID, memberKTHID string,
	startDate time.Time,
	endDate time.Time,
) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
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

func (s *Admin) RemoveMember(
	ctx context.Context,
	kthID, roleID string,
	memberID uuid.UUID,
) error {
	if ok, err := s.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := s.db.ExecContext(ctx, `
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

func (s *Admin) CreateRole(
	ctx context.Context,
	kthID, id, displayName, description string,
) error {
	if ok, err := s.MayCreateRoles(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		insert into roles (id, display_name, description)
		values ($1, $2, $3)
	`, id, displayName, description)
	return err
}

var tableRexeg = regexp.MustCompile(`on table "roles_(.*)"$`)

func (s *Admin) DeleteRole(
	ctx context.Context,
	kthID, id string,
) error {
	if ok, err := s.MayDeleteRoles(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		delete from roles_roles
		where subrole_id = $1
	`, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(`
		delete from roles
		where id = $1
	`, id)
	if err != nil && strings.HasPrefix(err.Error(), "pq: update or delete on table") {
		match := tableRexeg.FindStringSubmatch(err.Error())
		if len(match) > 1 {
			table := match[1]
			_ = tx.Rollback()
			return errors.New("This role still has " + table + " connected. They must be removed first.")
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Admin) RemovePermission(
	ctx context.Context,
	kthID, roleID string,
	system, permission string,
) error {
	if ok, err := s.MayUpdatePermissions(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}

	_, err := s.db.ExecContext(ctx, `
		delete from roles_permissions
		where role_id = $1
		and system = $2
		and permission = $3
	`, roleID, system, permission)
	return err
}
