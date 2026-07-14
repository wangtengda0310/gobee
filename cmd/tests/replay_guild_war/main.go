// 工会城战用例重放工具（单文件，逻辑自包含）
//
// TeamMatchGuildCityReq 后服务端经 TransportRawNtf 推送 PveGuildCityDataNtf，
// 解析 matchedCities 并动态改写 TeamSelectGuildCityReq.cityId。
//
// 用法:
//
//	go run ./cmd/tests/replay_guild_war -openid test1
//	go run ./cmd/tests/replay_guild_war -case cases/proto_cases/工会战.json -openid test1 -repeat 100
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	sp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
)

const (
	defaultSendIntervalMs = 1000
	defaultAckWaitMs      = 2000
	defaultMatchWait      = 15 * time.Second
	defaultAckWait        = 3 * time.Second

	msgIDTransportRawNtf        = uint16(pb.EGameMsgID_TransportRawNtf_id)
	msgIDPveGuildCityDataNtf    = uint16(pb.EGameMsgID_PveGuildCityDataNtf_id)
	msgIDTeamMatchGuildCityReq  = uint16(pb.EGameMsgID_TeamMatchGuildCityReq_id)
	msgIDTeamSelectGuildCityReq = uint16(pb.EGameMsgID_TeamSelectGuildCityReq_id)
	msgIDResponseErrorNtf       = uint16(pb.EGameMsgID_ResponseErrorNtf_id)
)

type replayOptions struct {
	caseFile         string
	openID           string
	httpAddr         string
	serverAddr       string
	sendInterval     time.Duration
	matchWaitTimeout time.Duration
	ackWaitTimeout   time.Duration
	printAckJSON     bool
	loginPayloadB64  string
}

type caseMessage struct {
	msgID       uint16
	msgName     string
	payloadJSON string
	direction   string
}

type frameMux struct {
	conn   net.Conn
	frames chan *sp.DecodedFrame
	done   chan struct{}
	once   sync.Once
}

