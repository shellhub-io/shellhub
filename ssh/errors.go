package main

import (
	"fmt"

	"github.com/shellhub-io/shellhub/ssh/pkg/errors"
)

var (
	ErrInvalidSessionTarget = errors.New(fmt.Errorf("invalid session target"), fmt.Errorf("invalid session target"))
	ErrBillingBlock         = errors.New(fmt.Errorf("reached the device limit"), fmt.Errorf("you cannot connect to this device because the namespace is not eligible for the free plan.\\nPlease contact the namespace owner's to upgrade the plan.\\nSee our pricing plans on https://www.shellhub.io/pricing to estimate the cost of your use cases on ShellHub Cloud or go to https://cloud.shellhub.io/settings/billing to upgrade the plan"))
	ErrFirewallBlock        = errors.New(fmt.Errorf("a firewall rule block this action"), fmt.Errorf("a firewall rule block this action"))
	ErrFindDevice           = errors.New(fmt.Errorf("cloud not find the device"), fmt.Errorf("cloud not find the device"))
	ErrLookupDevice         = errors.New(fmt.Errorf("could not lookup for device data"), fmt.Errorf("could not lookup for device data"))
)
