package postgres


import (
	"context"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type GoalRepo struct {
	db *gorm.DB
}

type Goal struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	Title         string
	Description   *string
	TargetAmount  float64
	CurrentAmount float64   `gorm:"default:0"`
	Deadline      *time.Time
	Status        string    `gorm:"default:'ACTIVE'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Goal) TableName() string {
	return "user_goals"
}

// 2. Mappers
func goalToDomain(model *Goal) *domain.Goal {
	return &domain.Goal{
		ID:            model.ID,
		UserID:        model.UserID,
		Title:         model.Title,
		TargetAmount:  model.TargetAmount,
		Description:   model.Description,
		CurrentAmount: model.CurrentAmount,
		Deadline:      model.Deadline,
		Status:        domain.GoalStatus(model.Status),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

func goalToModel(domainGoal *domain.Goal) *Goal {
	return &Goal{
		ID:            domainGoal.ID,
		UserID:        domainGoal.UserID,
		Title:         domainGoal.Title,
		Description:   domainGoal.Description,
		TargetAmount:  domainGoal.TargetAmount,
		CurrentAmount: domainGoal.CurrentAmount,
		Deadline:      domainGoal.Deadline,
		Status:        string(domainGoal.Status),
		CreatedAt:     domainGoal.CreatedAt,
		UpdatedAt:     domainGoal.UpdatedAt,
	}
}

func NewGoalRepo(db *gorm.DB) repository.GoalRepository {
	return &GoalRepo{db: db}
}

func (r *GoalRepo) Create(ctx context.Context, goal *domain.Goal) error {
	model := goalToModel(goal)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GoalRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	var model Goal
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}
	return goalToDomain(&model), nil
}

func (r *GoalRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Goal, error) {
	var models []Goal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	var goals []*domain.Goal
	for _, m := range models {
		goals = append(goals, goalToDomain(&m))
	}
	return goals, nil
}

func (r *GoalRepo) Update(ctx context.Context, goal *domain.Goal) error {
	model := goalToModel(goal)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GoalRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Goal{}, "id = ?", id).Error
}
