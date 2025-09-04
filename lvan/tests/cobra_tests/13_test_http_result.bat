@echo off
echo 测试HTTP API /result/{id} 接口
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8893

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT%"
timeout /t 3 /nobreak > nul

echo 创建测试任务...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=version&onlyid=true" > task_id.txt
set /p TASK_ID=<task_id.txt
echo 获取到任务ID: %TASK_ID%
timeout /t 2 /nobreak > nul

echo 测试 /result/{id} 接口...
curl -s "http://localhost:%TEST_PORT%/result/%TASK_ID%" > result.txt
type result.txt
findstr /i "version" result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] /result/{id} 接口正确返回了任务结果
) else (
    echo [失败] /result/{id} 接口未正确返回任务结果
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del task_id.txt result.txt

echo.
echo 测试完成