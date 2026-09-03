package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// AccessPolicy is a row of access_policies. TenantID on the model is namespace_id here, and
// the policy's tags are a join table rather than a column.
type AccessPolicy struct {
	bun.BaseModel `bun:"table:access_policies"`

	ID            string    `bun:"id,pk,type:uuid"`
	NamespaceID   string    `bun:"namespace_id"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
	Name          string    `bun:"name"`
	SubjectType   string    `bun:"subject_type"`
	SubjectValue  string    `bun:"subject_value"`
	Logins        []string  `bun:"logins,array"`
	SourceIP      []string  `bun:"source_ip,array"`
	RequireReauth bool      `bun:"require_reauth"`
	ReauthPeriod  *int      `bun:"reauth_period"`
	Action        string    `bun:"action"`

	Tags []*Tag `bun:"m2m:access_policy_tags,join:AccessPolicy=Tag"`
}

// AccessPolicyTag joins a policy to a tag. It is declared because bun needs the join model
// registered, even though nothing reads it directly.
type AccessPolicyTag struct {
	bun.BaseModel  `bun:"table:access_policy_tags"`
	AccessPolicyID string    `bun:"access_policy_id,pk"`
	TagID          string    `bun:"tag_id,pk"`
	CreatedAt      time.Time `bun:"created_at"`

	AccessPolicy *AccessPolicy `bun:"rel:belongs-to,join:access_policy_id=id"`
	Tag          *Tag          `bun:"rel:belongs-to,join:tag_id=id"`
}

// NewAccessPolicyTag returns the join row binding tagID to accessPolicyID.
func NewAccessPolicyTag(tagID, accessPolicyID string) *AccessPolicyTag {
	return &AccessPolicyTag{TagID: tagID, AccessPolicyID: accessPolicyID}
}

// AccessPolicyFromModel projects a policy into its row form, flattening the subject into its
// type and value columns.
func AccessPolicyFromModel(model *models.AccessPolicy) *AccessPolicy {
	accessPolicy := &AccessPolicy{
		ID:            model.ID,
		NamespaceID:   model.TenantID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Name:          model.Name,
		SubjectType:   string(model.Subject.Type),
		SubjectValue:  model.Subject.Value,
		Logins:        model.Logins,
		SourceIP:      model.SourceIP,
		RequireReauth: model.RequireReauth,
		ReauthPeriod:  model.ReauthPeriod,
		Action:        string(model.Action),
		Tags:          []*Tag{},
	}

	if len(model.Filter.Tags) > 0 {
		accessPolicy.Tags = make([]*Tag, len(model.Filter.Tags))
		for i, t := range model.Filter.Tags {
			accessPolicy.Tags[i] = TagFromModel(&t)
		}
	} else if len(model.Filter.TagIDs) > 0 {
		accessPolicy.Tags = make([]*Tag, len(model.Filter.TagIDs))
		for i, tagID := range model.Filter.TagIDs {
			accessPolicy.Tags[i] = &Tag{ID: tagID}
		}
	}

	return accessPolicy
}

// AccessPolicyToModel rebuilds a policy from its row.
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
			Taggable: models.Taggable{
				Tags: []models.Tag{},
			},
		},
		Logins:        entity.Logins,
		SourceIP:      entity.SourceIP,
		RequireReauth: entity.RequireReauth,
		ReauthPeriod:  entity.ReauthPeriod,
		Action:        models.PolicyAction(entity.Action),
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
