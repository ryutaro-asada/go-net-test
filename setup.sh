#!/bin/bash

http-client() {
go run ./cmd/ws_test/simple-http-client/main.go
}

http-server() {
go run ./cmd/ws_test/simple-http-server/main.go
}

websocket-client() {
go run ./cmd/ws_test/simple-websocket-client
}


websocket-server() {
go run ./cmd/ws_test/simple-websocket-server
}

net-websocket-client() {
go run ./cmd/ws_test/net-websocket/main.go
}

clean-up-exe() {
rm $(find . -maxdepth 1 -perm -111 -type f)
}

net-websocket-client() {
for i in {0..2}; do
    curl "http://127.0.0.1:8080"
    sleep 2
done
}

case "$1" in
  curl) http-client;;
  hc) http-client;;
  hs) http-server;;
  wc) websocket-client;;
  ws) websocket-server;;
  nw) net-websocket-client;;
  clean) clean-up-exe;;
    *) echo "opps";;
esac
