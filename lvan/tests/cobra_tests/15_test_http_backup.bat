@echo off
echo 测试HTTP API /backup/ 接口
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8895
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\tasks mkdir %TEST_WORKDIR%\tasks

echo 创建测试任务文件...
echo test > %TEST_WORKDIR%\tasks\test_task_1.txt

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 测试 /backup 接口...
curl -s "http://localhost:%TEST_PORT%/backup" -o backup.zip

echo 检查备份文件是否创建...
if exist backup.zip (
    echo [成功] 备份文件已创建
    
    echo 检查备份文件内容...
    mkdir backup_extract
    powershell -command "Expand-Archive -Path backup.zip -DestinationPath backup_extract -Force"
    
    if exist backup_extract\tasks\test_task_1.txt (
        echo [成功] 备份文件包含正确的任务文件
    ) else (
        echo [失败] 备份文件不包含正确的任务文件
    )
    
    rmdir /s /q backup_extract
    del backup.zip
) else (
    echo [失败] 备份文件未创建
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成