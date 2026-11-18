package services

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithHooksKeepsTheHooksOfEveryEvent(t *testing.T) {
	oneOfEach := func() Hooks {
		var hooks Hooks
		v := reflect.ValueOf(&hooks).Elem()
		for _, field := range v.Fields() {
			field.Set(reflect.MakeSlice(field.Type(), 1, 1))
		}

		return hooks
	}

	svc := &APIService{service: &service{}}
	WithHooks(oneOfEach())(svc)
	WithHooks(oneOfEach())(svc)

	got := reflect.ValueOf(svc.hooks)
	for i := range got.NumField() {
		assert.Equal(t, 2, got.Field(i).Len(), "Hooks.%s", got.Type().Field(i).Name)
	}
}
