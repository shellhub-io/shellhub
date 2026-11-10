package query

import (
	"net/url"
	"slices"
	"strings"
)

const filtersParamPrefix = "filters["

// UnmarshalQuery fills Data from the bracket filter parameters in params, each one shaped
// filters[<field>][<operator>]=<value>, and ignores every other parameter. The properties
// combine with AND and come out in key order. It does not check the field or the operator
// against anything: the endpoint does that with [ValidateFilters]. It returns
// [ErrFilterInvalid] for a key missing its field or its operator, with a bracket inside
// either, carrying anything after them, or sent more than once, and leaves Data untouched
// then.
func (fs *Filters) UnmarshalQuery(params url.Values) error {
	keys := make([]string, 0, len(params))
	for key := range params {
		if strings.HasPrefix(key, filtersParamPrefix) {
			keys = append(keys, key)
		}
	}

	if len(keys) == 0 {
		return nil
	}

	slices.Sort(keys)

	data := make([]Filter, 0, len(keys)+1)
	data = append(data, Filter{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}})

	for _, key := range keys {
		field, operator, ok := parseFiltersKey(key)
		if !ok || len(params[key]) != 1 {
			return ErrFilterInvalid
		}

		data = append(data, Filter{
			Type:   FilterTypeProperty,
			Params: &FilterProperty{Name: field, Operator: operator, Value: params[key][0]},
		})
	}

	fs.Data = data

	return nil
}

func parseFiltersKey(key string) (string, string, bool) {
	rest := strings.TrimPrefix(key, filtersParamPrefix)

	field, rest, ok := strings.Cut(rest, "][")
	if !ok || field == "" {
		return "", "", false
	}

	operator, rest, ok := strings.Cut(rest, "]")
	if !ok || operator == "" || rest != "" || strings.ContainsAny(operator, "[]") || strings.ContainsAny(field, "[]") {
		return "", "", false
	}

	return field, operator, true
}
