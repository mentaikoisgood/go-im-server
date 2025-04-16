package main

import "go-im-server/server"

func main() {
	s := server.NewServer("127.0.0.1", 8888)
	s.Start()
}


