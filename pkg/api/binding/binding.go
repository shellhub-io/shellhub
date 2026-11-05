// Package binding states where each field of an API request comes from, and refuses a request
// type that leaves the answer open.
//
// A field reads from exactly one place. A field the body fills names its key with `json:"key"`
// and carries no other source tag. Every other field declares `json:"-"`, which keeps encoding/json
// from filling it by its Go name, and names at most one of the sources in [Sources]: none means the
// handler fills it. An embedded struct with no json tag contributes its fields as if they were
// declared in place, as encoding/json promotes them. A struct the body fills holds only body
// fields, since echo also walks into it for the path, query and headers.
package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
)

// Sources are the struct tags that name a non-body source of a request field.
var Sources = []string{"param", "query", "header", "identity", "form"}

// IdentityClaims are the values an identity tag may name: the caller's user ID, tenant, device UID,
// role, whether they are an instance administrator, and the administrator's ID.
var IdentityClaims = []string{"id", "tenant", "device", "role", "admin", "admin_id"}

// Source is where one request field reads its value from.
type Source struct {
	// Tag is "json" for the body, one of [Sources], or empty for a field the handler fills.
	Tag string
	// Name is the body key, path parameter, query parameter, header, identity claim or form field
	// the field reads.
	Name string
}

// IsBody reports whether the request body fills the field.
func (s Source) IsBody() bool {
	return s.Tag == "json"
}

// NamedSources returns every one of [Sources] the tag names with a non-empty value, in the order of
// [Sources]. A tag naming none of them returns nil.
func NamedSources(tag reflect.StructTag) []Source {
	var named []Source
	for _, name := range Sources {
		if value, ok := tag.Lookup(name); ok && value != "" {
			named = append(named, Source{Tag: name, Name: value})
		}
	}

	return named
}

// SourceOf reads the source a field declares through its tag. It returns an error when the tag
// names no json key and lacks `json:"-"`, which leaves encoding/json free to fill the field from a
// body key matching its Go name; when it names the body and another source; when it names more
// than one of [Sources]; and when an identity tag names none of [IdentityClaims].
func SourceOf(tag reflect.StructTag) (Source, error) {
	others := NamedSources(tag)

	jsonTag, hasJSON := tag.Lookup("json")
	if !hasJSON {
		return Source{}, errors.New("no json tag: declare json:\"-\" unless the body fills it")
	}

	key, _, _ := strings.Cut(jsonTag, ",")
	if jsonTag != "-" {
		if key == "" {
			return Source{}, fmt.Errorf("json tag %q names no key", jsonTag)
		}

		if len(others) > 0 {
			return Source{}, fmt.Errorf("read from both the body and %s", others[0].Tag)
		}

		return Source{Tag: "json", Name: key}, nil
	}

	switch len(others) {
	case 0:
		return Source{}, nil
	case 1:
		if others[0].Tag == "identity" && !slices.Contains(IdentityClaims, others[0].Name) {
			return Source{}, fmt.Errorf("identity tag %q names no claim of %v", others[0].Name, IdentityClaims)
		}

		return others[0], nil
	default:
		return Source{}, fmt.Errorf("read from both %s and %s", others[0].Tag, others[1].Tag)
	}
}

// Schema is what a request type takes from the body.
type Schema struct {
	body    map[string]struct{}
	foreign map[string]struct{}
}

// Foreign returns the keys of a JSON object body that would name a request field read from
// somewhere else, matching them as encoding/json does: without regard to case, and only when no
// body field claims the key first.
func (s *Schema) Foreign(object map[string]json.RawMessage) []string {
	var keys []string
	for key := range object {
		folded := strings.ToLower(key)
		if _, ok := s.body[folded]; ok {
			continue
		}

		if _, ok := s.foreign[folded]; ok {
			keys = append(keys, key)
		}
	}

	slices.Sort(keys)

	return keys
}

type cached struct {
	schema *Schema
	err    error
}

var schemas sync.Map

// SchemaOf returns the schema of t, a struct or a pointer to one. It returns a nil schema and no
// error for any other type, which binds no fields, and the first field that breaks the rule as an
// error naming its path. The answer is computed once per type.
func SchemaOf(t reflect.Type) (*Schema, error) {
	if hit, ok := schemas.Load(t); ok {
		c := hit.(cached) //nolint:forcetypeassert // the map holds nothing else

		return c.schema, c.err
	}

	schema, err := build(t)
	schemas.Store(t, cached{schema: schema, err: err})

	return schema, err
}

// Check returns the error [SchemaOf] returns for t.
func Check(t reflect.Type) error {
	_, err := SchemaOf(t)

	return err
}

// MustCheck panics with the error [Check] returns for T, so a request type that breaks the rule
// stops the server from mounting the route that binds it.
func MustCheck[T any]() {
	if err := Check(reflect.TypeFor[T]()); err != nil {
		panic(err)
	}
}

func build(t reflect.Type) (*Schema, error) {
	t = indirect(t)
	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	schema := &Schema{body: map[string]struct{}{}, foreign: map[string]struct{}{}}
	if err := schema.walk(t, t.Name()); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *Schema) walk(t reflect.Type, path string) error {
	for field := range t.Fields() {
		if !field.IsExported() && !field.Anonymous {
			continue
		}

		fieldPath := path + "." + field.Name

		if _, tagged := field.Tag.Lookup("json"); field.Anonymous && !tagged && indirect(field.Type).Kind() == reflect.Struct {
			if err := s.walk(indirect(field.Type), fieldPath); err != nil {
				return err
			}

			continue
		}

		if !field.IsExported() {
			continue
		}

		source, err := SourceOf(field.Tag)
		if err != nil {
			return fmt.Errorf("%s: %w", fieldPath, err)
		}

		if !source.IsBody() {
			s.foreign[strings.ToLower(field.Name)] = struct{}{}

			continue
		}

		s.body[strings.ToLower(source.Name)] = struct{}{}

		if err := bodyOnly(field.Type, fieldPath, map[reflect.Type]bool{}); err != nil {
			return err
		}
	}

	return nil
}

func bodyOnly(t reflect.Type, path string, seen map[reflect.Type]bool) error {
	t = elem(t)
	if t.Kind() != reflect.Struct || seen[t] {
		return nil
	}

	seen[t] = true

	for field := range t.Fields() {
		fieldPath := path + "." + field.Name

		if named := NamedSources(field.Tag); len(named) > 0 {
			return fmt.Errorf("%s: %s tag inside a body field", fieldPath, named[0].Tag)
		}

		if err := bodyOnly(field.Type, fieldPath, seen); err != nil {
			return err
		}
	}

	return nil
}

func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

func elem(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			t = t.Elem()
		default:
			return t
		}
	}
}
