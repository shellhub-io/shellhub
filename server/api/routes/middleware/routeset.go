package middleware

import (
	"strings"

	"github.com/labstack/echo/v5"
)

type routeSet map[string]struct{}

func routeKey(method, path string) string {
	return method + " " + path
}

func (s routeSet) add(method, path string) {
	s[routeKey(method, path)] = struct{}{}
}

func (s routeSet) contains(c *echo.Context) bool {
	if _, ok := s[routeKey(c.Request().Method, c.Path())]; ok {
		return true
	}

	_, ok := s[routeKey(AnyMethod, c.Path())]

	return ok
}

func (s routeSet) unregistered(routes echo.Routes) []string {
	registered := make(map[string]struct{}, len(routes))
	paths := make(map[string]struct{}, len(routes))

	for _, route := range routes {
		registered[routeKey(route.Method, route.Path)] = struct{}{}
		paths[route.Path] = struct{}{}
	}

	var dead []string

	for entry := range s {
		method, path, _ := strings.Cut(entry, " ")

		if method == AnyMethod {
			if _, ok := paths[path]; !ok {
				dead = append(dead, entry)
			}

			continue
		}

		if _, ok := registered[entry]; !ok {
			dead = append(dead, entry)
		}
	}

	return dead
}
