// Package exampledata 内置示例数据服务(独立包,方便整体推翻删除)。
//
// C 方案:go:embed rain-robot resources(.bytes,1.4M) + 精简 fight_cases(10 个),
// "加载示例"按钮释放到私有目录 + 配置 function-test 指向,让战斗测试页开箱可用。
//
// 推翻/删除步骤(整体清理):
//  1. 删本包 backend/pkg/exampledata/
//  2. 去 cmd/rain-qa-func/wails.go 的 exampledata 注册(InitWithApp 无,仅 NewService)
//  3. 去 frontend/src/pages/settings/index.vue 的 <ExampleDataCard/> + import
//  4. 删前端 frontend/src/pages/settings/composables/use-example-data.ts
//  5. 删前端 frontend/src/pages/settings/components/ExampleDataCard.vue
//
// 数据来源:rain-robot 源码 project/xcard/xcard_excel/resources(1.4M,189 .bytes),
// 版本匹配 go.mod v0.0.0-20260702143628(2026-07-02)。
package exampledata

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:embed
var embedded embed.FS

const embedFSRoot = "embed"

// ReleaseTo 把 embed 数据递归释放到 destDir,返回写入文件数。
// 系统干净:先 RemoveAll 旧 destDir 再写,多次"加载示例"不累积存储。
func ReleaseTo(destDir string) (files int, err error) {
	if err = os.RemoveAll(destDir); err != nil {
		return 0, fmt.Errorf("清理旧示例数据失败: %w", err)
	}
	if err = os.MkdirAll(destDir, 0755); err != nil {
		return 0, fmt.Errorf("创建示例目录失败: %w", err)
	}
	err = fs.WalkDir(embedded, embedFSRoot, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		data, e := embedded.ReadFile(path)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(embedFSRoot, path)
		if e != nil {
			return e
		}
		target := filepath.Join(destDir, rel)
		if e = os.MkdirAll(filepath.Dir(target), 0755); e != nil {
			return e
		}
		if e = os.WriteFile(target, data, 0644); e != nil {
			return e
		}
		files++
		return nil
	})
	return files, err
}
