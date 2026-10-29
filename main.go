package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", newHandler()))
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	rooms := newRoomHub(16)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/ws", websocketEcho)
	mux.HandleFunc("/rooms/", rooms.serve)
	return mux
}

const roomQueueSize = 16

type roomHub struct {
	mu        sync.Mutex
	rooms     map[string]map[*roomClient]struct{}
	queueSize int
}

type roomClient struct {
	conn *websocket.Conn
	send chan []byte
}

func newRoomHub(queueSize int) *roomHub {
	return &roomHub{rooms: make(map[string]map[*roomClient]struct{}), queueSize: queueSize}
}

func (h *roomHub) serve(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/rooms/")
	if room == "" || strings.Contains(room, "/") {
		http.NotFound(w, r)
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	c := &roomClient{conn: conn, send: make(chan []byte, h.queueSize)}
	h.join(room, c)
	defer func() {
		h.leave(room, c)
		conn.Close(websocket.StatusNormalClosure, "disconnected")
	}()
	ctx := r.Context()
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			select {
			case msg, ok := <-c.send:
				if !ok {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		typ, msg, err := conn.Read(ctx)
		if err != nil {
			break
		}
		if typ == websocket.MessageText && !h.broadcast(room, c, msg) {
			break
		}
	}
	conn.Close(websocket.StatusPolicyViolation, "room delivery queue full")
	<-writerDone
}

func (h *roomHub) join(room string, c *roomClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*roomClient]struct{})
	}
	h.rooms[room][c] = struct{}{}
}

func (h *roomHub) leave(room string, c *roomClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[room]
	if _, ok := clients[c]; ok {
		delete(clients, c)
		close(c.send)
	}
	if len(clients) == 0 {
		delete(h.rooms, room)
	}
}

func (h *roomHub) broadcast(room string, sender *roomClient, msg []byte) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.rooms[room] {
		if c == sender {
			continue
		}
		copyMsg := append([]byte(nil), msg...)
		select {
		case c.send <- copyMsg:
		default:
			delete(h.rooms[room], c)
			close(c.send)
		}
	}
	if len(h.rooms[room]) == 0 {
		delete(h.rooms, room)
	}
	return true
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
