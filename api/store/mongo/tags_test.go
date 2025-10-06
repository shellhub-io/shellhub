package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStore_TagCreate(t *testing.T) {
	now := time.Now()

	clockMock := new(clockmocks.Clock)
	clockMock.On("Now").Return(now)
	clock.DefaultBackend = clockMock

	cases := []struct {
		description string
		tag         *models.Tag
		expected    error
	}{
		{
			description: "succeeds when tag data is valid",
			tag:         &models.Tag{Name: "staging", TenantID: "00000000-0000-4000-0000-000000000000"},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			insertedID, err := s.TagCreate(ctx, tc.tag)
			require.Equal(tt, tc.expected, err)

			if err == nil {
				objID, _ := primitive.ObjectIDFromHex(insertedID)

				tag := make(map[string]interface{})
				require.NoError(tt, db.Collection("tags").FindOne(ctx, bson.M{"_id": objID}).Decode(tag))

				require.Equal(
					tt,
					map[string]interface{}{
						"_id":        objID,
						"created_at": primitive.NewDateTimeFromTime(now),
						"updated_at": primitive.NewDateTimeFromTime(now),
						"name":       "staging",
						"tenant_id":  "00000000-0000-4000-0000-000000000000",
					},
					tag,
				)
			}
		})
	}
}

