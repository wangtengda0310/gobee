@echo off
echo 测试无参数时默认启动HTTP服务器功能
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set DEFAULT_PORT=8080

echo 测试无参数启动
start /b cmd /c "go run %EXPORTER_PATH%"
timeout /t 3 /nobreak > nul

echo 检查默认端口 %DEFAULT_PORT% 是否被监听...
netstat -ano | findstr ":%DEFAULT_PORT% " > nul
if %errorlevel% equ 0 (
    echo [成功] 默认端口 %DEFAULT_PORT% 已被监听
) else (
    echo [失败] 默认端口 %DEFAULT_PORT% 未被监听
)

echo 测试HTTP服务是否正常响应...
curl -s http://localhost:%DEFAULT_PORT% > http_response.txt
type http_response.txt
findstr /i "exporter" http_response.txt > nul
if %errorlevel% equ 0 (
    echo [成功] HTTP服务正常响应
) else (
    echo [失败] HTTP服务未正常响应
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%DEFAULT_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del http_response.txt

echo.
echo 测试完成