func main() {
	caseFile := flag.String("case", "cases/proto_cases/工会战.json", "工会城战用例 JSON 路径")
	openID := flag.String("openid", "", "登录账号（默认 replay_gw_<时间戳>）")
	httpAddr := flag.String("http", "10.254.114.204:20144", "HTTP 认证地址")
	serverAddr := flag.String("tcp", "", "TCP 地址（默认从用例文件 server_addr 读取）")
	intervalMs := flag.Int("interval", defaultSendIntervalMs, "消息发送间隔（毫秒），0 表示不等待")
	matchWaitSec := flag.Int("match-wait", 15, "等待 PveGuildCityDataNtf 超时（秒）")
	repeat := flag.Int("repeat", 1, "重复执行次数")
	printAck := flag.Bool("print-ack", true, "将每条 Req 对应的 Ack 以 JSON 输出到控制台")
	ackWaitMs := flag.Int("ack-wait", 3000, "等待 Ack 超时（毫秒）")
	flag.Parse()

	account := *openID
	if account == "" {
		account = fmt.Sprintf("replay_gw_%d", time.Now().Unix()%100000)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	opts := replayOptions{
		caseFile:         *caseFile,
		openID:           account,
		httpAddr:         *httpAddr,
		serverAddr:       *serverAddr,
		sendInterval:     time.Duration(*intervalMs) * time.Millisecond,
		matchWaitTimeout: time.Duration(*matchWaitSec) * time.Second,
		ackWaitTimeout:   time.Duration(*ackWaitMs) * time.Millisecond,
		printAckJSON:     *printAck,
	}

	var ok, fail int
	start := time.Now()
	for i := 1; i <= *repeat; i++ {
		if *repeat > 1 {
			log.Printf("[工会战重放] ===== 第 %d/%d 轮 =====", i, *repeat)
		}
		if err := replayGuildCityCase(opts); err != nil {
			fail++
			log.Printf("[工会战重放] 第 %d/%d 轮失败: %v", i, *repeat, err)
			continue
		}
		ok++
		if *repeat > 1 {
			log.Printf("[工会战重放] 第 %d/%d 轮成功", i, *repeat)
		}
	}

	if *repeat > 1 {
		log.Printf("[工会战重放] 汇总: 成功=%d 失败=%d 总计=%d 耗时=%v", ok, fail, *repeat, time.Since(start).Round(time.Millisecond))
	}
	if fail > 0 {
		os.Exit(1)
	}
}

func replayGuildCityCase(opts replayOptions) error {
	if opts.caseFile == "" {
		return fmt.Errorf("caseFile 不能为空")
	}
	if opts.openID == "" {
		return fmt.Errorf("openID 不能为空")
	}
	if opts.httpAddr == "" {
		return fmt.Errorf("httpAddr 不能为空")
	}
	if opts.matchWaitTimeout <= 0 {
		opts.matchWaitTimeout = defaultMatchWait
	}
	if opts.ackWaitTimeout <= 0 {
		opts.ackWaitTimeout = defaultAckWait
	}

	rec, err := prototest.LoadTestCaseFromFile(opts.caseFile)
	if err != nil {
		return err
	}
	messages := recordingToMessages(rec)
	if len(messages) == 0 {
		return fmt.Errorf("用例无消息: %s", opts.caseFile)
	}

	serverAddr := opts.serverAddr
	if serverAddr == "" {
		serverAddr = rec.ServerAddr
	}

	log.Printf("[工会战重放] 用例=%s 账号=%s TCP=%s HTTP=%s 消息数=%d",
		opts.caseFile, opts.openID, serverAddr, opts.httpAddr, len(messages))

	token, authServerAddr, err := sp.AuthLogin(opts.httpAddr, opts.openID)
	if err != nil {
		return fmt.Errorf("HTTP 登录失败: %w", err)
	}
	if serverAddr == "" {
		serverAddr = authServerAddr
	}
	if serverAddr == "" {
		return fmt.Errorf("无法确定 TCP 服务器地址")
	}

	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("TCP 连接失败: %w", err)
	}
	defer conn.Close()

	loginPayload := opts.loginPayloadB64
	if loginPayload == "" {
		loginPayload = rec.LoginPayloadB64
	}
	if err := sendLoginReq(conn, opts.openID, token, loginPayload); err != nil {
		return fmt.Errorf("发送 LoginReq 失败: %w", err)
	}
	if err := waitLoginResp(conn); err != nil {
		return fmt.Errorf("等待 LoginResp 失败: %w", err)
	}
	log.Printf("[工会战重放] 登录成功")

	mux := newFrameMux(conn)
	mux.start()
	defer mux.stop()

	stopHeartbeat := make(chan struct{})
	go heartbeat(conn, stopHeartbeat)
	defer close(stopHeartbeat)

	time.Sleep(2 * time.Second)

	seqID := uint32(1)
	var selectedCityID uint64

	for i, msg := range messages {
		if msg.direction != "" && msg.direction != "→" {
			log.Printf("[工会战重放] [%d/%d] 跳过 %s direction=%q", i+1, len(messages), msg.msgName, msg.direction)
			continue
		}

		payloadJSON := msg.payloadJSON
		if msg.msgID == msgIDTeamSelectGuildCityReq {
			if selectedCityID == 0 {
				return fmt.Errorf("TeamSelectGuildCityReq 前未收到 PveGuildCityDataNtf，无法选城")
			}
			patched, patchErr := patchTeamSelectGuildCityPayload(payloadJSON, selectedCityID)
			if patchErr != nil {
				return patchErr
			}
			log.Printf("[工会战重放] TeamSelectGuildCityReq 使用动态 cityId=%d", selectedCityID)
			payloadJSON = patched
		}

		if err := sendRawMessage(conn, msg.msgID, payloadJSON, seqID); err != nil {
			return fmt.Errorf("发送 %s 失败: %w", msg.msgName, err)
		}
		log.Printf("[工会战重放] [%d/%d] → %s (MsgID=%d, SeqID=%d)", i+1, len(messages), msg.msgName, msg.msgID, seqID)
		seqID++

		if msg.msgID == msgIDTeamMatchGuildCityReq {
			ntf, waitErr := mux.waitPveGuildCityData(opts.matchWaitTimeout)
			if waitErr != nil {
				return fmt.Errorf("等待 PveGuildCityDataNtf 失败: %w", waitErr)
			}
			cityID, pickErr := pickSelectableCityID(ntf)
			if pickErr != nil {
				return pickErr
			}
			selectedCityID = cityID
			log.Printf("[工会战重放] 已解析城池列表（TransportRawNtf/PveGuildCityDataNtf），matched=%d 选用 cityId=%d",
				len(ntf.MatchedCities), selectedCityID)
		} else if opts.printAckJSON {
			if ackID, ok := expectedAckMsgID(msg.msgID); ok {
				frame, waitErr := mux.waitFrameMsgID(ackID, opts.ackWaitTimeout)
				if waitErr != nil {
					log.Printf("[工会战重放] 等待 %s 失败: %v", sp.GetMsgName(ackID), waitErr)
				} else if jsonStr, jsonErr := framePayloadToJSON(frame.MsgID, frame.Payload); jsonErr != nil {
					log.Printf("[工会战重放] ← %s 解析失败: %v", sp.GetMsgName(ackID), jsonErr)
				} else {
					log.Printf("[工会战重放] ← %s (MsgID=%d, SeqID=%d)\n%s",
						sp.GetMsgName(ackID), frame.MsgID, frame.SeqID, jsonStr)
					if errCode := extractAckErrCode(frame.MsgID, frame.Payload); errCode != 0 {
						log.Printf("[工会战重放] ⚠ %s errCode=%d", sp.GetMsgName(ackID), errCode)
					}
				}
			}
		}

		if opts.sendInterval > 0 {
			time.Sleep(opts.sendInterval)
		}
	}

	time.Sleep(defaultAckWaitMs * time.Millisecond)
	log.Printf("[工会战重放] 完成")
	return nil
}

