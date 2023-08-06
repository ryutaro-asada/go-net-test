package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	defer conn.Close()

	if err != nil {
		fmt.Println("error: ", err)
	}

	httpGetStr := "GET / HTTP/1.1\r\nHost: localhost:8000\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	data := []byte(httpGetStr)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("error: ", err)
	}

	// ここから読み取り
	readdata := make([]byte, 1024)
	count, _ := conn.Read(readdata)
	fmt.Println(string(readdata[:count]))
}
