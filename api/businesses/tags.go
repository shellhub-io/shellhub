package businesses

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
)

type TagBuilder struct {
	ctx    context.Context
	store  store.Store
	tenant string
	tagNew string
	tagOld string
	err    error
}

func Tag(ctx context.Context, store store.Store) *TagBuilder {
	return &TagBuilder{
		ctx:   ctx,
		store: store,
	}
}

func (b *TagBuilder) FromTenant(tenant string) *TagBuilder {
	if b.err != nil {
		return b
	}

	namespace, err := b.store.NamespaceGet(b.ctx, tenant)
	if err != nil || namespace == nil {
		b.err = NewErrNamespaceNotFound(tenant, err)

		return b
	}

	b.tenant = tenant

	return b
}

func (b *TagBuilder) FromTag(tag string) *TagBuilder {
	if b.err != nil {
		return b
	}

	tags, count, err := b.store.TagsGet(b.ctx, b.tenant)
	if err != nil || count == 0 {
		b.err = NewErrTagEmpty(b.tenant, err)

		return b
	}

	var found bool
	for _, t := range tags {
		if t == tag {
			b.tagOld = tag

			found = true
		}
	}

	if !found {
		b.err = NewErrTagNotFound(tag, nil)

		return b
	}

	b.tagOld = tag

	return b
}

func (b *TagBuilder) ToTag(tag string) *TagBuilder {
	if b.err != nil {
		return b
	}

	tags, count, err := b.store.TagsGet(b.ctx, b.tenant)
	if err != nil || count == 0 {
		b.err = NewErrTagEmpty(b.tenant, err)

		return b
	}

	var found bool
	for _, t := range tags {
		if t == tag {
			found = true
		}
	}

	if found {
		b.err = NewErrTagDuplicated(tag, nil)

		return b
	}

	b.tagNew = tag

	return b
}

func (b *TagBuilder) Get() ([]string, int, error) {
	if b.err != nil {
		return nil, 0, b.err
	}

	return b.store.TagsGet(b.ctx, b.tenant)
}

func (b *TagBuilder) Delete() error {
	if b.err != nil {
		return b.err
	}

	return b.store.TagDelete(b.ctx, b.tenant, b.tagOld)
}

func (b *TagBuilder) Rename() error {
	if b.err != nil {
		return b.err
	}

	return b.store.TagRename(b.ctx, b.tenant, b.tagOld, b.tagNew)
}
