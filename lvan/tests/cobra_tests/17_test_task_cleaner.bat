@echo off
echo 测试任务清理功能
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8897
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\tasks mkdir %TEST_WORKDIR%\tasks

echo 创建测试任务文件（模拟过期任务）...
for /l %%i in (1, 1, 10) do (
    mkdir %TEST_WORKDIR%\tasks\old_task_%%i
    echo test > %TEST_WORKDIR%\tasks\old_task_%%i\output.txt
    
    REM 修改文件时间为30天前
    powershell -command "(Get-Item '%TEST_WORKDIR%\tasks\old_task_%%i').LastWriteTime = (Get-Date).AddDays(-30)"
)

echo 启动服务（启用自动清理）...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 10 /nobreak > nul

echo 检查过期任务是否被清理...
dir /b %TEST_WORKDIR%\tasks > task_files.txt
findstr /i "old_task" task_files.txt > nul
if %errorlevel% equ 0 (
    echo [失败] 过期任务未被清理
) else (
    echo [成功] 过期任务已被清理
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del task_files.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成