// 监听tcp7324端口，将接收到的客户端数据以16进制打印在控制台后将源数据转发到logon.5awow.com，将logon.5awow.com返回的数据以16进制打印在控制台后将返回的源数据返回客户端

package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
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

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(clientConn)
	}
}
func closeConnection(clientConn net.Conn) {
	err := clientConn.Close()
	if err != nil {
		log.Fatal(err)
	}
}
func handleConnection(clientConn net.Conn) {
	defer closeConnection(clientConn)

	serverConn, err := net.Dial("tcp", loginServer)
	fmt.Println("Connected to logon.5awow.com:3724")
	if err != nil {
		log.Fatal(err)
	}
	defer closeConnection(serverConn)
	for {
		go copyAndPrint(serverConn, clientConn, "src")
		copyAndPrint(clientConn, serverConn, "dst")
	}
}

var timestamp = time.Now().Format("20060102150405")

func copyAndPrint(dst net.Conn, src net.Conn, prefix string) {
	log.Println("\n\n\nCopying from", src.RemoteAddr(), "to", dst.RemoteAddr())
	buf := make([]byte, 1024)
	for {
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

		_, err = dst.Write(data)
		if err != nil {
			log.Fatal(err)
		}
	}
}
