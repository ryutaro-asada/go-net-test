package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 3; i++ {
		func() {
			resp, err := http.Get("http://127.0.0.1:8080")
			if err != nil {
				fmt.Println(err)
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("severRespons is %s \n", b)
		}()
		time.Sleep(time.Second * 2)
	}
}
