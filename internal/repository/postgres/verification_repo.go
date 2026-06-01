package postgres

import (
	"context"
	"errors"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VerificationRepo struct {
	db *gorm.DB
}

func NewVerificationRepo(db *gorm.DB) repository.VerificationRepository {
	return &VerificationRepo{db: db}
}

func (r *VerificationRepo) Create(ctx context.Context, code *domain.VerificationCode) error {
	return r.db.WithContext(ctx).Table("verification_codes").Create(code).Error
}

func (r *VerificationRepo) GetValidCode(ctx context.Context, userID uuid.UUID, codeStr string, codeType string) (*domain.VerificationCode, error) {
	var code domain.VerificationCode

	err := r.db.WithContext(ctx).
		Table("verification_codes").
		Where("user_id = ? AND code = ? AND code_type = ? AND used = false", userID, codeStr, codeType).
		First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &code, nil
}

func (r *VerificationRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Table("verification_codes").
		Where("id = ?", id).
		Update("used", true).Error
}

func (r *VerificationRepo) MarkUnusedAsUsed(ctx context.Context, userID uuid.UUID, codeType string) error {
	return r.db.WithContext(ctx).
		Table("verification_codes").
		Where("user_id = ? AND code_type = ? AND used = false", userID, codeType).
		Update("used", true).Error
}
