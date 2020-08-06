package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/firewall"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetFirewallRuleListURL = "/firewall/rules"
	GetFirewallRuleURL     = "/firewall/rules/:id"
	CreateFirewallRuleURL  = "/firewall/rules"
	UpdateFirewallRuleURL  = "/firewall/rules/:id"
	DeleteFirewallRuleURL  = "/firewall/rules/:id"
)

func GetFirewallRuleList(c apicontext.Context) error {
	svc := firewall.NewService(c.Store())

	query := paginator.NewQuery()
	c.Bind(query)

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	rules, count, err := svc.ListRules(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, rules)
}

func GetFirewallRule(c apicontext.Context) error {
	svc := firewall.NewService(c.Store())

	rule, err := svc.GetRule(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, rule)
}

func CreateFirewallRule(c apicontext.Context) error {
	svc := firewall.NewService(c.Store())

	var rule models.FirewallRule
	if err := c.Bind(&rule); err != nil {
		return err
	}

	if tenant := c.Tenant(); tenant != nil {
		rule.TenantID = tenant.ID
	}

	if err := svc.CreateRule(c.Ctx(), &rule); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, rule)
}

func UpdateFirewallRule(c apicontext.Context) error {
	svc := firewall.NewService(c.Store())

	var rule models.FirewallRuleUpdate
	if err := c.Bind(&rule); err != nil {
		return err
	}

	value, err := svc.UpdateRule(c.Ctx(), c.Param("id"), rule)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, value)
}

func DeleteFirewallRule(c apicontext.Context) error {
	svc := firewall.NewService(c.Store())

	if err := svc.DeleteRule(c.Ctx(), c.Param("id")); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
