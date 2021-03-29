package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type FirewallStore interface {
	CreateFirewallRule(ctx context.Context, rule *models.FirewallRule) error
	ListFirewallRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error)
	GetFirewallRule(ctx context.Context, id string) (*models.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, id string) error
}
