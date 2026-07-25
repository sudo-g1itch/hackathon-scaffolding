package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
)

// patientCheckinLimit bounds the check-in history a caregiver sees.
const patientCheckinLimit = 30

// CaregiverPatient is what a caregiver sees in their list of the people they
// support.
//
// PRIVACY: this list view carries only signals — risk, mood, craving, streak,
// goal progress and emergency history. It has never carried a transcript, a
// summary or a trigger, and it still does not. The narrative of a check-in is
// only ever reachable through PatientOverview, and only when the person in
// recovery has switched sharing on themselves.
type CaregiverPatient struct {
	UserID         uuid.UUID  `json:"user_id"`
	Name           string     `json:"name"`
	Goal           string     `json:"goal"`
	Substance      string     `json:"substance"`
	Risk           string     `json:"risk"`
	Emotion        string     `json:"emotion"`
	Craving        int        `json:"craving"`
	RecoveryStreak int        `json:"recovery_streak"`
	LastCheckinAt  *time.Time `json:"last_checkin_at"`
	EmergencyCount int64      `json:"emergency_count"`

	// Plan progress, so a caregiver can see momentum rather than only risk.
	ActiveGoals     int `json:"active_goals"`
	CompletedGoals  int `json:"completed_goals"`
	AverageProgress int `json:"average_progress"`

	UnreadMessages int64 `json:"unread_messages"`
}

