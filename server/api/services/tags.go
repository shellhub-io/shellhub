package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// TagFilterFields maps each filter field the tag list endpoint accepts to the
// set of operators valid for it. Tags currently have no filterable fields, so
// this is an empty FieldConstraints that causes all filter attempts to be
// rejected at the handler level.
var TagFilterFields = query.NewFieldConstraints(map[string][]string{})

// TagSortFields is the set of field names accepted in the sort_by query
// parameter when listing tags.
var TagSortFields = query.NewFieldSet(
	"name",
	"created_at",
	"updated_at",
)

// TagsService owns tags, which are namespace-scoped labels attached to devices and keys.
type TagsService interface {
	// CreateTag creates a new tag in the specified tenant namespace.
	//
	// Tags can share the same attributes (e.g. name) if they belong to different tenants.
	// For example, tenant1 and tenant2 can each have a tag named "production".
	//
	// It returns the insertedID and an error if any. A name already taken yields
	// [ErrDuplicateTagName], carrying the conflicting field name(s).
	CreateTag(ctx context.Context, req *requests.CreateTag) (insertedID string, err error)

	// PushTagTo adds an existing tag in a namespace to a target document (e.g. Device, PublicKey, FirewallRule).
	//
	// Returns ErrNamespaceNotFound if namespace not found, ErrTagNotFound if tag not found, or other errors if operation fails.
	PushTagTo(ctx context.Context, target store.TagTarget, req *requests.PushTag) (err error)

	// PullTagFrom removes a tag from a target document in a namespace. The tag itself is not deleted.
	// If no other documents reference the tag, it becomes orphaned but remains available for future use.
	//
	// Returns ErrNamespaceNotFound if namespace not found, ErrTagNotFound if tag not found, or other errors if operation fails.
	PullTagFrom(ctx context.Context, target store.TagTarget, req *requests.PullTag) (err error)

	// ListTags retrieves a batch of tags that belong to the given namespace.
	//
	// It returns the list of tags with pagination, an integer representing the total count of tags in the
	// database, ignoring pagination, and an error if any.
	ListTags(ctx context.Context, req *requests.ListTags) (tags []models.Tag, totalCount int, err error)

	// UpdateTag updates a tag with the specified name in the specified namespace.
	//
	// It returns an error if any. A name already taken yields [ErrDuplicateTagName], carrying the
	// conflicting field name(s).
	UpdateTag(ctx context.Context, req *requests.UpdateTag) (err error)

	// DeleteTag deletes a tag with the specified name in the specified namespace.
	//
	// It returns an error if any.
	DeleteTag(ctx context.Context, req *requests.DeleteTag) (err error)
}

func (s *service) CreateTag(ctx context.Context, req *requests.CreateTag) (string, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return "", err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return "", NewErrNamespaceNotFound(req.TenantID, err)
	}

	if conflicts, has, err := s.store.TagConflicts(ctx, sc, &models.TagConflicts{Name: req.Name}); has || err != nil {
		if !has {
			return "", err
		}

		return "", NewErrTagDuplicated(conflicts, err)
	}

	insertedID, err := s.store.TagCreate(ctx, &models.Tag{Name: req.Name, TenantID: req.TenantID})
	if err != nil {
		return "", err
	}

	return insertedID, nil
}

func (s *service) PushTagTo(ctx context.Context, target store.TagTarget, req *requests.PushTag) (err error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, req.Name)
	if err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	return s.store.TagPushToTarget(ctx, tag.ID, target, req.TargetID)
}

func (s *service) PullTagFrom(ctx context.Context, target store.TagTarget, req *requests.PullTag) (err error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, req.Name)
	if err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	return s.store.TagPullFromTarget(ctx, tag.ID, target, req.TargetID)
}

func (s *service) ListTags(ctx context.Context, req *requests.ListTags) ([]models.Tag, int, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return []models.Tag{}, 0, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return []models.Tag{}, 0, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	if req.Sorter.Order == "" {
		req.Sorter.Order = query.OrderDesc
	}

	req.Sorter.Tiebreak = "id"

	opts := []store.QueryOption{
		s.store.Options().Match(&req.Filters),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	}

	tags, totalCount, err := s.store.TagList(ctx, sc, opts...)
	if err != nil {
		return []models.Tag{}, 0, err
	}

	return tags, totalCount, nil
}

func (s *service) UpdateTag(ctx context.Context, req *requests.UpdateTag) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, req.Name)
	if err != nil {
		return NewErrTagNotFound(req.Name, err)
	}

	conflictsAttrs := &models.TagConflicts{}
	if req.NewName != "" && req.NewName != req.Name {
		conflictsAttrs.Name = req.NewName
	}

	if conflicts, has, err := s.store.TagConflicts(ctx, sc, conflictsAttrs); has || err != nil {
		if !has {
			return err
		}

		return NewErrTagDuplicated(conflicts, err)
	}

	if req.NewName != "" && !strings.EqualFold(req.NewName, tag.Name) {
		tag.Name = req.NewName
	}

	return s.store.TagUpdate(ctx, tag)
}

func (s *service) DeleteTag(ctx context.Context, req *requests.DeleteTag) error {
	return s.store.WithTransaction(ctx, s.deleteTagCallback(req))
}

func (s *service) deleteTagCallback(req *requests.DeleteTag) store.TransactionCb {
	return func(ctx context.Context) error {
		sc, err := BoundTo(req.TenantID)
		if err != nil {
			return err
		}

		if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
			return NewErrNamespaceNotFound(req.TenantID, err)
		}

		tag, err := s.store.TagResolve(ctx, sc, store.TagNameResolver, req.Name)
		if err != nil {
			return NewErrTagNotFound(req.Name, err)
		}

		for _, target := range store.TagTargets() {
			if err := s.store.TagPullFromTarget(ctx, tag.ID, target); err != nil {
				return err
			}
		}

		return s.store.TagDelete(ctx, tag)
	}
}
