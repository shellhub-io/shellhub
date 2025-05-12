package pg

import (
	"database/sql"
	"io"

	"github.com/shellhub-io/shellhub/api/store"
)

func fromSqlError(err error) error {
	switch {
	case err == nil:
		return nil
	case err == sql.ErrNoRows, err == io.EOF:
		return store.ErrNoDocuments
	default:
		return err
	}
}
