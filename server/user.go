package server

import (
	"fmt"
	"net"
	"strings"
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

func (u *User) SendMsg(msg string) {
	u.Conn.Write([]byte(msg + "\n"))
}

// 監聽該用戶的 Channel, 並寫入到連線
func (u *User) ListenMessage() {
	for msg := range u.C {
		u.SendMsg(msg)
	}
}

// 處理 Client 傳來的訊息
func (u *User) DoMessage(msg string) {

	msg = strings.TrimSpace(msg) // 移除多餘換行及空格

	if strings.HasPrefix(msg, "to=") {
		// 處理私聊： 格式為 "to=目標用戶＝消息內容"
		parts := strings.SplitN(msg, "=", 3)
		if len(parts) != 3 {
			u.SendMsg("錯誤：私聊格式錯誤，請使用 to=用戶名=消息")
			return
		}
		targetName := strings.TrimSpace(parts[1])
		content := strings.TrimSpace(parts[2])
		if targetName == "" || content == "" {
			u.SendMsg("錯誤：私聊格式錯誤，請使用 to=用戶名=消息")
			return
		}

		// 取得目標 USER 讀鎖保護在線地圖
		u.Server.mapLock.RLock()
		target, ok := u.Server.OnlineMap[targetName]
		u.Server.mapLock.RUnlock()
		if !ok {
			u.SendMsg(fmt.Sprintf("錯誤：用戶 [%s] 不在線", targetName))
			return
		}
		// 發送私聊消息
		target.SendMsg(fmt.Sprintf("[私聊] %s 說: %s", u.Name, content))

		// 同時回饋給發訊這確認
		u.SendMsg(fmt.Sprintf("[私聊] 您對 [%s] 說: %s", targetName, content))
		return
	}

	if strings.HasPrefix(msg, "rename=") {
		newName := strings.TrimSpace(strings.TrimPrefix(msg, "rename="))
		if newName == "" {
			u.SendMsg("錯誤：新名稱不能為空")
			return
		}

		// 檢查新名稱是否已被使用
		u.Server.mapLock.Lock()
		_, exists := u.Server.OnlineMap[newName]
		if exists {
			u.Server.mapLock.Unlock()
			u.SendMsg(fmt.Sprintf("錯誤：名稱 [%s] 已經被使用", newName))
			return
		}

		// 修改名稱： 先從 OnlineMap 移除舊的，然後加入新的名稱
		delete(u.Server.OnlineMap, u.Name)
		oldName := u.Name
		u.Name = newName
		u.Server.OnlineMap[u.Name] = u
		u.Server.mapLock.Unlock()

		u.SendMsg(fmt.Sprintf("成功：您的名稱已更新為 [%s]", u.Name))
		u.Server.Message <- fmt.Sprintf("公告：[%s] 改名為 [%s]", oldName, u.Name)
		return
	}

	if msg == "who" {
		u.Server.mapLock.RLock()
		fmt.Println("Debug: OnlineMap =", u.Server.OnlineMap) // 調試輸出

		for _, user := range u.Server.OnlineMap {
			onlineMsg := fmt.Sprintf("[在線] %s\n", user.Name)
			u.SendMsg(onlineMsg)

		}
		u.Server.mapLock.RUnlock()
	} else {
		broadcast := fmt.Sprintf("[%s] 說 %s", u.Name, msg)
		u.Server.Message <- broadcast
	}
}
