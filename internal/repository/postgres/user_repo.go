package postgres

import (
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"context"
	"gorm.io/gorm"
	"github.com/google/uuid"
	"time"
)


type UserRepo struct {
	db *gorm.DB
}
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func toDomain(model *User) *domain.User {
	return &domain.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func toModel(user *domain.User) *User {
	return &User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func NewUserRepo(db *gorm.DB) repository.UserRepository {
	return &UserRepo{db: db }
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model User

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error 

	if err != nil {
		return  nil, err
	}

	return toDomain(&model), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	
	var model User 
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		return nil, err
	}

	return toDomain(&model), nil
}