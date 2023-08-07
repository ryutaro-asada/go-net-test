package main

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"net"
	"net/http"
)

// opcode represents a WebSocket opcode.
type opcode int

// https://tools.ietf.org/html/rfc6455#section-11.8.
const (
	opContinuation opcode = iota
	opText
	opBinary
	// 3 - 7 are reserved for further non-control frames.
	_
	_
	_
	_
	_
	opClose
	opPing
	opPong
	// 11-16 are reserved for further control frames.
)

type webSocketHeader struct {
	fin    bool
	rsv1   bool
	rsv2   bool
	rsv3   bool
	opcode opcode

	payloadLength int64

	masked  bool
	maskKey uint32
}

func main() {
	// CREATE TCP Connection
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error: ", err)
	}
	defer conn.Close()

	// REQUEST upgrade to websocket
	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	// カギは本来乱数だが適当な固定値
	req.Header.Set("Sec-WebSocket-Key", "bIk46puofbgF8UhhxaiSHw==")
	req.Write(conn)

	// サーバからのACCEPT読み込むが無視(printするだけ)
	// 本来はsec_websocket_acceptの検証が入るはず
	readdata := make([]byte, 1024)
	count, _ := conn.Read(readdata)
	fmt.Println(string(readdata[:count]))

	// websocket frame(messageの送信)
	p := []byte("Hi!")
	wh := webSocketHeader{}
	wh.fin = false
	wh.opcode = opText
	wh.payloadLength = int64(len(p))
	wh.masked = true
	// カギは本来乱数だが適当な固定値
	tk := []byte("b429a94c")
	wh.maskKey = binary.LittleEndian.Uint32(tk)
	f, _ := newFrame(wh, p)
	conn.Write(f)

	// websocket frame(FIN送信)(送信終了のお知らせ)
	p = make([]byte, 1)
	wh = webSocketHeader{}
	wh.fin = true
	wh.opcode = opContinuation
	wh.payloadLength = int64(len(p))
	wh.masked = true
	// カギは本来乱数だが適当な固定値
	tk = []byte("b429a94c")
	wh.maskKey = binary.LittleEndian.Uint32(tk)
	f, _ = newFrame(wh, p)
	conn.Write(f)

	// 受信処理なし
}

func newFrame(h webSocketHeader, payload []byte) ([]byte, error) {
	/*
		websocket frameを作成する関数
		paload が126以上は対応していない。
		ref: https://github.com/nhooyr/websocket/blob/v1.8.7/frame.go#L109
		fig: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
				      0                   1                   2                   3
			      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
			     +-+-+-+-+-------+-+-------------+-------------------------------+
			     |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
			     |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
			     |N|V|V|V|       |S|             |   (if payload len==126/127)   |
			     | |1|2|3|       |K|             |                               |
			     +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
			     |     Extended payload length continued, if payload len == 127  |
			     + - - - - - - - - - - - - - - - +-------------------------------+
			     |                               |Masking-key, if MASK set to 1  |
			     +-------------------------------+-------------------------------+
			     | Masking-key (continued)       |          Payload Data         |
			     +-------------------------------- - - - - - - - - - - - - - - - +
			     :                     Payload Data continued ...                :
			     + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
			     |                     Payload Data continued ...                |
			     +---------------------------------------------------------------+
	*/
	buf := []byte{}

	// 1byte目
	var b byte
	if h.fin {
		b |= 1 << 7
	}
	if h.rsv1 {
		b |= 1 << 6
	}
	if h.rsv2 {
		b |= 1 << 5
	}
	if h.rsv3 {
		b |= 1 << 4
	}
	b |= byte(h.opcode)
	buf = append(buf, b)

	// 2byte目
	lengthByte := byte(0)
	if h.masked {
		lengthByte |= 1 << 7
	}
	//payload len less than 7
	lengthByte |= byte(h.payloadLength)
	buf = append(buf, lengthByte)

	// 3~6byte目
	// masked payloadの投入
	if h.masked {
		keybuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(keybuf, h.maskKey)
		buf = append(buf, keybuf...)
		mask(h.maskKey, payload)
	}
	buf = append(buf, payload...)

	return buf, nil
}

