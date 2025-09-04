@echo off
echo 测试HTTP API /cancel/ 接口
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8894

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT%"
timeout /t 3 /nobreak > nul

echo 创建一个长时间运行的测试任务...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=ping&args=localhost&onlyid=true" > task_id.txt
set /p TASK_ID=<task_id.txt
echo 获取到任务ID: %TASK_ID%
timeout /t 1 /nobreak > nul

echo 测试 /cancel/{id} 接口...
curl -s "http://localhost:%TEST_PORT%/cancel/%TASK_ID%" > cancel_result.txt
type cancel_result.txt
findstr /i "success" cancel_result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] /cancel/{id} 接口正确取消了任务
) else (
    echo [失败] /cancel/{id} 接口未正确取消任务
)

echo 验证任务是否被取消...
timeout /t 1 /nobreak > nul
curl -s "http://localhost:%TEST_PORT%/result/%TASK_ID%" > cancel_check.txt
type cancel_check.txt
findstr /i "canceled" cancel_check.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 任务状态显示为已取消
) else (
    echo [失败] 任务状态未显示为已取消
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del task_id.txt cancel_result.txt cancel_check.txt

echo.
echo 测试完成