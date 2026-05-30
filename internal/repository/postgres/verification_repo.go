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
	// Step 1: Clean the house! Delete any older codes for this user
	err := r.db.WithContext(ctx).
		Table("verification_codes").
		Where("user_id = ? AND code_type = ?", code.UserID, code.CodeType).
		Delete(&domain.VerificationCode{}).Error
	if err != nil {
		return err
	}
	// Step 2: Save the brand new code in the clean database
	return r.db.WithContext(ctx).Table("verification_codes").Create(code).Error
}


func (r *VerificationRepo) GetValidCode(ctx context.Context, userID uuid.UUID, codeStr string, codeType string) (*domain.VerificationCode, error) {
	var code domain.VerificationCode
	
	// We look for a code that matches everything AND hasn't been used yet!
	err := r.db.WithContext(ctx).
		Table("verification_codes").
		Where("user_id = ? AND code = ? AND code_type = ? AND used = false", userID, codeStr, codeType).
		First(&code).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if code is wrong or already used
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