func mask(key uint32, b []byte) uint32 {
	// payloadをマスクする関数
	// key を元にxorを繰り返す
	// copy from: https://github.com/nhooyr/websocket/blob/v1.8.7/frame.go#L185
	if len(b) >= 8 {
		key64 := uint64(key)<<32 | uint64(key)

		// At some point in the future we can clean these unrolled loops up.
		// See https://github.com/golang/go/issues/31586#issuecomment-487436401

		// Then we xor until b is less than 128 bytes.
		for len(b) >= 128 {
			v := binary.LittleEndian.Uint64(b)
			binary.LittleEndian.PutUint64(b, v^key64)
			v = binary.LittleEndian.Uint64(b[8:16])
			binary.LittleEndian.PutUint64(b[8:16], v^key64)
			v = binary.LittleEndian.Uint64(b[16:24])
			binary.LittleEndian.PutUint64(b[16:24], v^key64)
			v = binary.LittleEndian.Uint64(b[24:32])
			binary.LittleEndian.PutUint64(b[24:32], v^key64)
			v = binary.LittleEndian.Uint64(b[32:40])
			binary.LittleEndian.PutUint64(b[32:40], v^key64)
			v = binary.LittleEndian.Uint64(b[40:48])
			binary.LittleEndian.PutUint64(b[40:48], v^key64)
			v = binary.LittleEndian.Uint64(b[48:56])
			binary.LittleEndian.PutUint64(b[48:56], v^key64)
			v = binary.LittleEndian.Uint64(b[56:64])
			binary.LittleEndian.PutUint64(b[56:64], v^key64)
			v = binary.LittleEndian.Uint64(b[64:72])
			binary.LittleEndian.PutUint64(b[64:72], v^key64)
			v = binary.LittleEndian.Uint64(b[72:80])
			binary.LittleEndian.PutUint64(b[72:80], v^key64)
			v = binary.LittleEndian.Uint64(b[80:88])
			binary.LittleEndian.PutUint64(b[80:88], v^key64)
			v = binary.LittleEndian.Uint64(b[88:96])
			binary.LittleEndian.PutUint64(b[88:96], v^key64)
			v = binary.LittleEndian.Uint64(b[96:104])
			binary.LittleEndian.PutUint64(b[96:104], v^key64)
			v = binary.LittleEndian.Uint64(b[104:112])
			binary.LittleEndian.PutUint64(b[104:112], v^key64)
			v = binary.LittleEndian.Uint64(b[112:120])
			binary.LittleEndian.PutUint64(b[112:120], v^key64)
			v = binary.LittleEndian.Uint64(b[120:128])
			binary.LittleEndian.PutUint64(b[120:128], v^key64)
			b = b[128:]
		}

		// Then we xor until b is less than 64 bytes.
		for len(b) >= 64 {
			v := binary.LittleEndian.Uint64(b)
			binary.LittleEndian.PutUint64(b, v^key64)
			v = binary.LittleEndian.Uint64(b[8:16])
			binary.LittleEndian.PutUint64(b[8:16], v^key64)
			v = binary.LittleEndian.Uint64(b[16:24])
			binary.LittleEndian.PutUint64(b[16:24], v^key64)
			v = binary.LittleEndian.Uint64(b[24:32])
			binary.LittleEndian.PutUint64(b[24:32], v^key64)
			v = binary.LittleEndian.Uint64(b[32:40])
			binary.LittleEndian.PutUint64(b[32:40], v^key64)
			v = binary.LittleEndian.Uint64(b[40:48])
			binary.LittleEndian.PutUint64(b[40:48], v^key64)
			v = binary.LittleEndian.Uint64(b[48:56])
			binary.LittleEndian.PutUint64(b[48:56], v^key64)
			v = binary.LittleEndian.Uint64(b[56:64])
			binary.LittleEndian.PutUint64(b[56:64], v^key64)
			b = b[64:]
		}

		// Then we xor until b is less than 32 bytes.
		for len(b) >= 32 {
			v := binary.LittleEndian.Uint64(b)
			binary.LittleEndian.PutUint64(b, v^key64)
			v = binary.LittleEndian.Uint64(b[8:16])
			binary.LittleEndian.PutUint64(b[8:16], v^key64)
			v = binary.LittleEndian.Uint64(b[16:24])
			binary.LittleEndian.PutUint64(b[16:24], v^key64)
			v = binary.LittleEndian.Uint64(b[24:32])
			binary.LittleEndian.PutUint64(b[24:32], v^key64)
			b = b[32:]
		}

		// Then we xor until b is less than 16 bytes.
		for len(b) >= 16 {
			v := binary.LittleEndian.Uint64(b)
			binary.LittleEndian.PutUint64(b, v^key64)
			v = binary.LittleEndian.Uint64(b[8:16])
			binary.LittleEndian.PutUint64(b[8:16], v^key64)
			b = b[16:]
		}

		// Then we xor until b is less than 8 bytes.
		for len(b) >= 8 {
			v := binary.LittleEndian.Uint64(b)
			binary.LittleEndian.PutUint64(b, v^key64)
			b = b[8:]
		}
	}

	// Then we xor until b is less than 4 bytes.
	for len(b) >= 4 {
		v := binary.LittleEndian.Uint32(b)
		binary.LittleEndian.PutUint32(b, v^key)
		b = b[4:]
	}

	// xor remaining bytes.
	for i := range b {
		b[i] ^= byte(key)
		key = bits.RotateLeft32(key, -8)
	}

	return key
}
