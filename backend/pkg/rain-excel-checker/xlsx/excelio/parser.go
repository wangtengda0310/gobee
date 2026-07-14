// Package excel_internal 提供 Excel 文件过滤和解析功能
// 本包负责识别和验证 Excel 配表的格式，过滤出符合项目规范的 Sheet
package excelio

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/xuri/excelize/v2"
)

// FilteredExcel 过滤后的 Excel 文件映射
// 将 Excel 文件映射到其包含的有效 Sheet 名称列表
type FilteredExcel map[*excelize.File][]string

// ExcelFilter 过滤 Excel 文件，识别出符合项目规范的 Sheet
// 遍历所有 Excel 文件的所有 Sheet，检查其表头格式是否符合规范
//
// 参数：
//   - excels: Excel 文件对象列表
//
// 返回值：
//   - res: Excel 文件到 Sheet 列表的映射，包含合法和非法的 Sheet
//   - err: 读取过程中的错误
func ExcelFilter(excels []*excelize.File) (res map[*excelize.File][]*Sheet, err error) {
	res = make(map[*excelize.File][]*Sheet)

	// 遍历每个 Excel 文件的所有 Sheet，检查前4行是否符合要求格式
	for _, excel := range excels {
		excelFileName := path.Base(excel.Path)
		for _, sheet := range excel.GetSheetList() {
			// 提前过滤：只处理"中文|英文"格式的业务配置表
			// 非"中文|英文"格式的 sheet 是策划注释表，跳过加载和解析
			if !IsValidBusinessSheetName(sheet) {
				continue
			}
			rows, err := excel.GetRows(sheet)
			if err != nil {
				return res, err
			}
			// 应用过滤规则检查 Sheet 格式
			if ok, sheetType, sheetHeader, reason := excelFilterRules(excelFileName, sheet, rows); ok {
				// 合法 Sheet
				if sheets, exist := res[excel]; exist {
					res[excel] = append(sheets, &Sheet{SheetType: sheetType, Name: sheet, Header: sheetHeader})
				} else {
					res[excel] = []*Sheet{
						{
							SheetType: sheetType,
							Name:      sheet,
							Header:    sheetHeader,
						},
					}
				}
			} else {
				// 非法 Sheet（记录错误信息）
				if sheets, exist := res[excel]; exist {
					res[excel] = append(sheets, &Sheet{SheetType: sheetType, Name: sheet, Header: sheetHeader, Error: reason})
				} else {
					res[excel] = []*Sheet{
						{
							SheetType: sheetType,
							Name:      sheet,
							Header:    sheetHeader,
							Error:     reason,
						},
					}
				}
			}
		}
	}

	return res, err
}

