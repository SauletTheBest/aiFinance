package dto


import "time"

type ProfileResponse struct {
	ID string `json:"ID"`
	Email string `json:"email"`
	Name string `json:"name"`
	Currency string `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateProfileRequest – For future PUT /profile endpoint
type UpdateProfileRequest struct {
    Name     string `json:"name,omitempty"`
    Currency string `json:"currency,omitempty"`
	//email can be added in future
}