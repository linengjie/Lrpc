package main

import "Lrpc/server"

func main() {
	server := new(server.Server)
	server.Accept("tcp", ":7788")
}
