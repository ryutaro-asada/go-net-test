package main

import (
	"context"
	"fmt"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:8080", nil)
	if err != nil {
		// ...
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	for i := 0; i < 3; i++ {
		wsjson.Write(ctx, c, "Hi!")
		_, b, _ := c.Read(ctx)
		fmt.Printf("respons is %s \n", string(b))
		time.Sleep(time.Second * 2)
	}

	c.Close(websocket.StatusNormalClosure, "")
}
