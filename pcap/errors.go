package pcap

import "errors"

// 本文件定义抓包包对外暴露的哨兵错误。
// 这些错误用于注册/注销校验，以及不支持的能力降级提示。

var (
	// ErrHandlerExists 注册同名处理函数时返回。
	ErrHandlerExists = errors.New("pcap: handler already exists")

	// ErrHandlerNotFound 注销不存在的处理函数时返回。
	ErrHandlerNotFound = errors.New("pcap: handler not found")

	// ErrHandlerNil 注册 nil 处理函数时返回。
	ErrHandlerNil = errors.New("pcap: handler is nil")

	// ErrEmptyName 注册空名称的处理函数时返回。
	ErrEmptyName = errors.New("pcap: handler name must not be empty")

	// ErrBPFNotSupported 当 Target 指定了 BPF 但当前 Source 不支持 BPF 时返回。
	// 例如：内存 mock 数据源、或已解码的只读 pcap 文件重放，无法回退到内核过滤。
	ErrBPFNotSupported = errors.New("pcap: bpf filter not supported by this source")

	// ErrSourceNil 当传入的 Source 为 nil 时返回。
	ErrSourceNil = errors.New("pcap: source is nil")
)
