package migrator

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Func is the up of a Go migration. It runs on the pool, outside any transaction, and its
// migration is recorded only after it returns nil.
type Func func(ctx context.Context, db *bun.DB) error

// Migration is one step of a set, parsed from an .up.sql file by [Load] or built by [Go].
type Migration struct {
	Name    string
	Comment string

	transactional bool
	chunks        []string
	up            Func
}

func (m Migration) String() string {
	return m.Name + "_" + m.Comment
}

// Go returns a Go migration named name, the number a file of its own would carry.
func Go(name, comment string, up Func) Migration {
	return Migration{Name: name, Comment: comment, up: up}
}

// Tables names the table that records the applied migrations of a set and the lock table bun's
// Init creates beside it.
type Tables struct {
	Migrations string
	Locks      string
}

var fileName = regexp.MustCompile(`^(\d{1,14})_([0-9a-z_\-]+)\.`)

// Load parses every .up.sql file at the root of sqlFS and returns those migrations together with
// goMigrations, sorted by name. A file is transactional when its name ends in .tx.up.sql, and
// --bun:split separates its chunks. It fails on a file name bun would not accept, on any other
// --bun: directive, on a Go migration without a function and on two migrations sharing a name.
func Load(sqlFS fs.FS, goMigrations ...Migration) ([]Migration, error) {
	files, err := fs.Glob(sqlFS, "*.up.sql")
	if err != nil {
		return nil, err
	}

	for _, m := range goMigrations {
		if m.up == nil {
			return nil, fmt.Errorf("go migration %s has no function", m)
		}
	}

	loaded := slices.Clone(goMigrations)

	for _, file := range files {
		matches := fileName.FindStringSubmatch(file)
		if matches == nil {
			return nil, fmt.Errorf("unsupported migration file name %q", file)
		}

		raw, err := fs.ReadFile(sqlFS, file)
		if err != nil {
			return nil, err
		}

		chunks, err := split(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}

		loaded = append(loaded, Migration{
			Name:          matches[1],
			Comment:       matches[2],
			transactional: strings.HasSuffix(file, ".tx.up.sql"),
			chunks:        chunks,
		})
	}

	slices.SortFunc(loaded, func(a, b Migration) int { return strings.Compare(a.Name, b.Name) })

	for i := 1; i < len(loaded); i++ {
		if loaded[i].Name == loaded[i-1].Name {
			return nil, fmt.Errorf("migrations %s and %s share a name", loaded[i-1], loaded[i])
		}
	}

	return loaded, nil
}

func split(raw []byte) ([]string, error) {
	const directive = "--bun:"

	var chunks []string

	var chunk []byte

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Bytes()

		if name, ok := bytes.CutPrefix(line, []byte(directive)); ok {
			if !bytes.Equal(name, []byte("split")) {
				return nil, fmt.Errorf("unknown directive %q", line)
			}

			chunks = append(chunks, string(chunk))
			chunk = chunk[:0]

			continue
		}

		chunk = append(chunk, line...)
		chunk = append(chunk, '\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(chunk) > 0 {
		chunks = append(chunks, string(chunk))
	}

	return slices.DeleteFunc(chunks, func(c string) bool { return strings.TrimSpace(c) == "" }), nil
}

// Apply runs, in name order, every migration of migrations not yet recorded in tables.Migrations,
// and returns how many it ran. It records each one as a row of bun's [migrate.Migration], all of
// them under one new group id, so bun's migrator reads and rolls back what Apply writes. A
// transactional migration is recorded in the transaction that runs its chunks, so either both
// commit or neither does. Any other migration runs first and is recorded after, so it must be safe
// to run again. Apply stops at the first migration that fails, with an error naming it; the
// migrations before it stay applied and recorded. It does not serialize concurrent callers.
func Apply(ctx context.Context, db *bun.DB, tables Tables, migrations []Migration) (int, error) {
	registry := migrate.NewMigrations()
	for _, m := range migrations {
		registry.Add(migrate.Migration{Name: m.Name, Comment: m.Comment})
	}

	migrator := migrate.NewMigrator(db, registry,
		migrate.WithTableName(tables.Migrations),
		migrate.WithLocksTableName(tables.Locks),
	)

	if err := migrator.Init(ctx); err != nil {
		return 0, err
	}

	applied, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return 0, err
	}

	recorded := make(map[string]bool, len(applied))
	for _, m := range applied {
		recorded[m.Name] = true
	}

	group := applied.LastGroupID() + 1
	ran := 0

	for _, m := range migrations {
		if recorded[m.Name] {
			continue
		}

		if err := apply(ctx, db, tables.Migrations, group, m); err != nil {
			return ran, fmt.Errorf("migration %s: %w", m, err)
		}

		ran++
	}

	return ran, nil
}

func apply(ctx context.Context, db *bun.DB, table string, group int64, m Migration) error {
	switch {
	case m.up != nil:
		if err := m.up(ctx, db); err != nil {
			return err
		}

		return record(ctx, db, table, group, m)
	case m.transactional:
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if err := exec(ctx, tx, m.chunks); err != nil {
				return err
			}

			return record(ctx, tx, table, group, m)
		})
	default:
		conn, err := db.Conn(ctx)
		if err != nil {
			return err
		}

		err = exec(ctx, conn, m.chunks)
		if closeErr := conn.Close(); err == nil {
			err = closeErr
		}

		if err != nil {
			return err
		}

		return record(ctx, db, table, group, m)
	}
}

func exec(ctx context.Context, db bun.IConn, chunks []string) error {
	for _, chunk := range chunks {
		if _, err := db.ExecContext(ctx, chunk); err != nil {
			return err
		}
	}

	return nil
}

func record(ctx context.Context, db bun.IDB, table string, group int64, m Migration) error {
	_, err := db.NewInsert().
		Model(&migrate.Migration{Name: m.Name, GroupID: group}).
		ModelTableExpr(table).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("record as applied: %w", err)
	}

	return nil
}
