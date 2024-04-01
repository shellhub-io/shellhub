package migrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration57(t *testing.T) {
	logrus.Info("Testing Migration 57")

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
		SubItem          string         `json:"sub_item_id" bson:"sub_item_id,omitempty"`
	}

	type Namespace struct {
		TenantID string   `json:"tenant_id" bson:"tenant_id"`
		Billing  *Billing `json:"billing" bson:"billing,omitempty"`
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	cases := []struct {
		description   string
		requiredMocks func() (func() error, error)
		run           func() error
		check         func() (string, error)
		expected      string
	}{
		{
			description: "Success to apply up on migration 57 when namespace has billing",
			requiredMocks: func() (func() error, error) {
				_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), Namespace{
					TenantID: "tenant",
					Billing: &Billing{
						State: "processed",
					},
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					_, err := db.Client().Database("test").Collection("namespaces").DeleteOne(context.TODO(), bson.M{
						"tenant_id": "tenant",
					})
					if err != nil {
						return err
					}

					return nil
				}, nil
			},
			run: func() error {
				migrations := GenerateMigrations()[56:57]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				return nil
			},
			check: func() (string, error) {
				namespace := new(models.Namespace)
				err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "tenant",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				return string(namespace.Billing.Status), nil
			},
			expected: "active",
		},
		{
			description: "Success to apply up on migration 57 when namespace has no billing",
			requiredMocks: func() (func() error, error) {
				_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), Namespace{
					TenantID: "tenant",
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					_, err := db.Client().Database("test").Collection("namespaces").DeleteOne(context.TODO(), bson.M{
						"tenant_id": "tenant",
					})
					if err != nil {
						return err
					}

					return nil
				}, nil
			},
			run: func() error {
				migrations := GenerateMigrations()[56:57]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				return nil
			},
			check: func() (string, error) {
				namespace := new(models.Namespace)
				err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "tenant",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				if namespace.Billing != nil {
					return "", errors.New("billing should be nil")
				}

				return "", nil
			},
			expected: "",
		},
		{
			description: "Success to apply down on migration 57 when namespace has billing",
			requiredMocks: func() (func() error, error) {
				_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), &models.Namespace{
					TenantID: "tenant",
					Billing: &models.Billing{
						Status: "active",
					},
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					_, err := db.Client().Database("test").Collection("namespaces").DeleteOne(context.TODO(), bson.M{
						"tenant_id": "tenant",
					})
					if err != nil {
						return err
					}

					return nil
				}, nil
			},
			run: func() error {
				migrations := GenerateMigrations()[56:57]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				return nil
			},
			check: func() (string, error) {
				namespace := new(Namespace)
				err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "tenant",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				return namespace.Billing.State, nil
			},
			expected: "processed",
		},
		{
			description: "Success to apply down on migration 57 when namespace has no billing",
			requiredMocks: func() (func() error, error) {
				_, err := db.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), &models.Namespace{
					TenantID: "tenant",
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					_, err := db.Client().Database("test").Collection("namespaces").DeleteOne(context.TODO(), bson.M{
						"tenant_id": "tenant",
					})
					if err != nil {
						return err
					}

					return nil
				}, nil
			},
			run: func() error {
				migrations := GenerateMigrations()[56:57]
				migrates := migrate.NewMigrate(db.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return err
				}

				return nil
			},
			check: func() (string, error) {
				namespace := new(Namespace)
				err := db.Client().Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "tenant",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				if namespace.Billing != nil {
					return "", errors.New("billing should be nil")
				}

				return "", nil
			},
			expected: "",
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			teardown, err := tc.requiredMocks()
			assert.NoError(t, err)

			err = tc.run()
			assert.NoError(t, err)

			result, err := tc.check()
			assert.Equal(t, tc.expected, result)
			assert.NoError(t, err)

			err = teardown()
			assert.NoError(t, err)
		})
	}
}
