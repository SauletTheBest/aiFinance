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
	Name 		 string
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string
	Currency     string    `gorm:"type:varchar(3);not null;default:'KZT'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func usertoDomain(model *User) *domain.User {
	return &domain.User{
		ID:           model.ID,
		Name: 		  model.Name,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Currency:     model.Currency,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func toModel(user *domain.User) *User {
	return &User{
		ID:           user.ID,
		Name: 		  user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Currency:     user.Currency,
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

	return usertoDomain(&model), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	
	var model User 
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		return nil, err
	}

	return usertoDomain(&model), nil
}
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *UserRepo) UpdateProfile(ctx context.Context, userID uuid.UUID, name, currency string) error {
	user, err := r.GetByID(ctx, userID)
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
	user.UpdatedAt = time.Now()
	
	return r.Update(ctx, user)
}