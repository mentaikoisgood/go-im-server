package server

import (
	"fmt"
	"net"
	"sync"
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
	defer conn.Close()

	conn.Write([]byte("WELCOME TO GO IM SERVER\n"))

	user := NewUser(conn, s)

	// 加入線上用戶
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 廣播新用戶加入
	s.Message <- fmt.Sprintf("[%s] 上線了", user.Name)

	// 保持連線不中斷
	select {}
}
