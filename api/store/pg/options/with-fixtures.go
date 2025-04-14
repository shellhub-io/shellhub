package options

import (
	"context"
	"io/fs"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
)

func WithFixtures(dir string) Option {
	return func(ctx context.Context, db *bun.DB) error {
		println("path: ", dir)

		fixture := dbfixture.New(db)

		fsys := os.DirFS(dir)
		files, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return err
		}

		names := make([]string, 0)
		for _, file := range files {
			println("\tname: ", file.Name)

			if !file.IsDir() {
				names = append(names, file.Name())
			}
		}

		return fixture.Load(ctx, fsys, names...)
	}
}
