// 命令 echo-client：Buf demo 的客户端。
//
// 连接 echo-server，发送 EchoRequest 和 SumRequest，打印响应。
// 演示如何用 Buf 生成的 protobuf 代码做序列化。
//
// 启动（先启动 server）：
//
//	go run ./examples/buf/client -addr localhost:19090
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/wangtengda0310/gobee/pcap/examples/buf/frame"
	"github.com/wangtengda0310/gobee/pcap/examples/buf/proto/go/echopb"
	"google.golang.org/protobuf/proto"
)

func main() {
	addr := flag.String("addr", "localhost:19090", "服务端地址")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	log.Printf("connected to %s", *addr)

	// --- 发送 3 个 EchoRequest ---
	for i := 1; i <= 3; i++ {
		req := &echopb.EchoRequest{
			Message: fmt.Sprintf("hello from client #%d", i),
			Seq:     uint32(i),
		}
		payload, _ := proto.Marshal(req)
		if err := frame.WriteFrame(conn, frame.MsgTypeEchoRequest, payload); err != nil {
			log.Fatalf("write EchoRequest: %v", err)
		}
		log.Printf("→ EchoRequest: message=%q seq=%d", req.GetMessage(), req.GetSeq())

		// 读取响应。
		msgType, respPayload, err := frame.ReadFrame(conn)
		if err != nil {
			log.Fatalf("read response: %v", err)
		}
		if msgType != frame.MsgTypeEchoResponse {
			log.Fatalf("unexpected msgType: %d", msgType)
		}
		var resp echopb.EchoResponse
		if err := proto.Unmarshal(respPayload, &resp); err != nil {
			log.Fatalf("unmarshal EchoResponse: %v", err)
		}
		log.Printf("← EchoResponse: message=%q seq=%d server_time=%d",
			resp.GetMessage(), resp.GetSeq(), resp.GetServerTime())

		time.Sleep(500 * time.Millisecond)
	}

	// --- 发送 1 个 SumRequest ---
	sumReq := &echopb.SumRequest{Numbers: []int32{10, 20, 30, 40, 50}}
	payload, _ := proto.Marshal(sumReq)
	if err := frame.WriteFrame(conn, frame.MsgTypeSumRequest, payload); err != nil {
		log.Fatalf("write SumRequest: %v", err)
	}
	log.Printf("→ SumRequest: numbers=%v", sumReq.GetNumbers())

	msgType, respPayload, err := frame.ReadFrame(conn)
	if err != nil {
		log.Fatalf("read response: %v", err)
	}
	if msgType != frame.MsgTypeSumResponse {
		log.Fatalf("unexpected msgType: %d", msgType)
	}
	var sumResp echopb.SumResponse
	if err := proto.Unmarshal(respPayload, &sumResp); err != nil {
		log.Fatalf("unmarshal SumResponse: %v", err)
	}
	log.Printf("← SumResponse: sum=%d count=%d", sumResp.GetSum(), sumResp.GetCount())

	log.Println("done")
}
