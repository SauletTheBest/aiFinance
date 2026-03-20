package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"time"
)
type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}
func (uc *UserUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, name, currency string) error {
    user, err := uc.userRepo.GetByID(ctx, userID)
    if err != nil || user == nil {
        return err
    }
    
    // Update fields if provided
    if name != "" {
        user.Name = name
    }
    if currency != "" {
        user.Currency = currency
    }
    user.UpdatedAt = time.Now() // Set to time.Now() in real code
    
    return uc.userRepo.Update(ctx, user)
}