package web

import (
	"errors"
	"net/http"
	"strconv"
)

func getToken(req *http.Request) (string, error) {
	token := req.URL.Query().Get("token")

	if token == "" {
		return "", ErrGetToken
	}

	return token, nil
}

func getDimensions(req *http.Request) (uint32, uint32, error) {
	toUint32 := func(text string) (uint64, error) {
		integer, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return 0, err
		}

		return integer, nil
	}

	cols, err := toUint32(req.URL.Query().Get("cols"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	rows, err := toUint32(req.URL.Query().Get("rows"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	//nolint: gosec // cols and rows are uint32, so we can safely convert them.
	return uint32(cols), uint32(rows), nil
}

func getIP(req *http.Request) (string, error) {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		return "", ErrGetIP
	}

	return ip, nil
}

func getDisplay(req *http.Request) (string, error) {
	display := req.URL.Query().Get("display")
	if display == "" {
		return "", ErrGetDisplay
	}

	return display, nil
}
