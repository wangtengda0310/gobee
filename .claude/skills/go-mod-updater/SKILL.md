---
name: go-mod-updater
description: |
  xcard-qa-tools 子模块 go.mod 依赖更新 — 当子模块代码被修改并推送后，按依赖链顺序更新所有下游模块的 go.mod，确保流水线和本地调试使用最新代码。

  触发条件（满足任一即触发）：
  - 子模块代码修改完成并推送后，用户要求更新依赖、同步依赖、更新 go.mod
  - 用户提到"更新 go.mod"、"更新依赖"、"同步依赖"、"go get @latest"
  - 用户完成某个子模块的开发任务，需要级联更新下游模块
  - 用户要求检查各模块依赖是否为最新版本
  - 用户提到"依赖过期"、"依赖不是最新的"、"版本不对"
  - 每次子模块 commit 并 push 后，作为收尾步骤自动触发

  不要触发的场景：
  - 仅在单个子模块内开发，未修改对外接口/类型
  - 与 go.mod 无关的纯前端修改
  - 用户明确表示不需要更新依赖

  适用领域：xcard-qa-tools 子模块依赖管理、Go 模块版本同步
---

# xcard-qa-tools 子模块 go.mod 依赖更新

## 一、模块依赖关系

```
feishu-lib ──→ rain-excel-checker（类型定义 json_rule）
rain-excel-checker ──→ feishu-lib（通知功能）
rain-resources-checker ──→ rain-excel-checker
rain-qa-func ──→ feishu-lib + rain-excel-checker + rain-resources-checker + rain-robot
rain-robot（独立，不依赖其他子模块）
```

**关键特性**：feishu-lib 和 rain-excel-checker 存在循环依赖。

## 二、模块信息表

| 子模块 | Go 模块名 | 远程仓库分支 |
|--------|----------|-------------|
| feishu-lib | `git.devcloud.ztgame.com/v-tangfangda/feishu-lib` | main |
| rain-excel-checker | `git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker` | main |
| rain-resources-checker | `git.devcloud.ztgame.com/v-tangfangda/rain-resources-checker` | main |
| rain-qa-func | `rain-qa-func`（本地 replace） | main |
| rain-robot | `rain-robot`（本地 replace） | master |

**注意**：rain-robot 使用 master 分支，其他子模块使用 main 分支。

## 三、更新触发矩阵

修改了某个子模块后，必须按以下顺序更新下游：

| 修改了 | 需要更新的下游（按顺序） |
|--------|----------------------|
| feishu-lib | rain-excel-checker → rain-resources-checker → rain-qa-func |
| rain-excel-checker | feishu-lib → rain-resources-checker → rain-qa-func |
| rain-resources-checker | rain-qa-func |
| rain-robot | rain-qa-func（仅 go.mod replace，不需要 go get） |

## 四、完整更新流程

### 步骤 1：确认触发源

先确认哪个模块刚被修改并推送：

```bash
# 在各子模块中检查是否有新的未同步提交
cd <子模块目录> && git log --oneline -3
```

### 步骤 2：按依赖链从底层到上层逐个更新

对每个下游模块执行：

```bash
cd <下游模块目录>
go get <被修改模块的Go模块名>@latest
go mod tidy
```

**完整的 go get 命令列表**：

```bash
# 更新 feishu-lib
go get git.devcloud.ztgame.com/v-tangfangda/feishu-lib@latest

# 更新 rain-excel-checker
go get git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest

# 更新 rain-resources-checker
go get git.devcloud.ztgame.com/v-tangfangda/rain-resources-checker@latest
```

**同时更新多个依赖**（当多个上游模块都变更时）：

```bash
go get git.devcloud.ztgame.com/v-tangfangda/feishu-lib@latest git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest git.devcloud.ztgame.com/v-tangfangda/rain-resources-checker@latest
```

### 步骤 3：每个模块更新后立即提交推送

```bash
git add go.mod go.sum
git commit -m "chore: 更新 <依赖名> 依赖

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
git push
```