func recordingToMessages(rec *sp.Recording) []caseMessage {
	out := make([]caseMessage, 0, len(rec.Messages))
	for _, e := range rec.Messages {
		if e == nil || !prototest.IsReqDirection(e.Direction) {
			continue
		}
		out = append(out, caseMessage{
			msgID:       e.MsgID,
			msgName:     e.MsgName,
			payloadJSON: string(e.PayloadJSON),
			direction:   e.Direction,
		})
	}
	return out
}

// --- frame mux ---

func newFrameMux(conn net.Conn) *frameMux {
	return &frameMux{
		conn:   conn,
		frames: make(chan *sp.DecodedFrame, 256),
		done:   make(chan struct{}),
	}
}

func (m *frameMux) start() {
	go func() {
		for {
			select {
			case <-m.done:
				return
			default:
			}

			frame, err := readFrameFromConn(m.conn)
			if err != nil {
				select {
				case <-m.done:
				default:
					log.Printf("[frame_mux] 读取退出: %v", err)
				}
				return
			}

			m.logFrame(frame)

			select {
			case m.frames <- frame:
			case <-m.done:
				return
			}
		}
	}()
}

func (m *frameMux) stop() {
	m.once.Do(func() { close(m.done) })
}

func (m *frameMux) logFrame(frame *sp.DecodedFrame) {
	msgName := sp.GetMsgName(frame.MsgID)
	if msgName == "" || msgName == "Unknown" {
		switch frame.MsgID {
		case 2:
			msgName = "LoginResp"
		case 4:
			msgName = "Pong"
		default:
			msgName = fmt.Sprintf("Unknown(%d)", frame.MsgID)
		}
	}

	if frame.MsgID == msgIDTransportRawNtf {
		if ntf, err := parsePveGuildCityDataFromFrame(frame); err == nil && ntf != nil {
			log.Printf("[frame_mux] ← TransportRawNtf → PveGuildCityDataNtf (%d 座城池)", len(ntf.MatchedCities))
			return
		}
		transport, err := unmarshalProtoPayload(frame.MsgID, frame.Payload)
		if err == nil {
			raw := transport.(*pb.TransportRawNtf)
			innerName := sp.GetMsgName(uint16(raw.GetMsgId()))
			if innerName == "" || innerName == "Unknown" {
				innerName = fmt.Sprintf("MsgID(%d)", raw.GetMsgId())
			}
			log.Printf("[frame_mux] ← TransportRawNtf(inner=%s, %dB)", innerName, len(raw.Data))
			return
		}
	}

	log.Printf("[frame_mux] ← %s (MsgID=%d, SeqID=%d, %dB)", msgName, frame.MsgID, frame.SeqID, len(frame.Payload))
}

