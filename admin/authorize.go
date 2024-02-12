package admin

import (
	"context"
	"strings"
)

func (s *Admin) MayUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "role-"+roleID)
}

func (s *Admin) MayCreateRoles(ctx context.Context, kthID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "create-role")
}

func (s *Admin) MayDeleteRole(ctx context.Context, kthID, roleID string) (bool, error) {
	granted, err := s.api.FilterForUser(ctx, kthID, "pls", []string{"role-" + roleID, "create-role"})
	return len(granted) == 2, err
}

func (s *Admin) MayDeleteRoles(ctx context.Context, kthID string, roleIDs []string) (map[string]struct{}, error) {
	perms := []string{"create-role"}
	for _, id := range roleIDs {
		perms = append(perms, "role-"+id)
	}
	granted, err := s.api.FilterForUser(ctx, kthID, "pls", perms)
	deletable := make(map[string]struct{})
	for _, perm := range granted {
		prefix := "role-"
		if strings.HasPrefix(perm, prefix) {
			deletable[perm[len(prefix):]] = struct{}{}
		}
	}
	return deletable, err
}

func (s *Admin) MayUpdatePermissions(ctx context.Context, kthID, system string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "system-"+system)
}
