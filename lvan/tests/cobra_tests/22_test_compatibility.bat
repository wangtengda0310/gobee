@echo off
echo 测试与现有脚本和工具的兼容性
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=9001
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TEST_WORKDIR%\cmd mkdir %TEST_WORKDIR%\cmd
if not exist %TEST_WORKDIR%\cmd\test mkdir %TEST_WORKDIR%\cmd\test
if not exist %TEST_WORKDIR%\cmd\test\latest mkdir %TEST_WORKDIR%\cmd\test\latest

echo 创建测试脚本...
echo @echo off > %TEST_WORKDIR%\cmd\test\latest\test_script.bat
echo echo This is a test script >> %TEST_WORKDIR%\cmd\test\latest\test_script.bat
echo echo Command arguments: %%* >> %TEST_WORKDIR%\cmd\test\latest\test_script.bat

echo 创建测试调用脚本...
echo @echo off > call_exporter.bat
echo go run %EXPORTER_PATH% cmd test_script arg1 arg2 --workdir=%TEST_WORKDIR% >> call_exporter.bat

echo 测试现有脚本调用方式...
call call_exporter.bat > script_output.txt
type script_output.txt
findstr /i "arg1 arg2" script_output.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 现有脚本调用方式兼容
) else (
    echo [失败] 现有脚本调用方式不兼容
)

echo.
echo 测试HTTP API兼容性...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 测试现有API调用方式...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=test_script&args=api_arg1,api_arg2" > api_output.txt
set /p API_ID=<api_output.txt
echo 获取到任务ID: %API_ID%
timeout /t 2 /nobreak > nul

curl -s "http://localhost:%TEST_PORT%/result/%API_ID%" > api_result.txt
type api_result.txt
findstr /i "api_arg1 api_arg2" api_result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] 现有API调用方式兼容
) else (
    echo [失败] 现有API调用方式不兼容
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del call_exporter.bat script_output.txt api_output.txt api_result.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成