注意推送命令：
- 大部分子模块：`git push`
- rain-qa-func：`git push origin HEAD:refs/heads/main`

### 步骤 4：同步主仓库子模块引用

所有子模块更新完成后，在主仓库 `xcard-qa-tools` 中：

```bash
cd /d/work/xcard-qa-tools
git add <变更的子模块>
git commit -m "chore: 同步子模块依赖更新

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
git push
```

## 五、更新前检查清单

更新前先确认以下事项，避免浪费操作：

1. **确认修改已推送**：被修改的子模块必须先 `git push`，否则下游 `go get @latest` 拉不到最新版本
2. **检查当前版本**：`cat go.mod | grep <模块名>` 查看当前依赖版本
3. **检查远程最新**：`git fetch origin && git log --oneline origin/main -3` 确认远程有新提交
4. **只在需要时更新**：对比当前 go.mod 中的 pseudo-version 和远程最新提交，相同则跳过

## 六、常见场景

### 场景 A：仅修改 feishu-lib

```bash
# 1. feishu-lib 已 push
cd /d/work/xcard-qa-tools/rain-excel-checker
go get git.devcloud.ztgame.com/v-tangfangda/feishu-lib@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新 feishu-lib 依赖" && git push

cd /d/work/xcard-qa-tools/rain-resources-checker
go get git.devcloud.ztgame.com/v-tangfangda/feishu-lib@latest git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新依赖" && git push

cd /d/work/xcard-qa-tools/rain-qa-func
go get git.devcloud.ztgame.com/v-tangfangda/feishu-lib@latest git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest git.devcloud.ztgame.com/v-tangfangda/rain-resources-checker@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新依赖" && git push origin HEAD:refs/heads/main

cd /d/work/xcard-qa-tools
git add feishu-lib rain-excel-checker rain-resources-checker rain-qa-func
git commit -m "chore: 同步子模块依赖更新" && git push
```

### 场景 B：仅修改 rain-excel-checker

```bash
# 1. rain-excel-checker 已 push
cd /d/work/xcard-qa-tools/feishu-lib
go get git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新 rain-excel-checker 依赖" && git push

cd /d/work/xcard-qa-tools/rain-resources-checker
go get git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新依赖" && git push

cd /d/work/xcard-qa-tools/rain-qa-func
go get git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker@latest git.devcloud.ztgame.com/v-tangfangda/rain-resources-checker@latest && go mod tidy
git add go.mod go.sum && git commit -m "chore: 更新依赖" && git push origin HEAD:refs/heads/main

cd /d/work/xcard-qa-tools
git add feishu-lib rain-excel-checker rain-resources-checker rain-qa-func
git commit -m "chore: 同步子模块依赖更新" && git push
```

### 场景 C：全量检查所有模块依赖

当不确定哪些模块需要更新时，逐个检查：

```bash
# 检查各模块当前版本 vs 远程最新
for module in feishu-lib rain-excel-checker rain-resources-checker; do
    echo "=== $module ==="
    cd /d/work/xcard-qa-tools/$module
    git fetch origin
    local=$(git rev-parse HEAD)
    remote=$(git rev-parse origin/main)
    if [ "$local" != "$remote" ]; then
        echo "需要 pull: local=$local remote=$remote"
    else
        echo "已是最新"
    fi
done
```

然后对比各下游模块 go.mod 中的 pseudo-version 与远程最新提交是否一致。

## 七、注意事项

1. **rain-robot 不参与 go get 更新**：使用本地 replace，不通过远程拉取
2. **go.work 文件不提交**：worktree 中的 go.work 是本地配置，不应提交到 git
3. **合并使用 rebase 而非 merge**：项目约定
4. **提交信息使用中文描述 + 英文前缀**：如 `chore: 更新 feishu-lib 依赖`
5. **主仓库提交时列出所有变更的子模块**：确保 git add 包含所有变更的子模块目录
