package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"strings"
	"time"
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
	if task.Status == "running" && task.CmdPath == "" {
		go pkg.ExecuteCommand(task)
	}

	// 发送任务ID
	fmt.Fprintf(w, "data: {\"id\": \"%s\", \"status\": \"%s\"}\n\n", task.ID, task.Status)
	w.(http.Flusher).Flush()

	// 发送输出流
	for output := range outputChan {
		// 确保每行输出都有正确的SSE格式
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Fprintf(w, "data: %s\n\n", line)
				w.(http.Flusher).Flush()
			}
		}
	}

	// 如果输出通道关闭但任务仍在运行，发送最终状态
	task.Mutex.Lock()
	if task.Status != "running" {
		fmt.Fprintf(w, "data: {\"status\": \"%s\", \"exitCode\": %d}\n\n", task.Status, task.ExitCode)
		w.(http.Flusher).Flush()
	}
	task.Mutex.Unlock()
}

// 处理只返回ID的请求
func handleOnlyIDRequest(w http.ResponseWriter, task *pkg.Task) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(task.ID))

	// 异步执行命令
	go pkg.ExecuteCommand(task)
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

		task = pkg.CreateTask(req)

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

		// 解析请求体
		var req internal.CommandRequest
		if bodyType == "json" {
			err = json.Unmarshal(body, &req)
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
		task = pkg.CreateTask(req)

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
	go pkg.ExecuteCommand(task)
}

// 处理结果请求
func HandleResultRequest(w http.ResponseWriter, r *http.Request) {
	// 解析路径中的任务ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	taskID := pathParts[2]

	// 处理帮助请求
	if taskID == "help" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>API Documentation</h1><p>This is the API documentation for the exporter service.</p></body></html>"))
		return
	}

	// 获取任务
	task, exists := pkg.GetTask(taskID)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// 检查是否使用SSE
	useSSE := r.URL.Query().Get("sse") == "true"
	if useSSE {
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
			logger.Info("结果SSE客户端 %s 连接已关闭", clientID)
		}()

		// 添加健康检查ping
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					// 发送ping消息
					fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().Unix())
					w.(http.Flusher).Flush()
				case <-r.Context().Done():
					return
				}
			}
		}()

		// 发送现有输出
		task.Mutex.Lock()
		existingOutput := task.Output
		task.Mutex.Unlock()

		if existingOutput != "" {
			fmt.Fprintf(w, "data: %s\n\n", existingOutput)
			w.(http.Flusher).Flush()
		}

		// 如果任务已完成，关闭连接
		task.Mutex.Lock()
		isCompleted := task.Status != "running"
		task.Mutex.Unlock()

		if isCompleted {
			return
		}

		// 发送新输出
		for output := range outputChan {
			// 确保每行输出都有正确的SSE格式
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Fprintf(w, "data: %s\n\n", line)
					w.(http.Flusher).Flush()
				}
			}
		}
	} else {
		HandleSyncResultRequest(w, task)
	}
}

// 处理同步执行命令请求
func handleSyncRequest(w http.ResponseWriter, task *pkg.Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	isRunning := task.Status
	task.Mutex.Unlock()

	var res *internal.CmdResponse
	// 根据任务状态设置HTTP状态码
	if isRunning == "failed" {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &internal.CmdResponse{
			Code: 1,
			Msg:  "任务执行失败",
			Id:   task.ID,
		}
	} else {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
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

// 处理同步执行命令请求 // todo 跟新见任务的方法合并
func HandleSyncResultRequest(w http.ResponseWriter, task *pkg.Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	isRunning := task.Status
	task.Mutex.Unlock()

	var res *internal.ResultResponse
	// 根据任务状态设置HTTP状态码
	if isRunning == "failed" {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &internal.ResultResponse{
			Code: 3,
			Msg:  "任务执行失败",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else if isRunning == "running" {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &internal.ResultResponse{
			Code: 2,
			Msg:  "任务处理中",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else if isRunning == "blocking" {
		w.WriteHeader(http.StatusAccepted) // 202 todo 是否需要新的状态码
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &internal.ResultResponse{
			Code: 1,
			Msg:  "任务等待处理",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else {
		w.Header().Set("X-Exit-Code", "0")
		res = &internal.ResultResponse{
			Code: 0,
			Msg:  "任务执行成功",
			Id:   task.ID,
			Job:  task.Request,
		}
	}

	// 返回结果
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(task.Output))
	w.Header().Set("Content-Type", "application/json")
	marshal, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		bytes, err := json.Marshal(task.Request)
		var job string
		if err == nil {
			job = string(bytes)
		}
		w.Write([]byte(fmt.Sprintf("{\"code\":1,\"msg\":\"序列化错误\",\"id\":\"%s\",\"job\":\"%s\"}", task.ID, job)))
		return
	}
	w.Write(marshal)
}

// 处理根路径请求
func HandleRootRequest(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径请求
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 使用嵌入的HTTP文档内容
	w.Write([]byte(httpDoc))
}

// 嵌入HTTP文档
//
//go:embed http-doc.txt
var httpDoc string
