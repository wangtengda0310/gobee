@echo off
rem =====================================================
rem 测试结果：失败
rem exec子命令及其别名run未正确执行
rem --encoding参数测试也未成功
rem 注意：timeout命令语法在Git Bash中不兼容
rem 注意：文件删除失败，可能是文件被占用
rem =====================================================
echo 测试子命令 exec/run 及其 --encoding 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_CMD=echo Hello World

echo 测试 exec 子命令
start /b cmd /c "go run %EXPORTER_PATH% exec %TEST_CMD% > exec_output.txt"
timeout /t 2 /nobreak > nul
type exec_output.txt
findstr /i "Hello World" exec_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] exec 子命令正确执行
) else (
    echo [失败] exec 子命令未正确执行
)

echo.
echo 测试 run 子命令（别名）
start /b cmd /c "go run %EXPORTER_PATH% run %TEST_CMD% > run_output.txt"
timeout /t 2 /nobreak > nul
type run_output.txt
findstr /i "Hello World" run_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] run 子命令别名正确执行
) else (
    echo [失败] run 子命令别名未正确执行
)

echo.
echo 测试 exec 子命令的 --encoding 参数
echo {"message":"Hello World"} > test_json.json
start /b cmd /c "go run %EXPORTER_PATH% exec \"type test_json.json\" --encoding=json > exec_encoding_output.txt"
timeout /t 2 /nobreak > nul
type exec_encoding_output.txt
findstr /i "message" exec_encoding_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] exec 子命令的 --encoding 参数正确执行
) else (
    echo [失败] exec 子命令的 --encoding 参数未正确执行
)

del exec_output.txt run_output.txt exec_encoding_output.txt test_json.json

echo.
echo 测试完成