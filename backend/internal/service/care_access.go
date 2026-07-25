package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
)

// Actor is the authenticated caller, as the middleware resolved them. Handlers
// build one from the request context; services never read gin.
type Actor struct {
	ID   uuid.UUID
	Role string
}

// CareRelation is how an actor is related to the patient whose record they are
// reaching for. It is the single vocabulary for "may I?" across goals, the
// support thread and the caregiver views — so the rule is written once here
// rather than re-derived in every service.
type CareRelation string

const (
	// RelationSelf — the person in recovery, acting on their own record.
	RelationSelf CareRelation = "self"

	// RelationCaregiver — the caregiver this patient linked, and only them.
	// Being a caregiver account is not enough; the link must exist.
	RelationCaregiver CareRelation = "caregiver"

	// RelationAdmin — an administrator. Deliberately distinct from
	// RelationCaregiver: an admin may oversee, but is not a party to a private
	// two-person conversation. Each service states which relations it accepts.
	RelationAdmin CareRelation = "admin"
)

// AuthorRole maps a relation onto the author vocabulary stored on goal updates,
// so a caregiver's note is never attributed to the person in recovery.
func (r CareRelation) AuthorRole() string {
	if r == RelationSelf {
		return model.AuthorRoleUser
	}
	return model.AuthorRoleCaregiver
}

// careAccess resolves the relation between an actor and a patient. It is
// embedded by every service that exposes a patient-scoped record.
type careAccess struct {
	profiles repository.RecoverAIRepository
}

// resolve reports how actor relates to patientID, or a domain error when the
// answer is "not at all".
//
// The check is deliberately link-based rather than role-based: a caregiver
// account may only reach the people who chose them, which is the promise the
// caregiver-linking UI makes.
func (a careAccess) resolve(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
) (CareRelation, error) {
	if actor.ID == uuid.Nil {
		return "", apperr.Unauthorized("unauthorized")
	}
	if actor.ID == patientID {
		return RelationSelf, nil
	}

	profile, err := a.profiles.GetProfileByUserID(ctx, patientID)
	if err != nil {
		return "", apperr.Internal(err)
	}
	if profile != nil && profile.CaregiverID != nil && *profile.CaregiverID == actor.ID {
		return RelationCaregiver, nil
	}

	if actor.Role == model.RoleAdmin {
		return RelationAdmin, nil
	}

	// Same message whether the patient does not exist or simply did not link
	// this caregiver: probing for who is in recovery is not something the API
	// should answer.
	return "", apperr.Forbidden("You do not have access to this person's recovery record.")
}

// requireOneOf resolves the relation and rejects anything not in `allowed`.
func (a careAccess) requireOneOf(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
	allowed ...CareRelation,
) (CareRelation, error) {
	relation, err := a.resolve(ctx, actor, patientID)
	if err != nil {
		return "", err
	}

	for _, candidate := range allowed {
		if relation == candidate {
			return relation, nil
		}
	}
	return "", apperr.Forbidden("You do not have access to this person's recovery record.")
}

// linkedCaregiver returns the caregiver a patient has linked, or nil when they
// have none. Used by the support thread, which cannot exist without a link.
func (a careAccess) linkedCaregiver(ctx context.Context, patientID uuid.UUID) (*uuid.UUID, error) {
	profile, err := a.profiles.GetProfileByUserID(ctx, patientID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile == nil {
		return nil, nil
	}
	return profile.CaregiverID, nil
}
