package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/SauletTheBest/BackendFinancialApplication/internal/domain"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/repository"
	"github.com/google/uuid"
)


type GoalUseCase interface {
	CreateGoal(ctx context.Context, userID uuid.UUID, title string, description *string, targetAmount float64, deadline *time.Time) (*domain.Goal, error)
	GetGoalsByUserID(ctx context.Context, userID uuid.UUID)([]*domain.Goal, error)
	GetGoalByID(ctx context.Context, goalID uuid.UUID, userID uuid.UUID) (*domain.Goal, error)
	AddProgress(ctx context.Context, goalID uuid.UUID, userID uuid.UUID, amountToAdd float64) (*domain.Goal, error)
	DeleteGoal(ctx context.Context, goalID uuid.UUID, userID uuid.UUID) error
}


type goalUseCase struct {
	goalRepo repository.GoalRepository
}

func NewGoalUseCase(goalRepo repository.GoalRepository) GoalUseCase {
	return &goalUseCase{
		goalRepo: goalRepo,
	}
}

func (u *goalUseCase) CreateGoal(ctx context.Context, userID uuid.UUID, title string, description *string, targetAmount float64, deadline *time.Time) (*domain.Goal, error) {
    // 💡 Business Rule 1: You can't have a goal to save $0!
	if targetAmount <= 0 {
		return nil, errors.New("target amount must be greater than zero")
	}

	goal := &domain.Goal{
		ID:            uuid.New(), // Generate a new barcode
		UserID:        userID,
		Title:         title,
		Description:   description,
		TargetAmount:  targetAmount,
		CurrentAmount: 0.0, // 💡 Business Rule 2: A new goal ALWAYS starts at zero!
		Deadline:      deadline,
		Status:        domain.GoalStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := u.goalRepo.Create(ctx, goal)
	if err != nil {
		return nil, err
	}
	return goal, nil
}

func (u *goalUseCase) GetGoalsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Goal, error) {
	
	goals, err := u.goalRepo.GetByUserID(ctx, userID)
    
	if err != nil {
        return nil, err
    }
    for _, goal := range goals {
        checkExpiry(goal)
    }
	
	return goals, nil
}
func (u *goalUseCase) GetGoalByID(ctx context.Context, goalID uuid.UUID, userID uuid.UUID) (*domain.Goal, error){
	goal, err := u.goalRepo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this goal")
	}

	checkExpiry(goal) // check deadline

	return goal, nil
}

func (u *goalUseCase) AddProgress(ctx context.Context, goalID uuid.UUID, userID uuid.UUID, amountToAdd float64) (*domain.Goal, error) {
    // You can't add negative money
	if amountToAdd <= 0 {
		return nil, errors.New("must add a positive amount")
	}
	// Step 1: Fetch the goal from the database using our Repository
	goal, err := u.goalRepo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	// Ensure this user actually owns this goal!
    // We don't want User A modifying User B's goals.
	if goal.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this goal")
	}
	// Step 2: Add the money
	goal.CurrentAmount += amountToAdd
	goal.UpdatedAt = time.Now()
	// Did they hit the target? Mark it completed!
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.Status = domain.GoalStatusCompleted
	}
	// Step 3: Save the updated goal back to the database
	err = u.goalRepo.Update(ctx, goal)
	if err != nil {
		return nil, err
	}
	return goal, nil
}

func (u *goalUseCase) DeleteGoal(ctx context.Context, goalID uuid.UUID, userID uuid.UUID) error {
	// 1. Fetch the goal
	goal, err := u.goalRepo.GetByID(ctx, goalID)
	if err != nil {
		return err // Goal doesn't exist
	}

	// 2. Security Check! Only the owner can delete it.
	if goal.UserID != userID {
		return errors.New("unauthorized: you do not own this goal")
	}

	// 3. Use the repo toolbox to delete it
	return u.goalRepo.Delete(ctx, goalID)
}

// checkExpiry automatically marks a goal as EXPIRED if the deadline passed
// and it was never completed. This is called every time we READ a goal.
func checkExpiry(goal *domain.Goal) {
    if goal.Status == domain.GoalStatusActive &&
        goal.Deadline != nil &&
        time.Now().After(*goal.Deadline) {
        goal.Status = domain.GoalStatusExpired
    }
}
