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
)

func (pg *pg) Save(ctx context.Context, model any) error {
	if err := setUpdatedAt(model); err != nil {
		log.Warnf("failed to set UpdatedAt: %v", err)
	}

	_, err := pg.driver.NewUpdate().Model(model).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *pg) Delete(ctx context.Context, model any) error {
	_, err := pg.driver.NewDelete().Model(model).WherePK().Exec(ctx)

	return fromSqlError(err)
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
