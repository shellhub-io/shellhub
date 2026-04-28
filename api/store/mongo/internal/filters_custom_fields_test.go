package internal

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestParseCustomFieldsFilter(t *testing.T) {
	cases := []struct {
		description string
		fp          *query.FilterProperty
		wantOk      bool
		wantErr     bool
		checkResult func(t *testing.T, result bson.M)
	}{
		{
			description: "returns not-ok for unsupported operator eq",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "eq", Value: "prod"},
			wantOk:      false,
			wantErr:     false,
			checkResult: func(t *testing.T, result bson.M) {
				assert.Nil(t, result)
			},
		},
		{
			description: "returns error when value is not a string",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: 42},
			wantOk:      false,
			wantErr:     true,
			checkResult: func(t *testing.T, result bson.M) {
				assert.Nil(t, result)
			},
		},
		{
			description: "returns $expr condition for contains with string value",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: "production"},
			wantOk:      true,
			wantErr:     false,
			checkResult: func(t *testing.T, result bson.M) {
				require.NotNil(t, result)
				// Top-level key must be $expr
				exprRaw, ok := result["$expr"]
				require.True(t, ok, "result must have $expr key")

				expr, ok := exprRaw.(bson.M)
				require.True(t, ok)

				// $expr.$gt must exist
				gtRaw, ok := expr["$gt"]
				require.True(t, ok, "$expr must have $gt")

				gt, ok := gtRaw.(bson.A)
				require.True(t, ok)
				require.Len(t, gt, 2)

				// Second element of $gt must be 0 (threshold)
				assert.Equal(t, 0, gt[1])

				// First element is the $size expression
				sizeExpr, ok := gt[0].(bson.M)
				require.True(t, ok)
				_, hasSz := sizeExpr["$size"]
				assert.True(t, hasSz, "$gt[0] must be a $size expression")
			},
		},
		{
			description: "regex contains the search value",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: "team-a"},
			wantOk:      true,
			wantErr:     false,
			checkResult: func(t *testing.T, result bson.M) {
				require.NotNil(t, result)
				// Walk down to the $regexMatch input
				expr := result["$expr"].(bson.M)
				gt := expr["$gt"].(bson.A)
				sizeExpr := gt[0].(bson.M)
				filterExpr := sizeExpr["$size"].(bson.M)
				filterMap := filterExpr["$filter"].(bson.M)
				cond := filterMap["cond"].(bson.M)
				regexMatch := cond["$regexMatch"].(bson.M)

				assert.Equal(t, "team-a", regexMatch["regex"])
				assert.Equal(t, "i", regexMatch["options"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			result, ok, err := ParseCustomFieldsFilter(tc.fp)
			assert.Equal(t, tc.wantOk, ok)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			tc.checkResult(t, result)
		})
	}
}
