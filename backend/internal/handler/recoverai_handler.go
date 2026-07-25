package handler

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

// audioFormField is the multipart field name both check-in and transcribe read.
const audioFormField = "audio"

// RecoverAIHandler exposes the recovery features over HTTP. It stays thin:
// bind and validate, call the service, write the envelope.
type RecoverAIHandler struct {
	svc service.RecoverAIService

	// maxAudioBytes caps an uploaded recording so a large or malicious upload
	// cannot exhaust API memory.
	maxAudioBytes int64
}

func NewRecoverAIHandler(svc service.RecoverAIService, maxAudioBytes int64) *RecoverAIHandler {
	return &RecoverAIHandler{svc: svc, maxAudioBytes: maxAudioBytes}
}

// GET /api/v1/capabilities — which optional integrations are configured.
func (h *RecoverAIHandler) Capabilities(c *gin.Context) {
	response.OK(c, h.svc.Capabilities())
}

// POST /api/v1/checkin — multipart form with an "audio" file.
func (h *RecoverAIHandler) Checkin(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	audio, mimeType, err := h.readAudio(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	checkin, err := h.svc.ProcessVoiceCheckin(c.Request.Context(), userID, audio, mimeType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, checkin)
}

type riskRequest struct {
	Transcript string `json:"transcript" binding:"required,min=2,max=5000"`
}

// POST /api/v1/risk — analyse a typed check-in.
//
// The same reasoning as the voice flow, for users who cannot speak right now
// (no microphone permission, a shared room) and as a demo fallback.
func (h *RecoverAIHandler) Risk(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req riskRequest
	if !response.BindJSON(c, &req) {
		return
	}

	checkin, err := h.svc.ProcessTextCheckin(c.Request.Context(), userID, req.Transcript)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, checkin)
}

// POST /api/v1/voice/transcribe — speech to text only, no analysis or storage.
func (h *RecoverAIHandler) Transcribe(c *gin.Context) {
	audio, mimeType, err := h.readAudio(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	transcript, err := h.svc.Transcribe(c.Request.Context(), audio, mimeType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, map[string]string{"transcript": transcript})
}

type speakRequest struct {
	Text string `json:"text" binding:"required,max=5000"`
}

// POST /api/v1/voice/speak — text to speech; streams MP3 audio.
//
// This endpoint answers with audio rather than the JSON envelope: the browser
// feeds the bytes straight to an <audio> element. Errors still use the envelope.
func (h *RecoverAIHandler) Speak(c *gin.Context) {
	var req speakRequest
	if !response.BindJSON(c, &req) {
		return
	}

	audio, err := h.svc.Speak(c.Request.Context(), req.Text)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, service.TTSMimeType, audio)
}

// GET /api/v1/dashboard
func (h *RecoverAIHandler) Dashboard(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	data, err := h.svc.GetDashboardData(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, data)
}

// POST /api/v1/emergency
func (h *RecoverAIHandler) Emergency(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	result, err := h.svc.TriggerEmergency(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

type chatRequest struct {
	Message string `json:"message" binding:"required,max=2000"`
}

// POST /api/v1/coach/chat
func (h *RecoverAIHandler) CoachChat(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req chatRequest
	if !response.BindJSON(c, &req) {
		return
	}

	history, err := h.svc.SendCoachMessage(c.Request.Context(), userID, req.Message)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, history)
}

// GET /api/v1/coach/history
func (h *RecoverAIHandler) CoachHistory(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	history, err := h.svc.GetCoachHistory(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, history)
}

type eduRequest struct {
	Query string `json:"query" binding:"required,max=500"`
}

// POST /api/v1/education
func (h *RecoverAIHandler) Education(c *gin.Context) {
	var req eduRequest
	if !response.BindJSON(c, &req) {
		return
	}

	result, err := h.svc.Educate(c.Request.Context(), req.Query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, map[string]string{"result": result})
}

// GET /api/v1/timeline
func (h *RecoverAIHandler) Timeline(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	events, err := h.svc.GetTimeline(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, events)
}

// GET /api/v1/profile
func (h *RecoverAIHandler) GetProfile(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

type profileRequest struct {
	Goal             string `json:"goal" binding:"max=255"`
	Substance        string `json:"substance" binding:"max=100"`
	CaregiverName    string `json:"caregiver_name" binding:"max=150"`
	CaregiverPhone   string `json:"caregiver_phone" binding:"max=50"`
	EmergencyContact string `json:"emergency_contact" binding:"max=150"`

	// Consent for the linked caregiver to read a check-in's narrative. Absent
	// means false, which is the private default.
	ShareCheckinDetails bool `json:"share_checkin_details"`
}

// PUT /api/v1/profile
func (h *RecoverAIHandler) UpdateProfile(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req profileRequest
	if !response.BindJSON(c, &req) {
		return
	}

	profile, err := h.svc.UpdateProfile(c.Request.Context(), userID, service.ProfileInput{
		Goal:                req.Goal,
		Substance:           req.Substance,
		CaregiverName:       req.CaregiverName,
		CaregiverPhone:      req.CaregiverPhone,
		EmergencyContact:    req.EmergencyContact,
		ShareCheckinDetails: req.ShareCheckinDetails,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

// GET /api/v1/caregivers — caregiver accounts the user may link to.
func (h *RecoverAIHandler) GetCaregivers(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	options, err := h.svc.ListAvailableCaregivers(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, options)
}

type setCaregiverRequest struct {
	// A null or omitted caregiver_id unlinks the current caregiver.
	CaregiverID *string `json:"caregiver_id" binding:"omitempty,uuid"`
}

// PUT /api/v1/profile/caregiver
func (h *RecoverAIHandler) SetCaregiver(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req setCaregiverRequest
	if !response.BindJSON(c, &req) {
		return
	}

	var caregiverID *uuid.UUID
	if req.CaregiverID != nil && *req.CaregiverID != "" {
		parsed, err := uuid.Parse(*req.CaregiverID)
		if err != nil {
			response.Error(c, apperr.Validation(apperr.Fields{"caregiver_id": {"must be a valid UUID"}}))
			return
		}
		caregiverID = &parsed
	}

	if err := h.svc.SetCaregiver(c.Request.Context(), userID, caregiverID); err != nil {
		response.Error(c, err)
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, profile)
}

// readAudio pulls the uploaded recording out of the multipart form, refusing
// anything larger than the configured cap.
func (h *RecoverAIHandler) readAudio(c *gin.Context) ([]byte, string, error) {
	// Bound the request body before touching it, so an oversized upload is
	// rejected instead of being buffered into memory or onto disk.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxAudioBytes)

	file, header, err := c.Request.FormFile(audioFormField)
	if err != nil {
		return nil, "", apperr.Validation(apperr.Fields{audioFormField: {"is required"}})
	}
	defer func() { _ = file.Close() }()

	if header.Size > h.maxAudioBytes {
		return nil, "", apperr.Unprocessable("That recording is too large. Please record a shorter check-in.")
	}

	audio, err := io.ReadAll(file)
	if err != nil {
		return nil, "", apperr.Unprocessable("That recording could not be read. Please try again.")
	}
	if len(audio) == 0 {
		return nil, "", apperr.Validation(apperr.Fields{audioFormField: {"is empty"}})
	}

	return audio, audioMimeType(header), nil
}

// audioMimeType reads the browser-declared content type for the uploaded part.
func audioMimeType(header *multipart.FileHeader) string {
	if header == nil {
		return ""
	}
	return header.Header.Get("Content-Type")
}
