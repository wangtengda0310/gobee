package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// CommandRequest 命令请求结构，与main.go中定义的结构保持一致
type CommandRequest struct {
	Cmd     string            `json:"cmd" yaml:"cmd"`
	Version string            `json:"version" yaml:"version"`
	Args    []string          `json:"args" yaml:"args"`
	Env     map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

// 测试结果结构
type TestResult struct {
	Name        string
	Success     bool
	StatusCode  int
	ResponseLen int
	Error       string
	Duration    time.Duration
}

const (
	baseURL = "http://localhost:80"
)

func main() {
	fmt.Println("开始测试 Exporter 服务...")
	fmt.Println("基础URL:", baseURL)

	// 确保服务已启动
	if !checkServiceAvailable() {
		fmt.Printf("错误: Exporter 服务未启动或无法访问，请确保服务在 %s 运行\n", baseURL)
		return
	}

	// 运行所有测试
	results := runAllTests()

	// 输出测试结果摘要
	printTestSummary(results)
}

// 截断字符串，用于输出显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 输出测试结果摘要
func printTestSummary(results []TestResult) {
	fmt.Println("\n测试结果摘要:")
	fmt.Println("====================")

	successCount := 0
	failCount := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failCount++
		}
		totalDuration += result.Duration
	}

	fmt.Printf("总测试数: %d\n", len(results))
	fmt.Printf("成功: %d\n", successCount)
	fmt.Printf("失败: %d\n", failCount)
	fmt.Printf("总耗时: %v\n", totalDuration)
	fmt.Println("====================")

	// 输出失败的测试详情
	if failCount > 0 {
		fmt.Println("\n失败的测试:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  - %s: %s\n", result.Name, result.Error)
			}
		}
	}
}

// 检查服务是否可用
func checkServiceAvailable() bool {
	_, err := http.Get(baseURL + "/result/help")
	return err == nil
}

// 运行所有测试
func runAllTests() []TestResult {
	var results []TestResult

	// 基本GET请求测试
	results = append(results, testGetCommand("ping", []string{"-n", "3", "127.0.0.1"}))

	// 基本POST请求测试 - JSON格式
	results = append(results, testPostCommandJSON(CommandRequest{
		Cmd:     "ping",
		Version: "0.1.0",
		Args:    []string{"-n", "3", "127.0.0.1"},
	}))

	// 基本POST请求测试 - YAML格式
	results = append(results, testPostCommandYAML(CommandRequest{
		Cmd:     "ping",
		Version: "0.1.0",
		Args:    []string{"-n", "3", "127.0.0.1"},
	}))

	// 测试onlyid参数
	results = append(results, testOnlyID())

	// 测试SSE流式输出
	results = append(results, testSSE())

	// 测试结果查询
	results = append(results, testResultQuery())

	// 测试IO密集型任务
	results = append(results, testIOIntensiveTask())

	// 测试并发请求
	results = append(results, testConcurrentRequests(5))

	// 测试特殊字符参数
	results = append(results, testSpecialCharacters())

	// 测试Windows特有路径格式
	results = append(results, testWindowsPathFormat())

	// 测试错误处理
	results = append(results, testErrorHandling())

	return results
}

// 测试IO密集型任务
func testIOIntensiveTask() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试IO密集型任务"}

	// 使用type命令作为IO密集型任务示例
	// 在Windows上，type命令可以显示文件内容，类似于Linux的cat
	// 我们尝试读取一个大文件或生成大量输出
	req := CommandRequest{
		Cmd:     "findstr",
		Version: "0.1.0",
		Args:    []string{"/s", "/i", ".", "*.*"},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=json", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应时间: %v\n", result.Duration)
	fmt.Printf("  响应内容前100字节: %s\n\n", truncateString(string(body), 100))

	return result
}

