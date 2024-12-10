package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册API路由
func RegisterRoutes(r *gin.Engine) {
	r.GET("/hello", helloHandler)
}

// helloHandler 处理/hello的GET请求
func helloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}
