package core

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type SocketMessage struct {
	Amount          int64  `json:"amount"`
	PaymentIntentID string `json:"paymentIntentId"`
	UserSelector    string `json:"userSelector"`
}

var Clients = make(map[*websocket.Conn]string)
var Broadcast = make(chan *SocketMessage)
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"binary"},
}

func WebsocketWriter(sm *SocketMessage) {
	log.Println("do writer!")
	Broadcast <- sm
}

func WsEndpoint(c echo.Context) error {
	userSelector := c.Param("userSelector")

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("oh no!")
	}

	Clients[ws] = userSelector
	log.Println("client connected!!")

	defer ws.Close()
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		log.Printf("recv: %s", message)
		err = ws.WriteMessage(mt, message)
		if err != nil {
			return err
		}
	}
}

func RunWebsocketBroker() {
	for {
		val := <-Broadcast
		// send to every client that is currently connected

		for client, i := range Clients {
			if val.UserSelector == i {
				err := client.WriteJSON(val)
				if err != nil {
					log.Printf("Websocket error: %s", err)
					client.Close()
					delete(Clients, client)
				} else {
					break
				}
			}

		}
	}
}
