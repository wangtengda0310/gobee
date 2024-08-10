package main

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"wow"
)

func Test5aWow(t *testing.T) {
	msg := wow.LoginChallengeMsg{}

	msgByte := []byte{0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48}

	msg.UnMarshal(msgByte)

	dial, err := net.Dial("tcp", "logon.5awow.com:3724")
	if err != nil {
		t.Fatal(err)

	}
	defer dial.Close()
	n, err := dial.Write(msgByte)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("write bytes len:", n)

	var buf [1024]byte
	n, err = dial.Read(buf[:])
	if err != nil {
		t.Fatal(err)

	}

	response := &wow.LoginChallengeResponse{}
	err = response.UnMarshal(buf[:n])
	assert.NoError(t, err)
	t.Log("N", response.N)
	t.Log("B", response.B)
	t.Log("S", response.S)
	t.Log("G", response.G)

}
