package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

func getValueFromOsRelease(key string) (string, error) {
	osReleaseFile, err := os.Open(etcOsRelease)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("Error opening %s: %v", etcOsRelease, err)
		}
	}
	defer osReleaseFile.Close()

	var value string
	keyWithTrailingEqual := key + "="
	scanner := bufio.NewScanner(osReleaseFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, keyWithTrailingEqual) {
			data := strings.SplitN(line, "=", 2)
			values, err := shellwords.Parse(data[1])
			if err != nil {
				return "", fmt.Errorf("%s is invalid: %s", key, err.Error())
			}
			if len(values) != 1 {
				return "", fmt.Errorf("%s needs to be enclosed by quotes if they have spaces: %s", key, data[1])
			}
			value = values[0]
		}
	}

	return value, nil
}
