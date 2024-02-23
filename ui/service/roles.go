package service

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
)

func (ui *UI) ListRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
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

func (ui *UI) GetRole(ctx context.Context, id string) (*models.Role, error) {
	rows := ui.db.QueryRowContext(ctx, `--sql
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

func (ui *UI) UpdateRole(ctx context.Context, kthID, roleID, displayName, description string) error {
	if ok, err := ui.MayUpdateRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	res, err := ui.db.ExecContext(ctx, `--sql
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

func (ui *UI) CreateRole(
	ctx context.Context,
	kthID, id, displayName, description, ownerID string,
) error {
	if ok, err := ui.MayCreateRoles(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	if roles, err := ui.GetUserRoles(ctx, kthID); err != nil {
		return err
	} else if !slices.ContainsFunc(roles, func(role models.Role) bool { return role.ID == ownerID }) {
		return errors.New("The user does not have the role " + ownerID + ".")
	}
	tx, err := ui.db.BeginTx(ctx, nil)
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

func (ui *UI) DeleteRole(
	ctx context.Context,
	kthID, roleID string,
) error {
	if ok, err := ui.MayDeleteRole(ctx, kthID, roleID); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}
	tx, err := ui.db.BeginTx(ctx, nil)
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

func (ui *UI) GetAllRoles(ctx context.Context) ([]string, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
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
