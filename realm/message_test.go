package realm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"wow"
)

func TestMessage(t *testing.T) {

	msgByte := []byte{0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48}

	msg := &wow.LoginChallengeRequest{}
	err := msg.UnMarshal(msgByte)
	t.Log(msg)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0), msg.Cmd)
	assert.Equal(t, uint8(3), msg.Error)
	assert.Equal(t, uint16(44), msg.Size)
	assert.Equal(t, "WoW", string(bytes.Trim(msg.GameName[:], "\x00")))
	assert.Equal(t, uint8(1), msg.Version1)
	assert.Equal(t, uint8(12), msg.Version2)
	assert.Equal(t, uint8(1), msg.Version3)
	assert.Equal(t, uint16(5875), msg.Build)
	assert.Equal(t, "x86", string(bytes.Trim(msg.Platform[:], "\x00")))
	assert.Equal(t, "Win", string(bytes.Trim(msg.Os[:], "\x00")))
	assert.Equal(t, "zhCN", string(bytes.Trim(msg.Country[:], "\x00")))
	assert.Equal(t, uint32(time.Hour.Minutes()*8), msg.TimeZoneBias)
	assert.Equal(t, [4]byte{127, 0, 0, 1}, msg.Ip)
	assert.Equal(t, uint8(14), msg.ILen)
	assert.Equal(t, "WANGTENGDA0310", string(msg.I[:]))

	data := make([]byte, 34+msg.ILen)
	err = msg.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, msgByte, data)
}

// 处理粘包
func TestMessage2(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte{0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48, 0, 3, 44, 0, 87, 111, 87, 0, 1, 12, 1, 243, 22, 54, 56, 120, 0, 110, 105, 87, 0, 78, 67, 104, 122, 224, 1, 0, 0, 127, 0, 0, 1, 14, 87, 65, 78, 71, 84, 69, 78, 71, 68, 65, 48, 51, 49, 48}))
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) < 4 {
			return 0, nil, nil
		}
		var size uint16
		_ = binary.Read(bytes.NewReader(data[2:4]), binary.LittleEndian, &size)
		if len(data) < int(size)+4 {
			return 0, nil, nil
		}
		return int(size) + 4, data[:int(size)+4], nil
	})
	for scanner.Scan() {
		msg := &wow.LoginChallengeRequest{}
		err := msg.UnMarshal(scanner.Bytes())
		t.Log(msg)
		assert.NoError(t, err)
		assert.Equal(t, uint8(0), msg.Cmd)
		assert.Equal(t, uint8(3), msg.Error)
		assert.Equal(t, uint16(44), msg.Size)
		assert.Equal(t, "WoW", string(bytes.Trim(msg.GameName[:], "\x00")))
		assert.Equal(t, uint8(1), msg.Version1)
		assert.Equal(t, uint8(12), msg.Version2)
		assert.Equal(t, uint8(1), msg.Version3)
		assert.Equal(t, uint16(5875), msg.Build)
		assert.Equal(t, "x86", string(bytes.Trim(msg.Platform[:], "\x00")))
		assert.Equal(t, "Win", string(bytes.Trim(msg.Os[:], "\x00")))
		assert.Equal(t, "zhCN", string(bytes.Trim(msg.Country[:], "\x00")))
		assert.Equal(t, uint32(time.Hour.Minutes()*8), msg.TimeZoneBias)
		assert.Equal(t, [4]byte{127, 0, 0, 1}, msg.Ip)
		assert.Equal(t, uint8(14), msg.ILen)
		assert.Equal(t, "WANGTENGDA0310", string(msg.I[:]))
	}
}

func TestLoginChallengeResponse_UnMarshal(t *testing.T) {
	response := wow.LoginChallengeResponse{}
	data := []byte{0, 0, 0, 166, 99, 11, 236, 133, 219, 49, 245, 17, 125, 166, 98, 253, 78, 127, 121, 219, 223, 161, 231, 175, 173, 91, 94, 36, 183, 31, 240, 58, 96, 112, 2, 1, 7, 32, 183, 155, 62, 42, 135, 130, 60, 171, 143, 94, 191, 191, 142, 177, 1, 8, 83, 80, 6, 41, 139, 91, 173, 189, 91, 83, 225, 137, 94, 100, 75, 137, 58, 232, 246, 237, 141, 205, 224, 57, 169, 88, 221, 249, 215, 235, 160, 78, 102, 28, 77, 65, 7, 196, 167, 190, 145, 42, 229, 110, 192, 220, 50, 90, 186, 163, 30, 153, 160, 11, 33, 87, 252, 55, 63, 179, 105, 205, 210, 241, 0}
	t.Log(response, len(data))
	err := response.UnMarshal(data)
	assert.Equal(t, uint8(0), response.Cmd)
	assert.Equal(t, byte(0), response.Error)
	assert.Equal(t, uint8(0), response.FailEnum)
	assert.Equal(t, byte(1), response.GLen)
	assert.Equal(t, byte(32), response.NLen)
	assert.Equal(t, wow.VersionChallenge, response.VersionChallenge)
	assert.Equal(t, byte(0), response.SecurityFlags)
	assert.NoError(t, err)
}
