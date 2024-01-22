package web

import (
	"errors"
	"net/http"
	"strconv"
)

var (
	ErrGetToken      = errors.New("token not found on request query")
	ErrGetIP         = errors.New("ip not found on request query")
	ErrGetDimensions = errors.New("failed to get a terminal dimension")
)

func GetToken(req *http.Request) (string, error) {
	token := req.URL.Query().Get("token")

	if token == "" {
		return "", ErrGetToken
	}

	return token, nil
}

func GetDimensions(req *http.Request) (int, int, error) {
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

func GetIP(req *http.Request) (string, error) {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		return "", ErrGetIP
	}

	return ip, nil
}
