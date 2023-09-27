package admin

import "context"

func (s *Admin) MayUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "role-" + roleID)
}

func (s *Admin) MayCreateRoles(ctx context.Context, kthID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "create-role")
}