func (m *frameMux) waitPveGuildCityData(timeout time.Duration) (*pb.PveGuildCityDataNtf, error) {
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("等待 PveGuildCityDataNtf 超时 (%v)", timeout)
		}
		select {
		case frame := <-m.frames:
			ntf, err := parsePveGuildCityDataFromFrame(frame)
			if err != nil {
				return nil, err
			}
			if ntf != nil {
				return ntf, nil
			}
		case <-time.After(remaining):
			return nil, fmt.Errorf("等待 PveGuildCityDataNtf 超时 (%v)", timeout)
		case <-m.done:
			return nil, fmt.Errorf("frame mux 已停止")
		}
	}
}

func expectedAckMsgID(reqMsgID uint16) (uint16, bool) {
	ackID := reqMsgID + 1
	if strings.HasSuffix(sp.GetMsgName(ackID), "Ack") {
		return ackID, true
	}
	return 0, false
}

func (m *frameMux) waitFrameMsgID(msgID uint16, timeout time.Duration) (*sp.DecodedFrame, error) {
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("等待 %s 超时 (%v)", sp.GetMsgName(msgID), timeout)
		}
		select {
		case frame := <-m.frames:
			if frame.MsgID == msgIDResponseErrorNtf {
				if jsonStr, err := framePayloadToJSON(frame.MsgID, frame.Payload); err == nil {
					log.Printf("[工会战重放] ← ResponseErrorNtf\n%s", jsonStr)
				}
				continue
			}
			if frame.MsgID != msgID {
				continue
			}
			return frame, nil
		case <-time.After(remaining):
			return nil, fmt.Errorf("等待 %s 超时 (%v)", sp.GetMsgName(msgID), timeout)
		case <-m.done:
			return nil, fmt.Errorf("frame mux 已停止")
		}
	}
}

func readFrameFromConn(conn net.Conn) (*sp.DecodedFrame, error) {
	header := make([]byte, sp.FrameHeaderSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("读取帧头失败: %w", err)
	}
	msgLen, _, err := sp.ParseFrameHeader(header)
	if err != nil {
		return nil, fmt.Errorf("解析帧头失败: %w", err)
	}
	body := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, fmt.Errorf("读取消息体失败: %w", err)
	}
	raw := make([]byte, sp.FrameHeaderSize+msgLen)
	copy(raw, header)
	copy(raw[sp.FrameHeaderSize:], body)
	frame, err := sp.DecodeFrame(raw, false)
	if err != nil {
		return nil, fmt.Errorf("解码失败: %w", err)
	}
	return frame, nil
}

// --- guild city ---

func stripByteStreamPrefix(payload []byte) []byte {
	if len(payload) >= 2 {
		dataLen := int(payload[0]) | int(payload[1])<<8
		if dataLen > 0 && dataLen <= len(payload)-2 {
			return payload[2 : 2+dataLen]
		}
	}
	return payload
}

