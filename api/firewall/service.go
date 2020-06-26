package firewall

import (
	"context"
	"regexp"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Service interface {
	Evaluate(ctx context.Context, req Request) (bool, error)
	CreateRule(ctx context.Context, rule *models.FirewallRule) error
	ListRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error)
	GetRule(ctx context.Context, id string) (*models.FirewallRule, error)
	UpdateRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error)
	DeleteRule(ctx context.Context, id string) error
}

type Request struct {
	Hostname  string
	Namespace string
	Username  string
	IPAddress string
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) Evaluate(ctx context.Context, req Request) (bool, error) {
	user, err := s.store.GetUserByUsername(ctx, req.Namespace)
	if err != nil {
		return false, err
	}

	ctx = context.WithValue(ctx, "tenant", user.TenantID)

	rules, count, err := s.store.ListFirewallRules(ctx, paginator.Query{-1, -1})
	if err != nil {
		return false, err
	}

	if count == 0 {
		return true, nil
	}

	allow := true

	for _, rule := range rules {
		if !rule.Active {
			continue
		}

		ok, err := regexp.MatchString(rule.SourceIP, req.IPAddress)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		ok, _ = regexp.MatchString(rule.Username, req.Username)
		if err != nil {
			return false, err

		}
		if !ok {
			continue
		}

		ok, err = regexp.MatchString(rule.Hostname, req.Hostname)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		allow = rule.Action == "allow"
	}

	return allow, nil
}

func (s *service) CreateRule(ctx context.Context, rule *models.FirewallRule) error {
	return s.store.CreateFirewallRule(ctx, rule)
}

func (s *service) ListRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error) {
	return s.store.ListFirewallRules(ctx, pagination)
}

func (s *service) GetRule(ctx context.Context, id string) (*models.FirewallRule, error) {
	return s.store.GetFirewallRule(ctx, id)
}

func (s *service) UpdateRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error) {
	return s.store.UpdateFirewallRule(ctx, id, rule)
}

func (s *service) DeleteRule(ctx context.Context, id string) error {
	return s.store.DeleteFirewallRule(ctx, id)
}
