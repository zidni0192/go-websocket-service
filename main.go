package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", newHandler()))
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/ws", websocketEcho)
	return mux
}

func websocketEcho(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		messageType, message, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		if err := conn.Write(r.Context(), messageType, message); err != nil {
			return
		}
	}
}
