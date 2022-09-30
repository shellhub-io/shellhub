package businesses

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type NamespaceBuilder struct {
	ctx           context.Context // nolint: containedctx
	store         store.Store
	tenant        string
	name          string
	id            string
	sessionRecord bool
	maxDevices    int
	err           error
}

const (
	// NamespaceDeviceLimit is the maximum number of devices allowed in a namespace at the community version.
	// -1 means ulimited.
	NamespaceDeviceLimit = -1
	// CloudNamespaceDeviceLimit is the maximum number of devices allowed in a namespace at the cloud.
	CloudNamespaceDeviceLimit = 3
)

// Namespace initializes a new NamespaceBuilder to create a new namespace.
func Namespace(ctx context.Context, store store.Store) *NamespaceBuilder {
	builder := &NamespaceBuilder{ // nolint: exhaustruct
		ctx:   ctx,
		store: store,
	}

	return builder
}

// FromUser sets the namespace owner's ID.
//
// id is the ID from the user that will be the owner of the namespace.
//
// FromUser builder method is required.
// FromUser builder method checks if the user exists. If it does not exist, the error is appended to errs on the
// builder.
func (b *NamespaceBuilder) FromUser(id string) *NamespaceBuilder { // nolint: varnamelen
	if b.err != nil {
		return b
	}

	if user, _, err := b.store.UserGetByID(b.ctx, id, false); err != nil || user == nil {
		b.err = NewErrUserNotFound(id, err)

		return b
	}

	b.id = id

	return b
}

// WithUsernem sets the namespace owner's ID.
//
// username is the username from the user that will be the owner of the namespace.
//
// WithUsername builder method is required if WithUser is not set.
// WithUsername builder method checks if the user exists. If it does not exist, the error is appended to errs on the
// builder.
func (b *NamespaceBuilder) WithUsername(username string) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	var user *models.User
	user, err := b.store.UserGetByUsername(b.ctx, username)
	if err != nil || user == nil {
		b.err = NewErrUserNotFound(username, err)

		return b
	}

	b.id = user.ID

	return b
}

// WithTenantID sets the tenant ID to the namespace builder.
//
// tenant is the tenant ID for the namespace.
//
// WithTenantID builder method is required.
// WithTenantID builder method checks if the tenant ID is empty, it will generate a new UUID.
func (b *NamespaceBuilder) WithTenantID(tenant string) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	if tenant == "" {
		tenant = uuid.Generate()
	}

	b.tenant = tenant

	return b
}

func (b *NamespaceBuilder) FromTenantID(tenant string) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	namespace, err := b.store.NamespaceGet(b.ctx, b.tenant)
	if err != nil || namespace == nil {
		b.err = err // NewErrNamespaceNotFound(b.tenant, err)

		return b
	}

	b.tenant = tenant

	return b
}

// WithName sets the name to the namespace builder.
//
// name is the name for the namespace.
//
// WithName builder method is required.
// WithName builder method checks if the namespace exists. If it does not exist, the error is appended to errs on the
// builder.
func (b *NamespaceBuilder) WithName(name string) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	if namespace, err := b.store.NamespaceGetByName(b.ctx, name); namespace != nil {
		b.err = NewErrNamespaceDuplicated(err)

		return b
	}

	b.name = name

	return b
}

// WithSessionRecord sets the session record to the namespace builder.
//
// status is the status for the session record.
//
// WithSessionRecord builder method is optional.
func (b *NamespaceBuilder) WithSessionRecord(status bool) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	b.sessionRecord = status

	return b
}

// WithMaxDevices sets the maximum number of devices allowed in a namespace.
//
// max is the maximum number of devices allowed in a namespace.
//
// WithMaxDevices builder method is optional.
// WithMaxDevices builder method overwrites the default maximum number of devices allowed in a namespace set on the
// builder's constructor.
func (b *NamespaceBuilder) WithMaxDevices(max int) *NamespaceBuilder {
	if b.err != nil {
		return b
	}

	b.maxDevices = max

	return b
}

// Create creates a new namespace.
func (b *NamespaceBuilder) Create() (*models.Namespace, error) {
	// Checks if it is a cloud's instance is running. If it is, the namespace will be created with the cloud's limits.
	// If it is not, the namespace will be created with community's limit, what is ilimited.
	if envs.IsCloud() {
		b.maxDevices = CloudNamespaceDeviceLimit
	} else {
		b.maxDevices = NamespaceDeviceLimit
	}

	if b.err != nil {
		return nil, b.err
	}

	namespace := &models.Namespace{ // nolint: exhaustruct
		Name:     b.name,
		TenantID: b.tenant,
		Owner:    b.id,
		Members: []models.Member{
			{
				ID:   b.id,
				Role: guard.RoleOwner,
			},
		},
		Settings:   &models.NamespaceSettings{SessionRecord: b.sessionRecord},
		MaxDevices: b.maxDevices,
		CreatedAt:  time.Now(),
	}

	if _, err := models.Validate(namespace); err != nil {
		return nil, err
	}

	return b.store.NamespaceCreate(b.ctx, namespace)
}

// Delete deletes a namespace.
func (b *NamespaceBuilder) Delete() (*models.Namespace, error) {
	if b.err != nil {
		return nil, b.err
	}

	namespace, err := b.store.NamespaceGet(b.ctx, b.tenant)
	if err != nil {
		// b.err = NewErrNamespaceNotFound(tenantID, err)

		return nil, err
	}

	err = b.store.NamespaceDelete(b.ctx, b.tenant)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

// AddMember adds a member to a namespace.
func (b *NamespaceBuilder) AddMember(memberID, role string) (*models.Namespace, error) {
	if b.err != nil {
		return nil, b.err
	}

	/*if _, err := validator.ValidateStruct(models.Member{Username: memberUsername, Role: memberRole}); err != nil {
		return nil, NewErrNamespaceMemberInvalid(err)
	}*/

	namespace, err := b.store.NamespaceGet(b.ctx, b.tenant)
	if err != nil || namespace == nil {
		return nil, NewErrNamespaceNotFound(b.tenant, err)
	}

	active, ok := guard.CheckMember(namespace, b.id)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(b.id, err)
	}

	passive, err := b.store.UserGetByUsername(b.ctx, memberID)
	if err != nil {
		return nil, NewErrUserNotFound(memberID, err)
	}

	_, ok = guard.CheckMember(namespace, passive.ID)
	if ok {
		return nil, NewErrNamespaceMemberDuplicated(passive.ID, nil)
	}

	if !guard.CheckRole(active.Role, role) {
		return nil, guard.ErrForbidden
	}

	return b.store.NamespaceAddMember(b.ctx, b.tenant, passive.ID, role)
}
