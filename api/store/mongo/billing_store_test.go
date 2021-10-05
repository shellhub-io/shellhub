package mongo

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestBillingUpdateInstance(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.BillingUpdateInstance(data.Context, &data.Namespace, &data.Subscription)
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceGet(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, &models.Billing{
		SubscriptionID:   "subc_1111x",
		CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		PriceID:          "pid_11x",
		Active:           true,
		State:            "pending",
	}, ns.Billing)
}

func TestBillingUpdatePaymentFailed(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	pf := &models.PaymentFailed{
		Status:  true,
		Details: "invalid",
		Date:    time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		Amount:  47.54,
	}

	_, err = mongostore.BillingUpdatePaymentFailed(data.Context, "subs_id", true, pf)
	assert.Error(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &models.Namespace{
		TenantID: "tenant2",
		Billing: &models.Billing{
			SubscriptionID: "subs_id_1",
		},
	})
	assert.NoError(t, err)

	ns2, err := mongostore.BillingUpdatePaymentFailed(data.Context, "subs_id_1", true, pf)
	assert.NoError(t, err)
	assert.Equal(t, pf, ns2.Billing.PaymentFailed)

	ns2, err = mongostore.BillingUpdatePaymentFailed(data.Context, "subs_id_1", false, nil)
	assert.NoError(t, err)
	assert.Nil(t, ns2.Billing.PaymentFailed)
}

func TestBillingUpdateDeviceLimit(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	_, err = mongostore.BillingUpdateDeviceLimit(data.Context, data.Namespace.TenantID, -1)
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceGet(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, ns.MaxDevices, -1)
}

func TestBillingDeleteSubscription(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.BillingDeleteSubscription(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceGet(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, false, ns.Billing.Active)
}

func TestBillingRemoveInstance(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(data.Context, &models.Namespace{
		TenantID: data.Namespace.TenantID,
		Billing: &models.Billing{
			CustomerID:      "cust_111x",
			PaymentMethodID: "pid_111x",
			SubscriptionID:  "sub_1x",
		},
	})
	assert.NoError(t, err)

	_, err = mongostore.NamespaceGet(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)

	err = mongostore.BillingRemoveInstance(data.Context, "cust_111x")
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceGet(data.Context, data.Namespace.TenantID)
	assert.NoError(t, err)
	assert.Empty(t, ns.Billing)
	assert.Nil(t, ns.Billing)
}
