// 监听tcp7324端口，将接收到的客户端数据以16进制打印在控制台后将源数据转发到logon.5awow.com，将logon.5awow.com返回的数据以16进制打印在控制台后将返回的源数据返回客户端

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const loginServer = "logon.5awow.com:3724"

var (
	messageID int
	mu        sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":3724")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Listening on :3724")

	ctx, cancelApplication := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGILL, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		listener.Close()
		cancelApplication()
	}()
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(ctx, clientConn)
	}
}
func closeConnection(clientConn net.Conn) {
	err := clientConn.Close()
	if err != nil {
		log.Fatal(err)
	}
}
func handleConnection(ctx context.Context, clientConn net.Conn) {
	defer closeConnection(clientConn)

	serverConn, err := net.Dial("tcp", loginServer)
	if err != nil {
		log.Fatal(err)
	}
	defer closeConnection(serverConn)

	fmt.Println("Connected to logon.5awow.com:3724")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { wg.Done(); copyAndPrint(ctx, serverConn, clientConn, "src") }()
	go func() { wg.Done(); copyAndPrint(ctx, clientConn, serverConn, "dst") }()
	wg.Wait()
}

var timestamp = time.Now().Format("20060102150405")

func copyAndPrint(ctx context.Context, dst net.Conn, src net.Conn, prefix string) {
	log.Println("\n\n\nCopying from", src.RemoteAddr(), "to", dst.RemoteAddr())
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			break
		}
		log.Println("Reading from", src.RemoteAddr())
		n, err := src.Read(buf)
		log.Println("Read", n, "bytes")
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		data := buf[:n]
		base64Data := base64.StdEncoding.EncodeToString(data)

		// Write to file
		filename := fmt.Sprintf("%s_%s.txt", prefix, timestamp)
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		mu.Lock()
		messageID++
		_, err = file.WriteString(fmt.Sprintf("%d\n%s\n", messageID, base64Data))
		mu.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		_, err = dst.Write(data)
		if err != nil {
			log.Fatal(err)
		}
	}
}
