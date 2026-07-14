package protocol

import (
	"io"
	"log"
	"net"
	"sync/atomic"
)

// 消息方向
const (
	DirClientToServer = "→" // 客户端→服务端
	DirServerToClient = "←" // 服务端→客户端
)

// 连接计数器
var ConnIDCounter atomic.Uint64

// LoginReqCallback 录制模式下解析到 LoginReq 时的回调
type LoginReqCallback func(connID uint64, payload []byte)

// WriteLocker 保护连接写入的互斥锁接口
type WriteLocker interface {
	Lock()
	Unlock()
}

// SeqIDRewriter 返回下一个代理级别的 seqId（递增）
type SeqIDRewriter func() uint32

// RelayAndParse 透传 TCP 数据并解析协议帧
// 从 src 读取数据，写入 dst，同时解析并输出协议信息
// recorder: 录制器（可选）
// onLoginParsed: 解析到 LoginReq 时的回调（可选，用于提取账号）
// writeMu: 保护 dst 写入的互斥锁（可选）
// seqRewriter: client→server 方向转发前的 seqId 重写器（可选，返回新 seqId）
//
// 当 seqRewriter 非 nil 且 isClientData=true 时，转发前会用返回的 seqId 替换 body 中的 seqId。
// LoginReq（MsgID=1）透传时不重写（seqId 由服务端忽略），但 seqRewriter 仍会被调用以保持计数器同步。
func RelayAndParse(src net.Conn, dst net.Conn, connID uint64, dir string, isClientData bool, recorder *Recorder, onLoginParsed LoginReqCallback, writeMu WriteLocker, seqRewriter SeqIDRewriter) int64 {
	var total int64

	for {
		// 读取帧头
		header := make([]byte, FrameHeaderSize)
		n, err := io.ReadFull(src, header)
		if err != nil {
			if err != io.EOF && total > 0 {
				log.Printf("[TCP] #%d %s 读取帧头结束: %v", connID, dir, err)
			}
			return total
		}
		total += int64(n)

		msgLen, _, herr := ParseFrameHeader(header)
		if herr != nil {
			log.Printf("[TCP] #%d %s 帧头解析失败: %v", connID, dir, herr)
			if writeMu != nil {
				writeMu.Lock()
			}
			_, _ = dst.Write(header)
			if writeMu != nil {
				writeMu.Unlock()
			}
			return total
		}

		// 读取消息体
		body := make([]byte, msgLen)
		n, err = io.ReadFull(src, body)
		if err != nil {
			log.Printf("[TCP] #%d %s 读取消息体失败: %v", connID, dir, err)
			if writeMu != nil {
				writeMu.Lock()
			}
			_, _ = dst.Write(header)
			if writeMu != nil {
				writeMu.Unlock()
			}
			return total
		}
		total += int64(n)

		// client→server 方向且启用了 seqId 重写：在转发前替换 seqId
		if isClientData && seqRewriter != nil {
			flags := header[3]
			newSeqID := seqRewriter()
			if _, err := RewriteSeqID(body, flags, newSeqID); err != nil {
				log.Printf("[TCP] #%d seqId 重写失败: %v", connID, err)
			}
		}

		// 转发（加写锁防止并发写入交错）
		if writeMu != nil {
			writeMu.Lock()
		}
		if _, werr := dst.Write(header); werr != nil {
			if writeMu != nil {
				writeMu.Unlock()
			}
			log.Printf("[TCP] #%d %s 写入帧头失败: %v", connID, dir, werr)
			return total
		}
		if _, werr := dst.Write(body); werr != nil {
			if writeMu != nil {
				writeMu.Unlock()
			}
			log.Printf("[TCP] #%d %s 写入消息体失败: %v", connID, dir, werr)
			return total
		}
		if writeMu != nil {
			writeMu.Unlock()
		}

		// 异步解析并输出（不影响转发性能）
		go func(hdr, bdy []byte) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[TCP] #%d %s panic: %v", connID, dir, r)
				}
			}()

			raw := make([]byte, len(hdr)+len(bdy))
			copy(raw, hdr)
			copy(raw[len(hdr):], bdy)

			frame, derr := DecodeFrame(raw, isClientData)
			if derr != nil {
				log.Printf("[TCP] #%d %s 解码失败: %v", connID, dir, derr)
				return
			}
			log.Print(FormatFrame(frame, connID, dir))

			// LoginReq：保存 payload + 通知回调
			if isClientData && frame.MsgID == 1 && recorder != nil {
				recorder.RecordLoginPayload(frame.Payload)
				log.Printf("[录制] LoginReq payload(%dB)", len(frame.Payload))
			}
			if isClientData && frame.MsgID == 1 && onLoginParsed != nil {
				onLoginParsed(connID, frame.Payload)
			}

			// 录制 Proto 消息（双向，两个方向都录）
			if recorder != nil && frame.MsgID >= 1000 {
				if err := recorder.RecordFrame(frame, dir); err != nil {
					log.Printf("[录制] 错误: %v", err)
				}
			}
		}(header, body)
	}
}
