package template

import "testing"

func TestRenderTemplate(t *testing.T) {
	type args struct {
		filePath string
		data     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{filePath: "testdata/template.txt", data: map[string]string{
				"Name":    "张三",
				"Company": "ABC公司",
				"Date":    "2023-07-15",
			}},
			want:    "尊敬的 张三 先生/女士：\r\n感谢您加入 ABC公司！\r\n您的入职日期是：2023-07-15",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.args.filePath, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
