package query

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

const (
	// MinPageSize is the smallest page a cursor read returns.
	MinPageSize = 1
	// DefaultPageSize is the page a cursor read returns when it asks for none.
	DefaultPageSize = 100
	// DefaultMaxPageSize is the largest page when the server sets no maximum.
	DefaultMaxPageSize = 1000
)

const maxEncodedCursorLength = 512

// ErrInvalidCursor is returned by [ParseCursor] for a string that is not a cursor [Cursor.String]
// produced.
var ErrInvalidCursor = errors.New("invalid cursor")

// Cursor is the position of a row in a read ordered by creation time, then id. A read continues
// with the rows strictly after it.
type Cursor struct {
	CreatedAt time.Time `json:"-"`
	ID        string    `json:"-"`
}

type encodedCursor struct {
	CreatedAtMicro int64  `json:"c"`
	ID             string `json:"i"`
}

// String encodes c as unpadded base64url, safe to place in a URL as is. It keeps CreatedAt to the
// microsecond, the precision PostgreSQL stores. Clients treat it as opaque: its format may change.
func (c Cursor) String() string {
	data, _ := json.Marshal(encodedCursor{CreatedAtMicro: c.CreatedAt.UnixMicro(), ID: c.ID}) //nolint:errchkjson // an int64 and a string always marshal

	return base64.RawURLEncoding.EncodeToString(data)
}

// ParseCursor decodes a cursor [Cursor.String] produced. It returns ErrInvalidCursor for anything
// else, including an empty string, a cursor with fields missing or added, and one longer than any
// cursor the server issues.
func ParseCursor(s string) (Cursor, error) {
	if s == "" || len(s) > maxEncodedCursorLength {
		return Cursor{}, ErrInvalidCursor
	}

	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var decoded encodedCursor
	if err := decoder.Decode(&decoded); err != nil || decoder.More() {
		return Cursor{}, ErrInvalidCursor
	}

	if decoded.CreatedAtMicro <= 0 || decoded.ID == "" {
		return Cursor{}, ErrInvalidCursor
	}

	return Cursor{CreatedAt: time.UnixMicro(decoded.CreatedAtMicro).UTC(), ID: decoded.ID}, nil
}

// CursorPaginator is the cursor mode of a list request. After is the encoded cursor of the last
// row the previous page returned, empty on the first page.
type CursorPaginator struct {
	Size  int64  `query:"page[size]" json:"-"`
	After string `query:"page[after]" json:"-"`
}

// Normalize clamps Size into [MinPageSize, maxSize]. A zero Size takes DefaultPageSize, and a
// maxSize below one takes DefaultMaxPageSize.
func (p *CursorPaginator) Normalize(maxSize int64) {
	if maxSize < MinPageSize {
		maxSize = DefaultMaxPageSize
	}

	if p.Size == 0 {
		p.Size = DefaultPageSize
	}

	p.Size = min(max(p.Size, MinPageSize), maxSize)
}
