package service

import "context"

func (ui *UI) GetAllSystems(ctx context.Context) ([]string, error) {
	rows, err := ui.db.QueryContext(ctx, `--sql
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

func (ui *UI) CreateSystem(ctx context.Context, id, kthID string) error {
	if ok, err := ui.MayCreateSystems(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	if _, err := ui.db.ExecContext(ctx, `--sql
		insert into systems (id)
		values ($1)
	`, id); err != nil {
		return err
	}
	return nil
}

func (ui *UI) DeleteSystem(ctx context.Context, id, kthID string) error {
	if ok, err := ui.MayDeleteSystems(ctx, kthID); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	if _, err := ui.db.ExecContext(ctx, `--sql
		delete from systems
		where id = $1
	`, id); err != nil {
		return err
	}
	return nil
}

func (ui *UI) CreatePermission(ctx context.Context, system, permission string, hasScope bool, kthID string) error {
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	if _, err := ui.db.ExecContext(ctx, `--sql
		insert into permissions (system_id, id, has_scope)
		values ($1, $2, $3)
	`, system, permission, hasScope); err != nil {
		return err
	}
	return nil
}

func (ui *UI) DeletePermission(ctx context.Context, system, permission, kthID string) error {
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	if _, err := ui.db.ExecContext(ctx, `--sql
		delete from permissions
		where system_id = $1
		and id = $2
	`, system, permission); err != nil {
		return err
	}
	return nil
}

func (ui *UI) AddScopeToPermission(ctx context.Context, system, permission, defaultScope, kthID string) error {
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	tx, err := ui.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`--sql
		update permissions
		set has_scope = true
		where system_id = $1
		and id = $2
	`, system, permission); err != nil {
		return err
	}
	if _, err := tx.Exec(`--sql
		update permission_instances
		set scope = $3
		where system_id = $1
		and permission_id = $2
	`, system, permission, defaultScope); err != nil {
		return err
	}
	return tx.Commit()
}

func (ui *UI) RemoveScopeFromPermission(ctx context.Context, system, permission, kthID string) error {
	if ok, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system); err != nil {
		return err
	} else if !ok {
		// TODO: return forbidden or something
		return nil
	}
	tx, err := ui.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`--sql
		update permissions
		set has_scope = false
		where system_id = $1
		and id = $2
	`, system, permission); err != nil {
		return err
	}
	if _, err := tx.Exec(`--sql
		update permission_instances
		set scope = null
		where system_id = $1
		and permission_id = $2
	`, system, permission); err != nil {
		return err
	}
	return tx.Commit()
}
