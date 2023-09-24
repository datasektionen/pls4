package admin

import "context"

func (s *Admin) CanUpdateRole(ctx context.Context, kthID, roleID string) (bool, error) {
	return s.api.CheckUser(ctx, kthID, "pls", "role-" + roleID)
}
