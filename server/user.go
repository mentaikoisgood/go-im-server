package server

import (
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	Server *Server // 指回去 Server, 可調用廣播等方法
}

// 創建 User 實例
func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()
	user := &User{
		Name:   addr,
		Addr:   addr,
		C:      make(chan string),
		Conn:   conn,
		Server: server,
	}

	// 啟動 Goroutine 監聽個人 channel, 傳訊息給Client
	go user.ListenMessage()

	return user
}

// 發送訊息到 Client
func (u *User) ListenMessage() {
	for msg := range u.C {
		u.Conn.Write([]byte(msg + "\n"))
	}
}
