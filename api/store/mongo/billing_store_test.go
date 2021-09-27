package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestBillingUpdateInstance(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	}
	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	subscription := &models.Billing{
		SubscriptionID:   "subc_1111x",
		CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		PriceID:          "pid_11x",
		Active:           true,
		State:            "pending",
	}

	err = mongostore.BillingUpdateInstance(ctx, ns, subscription)
	assert.NoError(t, err)

	ns, err = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, subscription, ns.Billing)
}

func TestBillingUpdatePaymentFailed(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	pf := &models.PaymentFailed{
		Status:  true,
		Details: "invalid",
		Date:    time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		Amount:  47.54,
	}

	ns := &models.Namespace{
		TenantID: "tenant",
	}

	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	_, err = mongostore.BillingUpdatePaymentFailed(ctx, "subs_id", true, pf)
	assert.Error(t, err)

	subsID := "subs_id_1"

	ns2 := &models.Namespace{
		TenantID: "tenant2",
		Billing: &models.Billing{
			SubscriptionID: subsID,
		},
	}

	_, err = mongostore.NamespaceCreate(ctx, ns2)
	assert.NoError(t, err)

	ns2, err = mongostore.BillingUpdatePaymentFailed(ctx, subsID, true, pf)
	assert.NoError(t, err)

	assert.Equal(t, pf, ns2.Billing.PaymentFailed)

	ns2, err = mongostore.BillingUpdatePaymentFailed(ctx, subsID, false, nil)
	assert.NoError(t, err)

	assert.Nil(t, ns2.Billing.PaymentFailed)
}

func TestBillingUpdateDeviceLimit(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: 3,
		Billing: &models.Billing{
			SubscriptionID:   "subc_1111x",
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
			PriceID:          "pid_11x",
		},
	}
	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	newDeviceLimit := -1

	_, err = mongostore.BillingUpdateDeviceLimit(ctx, ns.TenantID, newDeviceLimit)
	assert.NoError(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.Equal(t, ns.MaxDevices, newDeviceLimit)
}

func TestBillingDeleteSubscription(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	subsID := "subc_1111x"
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
		Billing: &models.Billing{
			SubscriptionID:   subsID,
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
			Active:           true,
			PriceID:          "pid_11x",
		},
	}

	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	err = mongostore.BillingDeleteSubscription(ctx, ns.TenantID)
	assert.NoError(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.Equal(t, false, ns.Billing.Active)
}

func TestBillingRemoveInstance(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	subsID := "sub_1x"
	billing := models.Billing{
		CustomerID:      "cust_111x",
		PaymentMethodID: "pid_111x",
		SubscriptionID:  subsID,
	}

	namespace := models.Namespace{
		TenantID: "tenant",
	}

	namespaceBill := models.Namespace{
		TenantID: namespace.TenantID,
		Billing:  &billing,
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, &namespaceBill)

	assert.NoError(t, err)

	_, _ = mongostore.NamespaceGet(ctx, namespace.TenantID)
	err = mongostore.BillingRemoveInstance(ctx, billing.CustomerID)
	assert.NoError(t, err)

	ns, _ := mongostore.NamespaceGet(ctx, namespace.TenantID)
	assert.Empty(t, ns.Billing)
	assert.Nil(t, ns.Billing)
}
