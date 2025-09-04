@echo off
rem =====================================================
rem 测试结果：失败
rem cmd子命令及其别名command未正确执行
rem -v和--version参数测试也未成功
rem 注意：timeout命令语法在Git Bash中不兼容
rem 注意：文件删除失败，可能是文件被占用
rem =====================================================
echo 测试子命令 cmd/command 及其 -v/--version 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_CMD=version

echo 测试 cmd 子命令
start /b cmd /c "go run %EXPORTER_PATH% cmd %TEST_CMD% > cmd_output.txt"
timeout /t 2 /nobreak > nul
type cmd_output.txt
findstr /i "version" cmd_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] cmd 子命令正确执行
) else (
    echo [失败] cmd 子命令未正确执行
)

echo.
echo 测试 command 子命令（别名）
start /b cmd /c "go run %EXPORTER_PATH% command %TEST_CMD% > command_output.txt"
timeout /t 2 /nobreak > nul
type command_output.txt
findstr /i "version" command_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] command 子命令别名正确执行
) else (
    echo [失败] command 子命令别名未正确执行
)

echo.
echo 测试 cmd 子命令的 -v 参数
start /b cmd /c "go run %EXPORTER_PATH% cmd -v %TEST_CMD% > cmd_v_output.txt"
timeout /t 2 /nobreak > nul
type cmd_v_output.txt
findstr /i "version" cmd_v_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] cmd 子命令的 -v 参数正确执行
) else (
    echo [失败] cmd 子命令的 -v 参数未正确执行
)

echo.
echo 测试 cmd 子命令的 --version 参数
start /b cmd /c "go run %EXPORTER_PATH% cmd --version %TEST_CMD% > cmd_version_output.txt"
timeout /t 2 /nobreak > nul
type cmd_version_output.txt
findstr /i "version" cmd_version_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] cmd 子命令的 --version 参数正确执行
) else (
    echo [失败] cmd 子命令的 --version 参数未正确执行
)

del cmd_output.txt command_output.txt cmd_v_output.txt cmd_version_output.txt

echo.
echo 测试完成