// 测试并发请求
func testConcurrentRequests(concurrency int) TestResult {
	start := time.Now()
	result := TestResult{Name: fmt.Sprintf("测试并发请求(%d个)", concurrency)}

	// 创建等待组
	var wg sync.WaitGroup
	// 创建互斥锁保护共享数据
	var mu sync.Mutex
	// 记录成功和失败的请求数
	successCount := 0
	failCount := 0
	// 记录总响应大小
	totalResponseSize := 0
	// 记录错误信息
	errors := []string{}

	// 启动并发请求
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 构建URL，每个请求使用不同的参数
			url := fmt.Sprintf("%s/cmd/ping/127.0.0.%d", baseURL, (index%254)+1)

			// 发送请求
			resp, err := http.Get(url)
			if err != nil {
				mu.Lock()
				failCount++
				errors = append(errors, fmt.Sprintf("请求%d错误: %v", index, err))
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			// 读取响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				mu.Lock()
				failCount++
				errors = append(errors, fmt.Sprintf("请求%d读取响应错误: %v", index, err))
				mu.Unlock()
				return
			}

			// 更新统计信息
			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
			} else {
				failCount++
				errors = append(errors, fmt.Sprintf("请求%d状态码错误: %d", index, resp.StatusCode))
			}
			totalResponseSize += len(body)
			mu.Unlock()

			fmt.Printf("  并发请求%d完成，状态码: %d, 响应大小: %d字节\n", index, resp.StatusCode, len(body))
		}(i)
	}

	// 等待所有请求完成
	wg.Wait()

	// 设置测试结果
	result.Success = failCount == 0
	result.ResponseLen = totalResponseSize
	result.Duration = time.Since(start)
	if !result.Success {
		result.Error = strings.Join(errors[:min(3, len(errors))], "; ")
		if len(errors) > 3 {
			result.Error += fmt.Sprintf("; 以及其他%d个错误", len(errors)-3)
		}
	}

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  成功请求: %d\n", successCount)
	fmt.Printf("  失败请求: %d\n", failCount)
	fmt.Printf("  总响应大小: %d字节\n", totalResponseSize)
	fmt.Printf("  总耗时: %v\n", result.Duration)
	fmt.Printf("  平均每个请求耗时: %v\n\n", result.Duration/time.Duration(concurrency))

	return result
}

