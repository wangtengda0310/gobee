// Package check_internal 提供校验规则的内部辅助工具
// 本包包含列检查、参数解析、表查找等通用辅助函数
package ruleconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// SaveCheck 保存校验规则到文件
//
// 执行流程：
//  1. 创建目标目录（如果不存在）
//  2. 初始化WaitGroup和错误通道用于并发控制
//  3. 并发保存每个校验规则：
//     a. 将规则序列化为JSON格式
//     b. 构造文件路径（将Sheet名中的"|"替换为"_"）
//     c. 写入JSON文件
//  4. 等待所有并发任务完成
//  5. 检查是否有错误，有则返回错误
//  6. 返回nil表示保存成功
func SaveCheck(dir string, checkList []*json_rule.SheetRule) error {
	// 步骤1: 创建目标目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 步骤2: 初始化WaitGroup和错误通道
	var wg sync.WaitGroup
	errCh := make(chan error, len(checkList))

	// 步骤3: 并发保存每个校验规则
	for _, rule := range checkList {
		wg.Add(1)
		go func(rule *json_rule.SheetRule) {
			defer wg.Done()

			// 步骤3a: 将规则序列化为JSON格式
			jsonData, err := json.MarshalIndent(rule, "", " ")
			if err != nil {
				errCh <- err
				return
			}

			// 步骤3b: 构造文件路径
			finalFilePath := fmt.Sprintf("%s/%s.json", dir,
				strings.ReplaceAll(rule.Sheet, "|", "_"))

			// 步骤3c: 写入JSON文件
			if err := os.WriteFile(finalFilePath, jsonData, 0644); err != nil {
				errCh <- err
			}
		}(rule)
	}

	// 步骤4: 等待所有并发任务完成
	wg.Wait()
	close(errCh)

	// 步骤5: 检查是否有错误
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	// 步骤6: 返回nil表示保存成功
	return nil
}

// LoadJsonRules 从目录加载校验规则 JSON 文件
//
// 执行流程：
//  1. 规范化目录路径（去除末尾的"/"）
//  2. 使用Glob模式查找目录下所有JSON文件
//  3. 初始化WaitGroup、规则切片和错误通道
//  4. 并发读取每个JSON文件：
//     a. 读取文件内容
//     b. 使用json.Decoder解析JSON（保持数字为json.Number类型）
//     c. 按原始顺序将规则存入切片
//  5. 等待所有并发任务完成
//  6. 检查是否有错误，有则返回错误
//  7. 过滤掉nil值，返回有效的规则列表
func LoadJsonRules(dir string) ([]*json_rule.SheetRule, error) {
	// 步骤1: 规范化目录路径
	okDir, _ := strings.CutSuffix(dir, "/")

	// 步骤2: 查找所有JSON文件
	files, err := filepath.Glob(fmt.Sprintf("%s/*.json", okDir))
	if err != nil {
		return nil, err
	}

	// 步骤3: 初始化并发控制
	var wg sync.WaitGroup
	loadedRules := make([]*json_rule.SheetRule, len(files))
	errors := make(chan error, len(files))

	// 步骤4: 并发读取每个JSON文件
	for i, file := range files {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()

			// 步骤4a: 读取文件内容
			data, err := os.ReadFile(filePath)
			if err != nil {
				errors <- fmt.Errorf("读取文件 %s 失败: %v", filePath, err)
				return
			}

			var rule *json_rule.SheetRule
			// 步骤4b: 使用json.Decoder解析JSON
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.UseNumber() // 保持数字为 json.Number 类型
			if err := decoder.Decode(&rule); err != nil {
				errors <- fmt.Errorf("解析JSON文件 %s 失败: %v", filePath, err)
				return
			}

			// 步骤4c: 按原始顺序将规则存入切片
			loadedRules[index] = rule
		}(i, file)
	}

	// 步骤5: 等待所有并发任务完成
	wg.Wait()
	close(errors)

	// 步骤6: 检查错误
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	// 步骤7: 过滤掉nil值
	var result []*json_rule.SheetRule
	for _, rule := range loadedRules {
		if rule != nil {
			result = append(result, rule)
		}
	}

	return result, nil
}
