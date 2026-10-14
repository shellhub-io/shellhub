package options

import (
	"context"
	"os"

	"github.com/shellhub-io/shellhub/server/api/store/pg/options/internal"
	"github.com/sirupsen/logrus"
)

// Log installs query logging at the given logrus level.
func Log(level string, verbose bool) Option {
	return func(_ context.Context, store *Store) error {
		level, err := logrus.ParseLevel(level)
		if err != nil {
			return err
		}

		logger := &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     level,
		}

		*store.DB = *store.DB.WithQueryHook(internal.NewQueryHook(
			internal.WithEnabled(true),
			internal.WithVerbose(verbose),
			internal.WithQueryHookOptions(internal.QueryHookOptions{
				Logger:     logger,
				QueryLevel: logrus.DebugLevel,
				ErrorLevel: logrus.ErrorLevel,
			}),
		))

		return nil
	}
}
