# 开发环境配置指南

## 概述
本文档描述了LVAN Dumper项目的开发环境要求和配置步骤，确保开发者能够快速搭建一致的开发环境。

## 系统要求

### 基础环境
- **操作系统**: Windows 10/11, Linux, macOS
- **Go版本**: 1.21 或更高版本
- **Git**: 用于版本控制

### 推荐开发工具
- **IDE**: VS Code, GoLand, 或其他支持Go的IDE
- **终端**: Windows Terminal, iTerm2, 或其他现代终端
- **数据库工具**: MySQL Workbench, Redis Desktop Manager

## Go环境配置

### 安装Go

#### Windows
```bash
# 1. 下载Go安装包
# 访问 https://golang.org/dl/
# 下载适合Windows的安装包

# 2. 运行安装程序
# 按照安装向导完成安装

# 3. 验证安装
go version
# 应该显示: go version go1.21.x windows/amd64

# 4. 配置环境变量
# 确保GOPATH和PATH配置正确
echo $GOPATH
echo $PATH
```

#### Linux/macOS
```bash
# 1. 下载并安装Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 2. 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 3. 验证安装
go version
```

### Go模块配置

#### 代理配置 (可选，中国用户推荐)
```bash
# 设置Go模块代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
go env -w GONOPROXY=github.com/wangtengda0310

# 验证配置
go env GOPROXY
go env GOSUMDB
```

#### 私有仓库配置
```bash
# 配置私有仓库访问
go env -w GOPRIVATE=github.com/wangtengda0310
go env -w GONOPROXY=github.com/wangtengda0310
```

## 项目初始化

### 克隆项目
```bash
# 克隆项目仓库
git clone https://github.com/wangtengda0310/lvan_dumper.git
cd lvan_dumper/lvan

# 验证项目结构
ls -la
# 应该看到: cmd/, pkg/, docs/, go.mod 等文件和目录
```

### 依赖管理
```bash
# 下载依赖
go mod tidy

# 验证依赖
go mod verify

# 更新依赖 (如需要)
go mod download
```

### 验证编译
```bash
# 编译项目
go build ./cmd/dumper

# 运行可执行文件
./dumper --help

# 清理编译产物
go clean
```

## IDE配置

### VS Code配置

#### 安装扩展
```json
// 推荐扩展列表
{
    "recommendations": [
        "golang.go",
        "ms-vscode.test-adapter-converter",
        "redhat.vscode-yaml",
        "ms-vscode.vscode-json"
    ]
}
```

#### 工作区设置
```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.gopath": "",
    "go.goroot": "",
    "go.toolsManagement.checkForUpdates": "local",
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testOnSave": true,
    "go.coverOnSave": true,
    "go.testFlags": ["-v"],
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

#### 调试配置
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}"
        },
        {
            "name": "Launch Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}"
        }
    ]
}
```

### GoLand配置

#### 项目设置
1. 打开项目目录
2. Go → GOPATH → 设置为项目目录
3. Go → GOROOT → 选择Go安装目录
4. Go → Go Modules → 启用Go模块集成

#### 代码风格
1. File → Settings → Editor → Code Style → Go
2. 导入配置使用 `goimports`
3. 格式化配置使用 `gofmt`

## 数据库环境配置

### MySQL配置

#### 安装MySQL
```bash
# Windows (使用Chocolatey)
choco install mysql

# Linux (Ubuntu)
sudo apt update
sudo apt install mysql-server

# macOS (使用Homebrew)
brew install mysql
```

#### 创建测试数据库
```sql
-- 连接到MySQL
mysql -u root -p

-- 创建测试数据库
CREATE DATABASE lvan_test;
CREATE DATABASE lvan_dev;

-- 创建测试用户
CREATE USER 'lvan_test'@'localhost' IDENTIFIED BY 'test_password';
GRANT ALL PRIVILEGES ON lvan_test.* TO 'lvan_test'@'localhost';
FLUSH PRIVILEGES;
```

#### 配置连接
```yaml
# config/mysql_test.yaml
mysql:
  server:
    host: localhost
    port: 3306
    user: lvan_test
    password: test_password
    database: lvan_test
```

### Redis配置

#### 安装Redis
```bash
# Windows (使用WSL或Docker)
# Windows原生Redis不推荐

# Linux
sudo apt install redis-server

# macOS
brew install redis
```

#### 启动Redis
```bash
# Linux/macOS
redis-server

# 验证连接
redis-cli ping
# 应该返回: PONG
```

