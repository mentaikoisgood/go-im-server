package chat

import (
	"fmt"
	"go-im-server/internal/protocol"
	"net"
	"strings"
	"sync"
	"time"
)

// ServerConfig 定義服務器配置
type ServerConfig struct {
	IP              string
	Port            int
	TimeoutDuration time.Duration
}

// Server 結構體定義聊天服務器
type Server struct {
	config    ServerConfig
	OnlineMap map[string]*User
	Message   chan string
	mapLock   sync.RWMutex
}

// User 結構體定義用戶
type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	Server *Server
}

// NewServer 創建新的服務器實例
func NewServer(config ServerConfig) *Server {
	if config.TimeoutDuration == 0 {
		config.TimeoutDuration = 90 * time.Second
	}
	return &Server{
		config:    config,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// Start 啟動服務器
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.IP, s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	go s.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept connection error: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// BroadCast 廣播消息給所有在線用戶
func (s *Server) BroadCast(msg string) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()

	for _, user := range s.OnlineMap {
		select {
		case user.C <- msg:
		default:
			user.Offline()
		}
	}
}

// ListenMessage 監聽並處理服務器消息
func (s *Server) ListenMessage() {
	for msg := range s.Message {
		s.BroadCast(msg)
	}
}

// handleConnection 處理新的連接
func (s *Server) handleConnection(conn net.Conn) {
	user := NewUser(conn, s)
	defer func() {
		user.Offline()
		conn.Close()
	}()

	user.SendMsg(protocol.WelcomeMessage)
	user.Online()

	active := make(chan bool)
	defer close(active)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			msg := string(buf[:n-1])
			user.DoMessage(msg)
			active <- true
		}
	}()

	for {
		select {
		case <-active:
		case <-time.After(s.config.TimeoutDuration):
			user.SendMsg(protocol.TimeoutMessage)
			return
		}
	}
}

// NewUser 創建新用戶
func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()
	user := &User{
		Name:   addr,
		Addr:   addr,
		C:      make(chan string),
		Conn:   conn,
		Server: server,
	}
	go user.ListenMessage()
	return user
}

// Online 處理用戶上線
func (u *User) Online() {
	u.Server.mapLock.Lock()
	u.Server.OnlineMap[u.Name] = u
	u.Server.mapLock.Unlock()
	u.Server.Message <- fmt.Sprintf(protocol.UserJoinedFormat, u.Name)
}

// Offline 處理用戶下線
func (u *User) Offline() {
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.mapLock.Unlock()
	u.Server.Message <- fmt.Sprintf(protocol.UserLeftFormat, u.Name)
	close(u.C)
}

// SendMsg 發送消息給用戶
func (u *User) SendMsg(msg string) {
	u.Conn.Write([]byte(msg + "\n"))
}

// ListenMessage 監聽用戶消息
func (u *User) ListenMessage() {
	for msg := range u.C {
		u.SendMsg(msg)
	}
}

// DoMessage 處理用戶發送的消息
func (u *User) DoMessage(msg string) {
	msg = strings.TrimSpace(msg)

	switch {
	case strings.HasPrefix(msg, protocol.CmdTo+"="):
		u.handlePrivateMessage(strings.SplitN(msg, "=", 3))
	case strings.HasPrefix(msg, protocol.CmdRename+"="):
		u.handleRename(strings.TrimPrefix(msg, protocol.CmdRename+"="))
	case msg == protocol.CmdWho:
		u.handleWho()
	default:
		u.Server.Message <- fmt.Sprintf("[%s] 說 %s", u.Name, msg)
	}
}

// handlePrivateMessage 處理私聊消息
func (u *User) handlePrivateMessage(parts []string) {
	if len(parts) != 3 {
		u.SendMsg(protocol.ErrPrivateFormat)
		return
	}

	targetName := strings.TrimSpace(parts[1])
	content := strings.TrimSpace(parts[2])
	if targetName == "" || content == "" {
		u.SendMsg(protocol.ErrPrivateFormat)
		return
	}

	u.Server.mapLock.RLock()
	target, ok := u.Server.OnlineMap[targetName]
	u.Server.mapLock.RUnlock()

	if !ok {
		u.SendMsg(fmt.Sprintf(protocol.ErrUserOffline, targetName))
		return
	}

	target.SendMsg(fmt.Sprintf(protocol.PrivateMsgFormat, u.Name, content))
	u.SendMsg(fmt.Sprintf(protocol.PrivateConfirmFormat, targetName, content))
}

// handleRename 處理更名請求
func (u *User) handleRename(newName string) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		u.SendMsg(protocol.ErrNameEmpty)
		return
	}

	u.Server.mapLock.Lock()
	defer u.Server.mapLock.Unlock()

	if _, exists := u.Server.OnlineMap[newName]; exists {
		u.SendMsg(fmt.Sprintf(protocol.ErrNameTaken, newName))
		return
	}

	delete(u.Server.OnlineMap, u.Name)
	oldName := u.Name
	u.Name = newName
	u.Server.OnlineMap[u.Name] = u

	u.SendMsg(fmt.Sprintf(protocol.RenameSuccessFormat, u.Name))
	u.Server.Message <- fmt.Sprintf(protocol.UserRenamedFormat, oldName, u.Name)
}

// handleWho 處理查詢在線用戶請求
func (u *User) handleWho() {
	u.Server.mapLock.RLock()
	defer u.Server.mapLock.RUnlock()

	for _, user := range u.Server.OnlineMap {
		u.SendMsg(fmt.Sprintf(protocol.OnlineFormat, user.Name))
	}
}
