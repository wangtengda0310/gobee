@echo off
rem =====================================================
rem 测试结果：部分成功
rem 更多帮助信息已显示，但findstr命令未能正确匹配
rem 注意：文件删除失败，可能是文件被占用
rem =====================================================
echo 测试全局选项 --morehelp 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go

echo 测试 --morehelp 参数
start /b cmd /c "go run %EXPORTER_PATH% --morehelp > morehelp.txt"
type morehelp.txt
findstr /i "详细" morehelp.txt > nul
if %errorlevel% equ 0 (
    echo [成功] --morehelp 参数正确显示更多帮助信息
) else (
    echo [失败] --morehelp 参数未正确显示更多帮助信息
)

del morehelp.txt

echo.
echo 测试完成