package handler

import (
	"net/http"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/delivery/http/dto"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoalHandler struct {
	GoalUseCase usecase.GoalUseCase
}

func NewGoalHandler(goalUseCase usecase.GoalUseCase) *GoalHandler {
	return &GoalHandler{GoalUseCase: goalUseCase}
}

// POST /api/goals
func (h *GoalHandler) CreateGoal(c *gin.Context) {
	// 1. Get secure User ID from JWT Token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// 2. Parse the JSON from the frontend
	var req dto.CreateGoalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Send data to UseCase
	goal, err := h.GoalUseCase.CreateGoal(c.Request.Context(), userID, req.Title, req.TargetAmount, req.Deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Send response back
	response := dto.GoalResponse{
		ID:            goal.ID,
		Title:         goal.Title,
		TargetAmount:  goal.TargetAmount,
		CurrentAmount: goal.CurrentAmount,
		Deadline:      goal.Deadline,
		Status:        string(goal.Status),
		CreatedAt:     goal.CreatedAt,
	}
	c.JSON(http.StatusCreated, response)
}

// PUT /api/goals/:id/contribute
func (h *GoalHandler) AddProgress(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var req dto.AddProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.GoalUseCase.AddProgress(c.Request.Context(), goalID, userID, req.Amount)
	if err != nil {
		// UseCase will return an error if the user tries to hack someone else's goal
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress added successfully!", "current_amount": goal.CurrentAmount})
}

// GET /api/goals
func (h *GoalHandler) GetGoals(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	goals, err := h.GoalUseCase.GetGoalsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert Domain goals to DTO responses
	var responses []dto.GoalResponse
	for _, goal := range goals {
		responses = append(responses, dto.GoalResponse{
			ID:            goal.ID,
			Title:         goal.Title,
			TargetAmount:  goal.TargetAmount,
			CurrentAmount: goal.CurrentAmount,
			Deadline:      goal.Deadline,
			Status:        string(goal.Status),
			CreatedAt:     goal.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// GET /api/goals/:id
func (h *GoalHandler) GetGoal(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	// Get the "id" from the URL, e.g. /api/goals/123e4567-e89b...
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	goal, err := h.GoalUseCase.GetGoalByID(c.Request.Context(), goalID, userID)
	if err != nil {
		// Returns an error if it doesn't exist OR if they don't own it
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := dto.GoalResponse{
		ID:            goal.ID,
		Title:         goal.Title,
		TargetAmount:  goal.TargetAmount,
		CurrentAmount: goal.CurrentAmount,
		Deadline:      goal.Deadline,
		Status:        string(goal.Status),
		CreatedAt:     goal.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DELETE /api/goals/:id
func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	err = h.GoalUseCase.DeleteGoal(c.Request.Context(), goalID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal deleted successfully"})
}
