package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	id := os.Getenv("NODE_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(id))
	})

	slog.Info("starting", "node", id, "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
