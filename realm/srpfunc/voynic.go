package srpfunc

import (
	"wow"
)

type myStruct struct {
}

func (m *myStruct) Challenge(request wow.LoginChallengeRequest) wow.LoginChallengeResponse {
	//TODO implement me
	panic("implement me")
}

func (m *myStruct) Proof(request wow.LoginProofRequest) wow.LoginProofResponse {
	//TODO implement me
	panic("implement me")
}
