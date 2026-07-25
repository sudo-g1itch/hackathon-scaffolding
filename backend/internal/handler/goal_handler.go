package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

// GoalHandler exposes the multi-goal recovery plan over HTTP. It stays thin:
// bind, resolve the acting user, call the service, write the envelope. Every
// authorisation decision belongs to the service, which is the only layer that
// knows whether a caregiver is linked to this person.
type GoalHandler struct {
	svc service.GoalService
}

func NewGoalHandler(svc service.GoalService) *GoalHandler {
	return &GoalHandler{svc: svc}
}

// GET /api/v1/goals — the acting user's own plan.
func (h *GoalHandler) List(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	goals, err := h.svc.List(c.Request.Context(), actor, actor.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, goals)
}

// GET /api/v1/goals/summary — the roll-up the dashboard shows.
func (h *GoalHandler) Summary(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	summary, err := h.svc.Summary(c.Request.Context(), actor.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, summary)
}

// GET /api/v1/patients/:patientID/goals — someone else's plan, if you may see it.
func (h *GoalHandler) ListForPatient(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	goals, err := h.svc.List(c.Request.Context(), actor, patientID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, goals)
}

// goalRequest is the create body. Category and unit are free-form on the wire
// and normalised by the service, so an unknown category degrades to "other"
// rather than 422-ing a user mid-form.
type goalRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=200"`
	Description string `json:"description" binding:"max=2000"`
	Category    string `json:"category" binding:"max=50"`
	TargetValue int    `json:"target_value" binding:"omitempty,min=1,max=100000"`
	Unit        string `json:"unit" binding:"max=50"`

	// RFC3339, or omitted for a goal with no deadline.
	TargetDate *time.Time `json:"target_date"`
}

// POST /api/v1/goals — add a goal to my own plan.
func (h *GoalHandler) Create(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	h.create(c, actor, actor.ID)
}

// POST /api/v1/patients/:patientID/goals — a caregiver suggesting a goal.
func (h *GoalHandler) CreateForPatient(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	h.create(c, actor, patientID)
}

func (h *GoalHandler) create(c *gin.Context, actor service.Actor, patientID uuid.UUID) {
	var req goalRequest
	if !response.BindJSON(c, &req) {
		return
	}

	goal, err := h.svc.Create(c.Request.Context(), actor, patientID, service.GoalInput{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
		TargetDate:  req.TargetDate,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, goal)
}

// GET /api/v1/goals/:goalID — a goal and its progress feed.
func (h *GoalHandler) Get(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	goalID, ok := uuidParam(c, "goalID")
	if !ok {
		return
	}

	detail, err := h.svc.Get(c.Request.Context(), actor, goalID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, detail)
}

// goalPatchRequest is a partial update: an omitted field is left alone, which
// is why every member is a pointer.
type goalPatchRequest struct {
	Title        *string `json:"title" binding:"omitempty,min=2,max=200"`
	Description  *string `json:"description" binding:"omitempty,max=2000"`
	Category     *string `json:"category" binding:"omitempty,max=50"`
	Status       *string `json:"status" binding:"omitempty,max=50"`
	TargetValue  *int    `json:"target_value" binding:"omitempty,min=1,max=100000"`
	CurrentValue *int    `json:"current_value" binding:"omitempty,min=0,max=100000"`
	Unit         *string `json:"unit" binding:"omitempty,max=50"`

	TargetDate *time.Time `json:"target_date"`

	// clear_target_date:true removes a deadline. Without it there is no way to
	// tell "leave the date alone" from "delete the date" in JSON.
	ClearTargetDate bool `json:"clear_target_date"`
}

// PUT /api/v1/goals/:goalID
func (h *GoalHandler) Update(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	goalID, ok := uuidParam(c, "goalID")
	if !ok {
		return
	}

	var req goalPatchRequest
	if !response.BindJSON(c, &req) {
		return
	}

	goal, err := h.svc.Update(c.Request.Context(), actor, goalID, service.GoalPatch{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Status:          req.Status,
		TargetValue:     req.TargetValue,
		CurrentValue:    req.CurrentValue,
		Unit:            req.Unit,
		TargetDate:      req.TargetDate,
		ClearTargetDate: req.ClearTargetDate,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, goal)
}

// DELETE /api/v1/goals/:goalID
func (h *GoalHandler) Delete(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	goalID, ok := uuidParam(c, "goalID")
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), actor, goalID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

type progressRequest struct {
	Delta *int   `json:"delta" binding:"omitempty,min=-100000,max=100000"`
	Value *int   `json:"value" binding:"omitempty,min=0,max=100000"`
	Note  string `json:"note" binding:"max=1000"`
	Kind  string `json:"kind" binding:"max=50"`
}

// POST /api/v1/goals/:goalID/progress — log movement, a note, or (from a
// caregiver) a word of encouragement.
func (h *GoalHandler) LogProgress(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	goalID, ok := uuidParam(c, "goalID")
	if !ok {
		return
	}

	var req progressRequest
	if !response.BindJSON(c, &req) {
		return
	}

	detail, err := h.svc.LogProgress(c.Request.Context(), actor, goalID, service.ProgressInput{
		Delta: req.Delta,
		Value: req.Value,
		Note:  req.Note,
		Kind:  req.Kind,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, detail)
}
