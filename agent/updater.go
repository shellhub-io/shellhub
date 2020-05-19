package main

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/parnurzeal/gorequest"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Updater interface {
	CurrentVersion() (*semver.Version, error)
	ApplyUpdate(v *semver.Version) error
	CompleteUpdate() error
}

func CheckUpdate(server string) (*semver.Version, error) {
	info := models.Info{}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("%s/info", server)).EndStruct(&info)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return semver.NewVersion(info.Version)
}
