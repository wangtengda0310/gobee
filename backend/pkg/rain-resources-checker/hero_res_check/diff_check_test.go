package hero_res_check

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
)

func TestGitlabAndLocalDiffInExcel(t *testing.T) {
	comparator := NewComparator("../mjs_excel/tmp/test.json")

	// 模拟当前获取的数据
	currentHeroes := &diff.DataContainer{
		HeroDiff: &[]hero.HeroDiff{
			{
				Id:              10001,
				Name:            "英雄A",
				IsOpen:          true,
				OpenDate:        "2026-01-01",
				Gender:          1,
				Point:           100,
				HpLimit:         1000,
				Skill:           []int{1001, 2, 3},
				ExcludeIdentity: []int{101},
			},
			{
				Id:       2,
				Name:     "a",
				IsOpen:   false,
				OpenDate: "2024-02-01",
				Gender:   2,
				Point:    200,
				HpLimit:  1500,
				Skill:    []int{4, 5, 7},
			},
			// ... 更多数据
		},
	}

	// 加载上一次的数据
	oldMap, err := comparator.LoadPreviousData()
	if err != nil {
		fmt.Printf("没有找到历史数据或加载失败: %v\n", err)
		oldMap = &diff.DataContainer{}
	}

	// 执行对比
	result := comparator.CompareAll(oldMap, currentHeroes)

	printer := NewReportGenerator()
	// 打印摘要
	printer.PrintSummary(result)

	// 打印详细修改
	printer.PrintDetailedChanges(result)

	// 生成报告
	if report, err := printer.GenerateJSONReport(result); err != nil {
		fmt.Printf("保存生成失败: %v\n", err)
	} else {
		reportJSON, _ := json.MarshalIndent(report, "", "  ")

		if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
		} else {
			err := os.WriteFile("./tmp/diff_report.json", reportJSON, 0644)
			if err != nil {
				fmt.Printf("保存报告失败: %v\n", err)
			}
		}
	}

	// 保存当前数据供下次对比使用
	//if err := comparator.SaveCurrentData(currentHeroes); err != nil {
	//	fmt.Printf("保存数据失败: %v\n", err)
	//}
}
