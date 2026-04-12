package httpapi

import (
	"context"

	domainuser "github.com/lpxxn/blink/domain/user"
)

// ResolveUserNames loads display names for the given user IDs (deduplicated).
// Returns nil if users is nil; missing users map to "".
func ResolveUserNames(ctx context.Context, users domainuser.Repository, ids []int64) map[int64]string {
	if users == nil || len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	out := make(map[int64]string, len(uniq))
	for _, id := range uniq {
		u, err := users.GetByID(ctx, id)
		if err != nil || u == nil {
			out[id] = ""
			continue
		}
		out[id] = u.Name
	}
	return out
}
