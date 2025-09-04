@echo off
echo 运行所有Cobra重构测试
echo.

set TEST_DIR=%~dp0
cd %TEST_DIR%

echo 确保Git Bash可用于运行Shell脚本...
where bash > nul 2>&1
if %errorlevel% neq 0 (
    echo [警告] 未找到bash命令，某些测试可能无法运行
    echo 请确保Git Bash已安装并添加到PATH环境变量中
)

echo.
echo ===== 开始测试全局选项 =====
echo.

call 01_test_port.bat
echo.

call 02_test_version.bat
echo.

call 03_test_help.bat
echo.

call 04_test_morehelp.bat
echo.

call 05_test_log_level.bat
echo.

call 06_test_clean.bat
echo.

call 07_test_workdir.bat
echo.

call 08_test_env_vars.bat
echo.

echo ===== 开始测试子命令 =====
echo.

call 09_test_cmd_command.bat
echo.

call 10_test_exec_run.bat
echo.

call 11_test_default_http.bat
echo.

echo ===== 开始测试HTTP API =====
echo.

call 12_test_http_cmd.bat
echo.

call 13_test_http_result.bat
echo.

call 14_test_http_cancel.bat
echo.

call 15_test_http_backup.bat
echo.

echo ===== 开始测试功能性 =====
echo.

call 16_test_cron.bat
echo.

call 17_test_task_cleaner.bat
echo.

call 18_test_multi_version.bat
echo.

echo ===== 开始测试Cobra特性 =====
echo.

call 19_test_cobra_structure.bat
echo.

call 20_test_viper_config.bat
echo.

echo 运行Shell脚本测试...
bash 21_test_completion.sh
echo.

echo ===== 开始测试兼容性 =====
echo.

call 22_test_compatibility.bat
echo.

call 23_test_config_compatibility.bat
echo.

call 24_test_performance.bat
echo.

echo.
echo ===== 所有测试完成 =====