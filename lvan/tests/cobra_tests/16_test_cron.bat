@echo off
echo 测试定时任务功能
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8896
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\tasks mkdir %TEST_WORKDIR%\tasks
if not exist %TEST_WORKDIR%\tasks\cron mkdir %TEST_WORKDIR%\tasks\cron

echo 创建测试定时任务...
echo @echo Hello from cron job > %TEST_WORKDIR%\tasks\cron\test_cron.bat
echo * * * * * > %TEST_WORKDIR%\tasks\cron\test_cron.bat.cron

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 等待定时任务执行（最多等待70秒）...
set /a max_wait=70
set /a waited=0

:wait_loop
if %waited% geq %max_wait% goto :wait_done

dir /b %TEST_WORKDIR%\tasks > task_files.txt
findstr /i "test_cron" task_files.txt > nul
if %errorlevel% equ 0 (
    echo 定时任务已执行，创建了任务文件
    goto :wait_done
)

timeout /t 5 /nobreak > nul
set /a waited+=5
echo 已等待 %waited% 秒...
goto :wait_loop

:wait_done
if %waited% geq %max_wait% (
    echo [失败] 定时任务未在预期时间内执行
) else (
    echo [成功] 定时任务已执行
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del task_files.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成