# gameconfig 开发指南

## 修改 Skill 文件

### 唯一修改位置

**所有 skill 内容都在这里修改**：
```
.claude/skills/gameconfig/
├── SKILL.md              ← 主 skill 文件
├── abilities/
│   └── AI指导.md          ← AI 能力指导
└── README.md
```

### 同步到安装工具

修改 skill 后，**必须同步**到嵌入文件：

```bash
cd cmd/install-skill
make generate    # 或: go generate
```

### 验证同步

```bash
cd cmd/install-skill
make check       # 检查文件是否同步
make test        # 运行完整测试
```

---

## 发布前检查清单

### 1. 同步 Skill 文件

```bash
cd cmd/install-skill
go generate
git add skills/
```

### 2. 运行测试

```bash
# 回归测试
go test ./tests/... -run TestInstallSkill -v

# 完整测试
go test ./... -v
```

### 3. 提交检查

- [ ] `cmd/install-skill/skills/` 目录已更新
- [ ] 所有测试通过
- [ ] README.md 中的说明已更新（如有需要）

---

## 自动提醒机制

### Pre-commit Hook

安装 git hook，提交时自动检查：

```bash
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### CI/CD 检查

GitHub Actions 会在 PR 时自动检查：
- [Skill Sync Check](.github/workflows/skill-sync.yml)

### Makefile 快捷命令

```bash
cd cmd/install-skill
make check       # 检查同步
make test        # 运行测试
make install     # 同步 + 构建 + 安装
```

---

## 常见问题

### Q: 提示 "文件不同步" 怎么办？

```bash
cd cmd/install-skill
go generate
```

### Q: CI 失败了怎么办？

查看 CI 日志，如果是同步问题：
```bash
bash scripts/check-sync.sh
```

### Q: 如何测试安装工具？

```bash
# 测试编译
go build ./cmd/install-skill

# 测试安装（到临时目录）
go run cmd/install-skill/main.go -target /tmp/test-skill-install

# 查看安装结果
ls /tmp/test-skill-install/
```
