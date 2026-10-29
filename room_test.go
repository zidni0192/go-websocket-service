package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func roomConn(t *testing.T, server *httptest.Server, room string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/rooms/" + room
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

func TestRoomFanoutIsolationAndReconnect(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first := roomConn(t, server, "alpha")
	second := roomConn(t, server, "alpha")
	other := roomConn(t, server, "beta")
	if err := first.Write(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	typ, msg, err := second.Read(ctx)
	if err != nil || typ != websocket.MessageText || string(msg) != "hello" {
		t.Fatalf("fanout = (%v, %q, %v)", typ, msg, err)
	}
	isolatedCtx, stopIsolation := context.WithTimeout(ctx, 50*time.Millisecond)
	defer stopIsolation()
	if _, _, err := other.Read(isolatedCtx); err == nil {
		t.Fatal("message crossed room boundary")
	}
	other.Close(websocket.StatusNormalClosure, "reconnect")
	reconnected := roomConn(t, server, "beta")
	if err := first.Write(ctx, websocket.MessageText, []byte("again")); err != nil {
		t.Fatal(err)
	}
	reconnectCtx, stopReconnect := context.WithTimeout(ctx, 50*time.Millisecond)
	defer stopReconnect()
	if _, _, err := reconnected.Read(reconnectCtx); err == nil {
		t.Fatal("reconnected client joined wrong room")
	}
}

func TestRoomHubDropsSlowClient(t *testing.T) {
	hub := newRoomHub(1)
	sender := &roomClient{send: make(chan []byte, 1)}
	slow := &roomClient{send: make(chan []byte, 1)}
	hub.join("x", sender)
	hub.join("x", slow)
	if !hub.broadcast("x", sender, []byte("one")) {
		t.Fatal("first broadcast failed")
	}
	if !hub.broadcast("x", sender, []byte("two")) {
		t.Fatal("second broadcast failed")
	}
	<-slow.send // The queued message remains readable after the channel closes.
	if _, ok := <-slow.send; ok {
		t.Fatal("slow client was not removed")
	}
}
