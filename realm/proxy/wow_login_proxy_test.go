package main

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/realm/srpfunc"
	"net"
	"testing"
	"wow"
)

func Test5aWow(t *testing.T) {
	requestBytes := []byte{0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48}
	request := wow.LoginChallengeRequest{}

	err := request.UnMarshal(requestBytes)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0), request.Cmd)
	assert.Equal(t, uint8(3), request.Error)
	assert.Equal(t, uint16(44), request.Size)
	assert.Equal(t, [4]byte{87, 111, 87, 0}, request.GameName)
	assert.Equal(t, uint8(1), request.Version1)
	assert.Equal(t, uint8(12), request.Version2)
	assert.Equal(t, uint8(1), request.Version3)
	assert.Equal(t, uint16(5875), request.Build)
	//assert.Equal(t, [4]uint8{0,120, 56, 54}, request.Platform)
	//assert.Equal(t, [4]uint8{78, 67, 104, 122}, request.Os)
	//assert.Equal(t, [4]uint8{224, 1, 0, 0}, request.Country)
	//assert.Equal(t, uint32(127), request.TimeZoneBias)
	//assert.Equal(t, [4]uint8{0, 0, 1, 14}, request.Ip)
	//assert.Equal(t, uint8(87), request.ILen)
	assert.Equal(t, "WANGTENGDA0310", request.I)

	responseBytes := []byte{0, 0, 0, 196, 237, 55, 139, 100, 170, 19, 197, 139, 108, 225, 96, 91, 125, 11, 104, 159, 194, 197, 173, 205, 189, 60, 169, 150, 92, 126, 86, 122, 84, 7, 41, 1, 7, 32, 183, 155, 62, 42, 135, 130, 60, 171, 143, 94, 191, 191, 142, 177, 1, 8, 83, 80, 6, 41, 139, 91, 173, 189, 91, 83, 225, 137, 94, 100, 75, 137, 58, 232, 246, 237, 141, 205, 224, 57, 169, 88, 221, 249, 215, 235, 160, 78, 102, 28, 77, 65, 7, 196, 167, 190, 145, 42, 229, 110, 192, 220, 50, 90, 186, 163, 30, 153, 160, 11, 33, 87, 252, 55, 63, 179, 105, 205, 210, 241, 0}
	response := &wow.LoginChallengeResponse{}
	err = response.UnMarshal(responseBytes)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0), response.Cmd)
	assert.Equal(t, uint8(0), response.Error)
	assert.Equal(t, uint8(0), response.FailEnum)
	assert.Equal(t, [16]byte{0xba, 0xa3, 0x1e, 0x99, 0xa0, 0xb, 0x21, 0x57, 0xfc, 0x37, 0x3f, 0xb3, 0x69, 0xcd, 0xd2, 0xf1}, response.VersionChallenge)
	assert.Equal(t, uint8(0), response.SecurityFlags)
	assert.Equal(t, [32]byte{183, 155, 62, 42, 135, 130, 60, 171, 143, 94, 191, 191, 142, 177, 1, 8, 83, 80, 6, 41, 139, 91, 173, 189, 91, 83, 225, 137, 94, 100, 75, 137}, response.N)
	assert.Equal(t, [32]byte{196, 237, 55, 139, 100, 170, 19, 197, 139, 108, 225, 96, 91, 125, 11, 104, 159, 194, 197, 173, 205, 189, 60, 169, 150, 92, 126, 86, 122, 84, 7, 41}, response.B)
	assert.Equal(t, [32]byte{58, 232, 246, 237, 141, 205, 224, 57, 169, 88, 221, 249, 215, 235, 160, 78, 102, 28, 77, 65, 7, 196, 167, 190, 145, 42, 229, 110, 192, 220, 50, 90}, response.S)
	assert.Equal(t, uint8(7), response.G)

	dial, err := net.Dial("tcp", "logon.5awow.com:3724")
	if err != nil {
		t.Fatal(err)

	}
	defer closeConnection(dial)

	n, err := dial.Write(requestBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("write bytes len:", n)

	var buf [1024]byte
	n, err = dial.Read(buf[:])
	if err != nil {
		t.Fatal(err)

	}

	err = response.UnMarshal(buf[:n])
	assert.NoError(t, err)
	t.Log("N", response.N)
	t.Log("B", response.B)
	t.Log("S", response.S)
	t.Log("G", response.G)

	var buf1 = make([]byte, 1024)
	err = response.Marshal(buf1)
	assert.NoError(t, err)
	assert.Equal(t, buf[:n], buf1[:n])
	// 1. n g s
	// 2. B =(v + g ^ b) mod n, u = H(A, B)
	// 3. S = (A - v^u)^b, K = H(S), M1 =H(A, B, K) if M1 = M' then success M2 = H(A, M1, K)
	//
	//
}
func TestProof(t *testing.T) {
	// AuthSocket:537
	decodeString, err := base64.StdEncoding.DecodeString("ATVbRJzQnPq4TG/GQCJIBbdx77doLxYRm/hYm+GnaBU+WmF4IuCl1J+Hf0q3KpL6n/HKXgnxRMihngC9r4mcX8bpB4koNi1KzAAA")
	assert.NoError(t, err)
	fmt.Println(decodeString)

	request := &wow.LoginProofRequest{}
	err = request.UnMarshal(decodeString)
	assert.NoError(t, err)
	t.Log(request)

	var buf = make([]byte, len(decodeString))
	err = request.Marshal(buf)
	assert.NoError(t, err)
	assert.Equal(t, decodeString, buf)

	decodeString, err = base64.StdEncoding.DecodeString("AQBq8Y6xHumkx5wb2q7Tn0bc5yvk/AAAAAA=")
	assert.NoError(t, err)
	fmt.Println(decodeString)

	response := &wow.LoginProofResponse{}
	err = response.UnMarshal(decodeString)
	assert.NoError(t, err)
	t.Log(response)

	buf = make([]byte, len(decodeString))
	err = response.Marshal(buf)
	assert.NoError(t, err)
	assert.Equal(t, decodeString, buf)

}

func TestLogin(t *testing.T) {
	challengeRequest := wow.LoginChallengeRequest{
		Cmd:          0,
		Error:        0,
		Size:         0,
		GameName:     [4]byte{},
		Version1:     0,
		Version2:     0,
		Version3:     0,
		Build:        0,
		Platform:     [4]uint8{},
		Os:           [4]uint8{},
		Country:      [4]uint8{},
		TimeZoneBias: 0,
		Ip:           [4]uint8{},
		ILen:         0,
		I:            "",
	}

	var srp wow.WoWAuth
	srp = &srpfunc.MyStruct1{}
	challengeResponse := srp.Challenge(challengeRequest)
	assert.NotNil(t, challengeResponse)

	proofRequest := wow.LoginProofRequest{
		Cmd:           0,
		A:             [32]byte{},
		M1:            [20]byte{},
		CRC1:          [20]byte{},
		NumberOfKeys:  0,
		SecurityFlags: 0,
	}

	proofResponse := srp.Proof(proofRequest)
	assert.NotNil(t, proofResponse)
}
