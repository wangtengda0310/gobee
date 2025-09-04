@echo off
echo 测试多版本工具调用功能
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8898
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\cmd mkdir %TEST_WORKDIR%\cmd
if not exist %TEST_WORKDIR%\cmd\test mkdir %TEST_WORKDIR%\cmd\test
if not exist %TEST_WORKDIR%\cmd\test\v1.0 mkdir %TEST_WORKDIR%\cmd\test\v1.0
if not exist %TEST_WORKDIR%\cmd\test\v2.0 mkdir %TEST_WORKDIR%\cmd\test\v2.0
if not exist %TEST_WORKDIR%\cmd\test\latest mkdir %TEST_WORKDIR%\cmd\test\latest

echo 创建测试命令文件...
echo @echo This is test command v1.0 > %TEST_WORKDIR%\cmd\test\v1.0\test.bat
echo @echo This is test command v2.0 > %TEST_WORKDIR%\cmd\test\v2.0\test.bat
echo @echo This is latest test command > %TEST_WORKDIR%\cmd\test\latest\test.bat

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 测试指定版本调用...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=test&version=v1.0" > v1_result.txt
set /p V1_ID=<v1_result.txt
echo 获取到v1.0任务ID: %V1_ID%
timeout /t 2 /nobreak > nul

curl -s "http://localhost:%TEST_PORT%/result/%V1_ID%" > v1_output.txt
type v1_output.txt
findstr /i "v1.0" v1_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 正确调用了v1.0版本命令
) else (
    echo [失败] 未正确调用v1.0版本命令
)

echo.
echo 测试默认使用latest版本...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=test" > latest_result.txt
set /p LATEST_ID=<latest_result.txt
echo 获取到latest任务ID: %LATEST_ID%
timeout /t 2 /nobreak > nul

curl -s "http://localhost:%TEST_PORT%/result/%LATEST_ID%" > latest_output.txt
type latest_output.txt
findstr /i "latest" latest_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 默认正确调用了latest版本命令
) else (
    echo [失败] 默认未正确调用latest版本命令
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del v1_result.txt v1_output.txt latest_result.txt latest_output.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成