func unmarshalProtoPayload(msgID uint16, payload []byte) (proto.Message, error) {
	msg, ok := sp.NewMessage(msgID)
	if !ok {
		return nil, fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	protoData := stripByteStreamPrefix(payload)
	if err := proto.Unmarshal(protoData, msg); err != nil {
		return nil, fmt.Errorf("反序列化 MsgID=%d 失败: %w", msgID, err)
	}
	return msg, nil
}

func parsePveGuildCityDataFromFrame(frame *sp.DecodedFrame) (*pb.PveGuildCityDataNtf, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame 为空")
	}
	if frame.MsgID == msgIDPveGuildCityDataNtf {
		msg, err := unmarshalProtoPayload(frame.MsgID, frame.Payload)
		if err != nil {
			return nil, err
		}
		return msg.(*pb.PveGuildCityDataNtf), nil
	}
	if frame.MsgID != msgIDTransportRawNtf {
		return nil, nil
	}
	transport, err := unmarshalProtoPayload(frame.MsgID, frame.Payload)
	if err != nil {
		return nil, fmt.Errorf("解析 TransportRawNtf 失败: %w", err)
	}
	rawNtf := transport.(*pb.TransportRawNtf)
	if rawNtf.GetMsgId() != uint32(msgIDPveGuildCityDataNtf) {
		return nil, nil
	}
	inner := &pb.PveGuildCityDataNtf{}
	innerData := stripByteStreamPrefix(rawNtf.Data)
	if err := proto.Unmarshal(innerData, inner); err != nil {
		if err2 := proto.Unmarshal(rawNtf.Data, inner); err2 != nil {
			return nil, fmt.Errorf("解析 TransportRawNtf 内层 PveGuildCityDataNtf 失败: %w", err)
		}
	}
	return inner, nil
}

func pickSelectableCityID(ntf *pb.PveGuildCityDataNtf) (uint64, error) {
	if ntf == nil || len(ntf.MatchedCities) == 0 {
		return 0, fmt.Errorf("matchedCities 为空")
	}
	for _, city := range ntf.MatchedCities {
		if city == nil || city.CityInfo == nil || city.CityInfo.SimpleInfo == nil {
			continue
		}
		if !city.IsAttack {
			return city.CityInfo.SimpleInfo.Id, nil
		}
	}
	for _, city := range ntf.MatchedCities {
		if city != nil && city.CityInfo != nil && city.CityInfo.SimpleInfo != nil {
			return city.CityInfo.SimpleInfo.Id, nil
		}
	}
	return 0, fmt.Errorf("matchedCities 中无有效城池")
}

func patchTeamSelectGuildCityPayload(payloadJSON string, cityID uint64) (string, error) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &fields); err != nil {
		return "", fmt.Errorf("解析 TeamSelectGuildCityReq JSON 失败: %w", err)
	}
	fields["cityId"] = cityID
	out, err := json.Marshal(fields)
	if err != nil {
		return "", fmt.Errorf("编码 TeamSelectGuildCityReq JSON 失败: %w", err)
	}
	return string(out), nil
}

// --- ack json ---

type protoErrCoder interface {
	GetErrCode() uint32
}

func extractAckErrCode(msgID uint16, payload []byte) uint32 {
	msg, err := unmarshalProtoPayload(msgID, payload)
	if err != nil {
		return 0
	}
	if ec, ok := msg.(protoErrCoder); ok {
		return ec.GetErrCode()
	}
	return 0
}

func extractProtoData(payload []byte) []byte {
	if len(payload) == 0 {
		return nil
	}
	if len(payload) < 2 {
		return payload
	}
	dataLen := int(payload[0]) | int(payload[1])<<8
	if dataLen > len(payload)-2 {
		return nil
	}
	return payload[2 : 2+dataLen]
}

