@echo off
rem =====================================================
rem 测试结果：失败
rem 帮助信息未能正确显示或匹配
rem 注意：文件删除失败，可能是文件被占用
rem =====================================================
echo 测试全局选项 -h/--help 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go

echo 测试短参数 -h
start /b cmd /c "go run %EXPORTER_PATH% -h > help_short.txt"
type help_short.txt
findstr /i "usage" help_short.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 短参数 -h 正确显示帮助信息
) else (
    echo [失败] 短参数 -h 未正确显示帮助信息
)

echo.
echo 测试长参数 --help
start /b cmd /c "go run %EXPORTER_PATH% --help > help_long.txt"
type help_long.txt
findstr /i "usage" help_long.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 长参数 --help 正确显示帮助信息
) else (
    echo [失败] 长参数 --help 未正确显示帮助信息
)

del help_short.txt help_long.txt

echo.
echo 测试完成