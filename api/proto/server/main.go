package main

import (
	"fmt"
	"therealbroker/api/proto/server/handler"
)

func main() {
	fmt.Println("starting server")
	handler.StartServer()
}
