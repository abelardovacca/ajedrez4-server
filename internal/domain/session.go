package domain

import "time"

// Session represents an ephemeral user identity for the server process lifetime.
type Session struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

// ValidateDisplayName returns false if the display name violates constraints:
// must be 1–32 printable (non-control) characters.
func ValidateDisplayName(name string) bool {
	if len(name) == 0 || len([]rune(name)) > 32 {
		return false
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
