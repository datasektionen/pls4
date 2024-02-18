package admin

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
)

func (s *Admin) ListRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
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
	rows := s.db.QueryRowContext(ctx, `--sql
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
	rows, err := s.db.QueryContext(ctx, `--sql
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

func (s *Admin) GetRolePermissions(ctx context.Context, id string) ([]models.SystemPermissionInstances, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
		select id, system_id, permission_id, coalesce(scope, '')
		from roles_permissions
		inner join permission_instances
			on id = permission_instance_id
		where role_id = $1
		order by system_id
	`, id)
	if err != nil {
		return nil, err
	}
	perms := make([]models.SystemPermissionInstances, 1)
	p := &perms[0]
	for rows.Next() {
		var id uuid.UUID
		var system, permission, scope string
		if err := rows.Scan(&id, &system, &permission, &scope); err != nil {
			return nil, err
		}
		if p.System != system {
			perms = append(perms, models.SystemPermissionInstances{
				System:      system,
				Permissions: []models.PermissionInstance{},
			})
			p = &perms[len(perms)-1]
		}
		p.Permissions = append(p.Permissions, models.PermissionInstance{
			ID:           id,
			PermissionID: permission,
			Scope:        scope,
		})
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
	res, err := s.db.ExecContext(ctx, `--sql
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
	_, err := s.db.ExecContext(ctx, `--sql
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
	res, err := s.db.ExecContext(ctx, `--sql
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
	res, err := s.db.ExecContext(ctx, `--sql
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
	res, err := s.db.ExecContext(ctx, `--sql
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
	res, err := s.db.ExecContext(ctx, `--sql
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
	kthID, id, displayName, description, ownerID string,
) error {
	if ok, err := s.MayCreateRoles(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	if roles, err := s.GetUserRoles(ctx, kthID); err != nil {
		return err
	} else if !slices.ContainsFunc(roles, func(role models.Role) bool { return role.ID == ownerID }) {
		return errors.New("The user does not have the role " + ownerID + ".")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`--sql
		insert into roles (id, display_name, description)
		values ($1, $2, $3)
	`, id, displayName, description); err != nil {
		return err
	}
	var instanceID uuid.UUID
	if err := tx.QueryRow(`--sql
		insert into permission_instances (system_id, permission_id, scope)
		values ('pls', 'role', $1)
		returning id
	`, id).Scan(&instanceID); err != nil {
		return err
	}
	if _, err := tx.Exec(`--sql
		insert into roles_permissions (permission_instance_id, role_id)
		values ($1, $2)
	`, instanceID, ownerID); err != nil {
		return err
	}
	return tx.Commit()
}

var tableRexeg = regexp.MustCompile(`on table "roles_(.*)"$`)

func (s *Admin) DeleteRole(
	ctx context.Context,
	kthID, roleID string,
) error {
	if ok, err := s.MayDeleteRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`--sql
		delete from roles_roles
		where subrole_id = $1
	`, roleID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`--sql
		delete from roles
		where id = $1
	`, roleID)
	if err != nil && strings.HasPrefix(err.Error(), "pq: update or delete on table") {
		match := tableRexeg.FindStringSubmatch(err.Error())
		if len(match) > 1 {
			table := match[1]
			return errors.New("This role still has " + table + " connected. They must be removed first.")
		}
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(`--sql
		delete from permission_instances
		where system_id = 'pls'
		and permission_id = 'role'
		and scope = $1
	`, roleID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Admin) RemovePermission(
	ctx context.Context,
	kthID string,
	permissionInstanceID uuid.UUID,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}
	var system string
	if row := tx.QueryRow(`--sql
		select system_id
		from permission_instances
		where id = $1
	`, permissionInstanceID); row.Err() != nil {
		return row.Err()
	} else if err := row.Scan(&system); err != nil {
		return err
	}
	if ok, err := s.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}

	slog.InfoContext(ctx, "Removing permission", "id", permissionInstanceID, "system", system)

	// TODO: this must cascade to delete from either roles_permissions or
	// api_tokens_permissions
	_, err = tx.Exec(`--sql
		delete from permission_instances
		where id = $1
	`, permissionInstanceID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Admin) AddPermissionToRole(
	ctx context.Context,
	kthID, roleID string,
	system, permission, scope string,
) error {
	if ok, err := s.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}

	id, err := createPermissionInstance(tx, system, permission, scope)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`--sql
		insert into roles_permissions (permission_instance_id, role_id)
		values ($1, $2)
	`, id, roleID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func createPermissionInstance(
	tx *sql.Tx,
	system, permission, scope string,
) (uuid.UUID, error) {
	var hasScope bool
	if err := tx.QueryRow(`--sql
		select has_scope from permissions
		where system_id = $1
		and id = $2
	`, system, permission).Scan(&hasScope); err != nil {
		return uuid.Nil, err
	}
	if hasScope != (scope != "") {
		return uuid.Nil, errors.New("Provided scope when there should be one or the other way around")
	}
	var id uuid.UUID
	if err := tx.QueryRow(`--sql
		insert into permission_instances (system_id, permission_id, scope)
		values ($1, $2, nullif($3, ''))
		returning id
	`, system, permission, scope).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Admin) GetUserRoles(
	ctx context.Context,
	kthID string,
) ([]models.Role, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
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

func (s *Admin) GetAllSystems(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
		select id
		from systems
	`)
	if err != nil {
		return nil, err
	}
	var systems []string
	for rows.Next() {
		systems = append(systems, "")
		rows.Scan(&systems[len(systems)-1])
	}
	return systems, nil
}

func (s *Admin) GetAllRoles(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
		select id
		from roles
	`)
	if err != nil {
		return nil, err
	}
	var roles []string
	for rows.Next() {
		roles = append(roles, "")
		rows.Scan(&roles[len(roles)-1])
	}
	return roles, nil
}

func (s *Admin) GetPermissions(ctx context.Context, system string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `--sql
		select id
		from permissions
		where system_id = $1
	`, system)
	if err != nil {
		return nil, err
	}
	var permissions []string
	for rows.Next() {
		permissions = append(permissions, "")
		rows.Scan(&permissions[len(permissions)-1])
	}
	return permissions, nil
}

func (s *Admin) PermissionHasScope(ctx context.Context, system, permission string) (bool, error) {
	var hasScope bool
	err := s.db.QueryRowContext(ctx, `--sql
		select has_scope
		from permissions
		where system_id = $1
		and id = $2
	`, system, permission).Scan(&hasScope)
	return hasScope, err
}
