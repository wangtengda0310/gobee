# 内层 rain-qa-func/ → backend/ 改名设计

> **日期**：2026-06-18
> **状态**：待用户审查
> **范围**：内层 `rain-qa-func/` 目录改名为 `backend/`（module 名保留）

## 1. 背景

项目结构有双层 `rain-qa-func` 同名混淆：

```
xcard-qa-tools/rain-qa-func/              # 外层（= worktree 根，module 根）
├── rain-qa-func/                          # 内层（与外层同名！含 pkg/）
│   └── pkg/
│       ├── proto-test/
│       ├── server-config/  (原 streamproxy)
│       └── ...
├── frontend/
├── cmd/
└── go.mod  # module git.devcloud.ztgame.com/v-tangfangda/rain-qa-func
```

导致 import 路径双层 `.../rain-qa-func/rain-qa-func/pkg/X`，AI/开发者混淆内外层。

## 2. 目标

- 内层 `rain-qa-func/` → `backend/`，与 `frontend/` 对称（`backend/pkg/` vs `frontend/`）
- module 名 `git.devcloud.ztgame.com/v-tangfangda/rain-qa-func` **保留不变**（外层）
- 消除双层同名混淆
- import 路径变为 `.../rain-qa-func/backend/pkg/X`（清晰）

## 3. 映射

| 当前 | 目标 |
|------|------|
| `rain-qa-func/`（内层目录） | `backend/` |
| import `.../rain-qa-func/rain-qa-func/pkg/X` | `.../rain-qa-func/backend/pkg/X` |
| bindings `frontend/bindings/.../rain-qa-func/rain-qa-func/` | `frontend/bindings/.../rain-qa-func/backend/` |
| 前端 `@bindings/.../rain-qa-func/rain-qa-func/` | `@bindings/.../rain-qa-func/backend/` |
| module `git.devcloud.ztgame.com/v-tangfangda/rain-qa-func` | **保留不变** |

**不变项**：各包 package 名（`proto-test`/`serverconfig` 等）不变，只改路径中段；`cmd/`、`frontend/` 顶层目录不变。

## 4. 影响范围

- **290 个 Go 文件** import 路径中段（`rain-qa-func/rain-qa-func/` → `rain-qa-func/backend/`）
- 前端 bindings import（`@bindings/.../rain-qa-func/rain-qa-func/` → `.../rain-qa-func/backend/`）
- 文档（各级 CLAUDE.md、layout docs、specs 等含路径引用）

## 5. 执行策略（分阶段，机械替换）

**阶段 A — Go 目录改名 + import 替换**
1. `git mv rain-qa-func backend`
2. sed 批量替换所有 `.go` import 路径中段：`find . -name "*.go" | xargs sed -i 's|rain-qa-func/rain-qa-func/|rain-qa-func/backend/|g'`（精确匹配双层 `rain-qa-func/rain-qa-func/`，唯一模式不误伤单层 module 名前缀；处理**所有 import 含 blank import**，比 goimports 可靠 —— goimports 基于符号匹配，漏 blank import 如 `_ ".../proto-test/msg"`）
3. `goimports -w`（批量，仅排序/清理兜底，不依赖符号匹配）
4. `gofmt -w` + `go build ./...` + `go test ./...`
5. 若 go build 报个别 import 未修复（goimports 边角），`gofiximports` 显式替换兜底

**阶段 B — bindings 重生成 + 前端 import**
1. 删除旧 bindings 目录 `frontend/bindings/.../rain-qa-func/rain-qa-func/`
2. `wails3 generate bindings -ts`（生成到 `.../rain-qa-func/backend/`）
3. 批量替换前端 `@bindings/.../rain-qa-func/rain-qa-func/` → `.../rain-qa-func/backend/`
4. 前端 `tsc --noEmit`

**阶段 C — 文档同步**
1. 各 CLAUDE.md、layout docs、specs 中 `rain-qa-func/rain-qa-func/` → `rain-qa-func/backend/`、`rain-qa-func/pkg/` → `backend/pkg/`（描述性引用）
2. 注意：外层 `xcard-qa-tools/rain-qa-func/`（项目名）引用**不改**，只改内层路径

**阶段 D — build + 验证**
1. `wails3 task build`（重建 rain-qa-func.exe）
2. 全量验证：`go build ./...` + `go test ./...` + 前端 `tsc`

## 6. 边界（YAGNI）

- module 名 `rain-qa-func` 保留（不改 go.mod module 行）
- 外层目录 `xcard-qa-tools/rain-qa-func/`（项目根）不改
- `cmd/`（含 cmd/tests/streamproxy）、`frontend/` 顶层目录不改
- 各包 package 名不改（只改目录路径中段）
- 不含命名重构（已完成）、不含页签目录重组

## 7. 风险与缓解

- **sed 误伤**：`rain-qa-func/rain-qa-func/` 是明确的双层模式，sed 精确匹配该子串，不误伤单层 `rain-qa-func/`（module 名前缀或外层引用）。替换后 grep 验证无残留双层。
- **bindings 生成**：用 `-ts` 标志（命名重构教训：漏 -ts 生成 .js）
- **290 文件改动**：机械替换，go build/test 全量验证兜底
