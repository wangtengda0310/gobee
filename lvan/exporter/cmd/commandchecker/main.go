package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
	"github.com/xuri/excelize/v2"
)

var (
	goDir     = pflag.StringP("godir", "g", "", "Go代码所在目录")
	excelPath = pflag.StringP("excel", "e", "", "Excel文件路径")
	help      = pflag.BoolP("help", "h", false, "显示帮助信息")
)

// Excel中的配置信息
type ExcelConfig struct {
	MethodName string   // R列，方法名
	ParamTypes []string // G列，参数类型列表（数字）
}

// 接口方法信息
type MethodInfo struct {
	Name       string   // 方法名
	ParamTypes []string // 参数类型列表（数字）
	FilePath   string   // 文件路径
}

func main() {
	pflag.Parse()

	if *help || *goDir == "" || *excelPath == "" {
		fmt.Println("用法: commandchecker -g <Go代码目录> -e <Excel文件路径>")
		fmt.Println("选项:")
		pflag.PrintDefaults()
		return
	}

	// 读取Excel配置
	excelConfigs, err := readExcelConfig(*excelPath)
	if err != nil {
		fmt.Printf("读取Excel文件失败: %v\n", err)
		return
	}

	// 解析Go接口文件
	methods, err := parseGoInterfaces(*goDir)
	if err != nil {
		fmt.Printf("解析Go接口文件失败: %v\n", err)
		return
	}

	// 检查方法参数是否匹配
	checkMethodParams(methods, excelConfigs)
}

// 读取Excel配置
func readExcelConfig(excelPath string) (map[string]ExcelConfig, error) {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetList()[0]

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	configs := make(map[string]ExcelConfig)

	// 从第二行开始读取数据（跳过表头）
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 18 { // 确保R列（索引17）存在
			continue
		}

		// 读取R列（方法名）
		methodName := row[17] // R列索引为17
		if methodName == "" {
			continue
		}

		// 读取G列（参数类型）
		paramTypesStr := ""
		if len(row) > 6 { // G列索引为6
			paramTypesStr = row[6]
		}

		// 解析参数类型
		var paramTypes []string
		if paramTypesStr != "" {
			paramTypes = strings.Split(paramTypesStr, ";")
		}

		// 存储配置
		configs[methodName] = ExcelConfig{
			MethodName: methodName,
			ParamTypes: paramTypes,
		}
	}

	return configs, nil
}

// 解析Go接口文件
func parseGoInterfaces(goDir string) ([]MethodInfo, error) {
	var methods []MethodInfo

	// 遍历目录下的所有.go文件
	err := filepath.Walk(goDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.go文件
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// 解析Go文件
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				fmt.Printf("解析文件 %s 失败: %v\n", path, err)
				return nil // 继续处理其他文件
			}

			// 遍历所有声明
			ast.Inspect(node, func(n ast.Node) bool {
				// 查找接口类型声明
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok || typeSpec.Type == nil {
					return true
				}

				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					return true
				}

				// 遍历接口方法
				for _, method := range interfaceType.Methods.List {
					if method.Names == nil || len(method.Names) == 0 {
						continue
					}

					methodName := method.Names[0].Name

					// 获取方法类型
					funcType, ok := method.Type.(*ast.FuncType)
					if !ok {
						continue
					}

					// 解析参数类型
					paramTypes := extractParamTypes(funcType)

					// 添加到方法列表
					methods = append(methods, MethodInfo{
						Name:       methodName,
						ParamTypes: paramTypes,
						FilePath:   path,
					})
				}

				return true
			})
		}

		return nil
	})

	return methods, err
}

// 提取方法参数中的类型数字
func extractParamTypes(funcType *ast.FuncType) []string {
	var paramTypes []string

	// 跳过第一个参数（apiCtx）
	if funcType.Params != nil && len(funcType.Params.List) > 1 {
		for i, param := range funcType.Params.List {
			// 跳过第一个参数（apiCtx）
			if i == 0 {
				continue
			}

			// 获取参数类型字符串
			typeStr := getTypeString(param.Type)

			// 提取类型中的数字
			number := extractTypeNumber(typeStr)
			if number != "" {
				paramTypes = append(paramTypes, number)
			}
		}
	}

	return paramTypes
}

// 获取类型的字符串表示
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	default:
		return ""
	}
}

// 从类型字符串中提取数字
func extractTypeNumber(typeStr string) string {
	// 使用正则表达式提取类型中的数字
	re := regexp.MustCompile(`\d+$`)
	match := re.FindString(typeStr)
	return match
}

// 检查方法参数是否匹配Excel配置
func checkMethodParams(methods []MethodInfo, excelConfigs map[string]ExcelConfig) {
	// 用于跟踪已检查的方法
	checkedMethods := make(map[string]bool)

	// 检查每个方法
	for _, method := range methods {
		checkedMethods[method.Name] = true

		// 查找对应的Excel配置
		config, ok := excelConfigs[method.Name]
		if !ok {
			fmt.Printf("警告: Excel中未找到方法 %s 的配置\n", method.Name)
			continue
		}

		// 检查参数数量是否匹配
		if len(method.ParamTypes) != len(config.ParamTypes) {
			fmt.Printf("警告: 方法 %s 的参数数量不匹配 (Go: %d, Excel: %d)\n",
				method.Name, len(method.ParamTypes), len(config.ParamTypes))
			continue
		}

		// 检查每个参数类型是否匹配
		mismatch := false
		for i, paramType := range method.ParamTypes {
			excelType := config.ParamTypes[i]
			if paramType != excelType {
				mismatch = true
				fmt.Printf("警告: 方法 %s 的参数 #%d 类型不匹配 (Go: %s, Excel: %s)\n",
					method.Name, i+1, paramType, excelType)
			}
		}

		if !mismatch {
			fmt.Printf("方法 %s 的参数匹配成功\n", method.Name)
		}
	}

	// 检查Excel中有但Go代码中没有的方法
	for methodName := range excelConfigs {
		if !checkedMethods[methodName] {
			fmt.Printf("警告: Go代码中未找到Excel配置的方法 %s\n", methodName)
		}
	}
}