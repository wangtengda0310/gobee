# rain-resources-checker 项目说明

游戏资源检查工具，包含英雄资源、Wiki 资源、图片、语音等检查功能及武将 Wiki 数据差异对比。

已合并原 `pkg/hero-voice-resource-check/` 的武将语音资源检查功能。

## 技术栈
- Go 1.25
- Import 路径前缀: `git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/`
- 依赖: `pkg/common/feishu` (飞书通知), `rain-excel-checker` (Excel 解析)

## 目录结构

```
pkg/rain-resources-checker/
├── check_log/             # 日志封装 (Debug/Info/Warn/Error/Panic)
├── diff/                  # 差异数据结构 (DiffResult/TypeDiffResult/FieldChange)
├── herowiki_def/          # Wiki 数据结构 (HeroWikiDiff/HeroCompleteData/HeroSkillInfo)
├── hero_res_check/        # 英雄资源检查核心
│   ├── voice_check.go     # 语音检查 (CheckGitlabAndLocalVoicesInExcel)
│   ├── img_check.go       # 图片检查 (CheckGitlabAndLocalImgInExcel)
│   └── diff_check.go      # 差异对比 (Comparator/ReportGenerator)
├── herowiki/              # Wiki 数据格式化
│   ├── format.go          # BuildHeroWikiDiff
│   └── drop_builder.go    # BuildHeroDropInfo
├── wails.go               # Wails 前端绑定（武将语音资源检查入口，从 pkg/hero-voice-resource-check 合并）
├── mjs_excel/             # Excel 数据映射（25+ 子模块，每个含 def/diff_map/voice_map）
│   ├── diff_excel_init.go # 差异相关初始化
│   ├── voice_excel_init.go
│   ├── img_excel_init.go
│   ├── hero/ skill/ buff/ country/ ...  # 各表结构定义和映射
│   └── utils/utils.go
└── resource_utils/        # 资源列表获取（本地 + GitLab）
    ├── get_list.go        # GetFilesLocal
    └── gitlab/            # GitLab 文件获取
```

## 核心类型

| 类型 | 文件位置 | 说明 |
|------|----------|------|
| `HeroResCheckService` | wails.go:8 | 武将语音资源检查服务（Wails 前端绑定） |
| `DiffResult` | diff/interface.go | 总体对比结果 |
| `TypeDiffResult` | diff/interface.go | 单类型对比（Added/Removed/Changed） |
| `DataContainer` | diff/interface.go | 数据容器 |
| `HeroWikiDiff` | herowiki_def/def.go | 整合所有武将数据 |
| `HeroCompleteData` | herowiki_def/def.go | 单个武将完整数据 |

## 核心函数

| 函数 | 文件位置 | 职责 |
|------|----------|------|
| `NewHeroResCheckService` | wails.go:11 | 创建武将语音资源检查服务 |
| `Check` | wails.go:18 | 检查武将语音资源（Wails 前端方法） |
| `CheckGitlabAndLocalVoicesInExcel` | hero_res_check/voice_check.go | 语音检查 |
| `CheckGitlabAndLocalImgInExcel` | hero_res_check/img_check.go | 图片检查 |
| `Comparator.CompareAll` | hero_res_check/diff_check.go | 差异对比 |
| `Comparator.CompareHeroWikiDiff` | hero_res_check/diff_check.go | Wiki 差异 |
| `ReportGenerator.PrintSummary` | hero_res_check/diff_check.go | 报告生成 |
| `BuildHeroWikiDiff` | herowiki/format.go | Wiki 构建 |

## 检查流程概览

- **语音检查**：并行加载本地音频列表和 Excel 配置 -> 遍历武将检查音频存在性 -> 返回报告
- **图片检查**：并行加载本地图片列表和 Excel 配置 -> 遍历皮肤检查资源存在性 -> 输出缺失
- **差异对比**：加载历史数据 -> 保存当前数据 -> 对比所有类型 -> 生成报告

## mjs_excel 子模块

每个子模块统一包含：`def.go`（结构定义）、`diff_map.go`（差异映射）、可选 `voice_map.go`/`img_map.go`。

已支持 25+ 张表：武将、技能、Buff、国家、武将台词、皮肤相关（物品/Spine/收藏册）、武将UI、技能相关（台词/熔炼/标签/UI）、掉落相关（组/物品/规则）、推荐布阵、机器人行为、赛季通行证奖励、竞技场奖励、成就、任务条件、音频等。

## 使用方式

作为库被 `rain-qa-func` 引用，不独立运行。

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/hero-voice-resource-check/hero-voice-resource-check.spec.ts`](../../../frontend/e2e/hero-voice-resource-check/hero-voice-resource-check.spec.ts) | 页面加载、资源检查执行、语音/图片检查、差异报告展示 |

## 依赖关系

```
rain-qa-func -> pkg/rain-resources-checker
                    ├── pkg/common/feishu (飞书通知)
                    └── rain-excel-checker (Excel 解析)
```