// PatientCheckin is one check-in as a caregiver may see it.
//
// Risk, mood, craving, source and time are always present — they are the
// signals the caregiver role exists to watch. Summary and Triggers are the
// check-in's narrative, and are populated ONLY when the person in recovery has
// turned on RecoveryProfile.ShareCheckinDetails. The raw transcript is never
// included under any setting.
type PatientCheckin struct {
	ID         uuid.UUID `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	Risk       string    `json:"risk"`
	Emotion    string    `json:"emotion"`
	Craving    int       `json:"craving"`
	Source     string    `json:"source"`
	Summary    string    `json:"summary,omitempty"`
	Triggers   []string  `json:"triggers,omitempty"`
}

// PatientOverview is the caregiver's detail view of one person: their signals,
// their recovery plan, and their recent check-in history.
type PatientOverview struct {
	Patient  CaregiverPatient `json:"patient"`
	Checkins []PatientCheckin `json:"checkins"`
	Goals    []GoalData       `json:"goals"`
	Summary  GoalSummary      `json:"goal_summary"`

	// SharesCheckinDetails tells the UI why summaries are missing, so it can
	// say "they have kept the details private" instead of rendering blanks.
	SharesCheckinDetails bool `json:"shares_checkin_details"`

	// Relation is how the viewer is related to this person, so one screen can
	// serve the caregiver and an admin's read-only oversight.
	Relation string `json:"relation"`
}

// CareService is the caregiver side of the product: who am I supporting, how
// are they doing, and what is on their plan.
//
// It composes GoalService and SupportService rather than re-querying goals and
// messages, so progress is computed in exactly one place.
type CareService interface {
	ListPatients(ctx context.Context, caregiverID uuid.UUID) ([]CaregiverPatient, error)
	GetPatientOverview(ctx context.Context, actor Actor, patientID uuid.UUID) (*PatientOverview, error)
}

type careService struct {
	careAccess
	repo    repository.RecoverAIRepository
	goals   GoalService
	support SupportService
	log     *zap.Logger
	now     func() time.Time
}

func NewCareService(
	repo repository.RecoverAIRepository,
	goals GoalService,
	support SupportService,
	log *zap.Logger,
) CareService {
	return &careService{
		careAccess: careAccess{profiles: repo},
		repo:       repo,
		goals:      goals,
		support:    support,
		log:        log,
		now:        time.Now,
	}
}

func (s *careService) ListPatients(
	ctx context.Context,
	caregiverID uuid.UUID,
) ([]CaregiverPatient, error) {
	profiles, err := s.repo.GetCaregiverPatients(ctx, caregiverID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// One grouped query for every patient's unread count, rather than one per
	// row inside the loop below.
	unread, err := s.support.UnreadByPatient(ctx, caregiverID)
	if err != nil {
		return nil, err
	}

	patients := make([]CaregiverPatient, 0, len(profiles))
	for i := range profiles {
		patient, err := s.summarisePatient(ctx, &profiles[i])
		if err != nil {
			return nil, err
		}
		patient.UnreadMessages = unread[profiles[i].UserID]
		patients = append(patients, *patient)
	}

	return patients, nil
}

func (s *careService) GetPatientOverview(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
) (*PatientOverview, error) {
	relation, err := s.requireOneOf(
		ctx, actor, patientID, RelationSelf, RelationCaregiver, RelationAdmin,
	)
	if err != nil {
		return nil, err
	}

	profile, err := s.repo.GetProfileByUserID(ctx, patientID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile == nil {
		return nil, apperr.NotFound("recovery profile")
	}

	patient, err := s.summarisePatient(ctx, profile)
	if err != nil {
		return nil, err
	}

	// The person themself always sees their own narrative; the consent flag
	// only governs what someone ELSE may read.
	shareNarrative := relation == RelationSelf || profile.ShareCheckinDetails

	checkins, err := s.repo.ListCheckins(ctx, patientID, patientCheckinLimit)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	history := make([]PatientCheckin, 0, len(checkins))
	for i := range checkins {
		history = append(history, toPatientCheckin(&checkins[i], shareNarrative))
	}

	goals, err := s.goals.List(ctx, actor, patientID)
	if err != nil {
		return nil, err
	}

	summary, err := s.goals.Summary(ctx, patientID)
	if err != nil {
		return nil, err
	}

	if relation == RelationCaregiver {
		unread, err := s.support.UnreadByPatient(ctx, actor.ID)
		if err != nil {
			return nil, err
		}
		patient.UnreadMessages = unread[patientID]
	}

	return &PatientOverview{
		Patient:              *patient,
		Checkins:             history,
		Goals:                goals,
		Summary:              *summary,
		SharesCheckinDetails: profile.ShareCheckinDetails,
		Relation:             string(relation),
	}, nil
}

// summarisePatient turns a profile into the signal row both the list and the
// detail header render, so the two can never disagree.
func (s *careService) summarisePatient(
	ctx context.Context,
	profile *model.RecoveryProfile,
) (*CaregiverPatient, error) {
	patient := &CaregiverPatient{
		UserID:    profile.UserID,
		Goal:      profile.Goal,
		Substance: profile.Substance,
		Risk:      model.RiskLow,
		Emotion:   "Unknown",
	}
	if profile.User != nil {
		patient.Name = profile.User.FullName()
	}

	lastCheckin, err := s.repo.GetLastCheckin(ctx, profile.UserID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if lastCheckin != nil {
		patient.Risk = model.NormalizeRisk(lastCheckin.Risk)
		patient.Craving = lastCheckin.Craving
		patient.LastCheckinAt = &lastCheckin.CreatedAt
		if lastCheckin.Emotion != "" {
			patient.Emotion = lastCheckin.Emotion
		}
		// Transcript, Summary and Triggers are intentionally not copied here —
		// see the privacy note on CaregiverPatient.
	}

	days, err := s.repo.CheckinDays(ctx, profile.UserID, streakLookbackDays)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	patient.RecoveryStreak = calculateStreak(days, s.now())

	emergencies, err := s.repo.CountEmergencyLogs(ctx, profile.UserID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	patient.EmergencyCount = emergencies

	summary, err := s.goals.Summary(ctx, profile.UserID)
	if err != nil {
		return nil, err
	}
	patient.ActiveGoals = summary.Active
	patient.CompletedGoals = summary.Completed
	patient.AverageProgress = summary.AverageProgress

	return patient, nil
}

// toPatientCheckin projects a check-in for someone other than its author.
// withNarrative is the consent decision, made by the caller.
func toPatientCheckin(checkin *model.Checkin, withNarrative bool) PatientCheckin {
	data := PatientCheckin{
		ID:         checkin.ID,
		OccurredAt: checkin.CreatedAt,
		Risk:       model.NormalizeRisk(checkin.Risk),
		Emotion:    checkin.Emotion,
		Craving:    checkin.Craving,
		Source:     checkin.Source,
	}

	if withNarrative {
		data.Summary = checkin.Summary
		data.Triggers = checkin.Triggers
	}

	// The transcript is never projected, at any consent level.
	return data
}
