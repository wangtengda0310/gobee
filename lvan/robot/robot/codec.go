package robot

import (
	"encoding/binary"
	"fmt"
)

const ver = 1

func decode(data []byte) (*pb.Respond, error) {
	if len(data) < 11 {
		return nil, fmt.Errorf("decode msg failed, data is too short")
	}

	var ver = data[0]
	messageId := binary.BigEndian.Uint32(data[1:5])
	seq := binary.BigEndian.Uint32(data[5:9])
	errCode := binary.BigEndian.Uint16(data[9:11])

	logger.Debug(msg: "respond", args...: "errCode", errCode, "seq", seq, "messageId", messageId)

	return &pb.Respond{
		SeqNo:   seq,
		ErrCode: pb.Errno(errCode),
		Message: &pb.MessageContent{
			MessageID: messageId,
			Payload:   data[11:],
		},
	}, nil
}

func encode(respond *pb.Request) []byte {
	ret := make([]byte, 9)
	ret[0] = ver

	binary.BigEndian.PutUint32(ret[1:5], respond.Message.MessageID)
	binary.BigEndian.PutUint32(ret[5:9], respond.SeqNo)

	ret = append(ret, respond.Message.Payload...)

	return ret
}
