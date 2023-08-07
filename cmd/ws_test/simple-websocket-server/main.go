package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

func main() {
	wsh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			// ...
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			defer cancel()

			_, message, err := c.Read(ctx)
			if err != nil {
				break
			}
			log.Printf("Received %s", message)

			err = c.Write(ctx, websocket.MessageText, []byte("Yo!"))
			if err != nil {
				break
			}
		}
	})
	http.Handle("/", wsh)
	fmt.Println("start ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
