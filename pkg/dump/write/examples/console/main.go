package main

import (
	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
	"github.com/wangtengda0310/gobee/lvan/pkg/dump/write"
)

func main() {
	// 准备示例数据
	records := []dump.Record{
		{
			"name": []byte("tom"),
			"id":   []byte("1"),
			"data": []byte("some small data"),
		},
		{
			"name": []byte("jack"),
			"id":   []byte("2"),
			"data": []byte("some very large data that exceeds twenty characters in length and should be displayed as byte size"),
		},
		{
			"name": []byte("alice"),
			"id":   []byte("3"),
			"data": []byte("short"),
		},
	}

	// 调用Console函数输出数据
	write.Console(records)
}