// 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 测试特殊字符参数
func testSpecialCharacters() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试特殊字符参数"}

	// 构建包含特殊字符的请求
	req := CommandRequest{
		Cmd:     "echo",
		Version: "0.1.0",
		Args:    []string{"Hello, World!", "&", "|", "<", ">", "^", "%", "*", "?"},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=json", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试错误处理
func testErrorHandling() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试错误处理"}

	// 测试不存在的命令
	req := CommandRequest{
		Cmd:     "nonexistentcommand",
		Version: "0.1.0",
		Args:    []string{},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=json", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 检查是否返回了错误状态码
	result.Success = resp.StatusCode == http.StatusInternalServerError
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试Windows特有路径格式
func testWindowsPathFormat() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试Windows路径格式"}

	// 构建包含Windows路径的请求
	req := CommandRequest{
		Cmd:     "dir",
		Version: "0.1.0",
		Args:    []string{"C:\\Windows\\System32"},
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=json", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试结果查询
func testResultQuery() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试结果查询"}

	// 首先创建一个任务
	resp, err := http.Get(baseURL + "/cmd/ping/127.0.0.1?onlyid=true")
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("创建任务请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取任务ID
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取任务ID错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	taskID := string(body)

	// 输出任务ID
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  创建的任务ID: %s\n", taskID)

	// 等待任务完成
	time.Sleep(2 * time.Second)

	// 查询任务结果
	resultResp, err := http.Get(baseURL + "/result/" + taskID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("查询结果错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resultResp.Body.Close()

	// 读取结果
	resultBody, err := io.ReadAll(resultResp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取结果错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 检查结果
	result.Success = resultResp.StatusCode == http.StatusOK
	result.StatusCode = resultResp.StatusCode
	result.ResponseLen = len(resultBody)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("  结果状态码: %d\n", resultResp.StatusCode)
	fmt.Printf("  结果长度: %d 字节\n", len(resultBody))
	fmt.Printf("  结果内容: %s\n", truncateString(string(resultBody), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试GET命令
func testGetCommand(cmd string, args []string) TestResult {
	start := time.Now()
	result := TestResult{Name: fmt.Sprintf("GET命令测试: %s %v", cmd, args)}

	// 构建URL
	url := fmt.Sprintf("%s/cmd/%s", baseURL, cmd)
	for _, arg := range args {
		// 处理特殊前缀，将/开头的参数转换为__slash__
		if strings.HasPrefix(arg, "/") {
			arg = "__slash__" + strings.TrimPrefix(arg, "/")
		}
		url += "/" + arg
	}

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试POST命令 - JSON格式
func testPostCommandJSON(req CommandRequest) TestResult {
	start := time.Now()
	result := TestResult{Name: fmt.Sprintf("POST命令测试(JSON): %s %v", req.Cmd, req.Args)}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=json", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试POST命令 - YAML格式
func testPostCommandYAML(req CommandRequest) TestResult {
	start := time.Now()
	result := TestResult{Name: fmt.Sprintf("POST命令测试(YAML): %s %v", req.Cmd, req.Args)}

	// 序列化请求
	reqBody, err := yaml.Marshal(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("序列化请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 发送请求
	resp, err := http.Post(baseURL+"/cmd?body=yaml", "application/yaml", bytes.NewBuffer(reqBody))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = resp.StatusCode == http.StatusOK
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  响应内容: %s\n", truncateString(string(body), 100))
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}

// 测试onlyid参数
func testOnlyID() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试onlyid参数"}

	// 发送请求
	resp, err := http.Get(baseURL + "/cmd/ping/127.0.0.1?onlyid=true")
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取响应错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 验证返回的是任务ID
	taskID := string(body)
	result.Success = resp.StatusCode == http.StatusOK && len(taskID) > 0
	result.StatusCode = resp.StatusCode
	result.ResponseLen = len(body)
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  状态码: %d\n", result.StatusCode)
	fmt.Printf("  任务ID: %s\n", taskID)
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	// 等待一段时间后查询结果
	time.Sleep(2 * time.Second)

	// 查询任务结果
	resultResp, err := http.Get(baseURL + "/result/" + taskID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("查询结果错误: %v", err)
		return result
	}
	defer resultResp.Body.Close()

	// 读取结果
	resultBody, err := io.ReadAll(resultResp.Body)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("读取结果错误: %v", err)
		return result
	}

	fmt.Printf("  结果状态码: %d\n", resultResp.StatusCode)
	fmt.Printf("  结果内容: %s\n\n", truncateString(string(resultBody), 100))

	return result
}

// 测试SSE流式输出
func testSSE() TestResult {
	start := time.Now()
	result := TestResult{Name: "测试SSE流式输出"}

	// 发送请求
	resp, err := http.Get(baseURL + "/cmd/ping/127.0.0.1?sse=true")
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("请求错误: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// 检查响应头
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		result.Success = false
		result.Error = fmt.Sprintf("响应头Content-Type不是text/event-stream，而是%s", resp.Header.Get("Content-Type"))
		result.Duration = time.Since(start)
		return result
	}

	// 读取SSE事件流
	scanner := bufio.NewScanner(resp.Body)
	var output strings.Builder
	eventCount := 0
	timeout := time.After(5 * time.Second)

	done := make(chan bool)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				eventCount++
				data := strings.TrimPrefix(line, "data: ")
				output.WriteString(data)
				output.WriteString("\n")
				fmt.Printf("  收到SSE事件: %s\n", truncateString(data, 50))
			}
		}
		done <- true
	}()

	// 等待超时或完成
	select {
	case <-timeout:
		fmt.Println("  SSE测试超时，已收到", eventCount, "个事件")
	case <-done:
	}

	result.Success = eventCount > 0
	result.ResponseLen = output.Len()
	result.Duration = time.Since(start)

	// 输出详细信息
	fmt.Printf("测试: %s\n", result.Name)
	fmt.Printf("  收到事件数: %d\n", eventCount)
	fmt.Printf("  响应长度: %d 字节\n", result.ResponseLen)
	fmt.Printf("  耗时: %v\n\n", result.Duration)

	return result
}
