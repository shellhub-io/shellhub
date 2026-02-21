package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type TagResolver uint

const (
	TagIDResolver TagResolver = iota + 1
	TagNameResolver
)

type TagTarget int

const (
	TagTargetDevice TagTarget = iota + 1
	TagTargetPublicKey
)

var extraTagTargets []TagTarget

// RegisterTagTarget adds a custom TagTarget to the global list.
// Cloud/enterprise layers use this to extend tag targets without modifying core.
func RegisterTagTarget(t TagTarget) {
	extraTagTargets = append(extraTagTargets, t)
}

func TagTargets() []TagTarget {
	targets := []TagTarget{TagTargetDevice, TagTargetPublicKey}

	return append(targets, extraTagTargets...)
}

type TagsStore interface {
	// TagCreate creates new tag.
	//
	// It returns the inserted ID or an error if any.
	TagCreate(ctx context.Context, tag *models.Tag) (insertedID string, err error)

	// TagConflicts checks for uniqueness violations of tag attributes within a namespace.
	// Only non-zero values in the target are checked for conflicts.
	//
	// Example:
	//     conflicts, _, _ := store.TagConflicts(context.Background(), "tenant123", &models.TagConflicts{Name: "development"})
	//     println(conflicts) // => []string{"name"}
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) (conflicts []string, has bool, err error)

	// TagList retrieves a list of tags based on the provided options.
	//
	// It returns the list of tags, the total count of matching documents (ignoring pagination), and an error if any.
	TagList(ctx context.Context, opts ...QueryOption) (tags []models.Tag, totalCount int, err error)

	// TagResolve fetches a tag using a specific resolver.
	//
	// It returns the resolved tag if found and an error, if any.
	TagResolve(ctx context.Context, resolver TagResolver, value string, opts ...QueryOption) (tag *models.Tag, err error)

	// TagUpdate updates a tag.
	//
	// It returns an error, if any, or store.ErrNoDocuments if the tag does not exist.
	TagUpdate(ctx context.Context, tag *models.Tag) error

	// TagPushToTarget pushs an existent tag to the provided target.
	//
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	TagPushToTarget(ctx context.Context, id string, target TagTarget, targetID string) (err error)

	// TagPullFromTarget removes a tag from tagged documents in a namespace. If targetsID is empty it removes the tag from
	// all documents of the selected target type. If targetsID contains specific target IDs it only removes the tag from those
	// documents.
	//
	// Returns ErrNoDocuments if no matching documents found or other errors from the operation.
	TagPullFromTarget(ctx context.Context, id string, target TagTarget, targetIDs ...string) (err error)

	// TagUpdate deletes a tag.
	//
	// It returns an error, if any, or store.ErrNoDocuments if the tag does not exist.
	TagDelete(ctx context.Context, tag *models.Tag) error
}
