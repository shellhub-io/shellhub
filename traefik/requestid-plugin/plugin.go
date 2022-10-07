package requestid_plugin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func init() {
	f, err := os.OpenFile("/tmp/log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags)
}

type Config struct {
	HeaderName string
}

func CreateConfig() *Config {
	return &Config{
		HeaderName: "X-Request-ID",
	}
}

type RequestID struct {
	next       http.Handler
	headerName string
	name       string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.HeaderName) == 0 {
		return nil, fmt.Errorf("HeaderName cannot be empty")
	}

	return &RequestID{
		next:       next,
		headerName: config.HeaderName,
		name:       name,
	}, nil
}

func (r *RequestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.Header.Set(r.headerName, uuid.NewString())

	r.next.ServeHTTP(rw, req)
}
