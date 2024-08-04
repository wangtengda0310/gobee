package realm

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type loginChallengeMsg struct {
	Cmd           uint8
	Error         uint8
	Size          uint16
	Gamename      [4]byte
	Version1      uint8
	Version2      uint8
	Version3      uint8
	Build         uint16
	Platform      [4]uint8
	Os            [4]uint8
	Country       [4]uint8
	Timezone_bias uint32
	Ip            [4]uint8
	I_len         uint8
	I             string
}

func readLoginChallengeMsg(msg *loginChallengeMsg, data []byte) error {
	msg.Cmd = data[0]
	msg.Error = data[1]
	reader := bytes.NewReader(data[2:4])
	_ = binary.Read(reader, binary.LittleEndian, &msg.Size)
	copy(msg.Gamename[:], data[4:8])
	msg.Version1 = data[8]
	msg.Version2 = data[9]
	msg.Version3 = data[10]
	reader = bytes.NewReader(data[11:13])
	_ = binary.Read(reader, binary.LittleEndian, &msg.Build)
	msg.Platform = [4]uint8{data[16], data[15], data[14], data[13]}
	msg.Os = [4]uint8{data[20], data[19], data[18], data[17]}
	msg.Country = [4]uint8{data[24], data[23], data[22], data[21]}
	reader = bytes.NewReader(data[25:29])
	_ = binary.Read(reader, binary.LittleEndian, &msg.Timezone_bias)
	copy(msg.Ip[:], data[29:33])
	msg.I_len = data[33]
	msg.I = string(data[34 : 34+msg.I_len])
	return nil
}

func (m *loginChallengeMsg) unMarshal(data []byte) error {
	return readLoginChallengeMsg(m, data)
}

func TestMessage(t *testing.T) {

	msgByte := []byte{0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48}

	msg := &loginChallengeMsg{}
	err := msg.unMarshal(msgByte)
	t.Log(msg)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0), msg.Cmd)
	assert.Equal(t, uint8(3), msg.Error)
	assert.Equal(t, uint16(44), msg.Size)
	assert.Equal(t, "WoW", string(bytes.Trim(msg.Gamename[:], "\x00")))
	assert.Equal(t, uint8(1), msg.Version1)
	assert.Equal(t, uint8(12), msg.Version2)
	assert.Equal(t, uint8(1), msg.Version3)
	assert.Equal(t, uint16(5875), msg.Build)
	assert.Equal(t, "x86", string(bytes.Trim(msg.Platform[:], "\x00")))
	assert.Equal(t, "Win", string(bytes.Trim(msg.Os[:], "\x00")))
	assert.Equal(t, "zhCN", string(bytes.Trim(msg.Country[:], "\x00")))
	assert.Equal(t, uint32(time.Hour.Minutes()*8), msg.Timezone_bias)
	assert.Equal(t, [4]byte{127, 0, 0, 1}, msg.Ip)
	assert.Equal(t, uint8(14), msg.I_len)
	assert.Equal(t, "WANGTENGDA0310", string(msg.I[:]))

}
