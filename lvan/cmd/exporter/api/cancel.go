package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/wangtengda0310/gobee/lvan/internal/execute"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

func HandleCancelRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	taskID := pathParts[2]

	task, b := execute.GetTask(taskID)
	if !b {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(fmt.Sprintf("{\"code\":0,\"msg\":\"任务不存在\",\"id\":\"%s\"}", taskID)))
		if err != nil {
			logger.Warn("cancel 序列化错误")
		}
		return
	}
	task.Cancel()
	_, err := w.Write([]byte(fmt.Sprintf("{\"code\":0,\"msg\":\"任务已取消\",\"id\":\"%s\"}", task.ID)))
	if err != nil {
		logger.Warn("cancel 序列化错误")
	}
}
