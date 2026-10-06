package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type received struct {
	Path      string `json:"path"`
	Signature string `json:"signature"`
	Body      string `json:"body"`
}

func main() {
	var mu sync.Mutex

	journal := json.NewEncoder(os.Stdout)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /{decision}", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		mu.Lock()
		err = journal.Encode(received{Path: r.URL.Path, Signature: r.Header.Get("X-ShellHub-Signature"), Body: string(body)})
		mu.Unlock()

		if err != nil {
			log.Printf("journaling the request: %v", err)
		}

		if raw := r.URL.Query().Get("delay"); raw != "" {
			delay, err := time.ParseDuration(raw)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": r.PathValue("decision")})
	})

	address := os.Getenv("ADDRESS")
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	log.Print("listening")
	log.Fatal(server.ListenAndServe())
}
