.PHONY: generate clean test build

generate:
	@echo "同步 skill 文件..."
	@cp -r ../../.claude/skills/gameactor ./skills
	@echo "✅ Skill 文件已同步"

clean:
	@echo "清理生成的文件..."
	@rm -rf skills
	@echo "✅ 清理完成"

test:
	@go test -v ./...

build:
	@echo "编译安装工具..."
	@go build -o ../../gameactor-install-skill.exe .
	@echo "✅ 编译完成: ../../gameactor-install-skill.exe"

install: build
	@echo "安装 skill..."
	@../../gameactor-install-skill.exe

help:
	@echo "可用命令:"
	@echo "  make generate   - 同步 skill 文件到嵌入目录"
	@echo "  make clean      - 清理生成的文件"
	@echo "  make test       - 运行测试"
	@echo "  make build      - 编译安装工具"
	@echo "  make install    - 编译并安装 skill"