// IsValidBusinessSheetName 检查 sheet 名称是否符合 "中文|英文" 业务配置格式
// 业务配置表格式：中文描述|英文名（如 "武将|Hero"）
// 非此格式的表为策划注释表，不需要加载和校验
func IsValidBusinessSheetName(name string) bool {
	pipeIndex := strings.Index(name, "|")
	if pipeIndex < 0 {
		return false
	}
	// 中文部分不能为空
	chineseName := strings.TrimSpace(name[:pipeIndex])
	if chineseName == "" {
		return false
	}
	// 英文名不能为空，且必须以字母或下划线开头
	englishName := strings.TrimSpace(name[pipeIndex+1:])
	if englishName == "" {
		return false
	}
	firstChar := englishName[0]
	return (firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_'
}

// excelFilterRules 应用过滤规则检查 Sheet 格式
// 根据项目类型调用对应的过滤规则函数
//
// 参数：
//   - xlsxName: Excel 文件名
//   - sheet: Sheet 名称
//   - rows: Sheet 的所有行数据
//
// 返回值：
//   - res: 是否符合规范
//   - sheetType: Sheet 类型
//   - sheetHeader: Sheet 表头信息
//   - reason: 不符合规范的原因
func excelFilterRules(xlsxName, sheet string, rows [][]string) (res bool, sheetType SheetType, sheetHeader *SheetHeader, reason string) {
	project := "xcard"
	switch project {
	case "xcard":
		// 名将杀项目使用名将杀规则
		return mjsExcelFilterRule(xlsxName, sheet, rows)
	default:
		// 未指定类型
	}
	return false, NONE, &SheetHeader{}, "项目未定义"
}

// mjsExcelFilterRule 名将杀项目的 Excel 过滤规则
// 检查 Sheet 是否符合名将杀配表格式要求
//
// 名将杀配表格式：
//  1. 普通配表：4行表头（中文、类型、字段名、导出标识）
//  2. 枚举配表：1行表头（name、value、sign、description）
//
// 参数：
//   - xlsxName: Excel 文件名
//   - sheet: Sheet 名称
//   - rows: Sheet 的所有行数据
//
// 返回值：
//   - res: 是否符合规范
//   - excelType: Excel 类型
//   - sheetHeader: Sheet 表头信息
//   - reason: 不符合规范的原因
func mjsExcelFilterRule(xlsxName, sheet string, rows [][]string) (res bool, excelType SheetType, sheetHeader *SheetHeader, reason string) {
	header := &SheetHeader{}

	// 检查是否为枚举表（以 _enum.xlsx 结尾）
	if strings.HasSuffix(xlsxName, "_enum.xlsx") {
		// ===== 枚举表检查 =====

		// 行数必须至少有表头行
		if len(rows) < MJS_FIXED_ENUM_ROWS_NUM {
			return false, NONE, header, "行数未达标"
		}

		row1 := rows[MJS_FIXED_ENUM_ROWS_CHS]

		// 检查3列格式：name、value、description
		if len(row1) == 3 && strings.ToLower(row1[0]) == "name" && strings.ToLower(row1[1]) == "value" && strings.ToLower(row1[2]) == "description" {
			header.Col = make([]*ColAttr, 0, len(row1))
			header.Col = append(header.Col, &ColAttr{
				AttrName:  row1[0],
				ColStatus: ENUM,
			}, &ColAttr{
				AttrName:  row1[1],
				ColStatus: ENUM,
			}, &ColAttr{
				AttrName:  row1[2],
				ColStatus: ENUM,
			})
			return true, MING_JIANG_SHA_ENUM, header, ""
		} else if len(row1) >= 4 && strings.ToLower(row1[0]) == "name" && strings.ToLower(row1[1]) == "value" && strings.ToLower(row1[2]) == "sign" && strings.ToLower(row1[3]) == "description" {
			// 检查4列格式：name、value、sign、description
			header.Col = make([]*ColAttr, 0, len(row1))
			header.Col = append(header.Col, &ColAttr{
				AttrName:  row1[0],
				ColStatus: ENUM,
			}, &ColAttr{
				AttrName:  row1[1],
				ColStatus: ENUM,
			}, &ColAttr{
				AttrName:  row1[2],
				ColStatus: ENUM,
			}, &ColAttr{
				AttrName: row1[3],
			})
			return true, MING_JIANG_SHA_ENUM, header, ""
		}

		return false, NONE, header, fmt.Sprintf("枚举表表头格式解析错误，小于3行或3行时1、2、3列不为name value description，")

	} else {
		// ===== 普通配表检查 =====

		// 行数必须至少包含4行表头
		if len(rows) < MJS_FIXED_ROWS_NUM {
			return false, NONE, header, "表头行数未达标，应为4行，分别为中文、类型、属性名、导出到客户端或服务器"
		}

		// 读取前4行表头
		row1 := rows[MJS_FIXED_ROWS_CHS]       // 第1行：中文备注（用于显示）
		row2Types := rows[MJS_FIXED_ROWS_TYPE] // 第2行：单元格类型（关键）
		row3Names := rows[MJS_FIXED_ROWS_NAME] // 第3行：单元格属性名（关键）
		row4 := rows[MJS_FIXED_ROWS_CAS]       // 第4行：导出到服务器还是客户端

		// 第一行不做检查，仅用于显示
		_ = row1

		// ---------------合法表检查---------------------

		// 检查属性和类型数量是否一致
		if len(row2Types) != len(row3Names) {
			return false, NONE, header, "类型和属性数量不一致"
		}

		// 检查是否每个属性都有类型，每个类型是否都有属性
		commentOrEmptyColsIndex := make([]int, 0, len(row4))
		for i := range row2Types {
			if strings.HasPrefix(row2Types[i], "E#") {
				// 枚举类型，枚举表在 enum 文件夹里（_enum.xlsx 后缀）
				continue
			}
			if row2Types[i] == "#" {
				// # 表示策划备注列，不参与导出
				commentOrEmptyColsIndex = append(commentOrEmptyColsIndex, i)
				continue
			}
			if row2Types[i] == "" && row3Names[i] == "" {
				// 两个都是空，可能是空列，不用管
				commentOrEmptyColsIndex = append(commentOrEmptyColsIndex, i)
				continue
			}
			if row2Types[i] == "" || row3Names[i] == "" {
				// 有一个为空则报错
				return false, NONE, header, fmt.Sprintf("类型:[%s]和名称:[%s]不对应", row2Types[i], row3Names[i])
			}
		}

		// 检查第4行的导出标识，只有五种情况：空、server、client、server/client、client/server
		for i, s := range row4 {
			if slices.Contains(commentOrEmptyColsIndex, i) {
				// 空或者注释列跳过检查
				continue
			}
			if s != "" && s != "server" && s != "client" && s != "server/client" && s != "client/server" {
				return false, NONE, header, fmt.Sprintf("[%s]不为空、server、client或server/client, client/server", s)
			}
		}
		// ---------------合法表检查结束---------------------

		// ---------------合法列检查---------------------
		// 收集所有列的属性信息
		header.Col = make([]*ColAttr, 0, len(row2Types))
		for i := range row2Types {
			// 检查导出标识
			if len(row4) > i {
				export := row4[i]
				if export != "" && export != "server" && export != "client" && export != "server/client" && export != "client/server" {
					header.Col = append(header.Col, &ColAttr{
						AttrName:  row3Names[i],
						AttrType:  row2Types[i],
						AttrCHS:   row1[i],
						ColStatus: ERROR,
						Error:     fmt.Sprintf("[%s]不为空、server、client或server/client,client/server", row4[i]),
					})
					continue
				}
			} else {
				header.Col = append(header.Col, &ColAttr{
					AttrName:  row3Names[i],
					AttrType:  row2Types[i],
					AttrCHS:   row1[i],
					ColStatus: ERROR,
					Error:     "不存在导出类型",
				})
				continue
			}

			// 根据类型判断列状态
			if strings.HasPrefix(row2Types[i], "E#") {
				// 枚举类型
				header.Col = append(header.Col, &ColAttr{
					AttrName:  row3Names[i],
					AttrType:  row2Types[i],
					AttrCHS:   row1[i],
					ColStatus: ENUM,
				})
				continue
			}
			if row2Types[i] == "#" {
				// # 表示策划备注列
				header.Col = append(header.Col, &ColAttr{
					AttrName:  row3Names[i],
					AttrType:  row2Types[i],
					AttrCHS:   row1[i],
					ColStatus: COMMENT,
				})
				continue
			}
			if row2Types[i] == "" && row3Names[i] == "" {
				// 空列
				header.Col = append(header.Col, &ColAttr{
					AttrName:  row3Names[i],
					AttrType:  row2Types[i],
					AttrCHS:   row1[i],
					ColStatus: EMPTY,
				})
				continue
			}
			if row2Types[i] == "" || row3Names[i] == "" {
				// 有一个为空则报错
				header.Col = append(header.Col, &ColAttr{
					AttrName:  row3Names[i],
					AttrType:  row2Types[i],
					AttrCHS:   row1[i],
					ColStatus: ERROR,
					Error:     fmt.Sprintf("类型:[%s]和名称:[%s]不对应", row2Types[i], row3Names[i]),
				})
				continue
			}

			// 正常列
			header.Col = append(header.Col, &ColAttr{
				AttrName:  row3Names[i],
				AttrType:  row2Types[i],
				AttrCHS:   row1[i],
				ColStatus: NORMAL,
			})
		}
		// ---------------合法列检查结束---------------------
	}
	return true, MING_JIANG_SHA, header, ""
}
