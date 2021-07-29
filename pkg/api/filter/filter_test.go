package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterTypeProperty(t *testing.T) {
	f := Filter{Type: "property", Params: &FilterTypeProperty{Name: "name", Operator: "contains", Value: "value"}}
	assert.NoError(t, f.IsValid())
	f = Filter{Type: "property", Params: &FilterTypeProperty{Name: "name", Operator: "eq", Value: "value"}}
	assert.NoError(t, f.IsValid())
	f = Filter{Type: "property", Params: &FilterTypeProperty{Name: "name", Operator: "bool", Value: "value"}}
	assert.NoError(t, f.IsValid())
	f = Filter{Type: "property", Params: &FilterTypeProperty{Name: "name", Operator: "gt", Value: "value"}}
	assert.NoError(t, f.IsValid())
	f = Filter{Type: "property", Params: &FilterTypeProperty{Name: "name", Operator: "lt", Value: "value"}}
	assert.NoError(t, f.IsValid())
}

func TestFilterTypeOperator(t *testing.T) {
	f := Filter{Type: "operator", Params: &FilterTypeOperator{Name: "and"}}
	assert.NoError(t, f.IsValid())

	f = Filter{Type: "operator", Params: &FilterTypeOperator{Name: "or"}}
	assert.NoError(t, f.IsValid())
}
