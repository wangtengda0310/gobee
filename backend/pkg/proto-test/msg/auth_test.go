package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractAccountFromLoginPayload 验证从 LoginReq payload 提取账号名
func TestExtractAccountFromLoginPayload(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		expect  string
		wantErr bool
	}{
		{
			name:    "正常账号",
			payload: append([]byte{4, 0}, []byte("test")...), // [2B len=4][test]
			expect:  "test",
			wantErr: false,
		},
		{
			name:    "空账号",
			payload: []byte{0, 0}, // [2B len=0]
			expect:  "",
			wantErr: false,
		},
		{
			name:    "长账号名",
			payload: append([]byte{10, 0}, []byte("testplayer")...), // [2B len=10][testplayer]
			expect:  "testplayer",
			wantErr: false,
		},
		{
			name:    "包含数字后缀的账号",
			payload: append([]byte{5, 0}, []byte("test1")...), // [2B len=5][test1]
			expect:  "test1",
			wantErr: false,
		},
		{
			name:    "payload太短",
			payload: []byte{1},
			expect:  "",
			wantErr: true,
		},
		{
			name:    "Account字段不完整",
			payload: append([]byte{10, 0}, []byte("short")...), // 声明10字节但只有5字节
			expect:  "",
			wantErr: true,
		},
		{
			name:    "空payload",
			payload: []byte{},
			expect:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := ExtractAccountFromLoginPayload(tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, account)
			}
		})
	}
}
