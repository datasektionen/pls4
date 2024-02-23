package ui

import (
	"context"
	"slices"
)

func (s *UI) MayUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	scopes, err := s.api.UserGetScopes(ctx, kthID, "pls", "role")
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(scopes, func(r string) bool { return r == "*" || r == roleID }), nil
}

func (s *UI) MayCreateRoles(ctx context.Context, kthID string) (bool, error) {
	return s.api.UserCheckPermission(ctx, kthID, "pls", "create-role")
}

func (s *UI) MayDeleteRole(ctx context.Context, kthID, roleID string) (bool, error) {
	mayCreate, err := s.MayCreateRoles(ctx, kthID)
	if err != nil {
		return false, err
	}
	mayUpdate, err := s.MayUpdateRole(ctx, kthID, roleID)
	if err != nil {
		return false, err
	}
	return mayCreate && mayUpdate, nil
}

func (s *UI) MayDeleteRoles(ctx context.Context, kthID string) (map[string]struct{}, error) {
	mayCreate, err := s.MayCreateRoles(ctx, kthID)
	if err != nil {
		return nil, err
	}
	if !mayCreate {
		return make(map[string]struct{}), nil
	}
	deletable := make(map[string]struct{})
	roles, err := s.api.UserGetScopes(ctx, kthID, "pls", "role")
	if err != nil {
		return nil, err
	}
	if slices.Contains(roles, "*") {
		roles, err = s.GetAllRoles(ctx)
		if err != nil {
			return nil, err
		}
	}
	for _, role := range roles {
		deletable[role] = struct{}{}
	}
	return deletable, nil
}

func (s *UI) MayUpdatePermissionsInSystem(ctx context.Context, kthID, system string) (bool, error) {
	systems, err := s.api.UserGetScopes(ctx, kthID, "pls", "system")
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(systems, func(s string) bool { return s == "*" || s == system }), nil
}

func (s *UI) MayUpdatePermissionsInSystems(ctx context.Context, kthID string, systems ...[]string) (map[string]struct{}, error) {
	var sys []string
	switch len(systems) {
	case 0:
		var err error
		sys, err = s.GetAllSystems(ctx)
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
		mayUpdate, err := s.MayUpdatePermissionsInSystem(ctx, kthID, system)
		if err != nil {
			return nil, err
		}
		if mayUpdate {
			granted[system] = struct{}{}
		}
	}
	return granted, nil
}

func (s *UI) MayAddPermissions(ctx context.Context, kthID string) (bool, error) {
	return s.api.UserCheckPermission(ctx, kthID, "pls", "system")
}
