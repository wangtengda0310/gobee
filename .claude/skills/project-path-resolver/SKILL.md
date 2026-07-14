---
name: project-path-resolver
description: |
  项目路径解析技能。当AI需要知道项目相关路径（客户端代码、服务端代码、策划配表、QA工具等）
  但不确定具体位置时，使用此技能进行路径发现和确认。

  触发场景：
  - 任何技能需要访问客户端代码、服务端代码、策划配表目录前
  - 用户提到"搜索客户端代码"、"查看服务端"、"读取配表"等操作
  - 当前会话尚未确认过项目路径
  - 发现硬编码路径与实际环境不符

  本技能负责路径发现、确认和缓存，不执行实际业务操作。
---

# 项目路径解析技能

## 概述

本技能用于在会话中确认和缓存项目相关路径。由于不同开发者的机器环境不同
（路径不同、仓库位置不同、多版本并存），不能依赖硬编码路径或本地配置文件。

## 路径发现工作流

### 步骤1：检查会话缓存

首先检查当前会话是否已确认过路径：

```bash
# 检查环境变量或临时标记
echo $PROJECT_CLIENT_PATH
echo $PROJECT_SERVER_PATH
echo $PROJECT_CONFIG_PATH
```

如果环境变量已设置，直接使用，跳过后续步骤。

### 步骤2：自动发现候选路径

如果会话中未缓存路径，尝试自动发现：

**2.1 检查当前工作目录**

```bash
# 获取当前git仓库根目录
git rev-parse --show-toplevel 2>/dev/null

# 检查是否在rain-qa-func目录下
pwd | grep -i "rain-qa-func"
```

**2.2 检查常见路径模式**

| 路径类型 | 常见位置 | 验证方式 |
|----------|----------|----------|
| 客户端代码 | `../client`, `../../client`, `d:/work/client` | 检查是否存在 `.cs` 文件或 Unity 目录 |
| 服务端代码 | `../server`, `../../server`, `d:/work/server` | 检查是否存在服务端代码文件 |
| 策划配表 | `../config`, `../../config`, `d:/work/config` | 检查是否存在 `.xlsx` 文件 |
| QA工具 | 当前目录或 `../rain-qa-func` | 检查是否存在 `wails.json` 或 `main.go` |

**2.3 验证候选路径**

对每个候选路径，验证其有效性：

```bash
# 验证客户端路径
[ -d "$CLIENT_PATH" ] && ls "$CLIENT_PATH" | head -5

# 验证服务端路径
[ -d "$SERVER_PATH" ] && ls "$SERVER_PATH" | head -5

# 验证配表路径
[ -d "$CONFIG_PATH" ] && ls "$CONFIG_PATH/excel" 2>/dev/null | head -5
```

### 步骤3：处理发现结果

**场景A：只找到一个有效路径**

向用户确认：
> 发现客户端代码目录：`D:/work/client`，是否正确？
> - 正确 → 缓存并使用
> - 不正确 → 进入步骤4

**场景B：找到多个候选路径**

列出所有候选，让用户选择或指定：
> 发现多个可能的客户端代码目录：
> 1. `D:/work/client`
> 2. `D:/work/client-v2`
> 3. 其他（手动输入）

**场景C：未找到任何候选路径**

直接进入步骤4。

### 步骤4：询问用户

当自动发现失败时，询问用户：

> 未能在常见位置找到项目路径，请提供以下信息（可直接回车跳过不需要的）：
> - 客户端代码目录（如 `D:/work/client`）：
> - 服务端代码目录（如 `D:/work/server`）：
> - 策划配表目录（如 `D:/work/config`）：

### 步骤5：缓存路径

将确认后的路径缓存到会话中：

```bash
# 设置环境变量（当前会话有效）
export PROJECT_CLIENT_PATH="D:/work/client"
export PROJECT_SERVER_PATH="D:/work/server"
export PROJECT_CONFIG_PATH="D:/work/config"
export PROJECT_QA_TOOLS_PATH="D:/work/xcard-qa-tools/rain-qa-func"
```

同时记录到会话记忆文件（供后续查看）：

```bash
# 写入临时配置文件（当前会话目录）
cat > .claude/session_paths.json << 'EOF'
{
  "client": "D:/work/client",
  "server": "D:/work/server",
  "config": "D:/work/config",
  "qa_tools": "D:/work/xcard-qa-tools/rain-qa-func",
  "confirmed_at": "2026-04-29T10:00:00",
  "source": "user_input"
}
EOF
```

## 路径使用规范

### 在其他技能中引用路径

其他技能不应直接硬编码路径，而应：

1. **优先检查环境变量**：
   ```bash
   CLIENT_PATH=${PROJECT_CLIENT_PATH:-""}
   if [ -z "$CLIENT_PATH" ]; then
       # 调用路径解析技能
   fi
   ```

2. **提供相对路径作为备选**：
   ```bash
   # 如果绝对路径未设置，尝试相对路径
   CLIENT_PATH=${PROJECT_CLIENT_PATH:-"../client"}
   ```

3. **在技能文档中说明**：
   > 路径解析：本技能需要知道客户端代码位置。如果会话中未确认过路径，
   > 会先执行路径发现流程。你也可以提前运行 `project-path-resolver` 技能确认路径。

### 路径格式转换

根据当前shell环境自动转换路径格式：

| 环境 | 输入格式 | 转换后 |
|------|----------|--------|
| Git Bash | `D:/work/client` | `/d/work/client` |
| PowerShell | `D:/work/client` | `D:\work\client` |
| CMD | `D:/work/client` | `D:\work\client` |

## 多仓库/多版本支持

当用户有多个版本或分支时：

1. **为每个仓库单独确认路径**：
   ```bash
   export PROJECT_CLIENT_PATH_V1="D:/work/client"
   export PROJECT_CLIENT_PATH_V2="D:/work/client-v2"
   ```

2. **在任务级别指定路径**：
   > 本次任务使用 `D:/work/client-v2` 作为客户端代码目录。

3. **路径别名**：
   ```bash
   export PROJECT_CLIENT_PATH="D:/work/client-v2"
   ```

## 与其他技能的协作

### 调用方式

其他技能在需要路径时，应：

1. 检查环境变量是否存在
2. 如果不存在，提示用户运行路径解析
3. 或者自动调用路径解析流程

### 示例：activity-wiki-dev 技能中的使用

```markdown
## 查看客户端代码

**前置条件**：确保路径已解析

如果 `PROJECT_CLIENT_PATH` 未设置，先执行路径发现：
> 需要访问客户端代码来分析活动预制体类型。
> 正在检查项目路径...

搜索方式：
```bash
CLIENT_PATH=${PROJECT_CLIENT_PATH:-"../client"}
cd "$CLIENT_PATH"
grep -r "ActTypeSkinRaffle" --include="*.cs" .
```
```

### 示例：excel-parser 技能中的使用

```markdown
## 获取配表目录

**前置条件**：确保路径已解析

```json
{
  "filePath": "${PROJECT_CONFIG_PATH}/excel",
  "sheetName": "Hero"
}
```

如果路径未设置，使用默认值并提示用户：
> 使用默认配表路径 `D:/work/config/excel`（可在会话中通过 project-path-resolver 修改）
```

## 注意事项

1. **会话级别**：路径缓存仅在当前会话有效，新会话需要重新确认
2. **用户确认**：自动发现的路径必须经用户确认，避免误操作
3. **相对路径优先**：在脚本中使用相对路径作为备选，提高可移植性
4. **验证存在性**：使用路径前验证目录是否存在
5. **友好提示**：路径未设置时给出清晰的引导，不要让用户困惑
