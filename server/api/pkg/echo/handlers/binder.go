package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/binding"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	errors "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// Binder binds a request into a struct for echo, percent-decoding path parameters first.
type Binder struct{}

// NewBinder returns the binder the API installs on its echo instance.
func NewBinder() *Binder {
	return &Binder{}
}

// Bind fills s from the request. Each field of s reads from the one source its tags name, as package
// binding states: the body fills only the fields that name a json key, and a field tagged identity
// takes the caller's identity on the request's context (id, tenant, device, role, admin or
// admin_id), staying empty on an anonymous request. A path value that will not decode is left as
// it arrived rather than failing the request, so that validation reports it instead of the binder.
//
// It panics when an identity tag sits on an unexported field, which package binding does not
// check, or on a field of a type the value does not convert to: a mistake in a request struct, not
// in a request.
//
// It returns an error, which answers 500, when the type of s breaks the binding rule, so that a
// misdeclared request fails the first request that reaches it rather than binding what a client
// chose. It returns the error of [errors.NewErrUnreadableBody] when reading the body failed,
// whatever the bytes read so far held, and the error of [errors.NewErrUnprocessableEntity] for a
// JSON object body naming a field read from elsewhere and for any other failure to bind, with
// echo's 400 for a decode error unwrapped to its cause so that it answers 422. An XML body, which
// encoding/xml would bind by Go field name whatever the json tags say, answers echo's 415; it and
// the body limit's 413 stay in the chain and outrank the route error's status in the error handler.
func (b *Binder) Bind(c *echo.Context, s any) error {
	values := c.PathValues()
	for i, v := range values {
		decoded, err := url.PathUnescape(v.Value)
		if err != nil {
			continue
		}

		values[i].Value = decoded
	}

	c.SetPathValues(values)

	schema, err := binding.SchemaOf(reflect.TypeOf(s))
	if err != nil {
		return fmt.Errorf("request type %T: %w", s, err)
	}

	body := &watchedBody{ReadCloser: c.Request().Body}
	c.Request().Body = body

	discardEmptyChunkedBody(c.Request())

	if err := refuseForeignKeys(c.Request(), schema); err != nil {
		if body.failure != nil {
			return errors.NewErrUnreadableBody(body.failure)
		}

		return err
	}

	binder := new(echo.DefaultBinder)
	bindErr := binder.Bind(c, s)

	if body.failure != nil {
		return errors.NewErrUnreadableBody(body.failure)
	}

	if bindErr != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(bindErr))
	}

	if err := echo.BindHeaders(c, s); err != nil {
		return errors.NewErrUnprocessableEntity(unwrapBindError(err))
	}

	if target := reflect.Indirect(reflect.ValueOf(s)); target.Kind() == reflect.Struct {
		caller, _ := identity.FromContext(c.Request().Context())
		bindIdentity(caller, target)
	}

	return nil
}

func refuseForeignKeys(r *http.Request, schema *binding.Schema) error {
	if r.ContentLength == 0 {
		return nil
	}

	switch mediaType(r) {
	case echo.MIMEApplicationXML, echo.MIMETextXML:
		return errors.NewErrUnprocessableEntity(echo.ErrUnsupportedMediaType)
	case echo.MIMEApplicationJSON:
	default:
		return nil
	}

	if schema == nil {
		return nil
	}

	raw, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) == nil {
		if keys := schema.Foreign(object); len(keys) > 0 {
			return errors.NewErrUnprocessableEntity(fmt.Errorf("the body names fields it cannot set: %s", strings.Join(keys, ", ")))
		}
	}

	return nil
}

func mediaType(r *http.Request) string {
	base, _, _ := strings.Cut(r.Header.Get(echo.HeaderContentType), ";")

	return strings.TrimSpace(base)
}

type watchedBody struct {
	io.ReadCloser
	failure error
}

func (b *watchedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && !stderrors.Is(err, io.EOF) {
		b.failure = err
	}

	return n, err
}

func unwrapBindError(err error) error {
	if cause := stderrors.Unwrap(err); cause != nil {
		return cause
	}

	return err
}

func discardEmptyChunkedBody(r *http.Request) {
	if r.ContentLength >= 0 {
		return
	}

	buffered := bufio.NewReader(r.Body)
	if _, err := buffered.Peek(1); err != nil {
		r.ContentLength = 0

		return
	}

	r.Body = io.NopCloser(buffered)
}
