package main

import (
	"fmt"
	"net/http"
)

func main() {
	resp, _ := http.Get("http://localhost:8000")
	defer resp.Body.Close()
	fmt.Println(resp.Body)
}
