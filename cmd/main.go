package main

import (
	"bufio"
	"flag"
	"fmt"
	"go-im-server/internal/protocol"
	"go-im-server/pkg/chat"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	// 命令行參數解析
	isServer := flag.Bool("server", false, "Run as server mode")
	ip := flag.String("ip", "127.0.0.1", "Server IP address")
	port := flag.Int("port", 8888, "Server port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *ip, *port)

	if *isServer {
		runServer(*ip, *port)
	} else {
		runClient(addr)
	}
}

func runServer(ip string, port int) {
	config := chat.ServerConfig{
		IP:              ip,
		Port:            port,
		TimeoutDuration: 90 * time.Second,
	}

	server := chat.NewServer(config)
	log.Printf("Starting server on %s:%d\n", config.IP, config.Port)

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func runClient(serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("連線失敗:", err)
		return
	}
	defer conn.Close()
	fmt.Println("連線到 Server 成功")

	// 接收服務器消息的goroutine
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("讀取 Server 訊息時發生錯誤:", err)
		}
	}()

	showMenu()
	handleClientInput(conn)
}

func showMenu() {
	fmt.Println("========== 菜單 ==========")
	fmt.Println("1. 公聊模式：直接輸入訊息（預設模式）")
	fmt.Println("2. 私聊模式：輸入 'private' 進入私聊子菜單")
	fmt.Println("3. 查詢在線用戶：輸入 'who'")
	fmt.Println("4. 修改用戶名稱：輸入 'rename=新名稱'")
	fmt.Println("5. 隨時輸入 'menu' 顯示此菜單")
	fmt.Println("==========================")
}

func handleClientInput(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("請輸入指令(或直接輸入訊息公聊): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("讀取失敗:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case protocol.CmdMenu:
			showMenu()
		case protocol.CmdPrivate:
			handlePrivateChat(conn, reader)
		default:
			fmt.Fprintln(conn, input)
		}
	}
}

func handlePrivateChat(conn net.Conn, reader *bufio.Reader) {
	fmt.Println("====== 私聊模式 ======")
	fmt.Println("1. 查詢在線用戶：輸入 'who'")
	fmt.Println("2. 發送私聊訊息，請使用格式：to=目標用戶=訊息內容")
	fmt.Println("3. 輸入 'exit' 返回公聊模式")
	fmt.Println("=====================")

	for {
		fmt.Print("私聊模式- 請輸入命令： ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("讀取失敗:", err)
			continue
		}

		cmd = strings.TrimSpace(cmd)
		if strings.ToLower(cmd) == protocol.CmdExit {
			fmt.Println("退出私聊模式")
			break
		}

		fmt.Fprintln(conn, cmd)
	}
}
