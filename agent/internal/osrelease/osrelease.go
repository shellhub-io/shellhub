package osrelease

import (
	"bufio"
	"os"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

func GetValue(key string) (string, error) {
	file, err := os.Open(osRelease)
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
