package environment

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

const envOverridePath = "../.env.override"

type envKey struct {
	name    string
	aliases []string
}

func (k envKey) spellings() []string {
	return append([]string{k.name}, k.aliases...)
}

func shellOrOverride(path string, keys ...envKey) (map[string]string, error) {
	override, err := godotenv.Read(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key.name] = firstSet(os.Getenv, key.spellings())
		if values[key.name] == "" {
			values[key.name] = firstSet(func(name string) string { return override[name] }, key.spellings())
		}
	}

	return values, nil
}

func firstSet(lookup func(string) string, names []string) string {
	for _, name := range names {
		if value := lookup(name); value != "" {
			return value
		}
	}

	return ""
}