## 测试环境配置

### 测试工具安装

#### 依赖包
```bash
# 安装测试相关依赖
go mod tidy

# 验证测试依赖
go test -help
```

#### 内存数据库配置
```go
// 测试配置示例
// pkg/testdb/configs/mysql_test.yaml
mysql:
  server:
    port: 3307  # 避免与生产数据库冲突
    database: testdb
    user: root
    password: ""
```

### 测试脚本
```bash
#!/bin/bash
# scripts/test.sh

echo "运行单元测试..."
go test ./...

echo "运行集成测试..."
go test -tags=integration ./...

echo "生成覆盖率报告..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "测试完成"
```

## 代码质量工具

### 安装代码检查工具

#### golangci-lint
```bash
# 安装
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 验证安装
golangci-lint version
```

#### 配置文件
```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - misspell
    - unconvert
    - dupl
    - goconst
    - gocyclo

linters-settings:
  gocyclo:
    min-complexity: 10
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 2
```

### 使用代码质量工具
```bash
# 运行代码检查
golangci-lint run

# 运行特定检查器
golangci-lint run --enable-only=errcheck

# 生成HTML报告
golangci-lint run --out-format=html > report.html
```

## 依赖管理

### 添加新依赖
```bash
# 添加新包
go get github.com/example/package

# 添加特定版本
go get github.com/example/package@v1.2.3

# 更新依赖
go get -u ./...
```

### 依赖安全检查
```bash
# 安装安全检查工具
go install github.com/sonatypecommunity/nancy@latest

# 检查依赖安全性
go list -json -m all | nancy sleuth
```

## 常用开发命令

### 项目管理
```bash
# 编译项目
go build ./cmd/dumper

# 运行项目
go run ./cmd/dumper --help

# 格式化代码
go fmt ./...

# 整理导入
goimports -w .
```

### 测试命令
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/dump

# 运行特定测试
go test -run TestSpecificFunction ./pkg/dump

# 运行基准测试
go test -bench=. ./pkg/dump

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 调试命令
```bash
# 使用Delve调试器
dlv debug ./cmd/dumper

# 测试调试
dlv test ./pkg/dump
```

## 故障排除

### 常见问题

#### Go模块问题
```bash
# 问题: go mod tidy 失败
# 解决: 清理模块缓存
go clean -modcache
go mod download
go mod tidy
```

#### 编译错误
```bash
# 问题: 找不到包
# 解决: 检查go.mod和go.sum
go mod verify
go mod tidy
```

#### 测试失败
```bash
# 问题: 测试数据库连接失败
# 解决: 检查测试配置和数据库状态
# 确保测试数据库正在运行
```

#### 依赖冲突
```bash
# 问题: 依赖版本冲突
# 解决: 查看依赖树
go mod why github.com/example/package

# 解决冲突
go get github.com/example/package@compatible-version
```

### 性能问题

#### 编译慢
```bash
# 禁用CGO编译器缓存
export CGO_ENABLED=0

# 使用缓存构建
go build -buildcache=dir
```

#### 测试慢
```bash
# 并行运行测试
go test -parallel 4 ./...

# 跳过长时间测试
go test -short ./...
```

## 开发最佳实践

### 工作流建议

1. **分支管理**
   - 使用功能分支进行开发
   - 定期合并到主分支
   - 保持主分支的稳定性

2. **提交规范**
   - 使用有意义的提交信息
   - 遵循Conventional Commits规范
   - 频繁提交，小步快跑

3. **代码审查**
   - 所有代码变更需要审查
   - 重点关注接口设计和错误处理
   - 确保测试覆盖率

### 环境一致性

#### Docker开发环境
```dockerfile
# Dockerfile.dev
FROM golang:1.21-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build ./cmd/dumper

CMD ["./dumper"]
```

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
    ports:
      - "8080:8080"

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: lvan_test
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

## 总结

通过遵循本指南，开发者可以快速搭建一致的开发环境，确保代码质量和开发效率。关键要点：

1. **环境一致性**: 使用Docker或配置文件确保环境一致性
2. **工具链完整**: 配置完整的开发和测试工具链
3. **自动化流程**: 建立自动化的测试和代码检查流程
4. **文档维护**: 及时更新配置文档和最佳实践

定期更新开发环境和工具链，保持与项目发展同步。

---

*文档版本: v1.0*
*最后更新: 2025-01-07*