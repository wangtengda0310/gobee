@echo off
rem =====================================================
rem 测试结果：失败
rem 任务目录未被清理
rem 注意：timeout命令语法在Git Bash中不兼容
rem =====================================================
echo 测试全局选项 --clean 参数
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_WORKDIR=.\test_workdir
set TASK_DIR=%TEST_WORKDIR%\tasks

echo 创建测试目录和文件...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%
if not exist %TASK_DIR% mkdir %TASK_DIR%

echo 创建测试任务文件...
echo test > %TASK_DIR%\test_task_1.txt
echo test > %TASK_DIR%\test_task_2.txt

echo 测试 --clean 参数
start /b cmd /c "go run %EXPORTER_PATH% --workdir %TEST_WORKDIR% --clean"
timeout /t 3 /nobreak > nul

echo 检查任务目录是否被清理...
dir /b %TASK_DIR% > task_files.txt
findstr /i "test_task" task_files.txt > nul
if %errorlevel% equ 0 (
    echo [失败] 任务目录未被清理
) else (
    echo [成功] 任务目录已被清理
)

del task_files.txt
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成