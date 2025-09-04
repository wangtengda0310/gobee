@echo off
rem =====================================================
rem 测试结果：失败
rem 日志文件不存在或无法访问
rem 注意：timeout命令语法在Git Bash中不兼容
rem =====================================================
echo 测试全局选项 --log-level 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8889
set LOG_FILE=..\logs\exporter.log

echo 测试 --log-level=debug 参数
if exist %LOG_FILE% del %LOG_FILE%
start /b cmd /c "go run %EXPORTER_PATH% --port %TEST_PORT% --log-level=debug"
timeout /t 3 /nobreak > nul

echo 检查日志文件是否包含 DEBUG 级别日志...
findstr /i "debug" %LOG_FILE% > nul
if %errorlevel% equ 0 (
    echo [成功] 日志文件包含 DEBUG 级别日志
) else (
    echo [失败] 日志文件不包含 DEBUG 级别日志
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)
timeout /t 2 /nobreak > nul

echo.

echo.
echo 测试 --log-level=info 参数
if exist %LOG_FILE% del %LOG_FILE%
start /b cmd /c "go run %EXPORTER_PATH% --port %TEST_PORT% --log-level=info"
timeout /t 3 /nobreak > nul

echo 检查日志文件是否包含 INFO 级别日志...
findstr /i "info" %LOG_FILE% > nul
if %errorlevel% equ 0 (
    echo [成功] 日志文件包含 INFO 级别日志
) else (
    echo [失败] 日志文件不包含 INFO 级别日志
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

echo.
echo 测试完成