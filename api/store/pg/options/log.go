package options

import (
	"context"
	"os"

	"github.com/oiime/logrusbun"
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

		db.AddQueryHook(logrusbun.NewQueryHook(
			logrusbun.WithEnabled(true),
			logrusbun.WithVerbose(verbose),
			logrusbun.WithQueryHookOptions(logrusbun.QueryHookOptions{
				Logger:     logger,
				QueryLevel: logrus.DebugLevel,
				ErrorLevel: logrus.ErrorLevel,
				SlowLevel:  logrus.WarnLevel,
			}),
		))

		return nil
	}
}
