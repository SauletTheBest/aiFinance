package usecase

import (
	"context"
	"errors" //
	"github.com/google/uuid"
	"time"
	"math/rand"
	"fmt"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/password"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain" //? do we really need that?
	"github.com/SauletTheBest/BackendFinancialApplication/internal/email"
)
 
type OAuthVerifier interface {
	VerifyToken(ctx context.Context, token string) (email string, name string, err error)
}

type AuthUsecase struct {
	userRepo  		 repository.UserRepository	
	jwtSvc 			 *jwt.Service
	verificationRepo repository.VerificationRepository 
	emailService     *email.EmailService    
	oauthVerifier    OAuthVerifier    
}


func NewAuthUsecase(
	userRepo repository.UserRepository, jwtSvc *jwt.Service,
	verificationRepo repository.VerificationRepository,
	emailService *email.EmailService,
	oauthVerifier OAuthVerifier,
	) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		jwtSvc: jwtSvc,
		verificationRepo: verificationRepo,
		emailService:     emailService,
		oauthVerifier: 	  oauthVerifier,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, name string, email string, passwordRaw string, currency string) error {
	//logic

	existingUser, err := u.userRepo.GetByEmail(ctx, email) //

	if err == nil && existingUser != nil { //
        return errors.New("user with this email already exists")
    }
	
	hash , err := password.Hash(passwordRaw)

	if err != nil {
		return err
	}
	user := &domain.User {
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Currency:     currency,
		IsVerified:   false, 
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	codeStr := generateVerificationCode()

	vCode := &domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      codeStr,
		CodeType:  domain.CodeTypeEmailVerify,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}

	err = u.verificationRepo.Create(ctx, vCode)
	if err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}

	go func() {
		// We use a fresh context for the goroutine
		err := u.emailService.SendVerificationEmail(email, codeStr)
		if err != nil {
			fmt.Printf("Error sending verification email to %s: %v\n", email, err)
		}
	}()


	return nil
}

func (u *AuthUsecase) Login(ctx context.Context, email string, passwordRaw string) (string, error) {
	//logic
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !user.IsVerified {
		return "", errors.New("unverified_email") // We return a specific error so Flutter knows to show the Verify screen
	}

	err = password.Compare(user.PasswordHash, passwordRaw)
	if err != nil {
		return "", err
	}
	token, err := u.jwtSvc.GenerateToken(user.ID.String()) 
	
	if err != nil {
		return "", err
	}

	return token, nil
}

// 🆕 Notice it accepts `email string` now!
func (u *AuthUsecase) VerifyEmail(ctx context.Context, email string, codeStr string) (string, error) {
	// 1. Fetch the user using the email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 2. Ask the repository if this code exists for THIS user's ID
	vCode, err := u.verificationRepo.GetValidCode(ctx, user.ID, codeStr, domain.CodeTypeEmailVerify)
	if err != nil || vCode == nil {
		return "", errors.New("invalid or already used code")
	}

	// 3. Check if it's expired
	if time.Now().After(vCode.ExpiresAt) {
		return "", errors.New("this code has expired, please request a new one")
	}

	// 4. Mark code as used
	err = u.verificationRepo.MarkAsUsed(ctx, vCode.ID)
	if err != nil {
		return "", err
	}

	// 5. Update the User to be verified!
	user.IsVerified = true
	user.UpdatedAt = time.Now()
	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return "", err
	}
	
	// 6. Return the JWT Token!
	return u.jwtSvc.GenerateToken(user.ID.String())
}


// ForgotPassword generates a code and sends the reset email
func (u *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	// 1. Check if user actually exists
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user with this email does not exist")
	}

	// 2. Generate 4-digit code
	codeStr := generateVerificationCode()

	// 3. Save to database
	vCode := &domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      codeStr,
		CodeType:  domain.CodeTypePasswordReset, // 🆕 Notice the different type!
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}
	err = u.verificationRepo.Create(ctx, vCode)
	if err != nil {
		return err
	}

	// 4. Send Email in background
	go func() {
		_ = u.emailService.SendPasswordResetEmail(email, codeStr)
	}()

	return nil
}

// ResetPassword checks the code and updates the password
func (u *AuthUsecase) ResetPassword(ctx context.Context, email, codeStr, newPasswordRaw string) error {
	// 1. Find the user
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// 2. Check if the code is valid
	vCode, err := u.verificationRepo.GetValidCode(ctx, user.ID, codeStr, domain.CodeTypePasswordReset)
	if err != nil || vCode == nil {
		return errors.New("invalid or already used reset code")
	}
	if time.Now().After(vCode.ExpiresAt) {
		return errors.New("this code has expired")
	}

	// 3. Hash the new password
	newHash, err := password.Hash(newPasswordRaw)
	if err != nil {
		return err
	}

	// 4. Update the user
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 5. Mark code as used
	return u.verificationRepo.MarkAsUsed(ctx, vCode.ID)
}

func generateVerificationCode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

// ResendVerificationCode generates a new code, cleans up old ones, and fires a background email
func (u *AuthUsecase) ResendVerificationCode(ctx context.Context, email string) error {
	// 1. Check if the user exists
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// 2. If already verified, we shouldn't send anything
	if user.IsVerified {
		return errors.New("email is already verified")
	}

	// 3. Generate a new 4-digit code
	codeStr := generateVerificationCode()

	vCode := &domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      codeStr,
		CodeType:  domain.CodeTypeEmailVerify,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}

	// 4. Save to database (this automatically triggers the clean-up we wrote in Repo!)
	err = u.verificationRepo.Create(ctx, vCode)
	if err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}

	// 5. Send email in a separate background routine so the user doesn't wait
	go func() {
		err := u.emailService.SendVerificationEmail(email, codeStr)
		if err != nil {
			fmt.Printf("Error sending verification email to %s: %v\n", email, err)
		}
	}()

	return nil
}

func (u *AuthUsecase) LoginWithGoogle(ctx context.Context, idToken string) (string, error) {
	// 1. Verify token & retrieve user details using our decoupled interface
	email, name, err := u.oauthVerifier.VerifyToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	// 2. Fetch user
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// If user doesn't exist, register them automatically
		if err.Error() == "record not found" {
			user = &domain.User{
				ID:           uuid.New(),
				Name:         name,
				Email:        email,
				PasswordHash: "", // OAuth2 user has no password hash
				Currency:     "KZT",
				BaseBalance:  0,
				IsVerified:   true, // Already validated via Google OAuth
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			err = u.userRepo.Create(ctx, user)
			if err != nil {
				return "", fmt.Errorf("failed to register user: %w", err)
			}
		} else {
			return "", fmt.Errorf("database lookup error: %w", err)
		}
	}
	// 3. Generate session JWT
	token, err := u.jwtSvc.GenerateToken(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("session token generation failed: %w", err)
	}
	return token, nil
}