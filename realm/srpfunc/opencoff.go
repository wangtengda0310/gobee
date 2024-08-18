package srpfunc

import (
	"fmt"
	"wow"
)
import "github.com/opencoff/go-srp"

type MyStruct1 struct {
}

func (m *MyStruct1) Challenge(request wow.LoginChallengeRequest) wow.LoginChallengeResponse {
	bits := 1024
	pass := []byte("password string that's too long")
	i := []byte("foouser")

	s, err := srp.New(bits)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	fmt.Println(pass)
	fmt.Println(i)
	return wow.LoginChallengeResponse{
		Cmd:              0,
		Error:            0,
		FailEnum:         0,
		B:                [32]byte{},
		GLen:             0,
		G:                0,
		NLen:             0,
		N:                [32]byte{},
		S:                [32]byte{},
		VersionChallenge: [16]byte{},
		SecurityFlags:    0,
	}
}

func (m *MyStruct1) Proof(request wow.LoginProofRequest) wow.LoginProofResponse {
	//TODO implement me
	panic("implement me")
}
