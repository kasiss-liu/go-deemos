package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

/**
 * 启动一个服务监听
 * 保存连接客户端
 * 接收到客户端消息时 广播到所有连接客户端
 */

type client chan<- string

var (
	leaving  = make(chan client)
	entering = make(chan client)
	messages = make(chan string)
)

func main() {
	//创建一个监听器
	listener, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Fatal("Create Server Failed")
	}

	//启动广播器
	go broadcaster()
	//监听客户端接入
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}

}

//广播器
func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		//消息通道内接收到数据以后 将数据发送到所有客户端
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		//有新客户端接入时 将连接保存
		case cli := <-entering:
			clients[cli] = true
		//有客户端关闭时 将客户端从连接池中去除 关闭连接
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}

}

//处理客户端
func handleConn(conn net.Conn) {
	ch := make(chan string)
	//启动一个响应器 将消息通道传入响应器
	go connWriter(conn, ch)
	who := conn.RemoteAddr().String()
	welcome := "Hello, " + who
	ch <- welcome
	entering <- ch
	messages <- who + " Has Entered"
	reader := bufio.NewScanner(conn)
	for reader.Scan() {
		messages <- who + ": " + reader.Text()
	}
	leaving <- ch
	messages <- who + " Has Left"
	conn.Close()
}

//响应器 将通道内的数据响应到客户端
func connWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
