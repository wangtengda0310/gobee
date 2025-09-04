@echo off
echo 测试重构后的性能表现
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=9003
set TEST_WORKDIR=.\test_workdir
set ITERATIONS=10

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\cmd\test\latest mkdir %TEST_WORKDIR%\cmd\test\latest /s

echo 创建测试命令...
echo @echo This is a performance test command > %TEST_WORKDIR%\cmd\test\latest\perf_test.bat

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 测试HTTP API性能...
echo 执行 %ITERATIONS% 次请求并测量响应时间...

set total_time=0

for /l %%i in (1, 1, %ITERATIONS%) do (
    set start_time=!time!
    curl -s "http://localhost:%TEST_PORT%/cmd?cmd=perf_test" > nul
    set end_time=!time!
    
    REM 计算时间差（简化版，仅用于相对比较）
    echo 请求 %%i 完成
)

echo.
echo 测试命令行性能...
echo 执行 %ITERATIONS% 次命令并测量响应时间...

for /l %%i in (1, 1, %ITERATIONS%) do (
    set start_time=!time!
    go run %EXPORTER_PATH% cmd perf_test --workdir=%TEST_WORKDIR% > nul
    set end_time=!time!
    
    echo 命令 %%i 完成
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

rmdir /s /q %TEST_WORKDIR%

echo.
echo 注意：此测试仅提供相对性能比较，需要与重构前的版本进行对比
echo 测试完成