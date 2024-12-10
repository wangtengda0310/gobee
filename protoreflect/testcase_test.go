package main

import (
	"testing"
)

func Test_marshal(t *testing.T) {
	type args struct {
		test *Testcase
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				test: &Testcase{
					"测试用例",
					[]testOneCase{
						{protoMessage{
							"sendReq",
							map[string]string{"test1": "test"},
						},
							protoMessage{
								"receiveAck",
								map[string]string{"test2": "test"},
							}},
					},
				},
			},
			want: `name: 测试用例
seq:
  - req:
      proto: sendReq
      msg: test1: test
    ack:
      proto: receiveAck
      msg: test2: test`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshal(tt.args.test); got != tt.want {
				t.Errorf("marshal() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
