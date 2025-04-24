package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
)

// 处理SSE请求
func handleSSERequest(w http.ResponseWriter, r *http.Request, task *pkg.Task) {
	// 设置SSE头
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建客户端ID
	clientID := uuid.New().String()
	outputChan := task.AddClient(clientID)

	// 如果无法添加客户端（例如任务已完成或达到最大连接数）
	if outputChan == nil {
		http.Error(w, "Cannot connect to task: either completed or connection limit reached", http.StatusServiceUnavailable)
		return
	}

	// 设置连接关闭时的清理
	go func() {
		<-r.Context().Done()
		task.RemoveClient(clientID)
		logger.Info("SSE客户端 %s 连接已关闭", clientID)
	}()

	// 异步执行命令（如果尚未执行）
	taskResult := task.Result
	if task.Status == pkg.Blocking && task.CmdPath == "" {
		go pkg.ExecuteTask(task)
	}

	jsonw := json.NewEncoder(w)
	jsonw.Encode(map[string]any{"id": task.ID, "status": taskResult})

	// 发送任务ID
	fmt.Fprintf(w, "data: {\"id\": \"%s\", \"status\": \"%v\"}\n\n", task.ID, taskResult)
	w.(http.Flusher).Flush()

	// 发送输出流
	for output := range outputChan {
		// 确保每行输出都有正确的SSE格式
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line != "" {
				_, _ = fmt.Fprintf(w, "data: %s\n\n", line)
				w.(http.Flusher).Flush()
			}
		}
	}

	// 如果输出通道关闭但任务仍在运行，发送最终状态
	task.Mutex.Lock()
	if task.Status != pkg.Running {
		jsonw.Encode(map[string]any{"status": task.Status, "exitCode": taskResult.ExitCode})
		_, _ = fmt.Fprint(w, "\n")
		_, _ = fmt.Fprintf(w, "data: {\"status\": \"%v\", \"exitCode\": %d}\n\n", task.Status, taskResult.ExitCode)
		w.(http.Flusher).Flush()
	}
	task.Mutex.Unlock()
}

// 处理只返回ID的请求
func handleOnlyIDRequest(w http.ResponseWriter, task *pkg.Task) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(task.ID))

	// 异步执行命令
	go pkg.ExecuteTask(task)
}

// 处理命令请求
func HandleCommandRequest(w http.ResponseWriter, r *http.Request) {
	var task *pkg.Task
	switch r.Method {
	// 处理GET请求，格式为/cmd/command/param1/param2...
	case http.MethodGet:
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}
		cmd := pathParts[2]
		args := []string{}
		if len(pathParts) > 3 {
			// 处理URL路径中的参数，支持/和-引导的参数
			for _, arg := range pathParts[3:] {
				if arg != "" {
					// 处理特殊前缀，将__slash__转换为/开头的参数
					if strings.HasPrefix(arg, "__slash__") {
						arg = "/" + strings.TrimPrefix(arg, "__slash__")
					}
					args = append(args, arg)
				}
			}
		}

		req := internal.CommandRequest{
			Cmd:     cmd,
			Version: "",
			Args:    args,
			Env:     make(map[string]string),
		}

		task = pkg.CreateTask(req, w, os.Stdout)

	// 处理POST请求
	case http.MethodPost:
		// 解析请求体格式
		bodyType := r.URL.Query().Get("body")
		if bodyType == "" {
			bodyType = "yaml" // 默认YAML
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		logger.Warn("header: %s", r.Header.Get("Content-Type"))
		logger.Warn("body: %s", string(body))
		// 解析请求体
		var req internal.CommandRequest
		if r.Header.Get("Content-Type") == "application/json" || bodyType == "json" {
			err = json.Unmarshal(body, &req)
			_, _ = w.Write(body)
		} else {
			err = yaml.Unmarshal(body, &req)
		}

		if err != nil {
			http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 确保环境变量字段已初始化
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		// 创建任务
		task = pkg.CreateTask(req, w, os.Stdout)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否只返回ID
	onlyID := r.URL.Query().Get("onlyid") == "true"
	if onlyID {
		handleOnlyIDRequest(w, task)
		return
	}

	// 检查是否使用SSE
	useSSE := r.URL.Query().Get("sse") == "true"
	if useSSE {
		handleSSERequest(w, r, task)
		return
	}

	handleSyncRequest(w, task)
	w.(http.Flusher).Flush()

	// 同步执行命令
	go pkg.ExecuteTask(task)
}

// 处理同步执行命令请求
func handleSyncRequest(w http.ResponseWriter, task *pkg.Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	status := task.Status
	task.Mutex.Unlock()

	var res *internal.CmdResponse
	// 根据任务状态设置HTTP状态码
	if status == pkg.Failed {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.Result.ExitCode))
		res = &internal.CmdResponse{
			Code: 1,
			Msg:  "任务执行失败",
			Id:   task.ID,
		}
	} else {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.Result.ExitCode))
		res = &internal.CmdResponse{
			Code: 0,
			Msg:  "任务处理中",
			Id:   task.ID,
		}
	}

	// 返回结果
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(task.Output))
	w.Header().Set("Content-Type", "application/json")
	marshal, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		w.Write([]byte(fmt.Sprintf("{\"code\":1,\"msg\":\"序列化错误\",\"id\":\"%s\"}", task.ID)))
		return
	}
	w.Write(marshal)
}
