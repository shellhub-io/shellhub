package dockerutils

import (
	"io"
	"os"
	"regexp"
)

// CurrentContainerID returns the current running container ID.
func CurrentContainerID() (string, error) {
	fMountInfo, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer fMountInfo.Close()

	return parseContainerIDv2(fMountInfo)
}

func parseContainerIDv2(rd io.Reader) (string, error) {
	data, err := io.ReadAll(rd)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\d+\s\d+\s\d+:\d+\s.+containers/?.+([0-9a-f]{64})/hostname`)
	match := re.FindSubmatch(data)
	if match == nil || len(match) != 2 {
		return "", nil
	}

	return string(match[1]), nil
}

func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")

	return err == nil
}
