package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDummy(t *testing.T) {
	fmt.Println("dummy")

	res := make(chan *http.Request)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Header)
			fmt.Fprintf(w, "okay")
			res <- r
		},
	))

	srv := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if !waitForConnection("tcp", "127.0.0.1:8080") {
		panic("Failed to start http server")
	}

	client := &http.Client{}

	req, _ := http.NewRequest("GET", "http://localhost:80/", nil)
	go client.Do(req)

	rr := <-res
	fmt.Println(rr.Host)

	// / => UI
}
