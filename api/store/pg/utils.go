package pg

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func (pg *pg) Save(ctx context.Context, models ...any) error {
	for _, m := range models {
		if err := setUpdatedAt(m); err != nil {
			log.Warnf("failed to set UpdatedAt: %v", err)
		}

		if _, err := pg.driver.NewUpdate().Model(ptr(m)).WherePK().Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func (pg *pg) Delete(ctx context.Context, models ...any) error {
	for _, m := range models {
		if _, err := pg.driver.NewDelete().Model(ptr(m)).WherePK().Exec(ctx); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

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

func applyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) error {
	ctxWithQuery := context.WithValue(ctx, "query", query)
	for _, opt := range opts {
		if err := opt(ctxWithQuery); err != nil {
			return fromSqlError(err)
		}
	}

	return nil
}

func setUpdatedAt(model any) error {
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	updatedField := value.FieldByName("UpdatedAt")
	if updatedField.IsValid() && updatedField.CanSet() {
		if updatedField.Type() == reflect.TypeOf(time.Time{}) {
			updatedField.Set(reflect.ValueOf(time.Now().UTC()))
			return nil
		}
		return fmt.Errorf("UpdatedAt field is not of type time.Time")
	}

	return nil
}

func ptr(arg any) any {
	argPtr := arg
	if val := reflect.ValueOf(argPtr); val.Kind() != reflect.Ptr {
		newPtr := reflect.New(val.Type())
		newPtr.Elem().Set(val)
		argPtr = newPtr.Interface()
	}

	return argPtr
}
