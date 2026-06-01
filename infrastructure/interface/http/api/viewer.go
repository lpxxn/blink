package httpapi

import domainuser "github.com/lpxxn/blink/domain/user"

// viewerBypassesPostVisibility is true when the viewer may read any non-deleted post
// (admin moderation), matching GetForViewer's elevated access.
func viewerBypassesPostVisibility(role string) bool {
	return role == domainuser.RoleSuperAdmin || role == domainuser.RoleAdmin
}
