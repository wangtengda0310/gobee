@echo off
rem =====================================================
rem 测试结果：部分成功
rem EXPORTER_PORT：未生效，端口未被监听
rem EXPORTER_WORKDIR：生效，创建了logs子目录
rem EXPORTER_CLEAN：未生效，任务目录未被清理
rem 注意：timeout命令语法在Git Bash中不兼容
rem =====================================================
echo 测试环境变量支持
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8891
set TEST_WORKDIR=.\test_workdir

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%

echo 测试环境变量 EXPORTER_PORT
set EXPORTER_PORT=%TEST_PORT%
start /b cmd /c "go run %EXPORTER_PATH%"
timeout /t 3 /nobreak > nul

echo 检查端口 %TEST_PORT% 是否被监听...
netstat -ano | findstr ":%TEST_PORT% " > nul
if %errorlevel% equ 0 (
    echo [成功] 环境变量 EXPORTER_PORT 生效，端口 %TEST_PORT% 已被监听
) else (
    echo [失败] 环境变量 EXPORTER_PORT 未生效，端口 %TEST_PORT% 未被监听
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)
timeout /t 2 /nobreak > nul

echo.
echo 测试环境变量 EXPORTER_WORKDIR
set EXPORTER_PORT=
set EXPORTER_WORKDIR=%TEST_WORKDIR%
rmdir /s /q %TEST_WORKDIR%
mkdir %TEST_WORKDIR%

start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT%"
timeout /t 3 /nobreak > nul

echo 检查是否在指定工作目录创建了必要的子目录...
if exist %TEST_WORKDIR%\logs (
    echo [成功] 环境变量 EXPORTER_WORKDIR 生效，创建了logs子目录
) else (
    echo [失败] 环境变量 EXPORTER_WORKDIR 未生效，未创建logs子目录
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

echo.
echo 测试环境变量 EXPORTER_CLEAN
set EXPORTER_WORKDIR=
set EXPORTER_CLEAN=true

echo 创建测试任务文件...
if not exist %TEST_WORKDIR%\tasks mkdir %TEST_WORKDIR%\tasks
echo test > %TEST_WORKDIR%\tasks\test_task_1.txt
echo test > %TEST_WORKDIR%\tasks\test_task_2.txt

start /b cmd /c "go run %EXPORTER_PATH% --workdir %TEST_WORKDIR%"
timeout /t 3 /nobreak > nul

echo 检查任务目录是否被清理...
dir /b %TEST_WORKDIR%\tasks > task_files.txt
findstr /i "test_task" task_files.txt > nul
if %errorlevel% equ 0 (
    echo [失败] 环境变量 EXPORTER_CLEAN 未生效，任务目录未被清理
) else (
    echo [成功] 环境变量 EXPORTER_CLEAN 生效，任务目录已被清理
)

del task_files.txt
rmdir /s /q %TEST_WORKDIR%
set EXPORTER_CLEAN=

echo.
echo 测试完成