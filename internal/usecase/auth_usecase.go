package usecase

import (
	"context"
	"errors" //
	"github.com/google/uuid"
	"time"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/jwt"
	"github.com/SauletTheBest/BackendFinancialApplication/pkg/password"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain" //? do we really need that?
)

type AuthUsecase struct {
	userRepo  repository.UserRepository	
	jwtSvc *jwt.Service
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtSvc *jwt.Service) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		jwtSvc: jwtSvc,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, name string, email string, passwordRaw string, currency string) (string, error) {
	//logic

	existingUser, err := u.userRepo.GetByEmail(ctx, email) //

	if err == nil && existingUser != nil { //
        return "", errors.New("user with this email already exists")
    }
	
	hash , err := password.Hash(passwordRaw)

	if err != nil {
		return "", err
	}
	user := &domain.User {
		ID:  uuid.New(),
		Name: name,
		Email: email,
		PasswordHash: hash,
		Currency: currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = u.userRepo.Create(ctx, user)
	
	if err != nil {
		return "", err
	}

	token, err := u.jwtSvc.GenerateToken(user.ID.String())
	
	if err != nil {
		return "", err
	}

	return token, nil //changed so it can also return string {maybe i will remove}
}

func (u *AuthUsecase) Login(ctx context.Context, email string, passwordRaw string) (string, error) {
	//logic
	user, err := u.userRepo.GetByEmail(ctx, email)

	if err != nil {
		return "", err
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