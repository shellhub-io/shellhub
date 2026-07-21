package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type AccessPolicy struct {
	bun.BaseModel `bun:"table:access_policies"`

	ID             string    `bun:"id,pk,type:uuid"`
	NamespaceID    string    `bun:"namespace_id"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
	Name           string    `bun:"name"`
	SubjectType    string    `bun:"subject_type"`
	SubjectValue   string    `bun:"subject_value"`
	FilterHostname string    `bun:"filter_hostname"`
	Logins         []string  `bun:"logins,array"`
	RequireStepUp  bool      `bun:"require_step_up"`
	Effect         string    `bun:"effect"`

	Tags []*Tag `bun:"m2m:access_policy_tags,join:AccessPolicy=Tag"`
}

type AccessPolicyTag struct {
	bun.BaseModel  `bun:"table:access_policy_tags"`
	AccessPolicyID string    `bun:"access_policy_id,pk"`
	TagID          string    `bun:"tag_id,pk"`
	CreatedAt      time.Time `bun:"created_at"`

	AccessPolicy *AccessPolicy `bun:"rel:belongs-to,join:access_policy_id=id"`
	Tag          *Tag          `bun:"rel:belongs-to,join:tag_id=id"`
}

func NewAccessPolicyTag(tagID, accessPolicyID string) *AccessPolicyTag {
	return &AccessPolicyTag{TagID: tagID, AccessPolicyID: accessPolicyID}
}

func AccessPolicyFromModel(model *models.AccessPolicy) *AccessPolicy {
	accessPolicy := &AccessPolicy{
		ID:             model.ID,
		NamespaceID:    model.TenantID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		Name:           model.Name,
		SubjectType:    string(model.Subject.Type),
		SubjectValue:   model.Subject.Value,
		FilterHostname: model.Filter.Hostname,
		Logins:         model.Logins,
		RequireStepUp:  model.RequireStepUp,
		Effect:         string(model.Effect),
		Tags:           []*Tag{},
	}

	// Handle Tags if fully populated (e.g., from API requests)
	if len(model.Filter.Tags) > 0 {
		accessPolicy.Tags = make([]*Tag, len(model.Filter.Tags))
		for i, t := range model.Filter.Tags {
			accessPolicy.Tags[i] = TagFromModel(&t)
		}
	} else if len(model.Filter.TagIDs) > 0 {
		// Handle TagIDs if only IDs are provided (e.g., from tests or internal operations)
		accessPolicy.Tags = make([]*Tag, len(model.Filter.TagIDs))
		for i, tagID := range model.Filter.TagIDs {
			accessPolicy.Tags[i] = &Tag{ID: tagID}
		}
	}

	return accessPolicy
}

func AccessPolicyToModel(entity *AccessPolicy) *models.AccessPolicy {
	accessPolicy := &models.AccessPolicy{
		ID:        entity.ID,
		TenantID:  entity.NamespaceID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Subject: models.PolicySubject{
			Type:  models.PolicySubjectType(entity.SubjectType),
			Value: entity.SubjectValue,
		},
		Filter: models.PublicKeyFilter{
			Hostname: entity.FilterHostname,
			Taggable: models.Taggable{
				Tags: []models.Tag{},
			},
		},
		Logins:        entity.Logins,
		RequireStepUp: entity.RequireStepUp,
		Effect:        models.PolicyEffect(entity.Effect),
	}

	if len(entity.Tags) > 0 {
		accessPolicy.Filter.Tags = make([]models.Tag, len(entity.Tags))
		accessPolicy.Filter.TagIDs = make([]string, len(entity.Tags))
		for i, t := range entity.Tags {
			accessPolicy.Filter.Tags[i] = *TagToModel(t)
			accessPolicy.Filter.TagIDs[i] = t.ID
		}
	}

	return accessPolicy
}
