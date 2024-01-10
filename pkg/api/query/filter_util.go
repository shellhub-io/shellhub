package query

const (
	FilterTypeProperty = "property"
	FilterTypeOperator = "operator"
)

// FilterProperty is a JSON representation of a property expression in a query.
//
// Name is the attribute to be observed in the operation, Operator is the operation, and
// Value is the value used in the operation. While Name can be any string, the Operator must be supported
// by the implementation, and the Value is the value used in the operation.
//
// Each operator has its own implementation, and one operator can have multiple implementations. For that reason,
// the operator must be converted to a useful value using build methods.
//
// Examples:
// A FilterProperty with Operator "gt", Name "count", and Value 12 will filter documents with the attribute "count" greater than 12.
// Another FilterProperty with Operator "eq", Name "alias", and Value "foobar" will filter documents with the attribute "alias" equal to "foobar".
type FilterProperty struct {
	// Name is the attribute to be observed in the operation.
	Name string `json:"name"`

	// Operator is the operation (e.g., "eq" for equal).
	Operator string `json:"operator"`

	// Value is the value used in the operation. (e.g., "eq" operations use Value to determine the value to be equal).
	Value interface{} `json:"value"`
}

// FilterOperator represents a JSON representation of a filter operator in a query (e.g., "and", "or" in MongoDB queries).
type FilterOperator struct {
	// Name is the filter operator (e.g., "and", "or").
	Name string `json:"name"`
}
