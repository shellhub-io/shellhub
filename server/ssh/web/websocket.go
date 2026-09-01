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
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		return "", ErrGetIP
	}

	return ip, nil
}
