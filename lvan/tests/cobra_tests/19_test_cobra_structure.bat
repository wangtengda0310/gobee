@echo off
echo 验证Cobra命令结构设计
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go

echo 检查根命令帮助信息...
go run %EXPORTER_PATH% --help > root_help.txt
type root_help.txt

echo 检查是否包含所有子命令...
findstr /i "serve" root_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 根命令包含 serve 子命令
) else (
    echo [失败] 根命令不包含 serve 子命令
)

findstr /i "cmd" root_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 根命令包含 cmd 子命令
) else (
    echo [失败] 根命令不包含 cmd 子命令
)

findstr /i "exec" root_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 根命令包含 exec 子命令
) else (
    echo [失败] 根命令不包含 exec 子命令
)

findstr /i "version" root_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 根命令包含 version 子命令
) else (
    echo [失败] 根命令不包含 version 子命令
)

findstr /i "clean" root_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 根命令包含 clean 子命令
) else (
    echo [失败] 根命令不包含 clean 子命令
)

echo.
echo 检查子命令帮助信息...
go run %EXPORTER_PATH% serve --help > serve_help.txt
type serve_help.txt
findstr /i "port" serve_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] serve 子命令包含 port 参数
) else (
    echo [失败] serve 子命令不包含 port 参数
)

go run %EXPORTER_PATH% cmd --help > cmd_help.txt
type cmd_help.txt
findstr /i "version" cmd_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] cmd 子命令包含 version 参数
) else (
    echo [失败] cmd 子命令不包含 version 参数
)

go run %EXPORTER_PATH% exec --help > exec_help.txt
type exec_help.txt
findstr /i "encoding" exec_help.txt > nul
if %errorlevel% equ 0 (
    echo [成功] exec 子命令包含 encoding 参数
) else (
    echo [失败] exec 子命令不包含 encoding 参数
)

del root_help.txt serve_help.txt cmd_help.txt exec_help.txt

echo.
echo 测试完成