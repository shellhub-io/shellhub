package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration108Up(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		description string
		setup       func() error
		verify      func(tt *testing.T)
	}{
		{
			description: "succeeds creating tags collection with unique IDs and correct mapping",
			setup: func() error {
				devices := []bson.M{
					{
						"uid":       "device-1",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "backend", "critical"},
					},
					{
						"uid":       "device-2",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "frontend"},
					},
					{
						"uid":       "device-3",
						"tenant_id": "tenant-2",
						"tags":      []string{"production", "database"},
					},
					{
						"uid":       "device-4",
						"tenant_id": "tenant-1",
						"tags":      []string{},
					},
					{
						"uid":       "device-5",
						"tenant_id": "tenant-2",
						"tags":      nil,
					},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2], devices[3], devices[4]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tags").Find(ctx, bson.M{})
				require.NoError(tt, err)

				tags := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &tags))
				require.Equal(tt, 6, len(tags))

				tagsMap := make(map[string]map[string]primitive.ObjectID)
				for _, tag := range tags {
					require.NotNil(tt, tag["created_at"])
					require.NotNil(tt, tag["updated_at"])

					tenantID := tag["tenant_id"].(string)
					name := tag["name"].(string)
					id := tag["_id"].(primitive.ObjectID)

					if tagsMap[tenantID] == nil {
						tagsMap[tenantID] = make(map[string]primitive.ObjectID)
					}
					tagsMap[tenantID][name] = id
				}

				require.Contains(tt, tagsMap["tenant-1"], "production")
				require.Contains(tt, tagsMap["tenant-1"], "backend")
				require.Contains(tt, tagsMap["tenant-1"], "critical")
				require.Contains(tt, tagsMap["tenant-1"], "frontend")
				require.Contains(tt, tagsMap["tenant-2"], "production")
				require.Contains(tt, tagsMap["tenant-2"], "database")
			},
		},
		{
			description: "succeeds migrating device tags to tag_ids with correct ObjectID references",
			setup: func() error {
				devices := []bson.M{
					{
						"uid":       "device-1",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "backend"},
					},
					{
						"uid":       "device-2",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "frontend", "backend"},
					},
					{
						"uid":       "device-3",
						"tenant_id": "tenant-2",
						"tags":      []string{"staging", "backend"},
					},
					{
						"uid":       "device-4",
						"tenant_id": "tenant-1",
						"tags":      []string{},
					},
					{
						"uid":       "device-5",
						"tenant_id": "tenant-1",
						"tags":      nil,
					},
				}

				_, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2], devices[3], devices[4]})

				return err
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tags").Find(ctx, bson.M{})
				require.NoError(tt, err)

				tags := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &tags))

				tagsMap := make(map[string]map[string]primitive.ObjectID)
				for _, tag := range tags {
					tenantID := tag["tenant_id"].(string)
					name := tag["name"].(string)
					id := tag["_id"].(primitive.ObjectID)

					if tagsMap[tenantID] == nil {
						tagsMap[tenantID] = make(map[string]primitive.ObjectID)
					}
					tagsMap[tenantID][name] = id
				}

				device1 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-1"}).Decode(&device1))
				require.Nil(tt, device1["tags"])
				require.Equal(tt, 2, len(device1["tag_ids"].(primitive.A)))
				require.Contains(tt, device1["tag_ids"].(primitive.A), tagsMap["tenant-1"]["production"])
				require.Contains(tt, device1["tag_ids"].(primitive.A), tagsMap["tenant-1"]["backend"])

				device2 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-2"}).Decode(&device2))
				require.Nil(tt, device2["tags"])
				require.Equal(tt, 3, len(device2["tag_ids"].(primitive.A)))
				require.Contains(tt, device2["tag_ids"].(primitive.A), tagsMap["tenant-1"]["production"])
				require.Contains(tt, device2["tag_ids"].(primitive.A), tagsMap["tenant-1"]["frontend"])
				require.Contains(tt, device2["tag_ids"].(primitive.A), tagsMap["tenant-1"]["backend"])

				device3 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-3"}).Decode(&device3))
				require.Nil(tt, device3["tags"])
				require.Equal(tt, 2, len(device3["tag_ids"].(primitive.A)))
				require.Contains(tt, device3["tag_ids"].(primitive.A), tagsMap["tenant-2"]["staging"])
				require.Contains(tt, device3["tag_ids"].(primitive.A), tagsMap["tenant-2"]["backend"])

				device4 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-4"}).Decode(&device4))
				require.Nil(tt, device4["tags"])
				require.Equal(tt, 0, len(device4["tag_ids"].(primitive.A)))

				device5 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("devices").FindOne(ctx, bson.M{"uid": "device-5"}).Decode(&device5))
				require.Nil(tt, device5["tags"])
				require.Equal(tt, 0, len(device5["tag_ids"].(primitive.A)))
			},
		},
		{
			description: "succeeds migrating firewall_rules filter.tags to filter.tag_ids",
			setup: func() error {
				devices := []bson.M{
					{
						"uid":       "device-1",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "backend"},
					},
					{
						"uid":       "device-2",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "frontend"},
					},
					{
						"uid":       "device-3",
						"tenant_id": "tenant-2",
						"tags":      []string{"production"},
					},
				}

				firewallRules := []bson.M{
					{
						"tenant_id": "tenant-1",
						"priority":  1,
						"filter": bson.M{
							"tags": []string{"production", "backend"},
						},
					},
					{
						"tenant_id": "tenant-1",
						"priority":  2,
						"filter": bson.M{
							"tags": []string{"production", "frontend"},
						},
					},
					{
						"tenant_id": "tenant-2",
						"priority":  1,
						"filter": bson.M{
							"tags": []string{"production"},
						},
					},
					{
						"tenant_id": "tenant-2",
						"priority":  2,
						"filter": bson.M{
							"tags": []string{},
						},
					},
					{
						"tenant_id": "tenant-1",
						"priority":  3,
						"filter": bson.M{
							"hostname": ".*",
						},
					},
				}

				if _, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2]}); err != nil {
					return err
				}

				if _, err := c.Database("test").Collection("firewall_rules").InsertMany(ctx, []any{firewallRules[0], firewallRules[1], firewallRules[2], firewallRules[3], firewallRules[4]}); err != nil { // nolint:revive
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tags").Find(ctx, bson.M{})
				require.NoError(tt, err)

				tags := make([]bson.M, 0)
				require.NoError(tt, cursor.All(ctx, &tags))

				tagsMap := make(map[string]map[string]primitive.ObjectID)
				for _, tag := range tags {
					tenantID := tag["tenant_id"].(string)
					name := tag["name"].(string)
					id := tag["_id"].(primitive.ObjectID)

					if tagsMap[tenantID] == nil {
						tagsMap[tenantID] = make(map[string]primitive.ObjectID)
					}
					tagsMap[tenantID][name] = id
				}

				rule1 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"priority": 1, "tenant_id": "tenant-1"}).Decode(&rule1))
				require.Nil(tt, rule1["filter"].(bson.M)["tags"])
				require.Equal(tt, 2, len(rule1["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, rule1["filter"].(bson.M)["tag_ids"].(primitive.A), tagsMap["tenant-1"]["production"])
				require.Contains(tt, rule1["filter"].(bson.M)["tag_ids"].(primitive.A), tagsMap["tenant-1"]["backend"])

				rule2 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"priority": 2, "tenant_id": "tenant-1"}).Decode(&rule2))
				require.Nil(tt, rule2["filter"].(bson.M)["tags"])
				require.Equal(tt, 2, len(rule2["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, rule2["filter"].(bson.M)["tag_ids"].(primitive.A), tagsMap["tenant-1"]["production"])
				require.Contains(tt, rule2["filter"].(bson.M)["tag_ids"].(primitive.A), tagsMap["tenant-1"]["frontend"])

				rule3 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"priority": 1, "tenant_id": "tenant-2"}).Decode(&rule3))
				require.Nil(tt, rule3["filter"].(bson.M)["tags"])
				require.Equal(tt, 1, len(rule3["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, rule3["filter"].(bson.M)["tag_ids"].(primitive.A), tagsMap["tenant-2"]["production"])

				rule4 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"priority": 2, "tenant_id": "tenant-2"}).Decode(&rule4))
				require.Nil(tt, rule4["filter"].(bson.M)["tags"])
				require.Equal(tt, 0, len(rule4["filter"].(bson.M)["tag_ids"].(primitive.A)))

				rule5 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"priority": 3, "tenant_id": "tenant-1"}).Decode(&rule5))
				require.Nil(tt, rule5["filter"].(bson.M)["tags"])
				require.Equal(tt, 0, len(rule5["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Equal(tt, ".*", rule5["filter"].(bson.M)["hostname"].(string))
			},
		},
		{
			description: "succeeds migrating public_keys filter.tags to filter.tag_ids",
			setup: func() error {
				devices := []bson.M{
					{
						"uid":       "device-1",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "backend"},
					},
					{
						"uid":       "device-2",
						"tenant_id": "tenant-1",
						"tags":      []string{"production", "frontend"},
					},
					{
						"uid":       "device-3",
						"tenant_id": "tenant-2",
						"tags":      []string{"production"},
					},
				}

				publicKeys := []bson.M{
					{
						"tenant_id":   "tenant-1",
						"fingerprint": "key-1",
						"filter": bson.M{
							"tags": []string{"production", "backend"},
						},
					},
					{
						"tenant_id":   "tenant-1",
						"fingerprint": "key-2",
						"filter": bson.M{
							"tags": []string{"production", "frontend"},
						},
					},
					{
						"tenant_id":   "tenant-2",
						"fingerprint": "key-3",
						"filter": bson.M{
							"tags": []string{"production"},
						},
					},
					{
						"tenant_id":   "tenant-2",
						"fingerprint": "key-4",
						"filter": bson.M{
							"tags": []string{},
						},
					},
					{
						"tenant_id":   "tenant-1",
						"fingerprint": "key-5",
						"filter": bson.M{
							"hostname": ".*",
						},
					},
				}

				if _, err := c.Database("test").Collection("devices").InsertMany(ctx, []any{devices[0], devices[1], devices[2]}); err != nil {
					return err
				}

				if _, err := c.Database("test").Collection("public_keys").InsertMany(ctx, []any{publicKeys[0], publicKeys[1], publicKeys[2], publicKeys[3], publicKeys[4]}); err != nil { // nolint:revive
					return err
				}

				return nil
			},
			verify: func(tt *testing.T) {
				cursor, err := c.Database("test").Collection("tags").Find(ctx, bson.M{})
				require.NoError(tt, err)

				var tags []bson.M
				require.NoError(tt, cursor.All(ctx, &tags))

				tagMap := make(map[string]map[string]primitive.ObjectID)
				for _, tag := range tags {
					tenantID := tag["tenant_id"].(string)
					name := tag["name"].(string)
					id := tag["_id"].(primitive.ObjectID)

					if tagMap[tenantID] == nil {
						tagMap[tenantID] = make(map[string]primitive.ObjectID)
					}
					tagMap[tenantID][name] = id
				}

				key1 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": "key-1"}).Decode(&key1))
				require.Nil(tt, key1["filter"].(bson.M)["tags"])
				require.Equal(tt, 2, len(key1["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, key1["filter"].(bson.M)["tag_ids"].(primitive.A), tagMap["tenant-1"]["production"])
				require.Contains(tt, key1["filter"].(bson.M)["tag_ids"].(primitive.A), tagMap["tenant-1"]["backend"])

				key2 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": "key-2"}).Decode(&key2))
				require.Nil(tt, key2["filter"].(bson.M)["tags"])
				require.Equal(tt, 2, len(key2["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, key2["filter"].(bson.M)["tag_ids"].(primitive.A), tagMap["tenant-1"]["production"])
				require.Contains(tt, key2["filter"].(bson.M)["tag_ids"].(primitive.A), tagMap["tenant-1"]["frontend"])

				key3 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": "key-3"}).Decode(&key3))
				require.Nil(tt, key3["filter"].(bson.M)["tags"])
				require.Equal(tt, 1, len(key3["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Contains(tt, key3["filter"].(bson.M)["tag_ids"].(primitive.A), tagMap["tenant-2"]["production"])

				key4 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": "key-4"}).Decode(&key4))
				require.Nil(tt, key4["filter"].(bson.M)["tags"])
				require.Equal(tt, 0, len(key4["filter"].(bson.M)["tag_ids"].(primitive.A)))

				key5 := bson.M{}
				require.NoError(tt, c.Database("test").Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": "key-5"}).Decode(&key5))
				require.Nil(tt, key5["filter"].(bson.M)["tags"])
				require.Equal(tt, 0, len(key5["filter"].(bson.M)["tag_ids"].(primitive.A)))
				require.Equal(tt, ".*", key5["filter"].(bson.M)["hostname"].(string))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() { require.NoError(tt, srv.Reset()) })

			require.NoError(tt, tc.setup())
			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[107])
			require.NoError(tt, migrates.Up(ctx, migrate.AllAvailable))
			tc.verify(tt)
		})
	}
}
