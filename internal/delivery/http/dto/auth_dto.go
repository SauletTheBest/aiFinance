package dto


type RegisterRequest struct { //basically two rows but i can extend in the future
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}