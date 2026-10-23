package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type received struct {
	Authorization string          `json:"authorization"`
	Body          json.RawMessage `json:"body"`
}

type envelope struct {
	CreatedAt string `json:"created_at"`
	Batch     []struct {
		UUID string `json:"uuid"`
	} `json:"batch"`
}

func main() {
	var mu sync.Mutex

	journal := json.NewEncoder(os.Stdout)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /i/v1/analytics/events", func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			unzipped, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			reader = unzipped
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		var batch envelope
		if err := json.Unmarshal(body, &batch); err != nil || batch.CreatedAt == "" {
			http.Error(w, "not a capture v1 batch", http.StatusBadRequest)

			return
		}

		mu.Lock()
		err = journal.Encode(received{Authorization: r.Header.Get("Authorization"), Body: body})
		mu.Unlock()

		if err != nil {
			log.Printf("journaling the request: %v", err)
		}

		results := map[string]map[string]string{}
		for _, event := range batch.Batch {
			results[event.UUID] = map[string]string{"result": "ok"}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{Addr: os.Getenv("ADDRESS"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	log.Print("listening")
	log.Fatal(server.ListenAndServe())
}
