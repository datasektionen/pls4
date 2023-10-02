package admin

import "context"

func (s *Admin) MayUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "role-" + roleID)
}

func (s *Admin) MayCreateRoles(ctx context.Context, kthID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "create-role")
}

func (s *Admin) MayDeleteRoles(ctx context.Context, kthID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "delete-role")
}

func (s *Admin) MayUpdatePermissions(ctx context.Context, kthID, system string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "system-" + system)
}
