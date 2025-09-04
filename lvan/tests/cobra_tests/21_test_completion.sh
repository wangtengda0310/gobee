#!/bin/bash

echo "测试命令自动补全功能"
echo

EXPORTER_PATH="../../cmd/exporter/main.go"

echo "生成 Bash 补全脚本..."
go run $EXPORTER_PATH completion bash > exporter_completion.bash

echo "检查补全脚本是否生成..."
if [ -s exporter_completion.bash ]; then
    echo "[成功] Bash 补全脚本已生成"
    
    echo "检查补全脚本内容..."
    grep -q "bash completion" exporter_completion.bash
    if [ $? -eq 0 ]; then
        echo "[成功] Bash 补全脚本内容正确"
    else
        echo "[失败] Bash 补全脚本内容不正确"
    fi
    
    echo "检查补全脚本是否包含所有子命令..."
    grep -q "serve" exporter_completion.bash
    if [ $? -eq 0 ]; then
        echo "[成功] 补全脚本包含 serve 子命令"
    else
        echo "[失败] 补全脚本不包含 serve 子命令"
    fi
    
    grep -q "cmd" exporter_completion.bash
    if [ $? -eq 0 ]; then
        echo "[成功] 补全脚本包含 cmd 子命令"
    else
        echo "[失败] 补全脚本不包含 cmd 子命令"
    fi
    
    grep -q "exec" exporter_completion.bash
    if [ $? -eq 0 ]; then
        echo "[成功] 补全脚本包含 exec 子命令"
    else
        echo "[失败] 补全脚本不包含 exec 子命令"
    fi
else
    echo "[失败] Bash 补全脚本未生成"
fi

echo "生成 Zsh 补全脚本..."
go run $EXPORTER_PATH completion zsh > exporter_completion.zsh

echo "检查补全脚本是否生成..."
if [ -s exporter_completion.zsh ]; then
    echo "[成功] Zsh 补全脚本已生成"
else
    echo "[失败] Zsh 补全脚本未生成"
fi

echo "生成 PowerShell 补全脚本..."
go run $EXPORTER_PATH completion powershell > exporter_completion.ps1

echo "检查补全脚本是否生成..."
if [ -s exporter_completion.ps1 ]; then
    echo "[成功] PowerShell 补全脚本已生成"
else
    echo "[失败] PowerShell 补全脚本未生成"
fi

# 清理生成的文件
rm -f exporter_completion.bash exporter_completion.zsh exporter_completion.ps1

echo
echo "测试完成"