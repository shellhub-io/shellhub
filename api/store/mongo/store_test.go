package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAddDevice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err := mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
}

func TestCreateFirewallRule(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
}

func TestListFirewallRules(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)

	rules, count, err := mongostore.ListFirewallRules(ctx, paginator.Query{-1, -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)
}
