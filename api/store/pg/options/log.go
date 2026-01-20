package options

import (
	"context"
	"os"

	"github.com/shellhub-io/shellhub/api/store/pg/options/internal"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func Log(level string, verbose bool) Option {
	return func(ctx context.Context, db *bun.DB) error {
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

		db.AddQueryHook(internal.NewQueryHook(
			internal.WithEnabled(true),
			internal.WithVerbose(verbose),
			internal.WithQueryHookOptions(internal.QueryHookOptions{
				Logger:     logger,
				QueryLevel: logrus.DebugLevel,
				ErrorLevel: logrus.ErrorLevel,
				SlowLevel:  logrus.WarnLevel,
			}),
		))

		return nil
	}
}
