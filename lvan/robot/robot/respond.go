package robot

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

type Respond struct {
	Reply   proto.Message
	Updates []proto.Message
	ErrCode pb.Errno
}
func (res *Respond) String() string { /* 张锦 */
	updates := make([]reflect.Name, len(res.Updates))
	for i, update := range res.Updates {
		updates[i] = update.ProtoReflect().Descriptor().Name()
	}

	name := "nil"
	if res.Reply != nil {
		name = string(res.Reply.ProtoReflect().Descriptor().Name())
	}

	return fmt.Sprintf(format: "[reply:%s updates:%v ErrCode:%s]",
		name, updates, res.ErrCode)
}

func (res *Respond) CheckErrCode(err pb.Errno) bool { return res.ErrCode == err }

func (res *Respond) CheckReply(reply proto.Message) bool { no usages  /* 张锦 */
	if reply == nil { return true }
	return pProto.Equal(res.Reply, reply)
}

func (res *Respond) CheckUpdates(updates []proto.Message) (bool, proto.Message) { no usages  /* 张锦 */
	for _, want := range updates {
		check := false
		for _, update := range res.Updates {
			if pProto.Equal(want, update) {
				check = true
				break
			}
		}

		if !check { return false, want }
	}

	return true, nil
}
