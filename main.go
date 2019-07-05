package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
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
	hashArr := sha256.Sum256([]byte("password"))
	hash := hashArr[:]
	hashString := hex.EncodeToString(hash)
	println(hashString)
	loginMsg := fmt.Sprintf(`{
    "msg": "method",
    "method": "login",
    "id":"1312331313123123123123",
    "params":[
        {
            "user": { "username": "user" },
            "password": {
                "digest": "%s",
                "algorithm":"sha-256"
            }
        }
    ]
}`, hashString)
	c.WriteMessage(websocket.TextMessage, []byte(loginMsg))

	subscribeMsg := `{
		"msg": "sub",
		"id": "unique-id",
		"name": "stream-room-messages",
		"params":[
			"roomid",
			false
		]
	}`
	c.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))

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
		var msg ChatMessage
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Fatalln("read:", err)
		}
		if len(msg.Fields.Args) > 0 {
			chatMsg := msg.Fields.Args[0].Msg
			log.Printf("%s", msg.Fields.Args[0].Msg)
			if chatMsg == "flemli gib aubergine" {
				err := sendEggplant("roomid", c)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		handlePing(message, c)
	}
}

func sendEggplant(roomID string, c *websocket.Conn) error {
	eggplantMsg := fmt.Sprintf(`{
		"msg": "method",
		"method": "sendMessage",
		"id": "42",
		"params": [
			{
				"_id": "%s",
				"rid": "%s",
				"msg": ":eggplant:"
			}
		]
	}`, uuid.NewV4().String(), roomID)

	return c.WriteMessage(websocket.TextMessage, []byte(eggplantMsg))
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

type ChatMessage struct {
	Msg        string `json:"msg"`
	Collection string `json:"collection"`
	ID         string `json:"id"`
	Fields     struct {
		EventName string `json:"eventName"`
		Args      []struct {
			ID  string `json:"_id"`
			Rid string `json:"rid"`
			Msg string `json:"msg"`
			Ts  struct {
				Date int64 `json:"$date"`
			} `json:"ts"`
			U struct {
				ID       string `json:"_id"`
				Username string `json:"username"`
				Name     string `json:"name"`
			} `json:"u"`
			Mentions  []interface{} `json:"mentions"`
			Channels  []interface{} `json:"channels"`
			UpdatedAt struct {
				Date int64 `json:"$date"`
			} `json:"_updatedAt"`
		} `json:"args"`
	} `json:"fields"`
}
