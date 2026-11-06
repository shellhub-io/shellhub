package services

import (
	"context"
	"fmt"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// DeviceSelector is one rule of a namespace that picks devices by the tags they carry or by a
// pattern over their name. Name is how a refused tag deletion lists the rule, such as
// `access policy "production"`. A Hostname of "" selects by no name.
type DeviceSelector struct {
	Name     string
	TagIDs   []string
	Hostname string
}

func (d DeviceSelector) holdsTag(tagID string) bool {
	return slices.Contains(d.TagIDs, tagID)
}

func (d DeviceSelector) tellsApart(from, to string) bool {
	if d.Hostname == "" {
		return false
	}

	before, errBefore := models.MatchPattern(d.Hostname, from)
	after, errAfter := models.MatchPattern(d.Hostname, to)

	return errBefore != nil || errAfter != nil || before != after
}

// DeviceSelectorSource lists one kind of rule that picks devices, such as access policies or the
// enterprise firewall rules. Those rules read the tags and the name of a device, so changing either
// changes what the rule covers. A caller moves a tag a rule holds on or off a device, or renames a
// device into or out of a rule's pattern, only when its role holds the source's EditPermission.
type DeviceSelectorSource interface {
	// EditPermission is the permission under which the source's rules are edited. [NewService] panics
	// when no owner holds it, or when every role that moves tags or renames devices holds it: such a
	// role could change what the rule covers without editing it.
	EditPermission() authorizer.Permission
	// DeviceSelectors returns every rule of the source in sc, or an error when they cannot be
	// listed. A store read through ctx joins the caller's transaction when there is one, and an
	// error fails the change being checked.
	DeviceSelectors(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error)
}

// DeviceSelectorSourceFactoryFunc constructs a DeviceSelectorSource from the core store and cache,
// returning nil when the build has none to add. Enterprise packages register one through
// [RegisterDeviceSelectorSource] in their init functions; it runs during server setup.
type DeviceSelectorSourceFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache) (DeviceSelectorSource, error)

var deviceSelectorSourceFactory DeviceSelectorSourceFactoryFunc

// RegisterDeviceSelectorSource registers the factory of the rules the core store does not know
// of. It must be called before the server's Setup runs.
func RegisterDeviceSelectorSource(f DeviceSelectorSourceFactoryFunc) {
	deviceSelectorSourceFactory = f
}

// DeviceSelectorSourceFactory returns the registered factory, or nil in Community Edition builds.
func DeviceSelectorSourceFactory() DeviceSelectorSourceFactoryFunc {
	return deviceSelectorSourceFactory
}

// WithDeviceSelectorSource adds a kind of rule that picks devices to those the core store holds.
// [NewService] panics when no owner holds the source's EditPermission, or when every role that
// moves tags or renames devices holds it.
func WithDeviceSelectorSource(source DeviceSelectorSource) Option {
	return func(service *APIService) {
		service.selectorSources = append(service.selectorSources, source)
	}
}

var deviceSelectorInputs = []authorizer.Permission{authorizer.TagCreate, authorizer.TagDelete, authorizer.DeviceRename, authorizer.DeviceUpdate}

func mustSetApart(source DeviceSelectorSource) {
	edit := source.EditPermission()

	if !authorizer.RoleOwner.HasPermission(edit) {
		panic(fmt.Sprintf("device selector source %T is edited under a permission no owner holds", source))
	}

	setsApart := slices.ContainsFunc(authorizer.AssignableRoles(), func(role authorizer.Role) bool {
		changesInputs := slices.ContainsFunc(deviceSelectorInputs, role.HasPermission)

		return changesInputs && !role.HasPermission(edit)
	})
	if !setsApart {
		panic(fmt.Sprintf("device selector source %T is edited under a permission every role that moves tags or renames devices holds", source))
	}
}

type accessPolicySelectors struct {
	store store.Store
}

func (accessPolicySelectors) EditPermission() authorizer.Permission {
	return authorizer.AccessPolicyManage
}

func (a accessPolicySelectors) DeviceSelectors(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error) {
	policies, _, err := a.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	selectors := make([]DeviceSelector, 0, len(policies))
	for _, policy := range policies {
		selectors = append(selectors, DeviceSelector{Name: fmt.Sprintf("access policy %q", policy.Name), TagIDs: policy.Filter.TagIDs})
	}

	return selectors, nil
}

type publicKeySelectors struct {
	store store.Store
}

func (publicKeySelectors) EditPermission() authorizer.Permission {
	return authorizer.PublicKeyEdit
}

func (p publicKeySelectors) DeviceSelectors(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error) {
	keys, _, err := p.store.PublicKeyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	selectors := make([]DeviceSelector, 0, len(keys))
	for _, key := range keys {
		selectors = append(selectors, DeviceSelector{Name: fmt.Sprintf("public key %q", key.Name), TagIDs: key.Filter.TagIDs, Hostname: key.Filter.Hostname})
	}

	return selectors, nil
}

func (s *service) coreSelectorSources() []DeviceSelectorSource {
	return []DeviceSelectorSource{publicKeySelectors{store: s.store}, accessPolicySelectors{store: s.store}}
}

func (s *service) selectorsOutOfReach(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error) {
	caller, _ := identity.FromContext(ctx)

	outOfReach := []DeviceSelector{}

	for _, source := range slices.Concat(s.coreSelectorSources(), s.selectorSources) {
		if caller.Role.HasPermission(source.EditPermission()) {
			continue
		}

		selectors, err := source.DeviceSelectors(ctx, sc)
		if err != nil {
			return nil, err
		}

		outOfReach = append(outOfReach, selectors...)
	}

	return outOfReach, nil
}

func (s *service) guardDeviceTag(ctx context.Context, sc scope.Scope, target store.TagTarget, tagID string) error {
	if target != store.TagTargetDevice {
		return nil
	}

	outOfReach, err := s.selectorsOutOfReach(ctx, sc)
	if err != nil {
		return err
	}

	if slices.ContainsFunc(outOfReach, func(d DeviceSelector) bool { return d.holdsTag(tagID) }) {
		return NewErrDeviceSelectorOutOfReach()
	}

	return nil
}

func (s *service) attachTag(ctx context.Context, sc scope.Scope, target store.TagTarget, tagID, targetID string) error {
	if err := s.guardDeviceTag(ctx, sc, target, tagID); err != nil {
		return err
	}

	return s.store.TagPushToTarget(ctx, tagID, target, targetID)
}

func (s *service) detachTag(ctx context.Context, sc scope.Scope, target store.TagTarget, tagID, targetID string) error {
	if err := s.guardDeviceTag(ctx, sc, target, tagID); err != nil {
		return err
	}

	return s.store.TagPullFromTarget(ctx, tagID, target, targetID)
}
