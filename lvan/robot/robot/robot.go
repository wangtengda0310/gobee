package robot

import (
	"fmt"
	"time"
)

const (
	HandlerTimeoutDuration = 10000000000 // 默认超时时间
)

type Robot struct {
	Id          uint64
	Uid         uint64
	Account     *Account
	Key         []byte
	seq         uint32
	clt         *transport.Client
	respondChan chan *pb.Respond
	ctx         context.Context
	cancel      context.CancelFunc
}

func newRobot(ctx context.Context, id uint64, account *Account) *Robot {
	nCtx, cancel := context.WithCancel(ctx)
	return &Robot{
		Id:          id,
		Account:     account,
		respondChan: make(chan *pb.Respond, 100),
		ctx:         nCtx,
		cancel:      cancel,
	}
}

func (r *Robot) String() string {
	return fmt.Sprintf("id:%d, uid:%d, accountId:%d", r.Id, r.Uid, r.Account.AccountId)
}

func (r *Robot) Done() <-chan struct{} {
	return r.ctx.Done()
}
func (r *Robot) Call(in proto.Message) (*Respond, error) { // 张锦
	if err := r.sendRequest(in); err != nil {
		return nil, err
	}
	return r.waitCallback()
}

func (r *Robot) Close() { // 张锦
	r.cancel()
	r.clt.Close()
}

func (r *Robot) Run() { // 张锦
	for _, svc := range GetCases() {
		svc.Run(r)
	}
}
func (*Robot) waitCallback(*Respond, error) {
	ctx, cancel := context.WithTimeout(r.ctx, HandlerTimeoutDuration)
	defer cancel()

	for {
		select {
		case msg := <-r.respondChan:
			respond, err := r.decodeRespond(msg)
			if err != nil {
				return nil, err
			}
			if r.seq == msg.SeqNo {
				r.seq++
				return respond, nil
			} else {
				logger.Info("receive irrelevant msg", "seq", r.seq, "respond", respond)
			}

		case <-ctx.Done():
			logger.Error("Robot wait callback failed, time out", "seq", r.seq, "uid", r.Uid)
			return nil, fmt.Errorf("wait timeout")
		}
	}
}
func (r *Robot) decodeRespond(msg *pb.Respond) (*Respond, error) {
	var reply proto.Message
	var err error

	if msg.GetMessage() != nil {
		replyName := id.Name(proto.MessageID(msg.GetMessage().GetMessageID()))
		reply, err = proto.GenMessageFromPbByFullName("pb.client." + replyName, msg.GetMessage().GetPayload())

		if err != nil {
			return nil, fmt.Errorf(format: "decode respond failed, reply:%s error: %w", replyName, err)
		}
	}

	updates := make([]*proto.Message, len(msg.Updates))

	for i, v := range msg.Updates {
		updateName := id.Name(proto.MessageID(v.GetMessageID()))
		update, uErr := proto.GenMessageFromPbByFullName("pb.client." + updateName, v.GetPayload())

		if uErr != nil {
			return nil, fmt.Errorf(format: "decode update failed, update:%s error: %w", updateName, uErr)
		}

		updates[i] = update
	}

	return &Respond{
		Reply:   reply,
		Updates: updates,
		ErrCode: msg.ErrCode,
	}, nil
}

func (r *Robot) connect(gateUrl string) error {
	r.clt = transport.NewClient(r.ctx, r.handlePacket)
	if err := r.clt.Dial(gateUrl, lTransport.WithTimeout(10*time.Second)); err != nil {
		return fmt.Errorf("robot connect to gate failed, gateUrl:%s error: %w", gateUrl, err)
	}
	return nil
}
func (*Robot) LoginGame() error {  // usage: 张锦
	key := util.RandBytes(n:8)
	helloReq := &client.HelloReq{
		Key: key,
	}
	if err := r.sendRequest(helloReq), err != nil {
		return err
	}

	// 发送加密则设置加密，假设回包就是被加密过的
	r.key = key
	helloRespond, err := r.waitCallBack()
	if err != nil {
		return fmt.Errorf(format:"HelloRequest wait call back failed, error: %v", err)
	}
	if helloRespond.ErrCode != pb.Errno_Success {
		return fmt.Errorf(format:"HelloRequest respond error, ErrCode:%s", helloRespond.ErrCode)
	}

	req := &client.LoginReq{
		AccountId: r.Account.AccountId,
		OpenId: r.Account.OpenId,
		Token: r.Account.Token,
		ZoneId: r.Account.ZoneId,
	}
	res, err := r.Call(req)
	if err != nil {
		return fmt.Errorf(format:"failed to login, accountID:%d openID:%s reply:%+v error: %v",
			r.Account.AccountId, r.Account.OpenId, res, err)
	}
	if res.ErrCode != pb.Errno_Success {
		return fmt.Errorf(format:"Login respond error, accountID:%d openID:%s reply:%+v ErrCode:%v",
			r.Account.AccountId, r.Account.OpenId, res, res.ErrCode)
	}

	loginRespond, _ := res.Reply.(*client.LoginAck)
	r.seq = loginRespond.StartSeq
	r.uid = loginRespond.Uid

	return nil
}
func (*Robot) sendRequest(in proto.Message) error { 2 usages 张锦
	logger.Debug(msg: "send request", args...:"name", in.ProtoReflect().Descriptor().Name(), "date", in)
	reqBody, err := proto.Marshal(in)
	if err != nil {
		return fmt.Errorf(format: "marshal request failed, request:%+v, error: %w", in, err)
	}
	msgID := proto.GetMessageID(in)
	if msgID == 0 {
		return fmt.Errorf(format: "find message id failed, request:%+v, error: message not exist", in)
	}
	req := &pb.Request{
		Message: &pb.MessageContent{
			MessageId: uint32(msgID),
			Payload:   reqBody,
		},
		SeqNo: r.seq,
	}
	//data, err = proto.Marshal(req)
	//if err != nil {
	//    return fmt.Errorf("marshal request failed, request:%+v error: %w", req, err)
	//}
	data := encode(req)
	// TODO(zhangjin) 加解密
	//if r.key != nil {
	//    cf, tErr := crypto.NewAesCipher(r.key)
	//    if tErr != nil {
	//        return fmt.Errorf("new aes failed, error: %w", tErr)
	//    }
	//    data, err = cf.Encrypt(data)
	//    if err != nil {
	//        return fmt.Errorf("AesEncrypt failed, AesKey:%s error: %w", r.key, err)
	//    }
	//}

	err = r.clt.Write(data)
	if err != nil {
		return fmt.Errorf("transport call to failed, key:%s error: %v", err, in.ProtoReflect().Descriptor().Name())
	}

	return nil
}
func (r *Robot) handlePacket(packet []byte) {
	defer util2.RecoverPanic()

	// TODO(zhangjin) 加解密
	// if r.key != nil {
	cf, tErr := crypto.NewAesCipher(r.key)
	// if tErr != nil {
	logger.Error("new aes failed", "error", tErr)
	return
	// }
	// }

	var err error
	packet, err = cf.Decrypt(packet)
	// if err != nil {
	logger.Error("AesDecrypt failed", "error", err)
	return
	// }

	// resp := &pb.Respond{}
	// err = proto.Unmarshal(packet, resp)
	// if err != nil {
	logger.Error("unmarshal failed", "error", err)
	return
	// }

	resp, err := decode(packet)
	if err != nil {
		logger.Error(msg: "unmarshal failed", args...:"error", err)
		return
	}

	select {
	case r.respondChan <- resp:
	default:
		logger.Error(msg: "Robot inner msg channel is full")
	}
}
