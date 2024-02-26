package service

import (
	"context"
	"slices"
)

func (ui *UI) MayUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	scopes, err := ui.api.UserGetScopes(ctx, kthID, "pls", "role")
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(scopes, func(r string) bool { return r == "*" || r == roleID }), nil
}

func (ui *UI) MayCreateRoles(ctx context.Context, kthID string) (bool, error) {
	return ui.api.UserCheckPermission(ctx, kthID, "pls", "create-role")
}

func (ui *UI) MayDeleteRole(ctx context.Context, kthID, roleID string) (bool, error) {
	mayCreate, err := ui.MayCreateRoles(ctx, kthID)
	if err != nil {
		return false, err
	}
	mayUpdate, err := ui.MayUpdateRole(ctx, kthID, roleID)
	if err != nil {
		return false, err
	}
	return mayCreate && mayUpdate, nil
}

func (ui *UI) MayDeleteRoles(ctx context.Context, kthID string) (map[string]struct{}, error) {
	mayCreate, err := ui.MayCreateRoles(ctx, kthID)
	if err != nil {
		return nil, err
	}
	if !mayCreate {
		return make(map[string]struct{}), nil
	}
	deletable := make(map[string]struct{})
	roles, err := ui.api.UserGetScopes(ctx, kthID, "pls", "role")
	if err != nil {
		return nil, err
	}
	if slices.Contains(roles, "*") {
		roles, err = ui.GetAllRoles(ctx)
		if err != nil {
			return nil, err
		}
	}
	for _, role := range roles {
		deletable[role] = struct{}{}
	}
	return deletable, nil
}

func (ui *UI) MayUpdatePermissionsInSystem(ctx context.Context, kthID, system string) (bool, error) {
	systems, err := ui.api.UserGetScopes(ctx, kthID, "pls", "system")
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(systems, func(s string) bool { return s == "*" || s == system }), nil
}

func (ui *UI) MayUpdatePermissionsInSystems(ctx context.Context, kthID string, systems ...[]string) (map[string]struct{}, error) {
	var sys []string
	switch len(systems) {
	case 0:
		var err error
		sys, err = ui.GetAllSystems(ctx)
		if err != nil {
			return nil, err
		}
	case 1:
		sys = systems[0]
	default:
		panic("`systems` is an optional parameter, but more than one was provided.")
	}

	granted := make(map[string]struct{})
	for _, system := range sys {
		mayUpdate, err := ui.MayUpdatePermissionsInSystem(ctx, kthID, system)
		if err != nil {
			return nil, err
		}
		if mayUpdate {
			granted[system] = struct{}{}
		}
	}
	return granted, nil
}

func (ui *UI) MayAddPermissions(ctx context.Context, kthID string) (bool, error) {
	return ui.api.UserCheckPermission(ctx, kthID, "pls", "system")
}

func (ui *UI) MayCreateSystems(ctx context.Context, kthID string) (bool, error) {
	return ui.api.UserCheckPermission(ctx, kthID, "pls", "manage-systems")
}

func (ui *UI) MayDeleteSystems(ctx context.Context, kthID string) (bool, error) {
	return ui.api.UserCheckPermission(ctx, kthID, "pls", "manage-systems")
}
