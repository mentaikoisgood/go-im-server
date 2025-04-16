package server

import (
	"fmt"
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

// 用戶上線
func (u *User) Online() {
	u.Server.mapLock.Lock()
	u.Server.OnlineMap[u.Name] = u
	u.Server.mapLock.Unlock()

	u.Server.Message <- fmt.Sprintf("✅ [%s] 上線了", u.Name)
}

func (u *User) Offline() {
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.mapLock.Unlock()

	u.Server.Message <- fmt.Sprintf("❌ [%s] 離線了", u.Name)

}

// 處理使用者傳來的訊息
func (u *User) DoMessage(msg string) {
	u.Server.Message <- fmt.Sprintf("💬 [%s] 說：%s", u.Name, msg)
}

// 監聽該用戶的 Channel, 並寫入到連線
func (u *User) ListenMessage() {
	for msg := range u.C {
		u.Conn.Write([]byte(msg + "\n"))
	}
}
