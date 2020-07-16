package firewall

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	req := Request{
		Hostname:  "host",
		Namespace: "namespace",
		Username:  "user",
		IPAddress: "127.0.0.1",
	}

	user := &models.User{TenantID: "tenant"}

	rules := []models.FirewallRule{
		models.FirewallRule{
			FirewallRuleFields: models.FirewallRuleFields{
				Priority: 1,
				Action:   "allow",
				Active:   true,
				SourceIP: ".*",
				Username: ".*",
				Hostname: ".*",
			},
		},
	}

	mock.On("GetUserByUsername", ctx, req.Namespace).
		Return(user, nil).Once()
	mock.On("ListFirewallRules", context.WithValue(ctx, "tenant", user.TenantID), paginator.Query{-1, -1}).
		Return(rules, len(rules), nil).Once()

	ok, err := s.Evaluate(ctx, req)
	assert.NoError(t, err)
	assert.True(t, ok)

	mock.AssertExpectations(t)
}

func TestCreateRule(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	rule := &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	}

	mock.On("CreateFirewallRule", ctx, rule).
		Return(nil).Once()

	err := s.CreateRule(ctx, rule)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestListRules(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	rules := []models.FirewallRule{
		models.FirewallRule{
			FirewallRuleFields: models.FirewallRuleFields{
				Priority: 1,
				Action:   "allow",
				Active:   true,
				SourceIP: ".*",
				Username: ".*",
				Hostname: ".*",
			},
		},
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	mock.On("ListFirewallRules", ctx, query).
		Return(rules, len(rules), nil).Once()

	returnedRules, count, err := s.ListRules(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, rules, returnedRules)
	assert.Equal(t, count, len(rules))

	mock.AssertExpectations(t)
}

func TestGetRule(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	rule := &models.FirewallRule{
		ID: "id",
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	}

	mock.On("GetFirewallRule", ctx, rule.ID).
		Return(rule, nil).Once()

	returnedRule, err := s.GetRule(ctx, rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, rule, returnedRule)

	mock.AssertExpectations(t)
}

func TestUpdateRule(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	rule := &models.FirewallRule{
		ID: "id",
	}

	updateRule := models.FirewallRuleUpdate{}

	mock.On("UpdateFirewallRule", ctx, rule.ID, updateRule).
		Return(rule, nil).Once()

	returnedRule, err := s.UpdateRule(ctx, rule.ID, updateRule)
	assert.NoError(t, err)
	assert.Equal(t, rule, returnedRule)

	mock.AssertExpectations(t)
}

func TestDeleteRule(t *testing.T) {
	mock := &mocks.Store{}

	s := NewService(store.Store(mock))

	ctx := context.TODO()

	rule := &models.FirewallRule{
		ID: "id",
	}

	mock.On("DeleteFirewallRule", ctx, rule.ID).
		Return(nil).Once()

	err := s.DeleteRule(ctx, rule.ID)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
