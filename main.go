package main

import (
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := openConn()
	defer c.Close()
	done := make(chan struct{})



	go func() {
		read(done, c)
	}()
	hash := sha256.New()
	hash.Write([]byte("secretstuff"))
	hashString := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	println(hashString)
	loginMsg := fmt.Sprintf(`{
    "msg": "method",
    "method": "login",
    "id":"1312331313123123123123",
    "params":[
        {
            "user": { "username": "username" },
            "password": {
                "digest": "%s",
                "algorithm":"sha-256"
            }
        }
    ]
}`, hashString)
	c.WriteMessage(websocket.TextMessage, []byte(loginMsg))

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}

func openConn() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "chat.tarent.de", Path: "/websocket"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	connect(c)
	return c
}

func read(done chan struct{}, c *websocket.Conn) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalln("read:", err)
		}
		log.Printf("recv: %s", message)
		handlePing(message, c)
	}
}

func handlePing(message []byte, c *websocket.Conn) {
	if string(message) == `{"msg":"ping"}` {
		err := c.WriteMessage(websocket.TextMessage, []byte(`{"msg":"pong"}`))
		if err != nil {
			log.Fatalln("write:", err)
		}
	}
}

func connect(c *websocket.Conn) {
	err := c.WriteMessage(websocket.TextMessage, []byte(`{
	    "msg": "connect",
	    "version": "1",
	    "support": ["1"]}`))
	if err != nil {
		log.Fatal("Could not connect: %v", err)
	}
}