func TestStore_TagConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		has       bool
		err       error
	}

	cases := []struct {
		description string
		tenantID    string
		target      *models.TagConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{Name: "nonexistent"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts when namespace is different",
			tenantID:    "00000000-0000-4001-0000-000000000000",
			target:      &models.TagConflicts{Name: "production"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{Name: "production"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{"name"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			conflicts, has, err := s.TagConflicts(ctx, tc.tenantID, tc.target)
			require.Equal(tt, tc.expected, Expected{conflicts, has, err})
		})
	}
}

func TestStore_TagList(t *testing.T) {
	type Expected struct {
		tags  []models.Tag
		count int
		err   error
	}

	cases := []struct {
		description string
		fixtures    []string
		options     []store.QueryOption
		expected    Expected
	}{
		{
			description: "succeeds when no filters applied",
			fixtures:    []string{fixtureTags},
			options:     []store.QueryOption{},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3c2a62aafaefe821ab3",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "owners",
						TenantID:  "00000000-0000-4001-0000-000000000000",
					},
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:        "6791d3be5a201d874c4c2885",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "development",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				count: 3,
				err:   nil,
			},
		},
		{
			description: "succeeds when tenant filter applied",
			fixtures:    []string{fixtureTags},
			options: []store.QueryOption{
				func(ctx context.Context) error {
					query := ctx.Value("query").(*[]bson.M)
					*query = append(*query, bson.M{
						"$match": bson.M{
							"tenant_id": "00000000-0000-4000-0000-000000000000",
						},
					})

					return nil
				},
			},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:        "6791d3be5a201d874c4c2885",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "development",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(ns []models.Tag) {
		sort.Slice(ns, func(i, j int) bool {
			return ns[i].Name < ns[j].Name
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			tags, count, err := s.TagList(ctx, tc.options...)

			sort(tc.expected.tags)
			sort(tags)

			require.Equal(tt, tc.expected, Expected{tags: tags, count: count, err: err})
		})
	}
}

func TestStore_TagResolve(t *testing.T) {
	type Expected struct {
		tag *models.Tag
		err error
	}

	cases := []struct {
		description string
		resolver    store.TagResolver
		value       string
		options     []store.QueryOption
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when invalid ObjectID format",
			resolver:    store.TagIDResolver,
			value:       "invalid-id",
			options:     []store.QueryOption{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: nil,
				err: primitive.ErrInvalidHex,
			},
		},
		{
			description: "fails when tag not found by ID",
			resolver:    store.TagIDResolver,
			value:       "000000000000000000000000",
			options:     []store.QueryOption{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving tag by ID",
			resolver:    store.TagIDResolver,
			value:       "6791d3ae04ba86e6d7a0514d",
			options:     []store.QueryOption{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: &models.Tag{
					ID:        "6791d3ae04ba86e6d7a0514d",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "production",
					TenantID:  "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
		{
			description: "fails when tag not found by name",
			resolver:    store.TagNameResolver,
			value:       "nonexistent",
			options:     []store.QueryOption{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving tag by name with tenant filter",
			resolver:    store.TagNameResolver,
			value:       "production",
			options: []store.QueryOption{
				func(ctx context.Context) error {
					query := ctx.Value("query").(*[]bson.M)
					*query = append(*query, bson.M{
						"$match": bson.M{
							"tenant_id": "00000000-0000-4000-0000-000000000000",
						},
					})

					return nil
				},
			},
			fixtures: []string{fixtureTags},
			expected: Expected{
				tag: &models.Tag{
					ID:        "6791d3ae04ba86e6d7a0514d",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "production",
					TenantID:  "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			tag, err := s.TagResolve(ctx, tc.resolver, tc.value, tc.options...)
			require.Equal(tt, tc.expected, Expected{tag: tag, err: err})
		})
	}
}

func TestStore_TagUpdate(t *testing.T) {
	cases := []struct {
		description   string
		tag           *models.Tag
		fixtures      []string
		expected      error
		assertChanges func(context.Context) error
	}{
		{
			description: "fails when tag is not found due to id",
			tag: &models.Tag{
				ID:       "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "edited-tag",
			},
			fixtures:      []string{fixtureTags},
			expected:      store.ErrNoDocuments,
			assertChanges: nil,
		},
		{
			description: "succeeds when tag is found",
			tag: &models.Tag{
				ID:       "6791d3ae04ba86e6d7a0514d",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "edited-tag",
			},
			fixtures: []string{fixtureTags},
			expected: nil,
			assertChanges: func(ctx context.Context) error {
				tag := new(models.Tag)
				objID, _ := primitive.ObjectIDFromHex("6791d3ae04ba86e6d7a0514d")
				err := db.Collection("tags").FindOne(ctx, bson.M{"_id": objID}).Decode(tag)
				if err != nil {
					return err
				}

				if tag.Name != "edited-tag" {
					return errors.New("tag name was not updated")
				}

				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.TagUpdate(ctx, tc.tag)
			require.Equal(tt, tc.expected, err)

			if err == nil && tc.assertChanges != nil {
				require.NoError(tt, tc.assertChanges(ctx))
			}
		})
	}
}

func TestStore_TagPushToTarget(t *testing.T) {
	cases := []struct {
		description string
		id          string
		target      store.TagTarget
		targetID    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag does not exist",
			id:          "000000000000000000000000",
			target:      store.TagTargetDevice,
			targetID:    "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device does not exist",
			id:          "6791d3be5a201d874c4c2885",
			target:      store.TagTargetDevice,
			targetID:    "nonexistent",
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds to push a tag to device",
			id:          "6791d3be5a201d874c4c2885",
			target:      store.TagTargetDevice,
			targetID:    "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.TagPushToTarget(ctx, tc.id, tc.target, tc.targetID)
			require.Equal(tt, tc.expected, err)

			if err != nil {
				return
			}

			var device struct {
				TagIDs []string `bson:"tag_ids"`
			}

			require.NoError(tt, db.Collection("devices").FindOne(ctx, bson.M{"uid": tc.targetID}).Decode(&device))
			fmt.Printf("tag_ids: %+v\n", device.TagIDs)
			require.True(tt, slices.Contains(device.TagIDs, tc.id))
		})
	}
}

func TestTagPullFromTarget(t *testing.T) {
	cases := []struct {
		description string
		id          string
		target      store.TagTarget
		targetIDs   []string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag does not exist",
			id:          "000000000000000000000000",
			target:      store.TagTargetDevice,
			targetIDs:   []string{"5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f"},
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device does not exist",
			id:          "6791d3ae04ba86e6d7a0514d",
			target:      store.TagTargetDevice,
			targetIDs:   []string{"nonexistent"},
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds to pull a tag from device",
			id:          "6791d3ae04ba86e6d7a0514d",
			target:      store.TagTargetDevice,
			targetIDs:   []string{"5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f"},
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    nil,
		},
		{
			description: "succeeds to pull a tag from all targets when no specific targets provided",
			id:          "6791d3ae04ba86e6d7a0514d",
			target:      store.TagTargetDevice,
			targetIDs:   []string{},
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.TagPullFromTarget(ctx, tc.id, tc.target, tc.targetIDs...)
			require.Equal(tt, tc.expected, err)

			if err != nil || len(tc.targetIDs) <= 0 {
				return
			}

			var device struct {
				TagIDs []string `bson:"tag_ids"`
			}

			require.NoError(tt, db.Collection("devices").FindOne(ctx, bson.M{"uid": tc.targetIDs[0]}).Decode(&device))
			require.False(tt, slices.Contains(device.TagIDs, tc.id))
		})
	}
}

func TestStore_TagDelete(t *testing.T) {
	cases := []struct {
		description string
		tag         *models.Tag
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag is not found due to id",
			tag: &models.Tag{
				ID:       "000000000000000000000000",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			fixtures: []string{fixtureTags},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tag is found",
			tag: &models.Tag{
				ID:       "6791d3ae04ba86e6d7a0514d",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			fixtures: []string{fixtureTags},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.TagDelete(ctx, tc.tag)
			require.Equal(tt, tc.expected, err)

			if err != nil {
				return
			}

			objID, _ := primitive.ObjectIDFromHex(tc.tag.ID)
			count, err := db.Collection("tags").CountDocuments(ctx, bson.M{"_id": objID})
			require.NoError(tt, err)
			require.Equal(tt, int64(0), count)
		})
	}
}
