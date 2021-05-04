package core

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type SocketMessage struct {
	Amount       int64  `json:"amount"`
	SessionID    string `json:"sessionId"`
	UserSelector string `json:"userSelector"`
}

var Clients = make(map[*websocket.Conn]string)
var Broadcast = make(chan *SocketMessage)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketWriter(sm *SocketMessage) {
	log.Println("do writer!")
	Broadcast <- sm
}

func reader(conn *websocket.Conn) {
	for {
		log.Println("reader is running")
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func WsEndpoint(c echo.Context) error {
	userSelector := c.Param("userSelector")
	r := c.Request()
	w := c.Response().Writer

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("oh no!")
	}

	Clients[ws] = userSelector
	log.Println("client connected!!")
	log.Println(len(Clients))
	reader(ws)

	log.Println("****************")
	return c.String(http.StatusOK, "ok")
}

func RunWebsocketBroker() {
	for {
		log.Println("RunWebsocketBroker")
		val := <-Broadcast
		log.Println(val)
		// send to every client that is currently connected

		for client, i := range Clients {
			log.Println("client message outgoing")
			log.Println(i)
			// only send to the user that initiated checkout
			if val.UserSelector == i {
				err := client.WriteJSON(val)
				if err != nil {
					log.Printf("Websocket error: %s", err)
					client.Close()
					delete(Clients, client)
				}
			}

		}
	}
}

// func MakeSocketServer() *socketio.Server {

// 	server := socketio.NewServer(nil)

// 	server.OnConnect("/", func(s socketio.Conn) error {
// 		log.Println("***********CONNECT***************")
// 		s.SetContext("")
// 		log.Println("connected:", s.ID())
// 		userSelector := "chatroom"
// 		s.Join(userSelector)
// 		return nil
// 	})

// 	server.OnError("/", func(s socketio.Conn, e error) {
// 		log.Println("meet error:", e)
// 	})

// 	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
// 		log.Println("closed", reason)
// 	})

// //create
// server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

// //handle connected
// server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
// 	log.Println("New client connected")
// 	selector := c.RequestHeader().Get("userSelector")
// 	log.Println("*****************")
// 	log.Println(selector)
// 	//join them to room
// 	c.Join(selector)
// })

// //handle disconnected
// server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
// 	log.Println("Client disconnected")
// 	// close room
// 	c.Close()
// })

// 	return server
// }
