package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/n9te9/aria"
)

func main() {
	a := aria.New(
		aria.WithComporessionMode(int(websocket.CompressionDisabled)),
	)

	// 接続イベント
	a.OnConnect(func(ctx context.Context, conn *aria.Conn) error {
		log.Println("new connection")
		return nil
	})

	// メッセージ受信イベント
	a.OnMessage(func(ctx context.Context, conn *aria.Conn, message []byte) error {
		log.Printf("received: %s\n", string(message))
		// 全員に送信
		return a.BroadCast(ctx, message)
	})

	// エラーイベント
	a.OnError(func(ctx context.Context, conn *aria.Conn, err error) {
		log.Println("error:", err)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Handle(w, r); err != nil {
			log.Println("handle error:", err)
		}
	})

	// 静的ファイル（index.html）を配信
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
