package dockerutils

import (
	"io"
	"io/ioutil"
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
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompilePOSIX(`([0-9]+:[a-z_,=]+.*docker[/-]| (/@)?/var/lib/docker/containers/)([0-9a-f]{64})`)
	match := re.FindSubmatch(content)
	if match == nil || len(match) != 4 {
		return "", nil
	}

	return string(match[3]), nil
}

func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")

	return err == nil
}
