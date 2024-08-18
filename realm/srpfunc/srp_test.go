package srpfunc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type srpfunc func(i string, password string) bool

func Test_myStruct1_test(t *testing.T) {
	type args struct {
		f        srpfunc
		username string
		password string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试opencoff的srp功能",
			args: args{srpfunc_opencoff, "wangtengda0310", "123456"},
		},
		{
			name: "测试voynic的srp功能",
			args: args{srpfunc_voynic, "wangtengda0310", "123456"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.args.f(tt.args.username, tt.args.username))
		})
	}

}
func FuzzSrpFunc(f *testing.F) {
	f.Fuzz(func(t *testing.T, i, password string) {
		assert.True(t, srpfunc_voynic(i, password))
		assert.True(t, srpfunc_opencoff(i, password))
	})
}
