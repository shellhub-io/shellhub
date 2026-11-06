package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
)

func getToken(req *http.Request) (string, error) {
	token := req.URL.Query().Get("token")

	if token == "" {
		return "", ErrGetToken
	}

	return token, nil
}

func getDimensions(req *http.Request) (uint16, uint16, error) {
	toUint16 := func(text string) (uint16, error) {
		integer, err := strconv.ParseUint(text, 10, 16)
		if err != nil {
			return 0, err
		}

		return uint16(integer), nil
	}

	cols, err := toUint16(req.URL.Query().Get("cols"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	rows, err := toUint16(req.URL.Query().Get("rows"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	return cols, rows, nil
}

func getIP(req *http.Request) (string, error) {
	resolved, ok := origin.FromContext(req.Context())
	if !ok || resolved.IP == "" {
		return "", ErrGetIP
	}

	return resolved.IP, nil
}
