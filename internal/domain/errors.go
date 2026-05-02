package domain

// Error code constants used in all API error responses.
const (
	ErrSessionNotFound       = "session_not_found"
	ErrMissingSessionHeader  = "missing_session_header"
	ErrInvalidDisplayName    = "invalid_display_name"
	ErrGameNotFound          = "game_not_found"
	ErrGameNotOpen           = "game_not_open"
	ErrCannotJoinOwnGame     = "cannot_join_own_game"
	ErrNotAParticipant       = "not_a_participant"
	ErrWrongTurn             = "wrong_turn"
	ErrGameNotActive         = "game_not_active"
	ErrIllegalMove           = "illegal_move"
	ErrInvalidNotation       = "invalid_notation"
	ErrDrawClaimIneligible   = "draw_claim_ineligible"
	ErrInvalidDrawClaimType  = "invalid_draw_claim_type"
	ErrCannotCancelGame      = "cannot_cancel_game"
)
