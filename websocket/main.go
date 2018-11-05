package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

//官方包 指定缓冲字节大小
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type webSocketServer struct{}

//websocket 监听函数
func (ws *webSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println(r.Proto)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	responMessage := ""
	for {
		messageType := websocket.TextMessage
		responMessage = `你好`
		if err := conn.WriteMessage(messageType, []byte(responMessage)); err != nil {
			log.Println(err)
			return
		}
		time.Sleep(2 * time.Second)

	}

}

var wsServer webSocketServer

func main() {
	c := make(chan bool)
	//启动一个文件服务器和websocket监听server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./test"))))
		mux.Handle("/", &wsServer)
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
	}()

	<-c

}
