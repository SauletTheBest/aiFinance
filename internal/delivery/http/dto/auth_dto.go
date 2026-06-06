package dto


type RegisterRequest struct { 
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Currency string `json:"currency"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code string `json:"code" binding:"required,len=4"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=4"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ResendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type GoogleLoginRequest struct{
	IDToken string `json:"id_token" binding:"required"`
}