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
	ErrFindPublicKey        = errors.New(fmt.Errorf("could not find the public key"), fmt.Errorf("it could not possible to get the public key from the server"))
	ErrEvaluatePublicKey    = errors.New(fmt.Errorf("could not be evaluate public key"), fmt.Errorf("it could not evaluate the public key in the server"))
	ErrForbiddenPublicKey   = errors.New(fmt.Errorf("could not be used to peform this action with public key"), fmt.Errorf("this public key could not be used to this action"))
	ErrDataPublicKey        = errors.New(fmt.Errorf("could not parse the public key data"), fmt.Errorf("it could not parse the public key data"))
	ErrSignaturePublicKey   = errors.New(fmt.Errorf("could not decode the public key signature"), fmt.Errorf("it could not decode the public key signature"))
	ErrVerifyPublicKey      = errors.New(fmt.Errorf("could not verify the public key"), fmt.Errorf("it could not verify the public key"))
	ErrSignerPublicKey      = errors.New(fmt.Errorf("could not signer the public key"), fmt.Errorf("it could not signer the public key"))
	ErrDialSSH              = errors.New(fmt.Errorf("could not dial to SSH"), fmt.Errorf("it could not dial to connect to SSH server"))
	ErrSession              = errors.New(fmt.Errorf("could not create a new session"), fmt.Errorf("it could not create the SSH session"))
	ErrEnvIPAddress         = errors.New(fmt.Errorf("could not set the env virable of ip address"), fmt.Errorf("it could not set the env virable of ip address to session"))
	ErrEnvWS                = errors.New(fmt.Errorf("could not set the env virable of web socket"), fmt.Errorf("it could not set the env virable of web socket to session"))
	ErrPipeStdin            = errors.New(fmt.Errorf("could not pipe session stdin"), fmt.Errorf("it could not pipe session stdin from client to agent"))
	ErrPipeStdout           = errors.New(fmt.Errorf("could not pipe session stdout"), fmt.Errorf("it could not pipe session stdout from client to agent"))
	ErrPty                  = errors.New(fmt.Errorf("could not request the pty"), fmt.Errorf("it could not request the pty from agent"))
	ErrShell                = errors.New(fmt.Errorf("could not get the shell"), fmt.Errorf("it could not get the shell from agent"))
	ErrHost                 = errors.New(fmt.Errorf("cloud not split the host into address and port"), fmt.Errorf("it could not get the device address"))
)
