@echo off
rem =====================================================
rem 测试结果：部分成功
rem 短参数-w：创建了logs子目录但未创建tasks子目录
rem 长参数--workdir：未创建任何子目录
rem 注意：timeout命令语法在Git Bash中不兼容
rem =====================================================
echo 测试全局选项 -w/--workdir 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_WORKDIR=.\test_workdir
set TEST_PORT=8890

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%

echo 测试短参数 -w
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% -w %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 检查是否在指定工作目录创建了必要的子目录...
if exist %TEST_WORKDIR%\logs (
    echo [成功] 在指定工作目录创建了logs子目录
) else (
    echo [失败] 未在指定工作目录创建logs子目录
)

if exist %TEST_WORKDIR%\tasks (
    echo [成功] 在指定工作目录创建了tasks子目录
) else (
    echo [失败] 未在指定工作目录创建tasks子目录
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)
timeout /t 2 /nobreak > nul

echo.
echo 测试长参数 --workdir
rmdir /s /q %TEST_WORKDIR%
mkdir %TEST_WORKDIR%

start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT% --workdir %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 检查是否在指定工作目录创建了必要的子目录...
if exist %TEST_WORKDIR%\logs (
    echo [成功] 在指定工作目录创建了logs子目录
) else (
    echo [失败] 未在指定工作目录创建logs子目录
)

if exist %TEST_WORKDIR%\tasks (
    echo [成功] 在指定工作目录创建了tasks子目录
) else (
    echo [失败] 未在指定工作目录创建tasks子目录
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成