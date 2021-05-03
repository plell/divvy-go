package core

import (
	"fmt"
	"log"

	socketio "github.com/googollee/go-socket.io"
)

func MakeSocketServer() *socketio.Server {

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("***********CONNECT***************")
		s.SetContext("")
		// fmt.Println("connected:", s.ID())
		userSelector := "asd"
		s.Join(userSelector)
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

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

	return server
}
