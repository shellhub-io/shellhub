package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type FirewallStore interface {
	FirewallRuleList(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error)
	FirewallRuleCreate(ctx context.Context, rule *models.FirewallRule) error
	FirewallRuleGet(ctx context.Context, id string) (*models.FirewallRule, error)
	FirewallRuleUpdate(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error)
	FirewallRuleDelete(ctx context.Context, id string) error
}
