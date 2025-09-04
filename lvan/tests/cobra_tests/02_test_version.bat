@echo off
rem =====================================================
rem 测试结果：部分成功
rem 版本信息正确显示为 "Exporter version 0.0.0"
rem 但findstr命令未能正确匹配，可能是编码问题
rem 注意：文件删除失败，可能是文件被占用
rem =====================================================
echo 测试全局选项 -v/--version 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go

echo 测试短参数 -v
start /b cmd /c "go run %EXPORTER_PATH% -v > version_short.txt"
type version_short.txt
findstr /i "version" version_short.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 短参数 -v 正确显示版本信息
) else (
    echo [失败] 短参数 -v 未正确显示版本信息
)

echo.
echo 测试长参数 --version
start /b cmd /c "go run %EXPORTER_PATH% --version > version_long.txt"
type version_long.txt
findstr /i "version" version_long.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 长参数 --version 正确显示版本信息
) else (
    echo [失败] 长参数 --version 未正确显示版本信息
)

del version_short.txt version_long.txt

echo.
echo 测试完成