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

## CLI (sheetcmp-merge)

```bash
go run ./cmd/sheetcmp-merge -key 0 left.xlsx right.xlsx   # 表格模式, ID 列作关键列
go run ./cmd/sheetcmp-merge a.txt b.txt                   # 文本模式 (行 diff)
go run ./cmd/sheetcmp-merge -doc                          # 查看本 README
```

文件按扩展名分流（.xlsx/.xlsm/.csv → 表格模式；其余 → 文本行 diff）。
退出码：0=无差异，1=有差异，2=用法/错误——可直接作 git difftool 使用；
merge tool 形态（GUI 解决冲突后写回）待桌面版完成后接入。

## 开发阶段

1. [x] engine：行对齐 + 单元格 diff（18 用例）
2. [x] xlsx 读写层（excelize：Load/ApplyCellEdits 原位写回/LoadCSV，9 用例）
3. [x] textdiff 行 diff（go-diff，6 用例）+ sheetcmp-merge CLI
4. [x] Wails v3 + Vue3 双栏界面（单滚动容器同步滚动/差异高亮/公式差异红框/文本模式）
5. [x] 行同步（点击差异格取对侧值, 攒批）/ 撤销重做 / 只看差异
6. [x] merge GUI 形态 + Git/UGit 接入

## Git / UGit 接入

```powershell
# 一键配置当前用户的差异/合并工具 (需先构建 bin/sheetcmp.exe)
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\configure-git.ps1
```

写入的配置（信任退出码，UGit/GitGUI 生效，已打开的 UGit 需重启）：

```ini
[difftool "sheetcmp"]   cmd = "<exe>" "$LOCAL" "$REMOTE"
[mergetool "sheetcmp"]  cmd = "<exe>" -MergeTool "$REMOTE" "$MERGED"
                       trustExitCode = true
diff.guitool = sheetcmp ; merge.guitool = sheetcmp
```

合并模式语义：左=REMOTE（对方改动）、右=MERGED（本地副本，写回目标）；
「✓ 完成合并」= 保存全部待写编辑 + 退出码 0；直接关窗 = 放弃（退出码 1，git 保留冲突状态）。

## 同步语义

- 点击差异格 = **目标侧此格取对侧值**（只写目标侧）；待写回格蓝色虚线框 + ✓。
- 编辑先攒批，**保存写回**时按目标文件分组一次性原位写回（xlsx/xlsm；CSV 只读比对），
  对侧无公式而本侧有公式时自动清公式。撤销/重做即攒批列表操作，写回前任意回退。
- 行插入/删除（单侧独有行的同步）暂未实现——单元格级同步仅覆盖配对行。

## 开发说明

```bash
# GUI 本机构建
cd frontend && npm install && npm run build && cd ..
go build -o bin/sheetcmp.exe ./cmd/sheetcmp

# 服务方法变更后重新生成前端 bindings
# ⚠ -d 必须指向 frontend/bindings 子目录: -clean 会清空整个 -d 目录,
#   误写 -d frontend 会把 node_modules/package.json 全部清掉 (实测踩坑)
wails3 generate bindings -clean -d frontend/bindings ./cmd/sheetcmp

# 前端开发热更 (先起 vite 再跑 go, 或用 wails3 dev)
cd frontend && npm run dev
```

架构注：根包 sheetcmp 集中 go:embed（README + frontend/dist），
GUI 主程序在 cmd/sheetcmp、CLI 在 cmd/sheetcmp-merge，二者共用根包资源与服务层。
