package environment

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

const envOverridePath = "../.env.override"

func shellOrOverride(path string, names ...string) (map[string]string, error) {
	override, err := godotenv.Read(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	values := make(map[string]string, len(names))
	for _, name := range names {
		values[name] = os.Getenv(name)
		if values[name] == "" {
			values[name] = override[name]
		}
	}

	return values, nil
}
