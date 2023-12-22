package dockerutils

import (
	"io"
	"os"
	"regexp"
)

// CurrentContainerID returns the current running container ID.
func CurrentContainerID() (string, error) {
	fCgroup, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer fCgroup.Close()

	fMountInfo, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer fMountInfo.Close()

	reader := io.MultiReader(fCgroup, fMountInfo)
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\d+\s\d+\s\d+:\d+\s/var/.+docker/?.+([0-9a-f]{64})/`)
	match := re.FindSubmatch(content)
	if match == nil || len(match) != 2 {
		reg_hostname := regexp.MustCompile(`\d+\s\d+\s\d+:\d+\s.+containers/?.+([0-9a-f]{64})/hostname`)
		match = reg_hostname.FindSubmatch(content)
		if match == nil || len(match) != 2 {
			return "", nil
		}
	}

	return string(match[1]), nil
}

func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")

	return err == nil
}
