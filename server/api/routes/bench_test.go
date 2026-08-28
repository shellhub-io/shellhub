package routes

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/sirupsen/logrus"
)

// nullWriter is a [http.ResponseWriter] that keeps the benchmarks measuring the router and its
// middleware rather than httptest's recorder, which allocates a buffer per request.
type nullWriter struct {
	header http.Header
}

func (w *nullWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}

	return w.header
}

func (w *nullWriter) Write(b []byte) (int, error) { return len(b), nil }

func (w *nullWriter) WriteHeader(int) {}

func benchRouter(b *testing.B) http.Handler {
	b.Helper()

	// The request-log middleware writes a line per request. Keep the formatting in the
	// measurement — it is real per-request work — but not the terminal write behind it.
	previous := logrus.StandardLogger().Out
	logrus.SetOutput(io.Discard)
	b.Cleanup(func() { logrus.SetOutput(previous) })

	service := serviceMocks.NewMockService(b)
	service.On("PublicKey").Return(nil).Maybe()

	return NewRouter(service)
}

// benchServe drives target through the router. It asserts the status once up front so a
// benchmark cannot quietly measure a 404 it did not mean to.
func benchServe(b *testing.B, target string, want int) {
	b.Helper()

	router := benchRouter(b)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil))

	if rec.Code != want {
		b.Fatalf("setup request returned %d, want %d", rec.Code, want)
	}

	b.ReportAllocs()

	for b.Loop() {
		router.ServeHTTP(&nullWriter{}, httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil))
	}
}

// BenchmarkRouterHealthcheck measures the cost of the middleware chain and the router with a
// handler that does nothing but set a status.
func BenchmarkRouterHealthcheck(b *testing.B) {
	benchServe(b, "/api"+HealthCheckURL, http.StatusOK)
}

// BenchmarkRouterNotFound measures a routing miss, which also runs the error handler.
func BenchmarkRouterNotFound(b *testing.B) {
	benchServe(b, "/api/no-such-route", http.StatusNotFound)
}

// BenchmarkRouterHealthcheckParallel looks for contention around the context pool and the
// router, which a single-goroutine benchmark cannot see.
func BenchmarkRouterHealthcheckParallel(b *testing.B) {
	router := benchRouter(b)
	target := "/api" + HealthCheckURL

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			router.ServeHTTP(&nullWriter{}, httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil))
		}
	})
}
