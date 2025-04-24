package mp

import "testing"

func Test_adaptContent(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			"结尾的分隔符",
			`a,b,c|;`,
			`a,b,c`,
		},
		{
			"首尾的引号",
			`"a,b,c"`,
			`a,b,c`,
		},
		{
			"首尾的p标签",
			`<p>a,b,c</p>`,
			`a,b,c`,
		},
		{
			"4个连续引号替换为一个",
			`""""a,""""b,""""c""""`,
			`a,"b,"c`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adaptContent(tt.args); got != tt.want {
				t.Errorf("adaptContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
