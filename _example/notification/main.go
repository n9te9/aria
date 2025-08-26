package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/n9te9/aria"
)

func main() {
	a := aria.New()

	a.OnConnect(func(ctx context.Context, c *aria.Conn) error {
		log.Println("connected")
		uid := uuid.NewString()
		c.Set("id", uid)
		return a.BroadCast(ctx, []byte(fmt.Sprintf("%s has connected", uid)))
	})

	a.OnDisconnect(func(ctx context.Context, c *aria.Conn) error {
		log.Println("disconnected")

		uid, _ := c.Get("id")
		return a.BroadCast(ctx, []byte(fmt.Sprintf("%s has disconnected", uid)))
	})

	a.OnMessage(func(ctx context.Context, c *aria.Conn, msg []byte) error {
		return a.BroadCast(ctx, msg)
	})

	a.OnClose(func(ctx context.Context, conn *aria.Conn) error {
		uid, _ := conn.Get("id")
		return a.BroadCast(ctx, []byte(fmt.Sprintf("%s has left", uid)))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		_ = a.Handle(w, r)
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
