package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDeleteCodes(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		username    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when trying to delete codes",
			username:    "username",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}

			doc := bson.M{
				"username": tc.username,
				"codes":    []string{"code1", "code2"}, // initialize with some sample codes
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.DeleteCodes(ctx, tc.username)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestAddStatusMFA(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		username    string
		statusMFA   bool
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when trying to add status MFA",
			username:    "username",
			statusMFA:   true,
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}

			doc := bson.M{
				"username":   tc.username,
				"status_mfa": false, // initialize with status MFA as false
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.AddStatusMFA(ctx, tc.username, tc.statusMFA)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestAddSecret(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		username    string
		secret      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when trying to add status MFA",
			username:    "username",
			secret:      "IOJDSFIAWMKXskdlmawOSDMCALWC",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}

			doc := bson.M{
				"username": tc.username,
				"secret":   "",
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.AddSecret(ctx, tc.username, tc.secret)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	type Expected struct {
		err error
	}
	cases := []struct {
		description string
		username    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success to delete a secret MFA",
			username:    "username",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}

			doc := bson.M{
				"username": tc.username,
				"secret":   "test_secret",
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.DeleteSecret(ctx, tc.username)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestGetStatusMFA(t *testing.T) {
	type Expected struct {
		status bool
		err    error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when getting status MFA",
			id:          "605aa8e1b5454f001e6c27c8",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{status: false, err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			id, err := primitive.ObjectIDFromHex(tc.id)
			assert.NoError(t, err)

			doc := bson.M{
				"_id": id,
				"MFA": false,
			}
			if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			status, err := mongostore.GetStatusMFA(ctx, tc.id)
			assert.Equal(t, tc.expected.status, status)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestGetSecret(t *testing.T) {
	type Expected struct {
		secret string
		err    error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when getting secret",
			id:          "605aa8e1b5454f001e6c27c8",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{secret: "test_secret", err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			id, err := primitive.ObjectIDFromHex(tc.id)
			assert.NoError(t, err)

			doc := bson.M{
				"_id":    id,
				"secret": "test_secret",
			}
			if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			secret, err := mongostore.GetSecret(ctx, tc.id)
			assert.Equal(t, tc.expected.secret, secret)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestAddCodes(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		username    string
		codes       []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when adding codes",
			username:    "username",
			codes:       []string{"code1", "code2"},
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown()

			doc := bson.M{
				"username": tc.username,
				"codes":    []string{"existing_code"},
			}
			if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.AddCodes(ctx, tc.username, tc.codes)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestGetCodes(t *testing.T) {
	type Expected struct {
		codes []string
		err   error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "success when getting codes",
			id:          "605aa8e1b5454f001e6c27c8",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    Expected{codes: []string{"code1", "code2"}, err: nil},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("users")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown()

			id, err := primitive.ObjectIDFromHex(tc.id)
			assert.NoError(t, err)

			doc := bson.M{
				"_id":   id,
				"codes": []string{"code1", "code2"},
			}
			if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			codes, err := mongostore.GetCodes(ctx, tc.id)
			assert.Equal(t, tc.expected.codes, codes)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}
