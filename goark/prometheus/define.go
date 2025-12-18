package metrics

var MetricGrpcReq    = &Counter{opt: opt{name: "client_grpc_total", help: "客户端grpc请求总数"}}
