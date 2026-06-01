package repository

import (
	"context"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ListIDs(ctx context.Context) ([]uuid.UUID, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, name, currency string) error
	UpdateBaseBalance(ctx context.Context, userID uuid.UUID, baseBalance float64) error
}
