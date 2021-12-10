package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration39(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	logrus.Info("Testing Migration 39 - Test if the token field was created on namespaces collection")

	type PaymentFailed struct {
		Status  bool      `json:"status" bson:"status,omitempty"`
		Amount  float64   `json:"amount" bson:"amount,omitempty"`
		Date    time.Time `json:"date" bson:"date,omitempty"`
		Details string    `json:"details" bson:"details,omitempty"`
	}

	type Billing struct {
		SubscriptionID   string         `json:"subscription_id" bson:"subscription_id,omitempty"`
		CurrentPeriodEnd time.Time      `json:"current_period_end" bson:"current_period_end,omitempty"`
		PriceID          string         `json:"price_id" bson:"price_id,omitempty"`
		CustomerID       string         `json:"customer_id" bson:"customer_id,omitempty"`
		PaymentMethodID  string         `json:"payment_method_id" bson:"payment_method_id,omitempty"`
		PaymentFailed    *PaymentFailed `json:"payment_failed" bson:"payment_failed,omitempty"`
		State            string         `json:"state" bson:"state,omitempty"`
		Active           bool           `json:"active" bson:"active,omitempty"`
	}

	type NamespaceSettings struct {
		SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
	}

	type Member struct {
		ID       string `json:"id,omitempty" bson:"id,omitempty"`
		Username string `json:"username,omitempty" bson:"username,omitempty" validate:"min=3,max=30,alphanum,ascii"`
		Type     string `json:"type" bson:"type" validate:"required,oneof=administrator operator observer"`
	}

	type Namespace struct {
		Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
		Owner        string             `json:"owner"`
		TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
		Members      []Member           `json:"members" bson:"members"`
		Settings     *NamespaceSettings `json:"settings"`
		Devices      int                `json:"-" bson:"devices,omitempty"`
		Sessions     int                `json:"-" bson:"sessions,omitempty"`
		MaxDevices   int                `json:"max_devices" bson:"max_devices"`
		DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
		CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
		Billing      *Billing           `json:"billing" bson:"billing,omitempty"`
	}

	namespaceWithoutToken := Namespace{
		Name:     "namespaceWithoutToken",
		Owner:    "owner",
		TenantID: "tenantWithoutToken",
	}

	namespaceWithToken := models.Namespace{
		Name:     "namespaceWithToken",
		Owner:    "owner",
		TenantID: "tenantWithToken",
		Tokens:   []models.Token{},
	}
	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespaceWithoutToken)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespaceWithToken)
	assert.NoError(t, err)

	migrations := GenerateMigrations()[38:39]
	migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			description: "Exec migration UP when namespace does not have token's field",
			Test: func(t *testing.T) {
				t.Helper()

				err = migrates.Up(migrate.AllAvailable)
				assert.NoError(t, err)

				var namespaceMigrated models.Namespace
				err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.D{{"tenant_id", namespaceWithoutToken.TenantID}}).Decode(&namespaceMigrated)
				assert.NoError(t, err)
				assert.Equal(t,
					models.Namespace{
						Name:     "namespaceWithoutToken",
						Owner:    "owner",
						TenantID: "tenantWithoutToken",
						Tokens:   []models.Token{},
					},
					namespaceMigrated)
			},
		},
		{
			description: "Exec migration DOWN when namespace has token's field",
			Test: func(t *testing.T) {
				t.Helper()

				err = migrates.Down(migrate.AllAvailable)
				assert.NoError(t, err)

				var namespaceMigrated Namespace
				err = db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.D{{"tenant_id", namespaceWithToken.TenantID}}).Decode(&namespaceMigrated)
				assert.NoError(t, err)
				assert.Equal(t,
					Namespace{
						Name:     "namespaceWithToken",
						Owner:    "owner",
						TenantID: "tenantWithToken",
					}, namespaceMigrated)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, tc.Test)
	}
}
