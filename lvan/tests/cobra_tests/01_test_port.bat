@echo off
rem =====================================================
rem 测试结果：部分成功
rem 短参数 -p：成功 - 端口被正确监听
rem 长参数 --port：失败 - 端口未被监听
rem 注意：在Git Bash中timeout命令语法不兼容
rem =====================================================
echo 测试全局选项 -p/--port 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8888

echo 测试短参数 -p
start /b go run %EXPORTER_PATH% -p %TEST_PORT%
timeout /t 2 /nobreak > nul

echo 检查端口 %TEST_PORT% 是否被监听...
netstat -ano | findstr ":%TEST_PORT% " > nul
if %errorlevel% equ 0 (
    echo [成功] 端口 %TEST_PORT% 已被监听
) else (
    echo [失败] 端口 %TEST_PORT% 未被监听
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)
timeout /t 2 /nobreak > nul

echo.
echo 测试长参数 --port
start /b go run %EXPORTER_PATH% --port %TEST_PORT%
timeout /t 2 /nobreak > nul

echo 检查端口 %TEST_PORT% 是否被监听...
netstat -ano | findstr ":%TEST_PORT% " > nul
if %errorlevel% equ 0 (
    echo [成功] 端口 %TEST_PORT% 已被监听
) else (
    echo [失败] 端口 %TEST_PORT% 未被监听
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

echo.
echo 测试完成