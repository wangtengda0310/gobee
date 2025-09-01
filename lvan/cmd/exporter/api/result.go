package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/internal/execute"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

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
		_, _ = w.Write([]byte("<html><body><h1>API Documentation</h1><p>This is the API documentation for the exporter service.</p></body></html>"))
		return
	}

	// 获取任务
	task, exists := execute.GetTask(taskID)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// 检查是否使用SSE
	useSSE := r.URL.Query().Get("sse") == "true"
	if useSSE {
		if resultSSE(w, r, task) {
			return
		}
	} else {
		HandleSyncResultRequest(w, task)
	}
}

func resultSSE(w http.ResponseWriter, r *http.Request, task *execute.Task) bool {
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
		return true
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
	existingOutput := task.Result.Output
	task.Mutex.Unlock()

	if existingOutput != "" {
		fmt.Fprintf(w, "data: %s\n\n", existingOutput)
		w.(http.Flusher).Flush()
	}

	// 如果任务已完成，关闭连接
	task.Mutex.Lock()
	isCompleted := task.Status == execute.Completed
	task.Mutex.Unlock()

	if isCompleted {
		return true
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
	return false
}

// 处理同步执行命令请求 // todo 跟新见任务的方法合并
func HandleSyncResultRequest(w http.ResponseWriter, task *execute.Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	isRunning := task.Status
	task.Mutex.Unlock()

	var res *internal.ResultResponse
	// 根据任务状态设置HTTP状态码
	if isRunning == execute.Failed {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.Result.ExitCode))
		res = &internal.ResultResponse{
			Code: 3,
			Msg:  strings.Join(task.Result.Stderr, "\n"),
			Id:   task.ID,
			Job:  task.Request,
		}
	} else if isRunning == execute.Running {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.Result.ExitCode))
		res = &internal.ResultResponse{
			Code: 2,
			Msg:  "任务处理中",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else if isRunning == execute.Blocking {
		w.WriteHeader(http.StatusAccepted) // 202 todo 是否需要新的状态码
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.Result.ExitCode))
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
		_, err = w.Write([]byte(fmt.Sprintf("{\"code\":1,\"msg\":\"序列化错误\",\"id\":\"%s\",\"job\":\"%s\"}", task.ID, job)))
		if err != nil {
			logger.Error("写入响应失败: %v", err)
		}
		return
	}
	_, _ = w.Write(marshal)
}
