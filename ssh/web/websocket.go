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

func getDimensions(req *http.Request) (int, int, error) {
	toUint8 := func(text string) (uint64, error) {
		integer, err := strconv.ParseUint(text, 10, 8)
		if err != nil {
			return 0, err
		}

		return integer, nil
	}

	cols, err := toUint8(req.URL.Query().Get("cols"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	rows, err := toUint8(req.URL.Query().Get("rows"))
	if err != nil {
		return 0, 0, errors.Join(ErrGetDimensions, err)
	}

	return int(cols), int(rows), nil
}

func getIP(req *http.Request) (string, error) {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		return "", ErrGetIP
	}

	return ip, nil
}
