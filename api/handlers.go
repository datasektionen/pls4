package api

import (
	"context"
	"fmt"
	"regexp"
)

var (
	systemRegex     = regexp.MustCompile("^[a-z0-9]+(-[a-z0-9]+)*$")
	permissionRegex = systemRegex
)

type Permission struct {
	PermissionID string   `json:"permission"`
	Scopes       []string `json:"scopes,omitempty"`
}

func (s *API) UserGetPermissions(ctx context.Context, kthID, system string) ([]Permission, error) {
	if !systemRegex.MatchString(system) {
		return nil, fmt.Errorf("Invalid permission %v. Must match %v", system, systemRegex)
	}
	rows, err := s.db.QueryContext(ctx, `--sql
		with recursive all_roles (role_id) as (
			select role_id from roles_users
			where kth_id = $1 and now() between start_date and end_date
			union
			select superrole_id from all_roles
			inner join roles_roles
				on subrole_id = role_id
		)
		select permission_id, coalesce(scope, '') from all_roles a
		inner join roles_permissions p
			using (role_id)
		inner join permission_instances i
			on i.id = p.permission_instance_id
		where i.system_id = $2
		order by permission_id
	`, kthID, system)
	if err != nil {
		return nil, err
	}
	perms := make([]Permission, 1)
	p := &perms[0]
	for rows.Next() {
		var id, scope string
		if err := rows.Scan(&id, &scope); err != nil {
			return nil, err
		}
		if p.PermissionID != id {
			perms = append(perms, Permission{
				PermissionID: id,
				Scopes:       []string{},
			})
			p = &perms[len(perms)-1]
		}
		p.Scopes = append(p.Scopes, scope)
	}
	return perms[1:], nil
}

func (s *API) UserCheckPermission(ctx context.Context, kthID, system string, permission string) (bool, error) {
	if !systemRegex.MatchString(system) {
		return false, fmt.Errorf("Invalid permission %v. Must match %v", system, systemRegex)
	}
	if !permissionRegex.MatchString(permission) {
		return false, fmt.Errorf("Invalid permission %v. Must match %v", permission, permissionRegex)
	}
	permissions, err := s.UserGetPermissions(ctx, kthID, system)
	if err != nil {
		return false, err
	}
	for _, perm := range permissions {
		if permission == perm.PermissionID {
			return true, nil
		}
	}
	return false, nil
}

func (s *API) UserGetScopes(ctx context.Context, kthID, system string, permission string) ([]string, error) {
	if !systemRegex.MatchString(system) {
		return nil, fmt.Errorf("Invalid permission %v. Must match %v", system, systemRegex)
	}
	if !permissionRegex.MatchString(permission) {
		return nil, fmt.Errorf("Invalid permission %v. Must match %v", permission, permissionRegex)
	}
	permissions, err := s.UserGetPermissions(ctx, kthID, system)
	if err != nil {
		return nil, err
	}
	for _, perm := range permissions {
		if permission == perm.PermissionID {
			return perm.Scopes, nil
		}
	}
	return nil, nil
}
