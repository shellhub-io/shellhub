package dockerutils

import (
	"io/ioutil"
	"os"
	"regexp"
)

// CurrentContainerID returns the current running container ID
func CurrentContainerID() (string, error) {
	const idLength = 64

	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompilePOSIX(`[0-9]+:[a-z_,=]+.*docker[/-]([0-9a-f]{64})`)
	line := re.FindAllSubmatch(content, -1)
	if len(line) <= 0 || len(line[0]) != 2 {
		return "", nil
	}

	return string(line[0][1]), nil
}
