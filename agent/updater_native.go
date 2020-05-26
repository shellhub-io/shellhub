// +build !docker

package main

import (
	"github.com/Masterminds/semver"
)

type nativeUpdater struct {
}

func (n *nativeUpdater) CurrentVersion() (*semver.Version, error) {
	return semver.NewVersion(AgentVersion)
}

func (n *nativeUpdater) ApplyUpdate(_ *semver.Version) error {
	return nil
}

func (n *nativeUpdater) CompleteUpdate() error {
	return nil
}

func NewUpdater() (Updater, error) {
	return &nativeUpdater{}, nil
}
