# sheetcmp — 表格双向比对与合并工具

对两份表格（xlsx/CSV）做双向比对：按关键列（或行序）对齐行、单元格级差异高亮、
单向行同步、解决冲突后写回。可作 Git/UGit 的差异与合并工具（`-MergeTool` 形态）。

## 设计

```
sheetcmp/
├── main.go               # Wails v3 桌面入口 (Go + Vue3)
├── cmd/sheetcmp-merge/   # 纯命令行合并入口 (git merge tool 形态, 无 GUI 依赖)
├── internal/engine/      # ★ 核心引擎: 行对齐 + 单元格 diff (纯逻辑, 无 IO, 单测主战场)
├── internal/xlsx/        # 表格读写 (excelize: 流式读 + 原生写回, 零 Excel 依赖)
├── internal/textdiff/    # 文本类文件按行 diff (代码/JSON/CSV 等普通文本)
└── frontend/             # Vue3 双栏表格界面 (同步滚动/差异高亮/行同步)
```

分层原则：engine 不依赖 GUI 与文件格式，输入输出全部为纯数据结构；
GUI、CLI、merge 工具三种形态共用同一 engine。

## 引擎语义（测试即规格）

- **行对齐**：无关键列时按行号位置配对（中间插行会连锁错位——处理插入/乱序请选关键列，
  曾实测整行 LCS 会让有差异的行全部落成删除+新增对、单元格 diff 空转，故弃用）；
  指定关键列时按关键列值匹配（行序无关），关键列全空的行与对方同类行按行序兜底配对，
  同 key 多行按出现顺序依次配对。
- **单元格差异**：值或公式任一不同即差异；双侧有值不同 = modified，
  仅左 = left-only，仅右 = right-only。
- **行同步**：只写目标侧，不动源侧（与界面上"左右"的语义解耦）。

## 开发阶段

1. [进行中] engine 测试用例 → 审核后实现
2. xlsx 读写层（excelize）
3. Wails v3 + Vue3 双栏界面
4. 行同步 / 撤销重做 / 只看差异（缓存复用）
5. merge CLI + Git/UGit 接入
