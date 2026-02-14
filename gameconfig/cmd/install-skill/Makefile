.PHONY: generate build clean check install test

generate:
	go generate

build: generate
	go build -o ../../bin/install-skill

install: build
	go install

clean:
	rm -rf skills/
	rm -f ../../bin/install-skill

check:
	@echo "🔍 检查 skill 文件同步..."
	@bash ../../scripts/check-sync.sh

test: check
	@echo "🧪 运行安装测试..."
	@cd ../.. && go test -v ./tests/... -run TestInstallSkill
