package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	address := os.Getenv("ALERT_SINK_ADDR")
	if address == "" {
		address = "127.0.0.1:8099"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	})
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "cannot read request body", http.StatusBadRequest)
			return
		}
		var pretty json.RawMessage
		if json.Valid(body) {
			_ = json.Unmarshal(body, &pretty)
		}
		log.Printf("received alert method=%s path=%s event=%q signature=%t body=%s", r.Method, r.URL.Path, r.Header.Get("X-ORP-Event"), r.Header.Get("X-ORP-Signature") != "", string(pretty))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true,"receivedAt":"`+time.Now().UTC().Format(time.RFC3339)+`"}`)
	})
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("alert test sink listening on http://%s", address)
	log.Fatal(server.ListenAndServe())
}
