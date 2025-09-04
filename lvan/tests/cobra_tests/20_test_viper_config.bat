@echo off
echo 验证Viper配置集成
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8899
set TEST_WORKDIR=.\test_workdir
set CONFIG_FILE=.\exporter_config.yaml

echo 创建测试配置文件...
echo port: %TEST_PORT% > %CONFIG_FILE%
echo workdir: %TEST_WORKDIR% >> %CONFIG_FILE%
echo log-level: debug >> %CONFIG_FILE%

echo 创建测试工作目录...
if not exist %TEST_WORKDIR% mkdir %TEST_WORKDIR%

echo 测试使用配置文件启动...
start /b cmd /c "go run %EXPORTER_PATH% --config %CONFIG_FILE%"
timeout /t 3 /nobreak > nul

echo 检查端口 %TEST_PORT% 是否被监听...
netstat -ano | findstr ":%TEST_PORT% " > nul
if %errorlevel% equ 0 (
    echo [成功] 配置文件中的端口设置生效
) else (
    echo [失败] 配置文件中的端口设置未生效
)

echo 检查是否在指定工作目录创建了必要的子目录...
if exist %TEST_WORKDIR%\logs (
    echo [成功] 配置文件中的工作目录设置生效
) else (
    echo [失败] 配置文件中的工作目录设置未生效
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

echo.
echo 测试环境变量优先级...
set EXPORTER_PORT=9000
start /b cmd /c "go run %EXPORTER_PATH% --config %CONFIG_FILE%"
timeout /t 3 /nobreak > nul

echo 检查端口 9000 是否被监听（环境变量应优先）...
netstat -ano | findstr ":9000 " > nul
if %errorlevel% equ 0 (
    echo [成功] 环境变量优先级高于配置文件
) else (
    echo [失败] 环境变量优先级不高于配置文件
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":9000 "') do (
    taskkill /f /pid %%a > nul 2>&1
)

echo.
echo 测试命令行参数优先级...
set EXPORTER_PORT=9000
start /b cmd /c "go run %EXPORTER_PATH% --config %CONFIG_FILE% --port 9001"
timeout /t 3 /nobreak > nul

echo 检查端口 9001 是否被监听（命令行参数应优先）...
netstat -ano | findstr ":9001 " > nul
if %errorlevel% equ 0 (
    echo [成功] 命令行参数优先级高于环境变量和配置文件
) else (
    echo [失败] 命令行参数优先级不高于环境变量和配置文件
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":9001 "') do (
    taskkill /f /pid %%a > nul 2>&1
)

set EXPORTER_PORT=
del %CONFIG_FILE%
rmdir /s /q %TEST_WORKDIR%

echo.
echo 测试完成