// 命令 echo-server：Buf demo 的服务端。
//
// 监听 TCP 端口，接收 EchoRequest/SumRequest，返回 EchoResponse/SumResponse。
// 使用 Buf 生成的 protobuf 代码（echopb）做序列化/反序列化。
//
// 启动：
//
//	go run ./examples/buf/server -port 19090
//
// 然后用 client 或 sniffer 连接。
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/wangtengda0310/gobee/pcap/examples/buf/frame"
	"github.com/wangtengda0310/gobee/pcap/examples/buf/proto/go/echopb"
	"google.golang.org/protobuf/proto"
)

func main() {
	port := flag.Int("port", 19090, "监听端口")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("Echo server listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

// handleConn 处理一个客户端连接：循环读帧 → 反序列化 → 处理 → 响应。
func handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Printf("[%s] connected", remote)

	for {
		msgType, payload, err := frame.ReadFrame(conn)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("[%s] disconnected", remote)
			} else {
				log.Printf("[%s] read frame: %v", remote, err)
			}
			return
		}

		switch msgType {
		case frame.MsgTypeEchoRequest:
			var req echopb.EchoRequest
			if err := proto.Unmarshal(payload, &req); err != nil {
				log.Printf("[%s] unmarshal EchoRequest: %v", remote, err)
				continue
			}
			log.Printf("[%s] EchoRequest: message=%q seq=%d", remote, req.GetMessage(), req.GetSeq())

			// 构造响应：原样回显 message + seq，加上服务端时间戳。
			resp := &echopb.EchoResponse{
				Message:    req.GetMessage(),
				Seq:        req.GetSeq(),
				ServerTime: time.Now().UnixMilli(),
			}
			respBytes, _ := proto.Marshal(resp)
			if err := frame.WriteFrame(conn, frame.MsgTypeEchoResponse, respBytes); err != nil {
				log.Printf("[%s] write EchoResponse: %v", remote, err)
				return
			}

		case frame.MsgTypeSumRequest:
			var req echopb.SumRequest
			if err := proto.Unmarshal(payload, &req); err != nil {
				log.Printf("[%s] unmarshal SumRequest: %v", remote, err)
				continue
			}
			var sum int32
			for _, n := range req.GetNumbers() {
				sum += n
			}
			log.Printf("[%s] SumRequest: %v → sum=%d", remote, req.GetNumbers(), sum)

			resp := &echopb.SumResponse{Sum: sum, Count: int32(len(req.GetNumbers()))}
			respBytes, _ := proto.Marshal(resp)
			if err := frame.WriteFrame(conn, frame.MsgTypeSumResponse, respBytes); err != nil {
				log.Printf("[%s] write SumResponse: %v", remote, err)
				return
			}

		default:
			log.Printf("[%s] unknown msgType: %d", remote, msgType)
		}
	}
}
