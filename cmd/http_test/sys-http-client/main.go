package main

import (
	"fmt"
	"syscall"
)

func main() {
	// Create a TCP socket
	sock, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	defer syscall.Close(sock)

	// Create Connection
	syscall.Connect(sock, &syscall.SockaddrInet4{
		Port: 8000,
		Addr: [4]byte{127, 0, 0, 1},
	})

	httpGet := "GET / HTTP/1.1\r\nHost: localhost:8000\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	syscall.Write(sock, []byte(httpGet))
	response := make([]byte, 1024)
	n, _ := syscall.Read(sock, response)
	fmt.Println(string(response[:n]))
}