func framePayloadToJSON(msgID uint16, payload []byte) (string, error) {
	msg, ok := sp.NewMessage(msgID)
	if !ok {
		return "", fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	protoData := extractProtoData(payload)
	if protoData == nil {
		return "", fmt.Errorf("payload 为空或格式无效")
	}
	if len(protoData) == 0 {
		return "{}", nil
	}
	if err := proto.Unmarshal(protoData, msg); err != nil {
		return "", fmt.Errorf("反序列化失败: %w", err)
	}
	jsonBytes, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// --- login / send ---

func sendRawMessage(conn net.Conn, msgID uint16, payloadJSON string, seqID uint32) error {
	frame, err := sp.EncodeClientMessage(msgID, seqID, payloadJSON)
	if err != nil {
		return err
	}
	if _, err := conn.Write(frame); err != nil {
		return fmt.Errorf("TCP 发送失败: %v", err)
	}
	return nil
}

func sendLoginReq(conn net.Conn, openID, token, loginPayloadB64 string) error {
	var payload []byte
	if loginPayloadB64 != "" {
		savedPayload, err := base64.StdEncoding.DecodeString(loginPayloadB64)
		if err != nil {
			return fmt.Errorf("解码 LoginReq payload 失败: %v", err)
		}
		if len(savedPayload) < 2 {
			return fmt.Errorf("LoginReq payload 太短")
		}
		accountLen := int(savedPayload[0]) | int(savedPayload[1])<<8
		tokenOffset := 2 + accountLen
		if tokenOffset+2 > len(savedPayload) {
			return fmt.Errorf("LoginReq payload 格式错误")
		}
		oldTokenLen := int(savedPayload[tokenOffset]) | int(savedPayload[tokenOffset+1])<<8
		var buf bytes.Buffer
		buf.Write(savedPayload[:tokenOffset])
		writeString(&buf, token)
		buf.Write(savedPayload[tokenOffset+2+oldTokenLen:])
		payload = buf.Bytes()
		log.Printf("[重放] 使用录制的 LoginReq payload (%d 字节)，已替换 token", len(payload))
	} else {
		var buf bytes.Buffer
		writeString(&buf, openID)
		writeString(&buf, token)
		binary.Write(&buf, binary.LittleEndian, uint64(0))
		writeString(&buf, "0.8.3001.1.1")
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint32(0))
		writeString(&buf, "")
		writeString(&buf, "")
		payload = buf.Bytes()
		log.Printf("[重放] 使用自行序列化的 LoginReq (%d 字节)", len(payload))
	}
	frame := sp.EncodeFrame(1, 0, sp.FlagEncrypt, payload, true)
	if _, err := conn.Write(frame); err != nil {
		return fmt.Errorf("TCP 发送失败: %v", err)
	}
	return nil
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

func waitLoginResp(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		header := make([]byte, sp.FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			return fmt.Errorf("读取帧头失败: %v", err)
		}
		msgLen, _, err := sp.ParseFrameHeader(header)
		if err != nil {
			return fmt.Errorf("解析帧头失败: %v", err)
		}
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			return fmt.Errorf("读取消息体失败: %v", err)
		}
		raw := make([]byte, sp.FrameHeaderSize+msgLen)
		copy(raw, header)
		copy(raw[sp.FrameHeaderSize:], body)
		frame, err := sp.DecodeFrame(raw, false)
		if err != nil {
			return fmt.Errorf("解码失败: %v", err)
		}
		if frame.MsgID == 2 {
			if len(frame.Payload) >= 12 {
				result := binary.LittleEndian.Uint32(frame.Payload[8:12])
				if result != 0 {
					return fmt.Errorf("登录失败: Result=%d", result)
				}
			}
			return nil
		}
		msgName := sp.GetMsgName(frame.MsgID)
		if msgName == "" || msgName == "Unknown" {
			switch frame.MsgID {
			case 4:
				msgName = "Pong"
			case 10:
				msgName = "KickOut"
			default:
				msgName = fmt.Sprintf("Unknown(%d)", frame.MsgID)
			}
		}
		log.Printf("[重放] 跳过消息，等待 LoginResp: %s (MsgID=%d)", msgName, frame.MsgID)
	}
}

func heartbeat(conn net.Conn, stop chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			pingFrame := sp.EncodeFrame(3, 0, sp.FlagEncrypt, []byte{}, true)
			if _, err := conn.Write(pingFrame); err != nil {
				log.Printf("[重放] 心跳发送失败: %v", err)
				return
			}
			log.Printf("[重放] 心跳已发送")
		}
	}
}
