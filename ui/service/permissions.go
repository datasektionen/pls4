package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/datasektionen/pls4/models"
	"github.com/google/uuid"
)

func (ui *UI) GetRolePermissions(ctx context.Context, id string) ([]models.SystemPermissionInstances, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
		select id, system_id, permission_id, coalesce(scope, '')
		from roles_permissions
		inner join permission_instances
			on id = permission_instance_id
		where role_id = $1
		order by system_id, permission_id
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

func (ui *UI) RemovePermission(
	ctx context.Context,
	kthID string,
	permissionInstanceID uuid.UUID,
) error {
	tx, err := ui.db.BeginTx(ctx, nil)
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
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
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

func (ui *UI) AddPermissionToRole(
	ctx context.Context,
	kthID, roleID string,
	system, permission, scope string,
) error {
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return an error
		return nil
	}

	tx, err := ui.db.BeginTx(ctx, nil)
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

func (ui *UI) GetPermissions(ctx context.Context, system string) ([]models.Permission, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
		select id, has_scope
		from permissions
		where system_id = $1
	`, system)
	if err != nil {
		return nil, err
	}
	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		rows.Scan(&perm.ID, &perm.HasScope)
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func (ui *UI) PermissionHasScope(ctx context.Context, system, permission string) (bool, error) {
	var hasScope bool
	err := ui.db.QueryRowContext(ctx, `--sql
		select has_scope
		from permissions
		where system_id = $1
		and id = $2
	`, system, permission).Scan(&hasScope)
	return hasScope, err
}
