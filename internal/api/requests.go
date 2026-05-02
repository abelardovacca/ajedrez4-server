package api

// CreateSessionRequest is the JSON body for POST /sessions.
type CreateSessionRequest struct {
	DisplayName string `json:"display_name"`
}

// SubmitMoveRequest is the JSON body for POST /games/{id}/moves.
type SubmitMoveRequest struct {
	Notation string `json:"notation"`
}

// ClaimDrawRequest is the JSON body for POST /games/{id}/draw.
type ClaimDrawRequest struct {
	ClaimType string `json:"claim_type"`
}
