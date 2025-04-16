package server

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User
	Message   chan string
	mapLock   sync.RWMutex // 讀寫鎖, 用於保護 OnlineMap
}

// 創建 Server 實例
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// 啟動 Server
func (s *Server) Start() {
	// 監聽
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("LISTEN ERROR:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server started at %s:%d\n", s.IP, s.Port)

	go s.ListenMessage()

	// 主循環: 接收 Client 連線
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ACCEPT　ERROR:", err)
			continue
		}

		// 處理 Client 連線 (暫時只打印訊息)
		go s.handleConnection(conn)
	}

}

func (s *Server) BroadCast(msg string) {
	s.mapLock.RLock()
	for _, user := range s.OnlineMap {
		user.C <- msg
	}
	s.mapLock.RUnlock()
}

// 監聽廣播訊息 channel 的 goroutine
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.BroadCast(msg)
	}
}

// 處理 Client 連線
func (s *Server) handleConnection(conn net.Conn) {
	fmt.Println("New connection from", conn.RemoteAddr().String())

	conn.Write([]byte("WELCOME TO GO IM SERVER\n"))

	user := NewUser(conn, s)
	user.Online()

	// 為每個連線建立一個活動CHANNEL， 當用戶收到任意消息時，向此 CHANNEL 發信號， 表示用戶活躍
	active := make(chan bool)

	// 啟動獨立 goroutine，處理 Client 傳來的訊息
	go func() {
		buf := make([]byte, 4096) // 4kb緩衝區
		for {
			n, err := conn.Read(buf)
			if err != nil {
				user.Offline()
				return
			}

			msg := string(buf[:n-1])
			user.DoMessage(msg)
			active <- true
		}
	}()

	// 超時檢查： 超過30a 秒沒有收到任何ACTIVE CHANNEL 的信號 ，強制踢出
	timeoutDuration := 90 * time.Second
	for {
		select {
		case <-active:
			// 收到活動信號，重置計時器
		case <-time.After(timeoutDuration):
			// 超時，關閉連線
			user.SendMsg("❌ 你已經超時了，將被踢出")
			user.Offline()
			conn.Close()
			return
		}
	}

}
