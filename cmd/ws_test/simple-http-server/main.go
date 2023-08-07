package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Yo!")
	})
	http.ListenAndServe("127.0.0.1:8080", nil)
}
