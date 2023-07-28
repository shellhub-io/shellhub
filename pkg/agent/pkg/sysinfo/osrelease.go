package sysinfo

import (
	"bufio"
	"os"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

var DefaultOSReleaseFilename = "/etc/os-release"

type OSRelease struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetOSRelease() (*OSRelease, error) {
	id, err := getValueFromOsRelease("ID")
	if err != nil {
		return nil, err
	}

	name, err := getValueFromOsRelease("PRETTY_NAME")
	if err != nil {
		return nil, err
	}

	if name == "" {
		name, err = getValueFromOsRelease("PRETTY_NAME")
		if err != nil {
			return nil, err
		}
	}

	return &OSRelease{id, name}, nil
}

func getValueFromOsRelease(key string) (string, error) {
	file, err := os.Open(DefaultOSReleaseFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	defer file.Close()

	var value string
	keyWithTrailingEqual := key + "="
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, keyWithTrailingEqual) {
			data := strings.SplitN(line, "=", 2)
			values, err := shellwords.Parse(data[1])
			if err != nil {
				return "", err
			}

			if len(values) != 1 {
				value = strings.Join(values, " ")
			} else {
				value = values[0]
			}
		}
	}

	return value, nil
}
