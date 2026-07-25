package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

// CareHandler serves the caregiver side of the product: the people a caregiver
// supports, one person's detailed picture, and the private conversation between
// the two of them.
//
// Who may see what is decided entirely in the service layer — a route guard can
// only check a role, and "is this caregiver linked to this person" is not a
// role question.
type CareHandler struct {
	care    service.CareService
	support service.SupportService
}

func NewCareHandler(care service.CareService, support service.SupportService) *CareHandler {
	return &CareHandler{care: care, support: support}
}

// GET /api/v1/caregiver — everyone who chose the acting caregiver.
func (h *CareHandler) ListPatients(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patients, err := h.care.ListPatients(c.Request.Context(), actor.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, patients)
}

// GET /api/v1/patients/:patientID — one person's signals, plan and check-in
// history. Serves the caregiver's detail screen; a user may also call it for
// themselves.
func (h *CareHandler) PatientOverview(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	overview, err := h.care.GetPatientOverview(c.Request.Context(), actor, patientID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, overview)
}

// GET /api/v1/patients/:patientID/messages — the shared conversation. Reading
// it marks the other side's messages as read.
func (h *CareHandler) Thread(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	thread, err := h.support.GetThread(c.Request.Context(), actor, patientID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, thread)
}

type supportMessageRequest struct {
	Body string `json:"body" binding:"required,max=2000"`
}

// POST /api/v1/patients/:patientID/messages
func (h *CareHandler) SendMessage(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	var req supportMessageRequest
	if !response.BindJSON(c, &req) {
		return
	}

	// The whole thread comes back, not just the new message: the reply lands in
	// the same shape the screen already renders, so there is no second fetch
	// and no client-side merge to get wrong.
	thread, err := h.support.Send(c.Request.Context(), actor, patientID, req.Body)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, thread)
}

// POST /api/v1/patients/:patientID/messages/read — clear the unread badge
// without loading the conversation.
func (h *CareHandler) MarkRead(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	patientID, ok := uuidParam(c, "patientID")
	if !ok {
		return
	}

	if err := h.support.MarkRead(c.Request.Context(), actor, patientID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// POST /api/v1/emergency/:logID/acknowledge — the caregiver confirming they
// have seen an alert. Only then can the app tell the sender somebody is there.
func (h *CareHandler) AcknowledgeEmergency(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	logID, ok := uuidParam(c, "logID")
	if !ok {
		return
	}

	alert, err := h.care.AcknowledgeEmergency(c.Request.Context(), actor, logID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, alert)
}

// GET /api/v1/messages/unread — one number for the navigation badge, whichever
// side of the conversation the caller is on.
func (h *CareHandler) UnreadCount(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}

	count, err := h.support.UnreadForUser(c.Request.Context(), actor)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{"unread": count})
}
