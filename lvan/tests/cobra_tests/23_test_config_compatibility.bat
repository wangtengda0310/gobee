@echo off
echo 测试与现有配置文件的兼容性
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=9002
set TEST_WORKDIR=.\test_workdir
set TEST_CMD_DIR=%TEST_WORKDIR%\cmd\test\latest

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\cmd\test\latest mkdir %TEST_WORKDIR%\cmd\test\latest /s

echo 创建测试命令...
echo @echo This is a test command > %TEST_CMD_DIR%\test.bat

echo 创建meta.yaml配置文件...
echo charset: utf-8 > %TEST_CMD_DIR%\meta.yaml
echo shell: cmd >> %TEST_CMD_DIR%\meta.yaml
echo resources: >> %TEST_CMD_DIR%\meta.yaml
echo   - resource1 >> %TEST_CMD_DIR%\meta.yaml
echo   - resource2 >> %TEST_CMD_DIR%\meta.yaml

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 测试meta.yaml配置文件兼容性...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=test" > cmd_output.txt
set /p CMD_ID=<cmd_output.txt
echo 获取到任务ID: %CMD_ID%
timeout /t 2 /nobreak > nul

curl -s "http://localhost:%TEST_PORT%/result/%CMD_ID%" > cmd_result.txt
type cmd_result.txt
findstr /i "test command" cmd_result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] meta.yaml配置文件兼容性测试通过
) else (
    echo [失败] meta.yaml配置文件兼容性测试失败
)

echo.
echo 测试cron配置文件兼容性...
if not exist %TEST_WORKDIR%\tasks\cron mkdir %TEST_WORKDIR%\tasks\cron

echo @echo This is a cron test > %TEST_WORKDIR%\tasks\cron\cron_test.bat
echo * * * * * > %TEST_WORKDIR%\tasks\cron\cron_test.bat.cron

echo 重启服务以加载cron配置...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)
timeout /t 1 /nobreak > nul

start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 检查cron配置是否被加载...
curl -s "http://localhost:%TEST_PORT%" > cron_check.txt
findstr /i "cron" cron_check.txt > nul
if %errorlevel% equ 0 (
    echo [成功] cron配置文件兼容性测试通过
) else (
    echo [失败] cron配置文件兼容性测试失败
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del cmd_output.txt cmd_result.txt cron_check.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成