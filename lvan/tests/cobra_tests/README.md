# Exporter Cobra重构测试脚本

本目录包含用于测试Exporter程序Cobra重构后功能的测试脚本集合。这些测试脚本旨在确保重构后的程序保持与原有功能的兼容性，同时验证新增的Cobra特性是否正常工作。

## 测试脚本列表

### 全局选项测试

- `01_test_port.bat` - 测试全局选项 -p/--port 参数
- `02_test_version.bat` - 测试全局选项 -v/--version 参数
- `03_test_help.bat` - 测试全局选项 -h/--help 参数
- `04_test_morehelp.bat` - 测试全局选项 --morehelp 参数
- `05_test_log_level.bat` - 测试全局选项 --log-level 参数
- `06_test_clean.bat` - 测试全局选项 --clean 参数
- `07_test_workdir.bat` - 测试全局选项 -w/--workdir 参数
- `08_test_env_vars.bat` - 测试环境变量支持

### 子命令测试

- `09_test_cmd_command.bat` - 测试子命令 cmd/command 及其 -v/--version 参数
- `10_test_exec_run.bat` - 测试子命令 exec/run 及其 --encoding 参数
- `11_test_default_http.bat` - 测试无参数时默认启动HTTP服务器功能

### HTTP API测试

- `12_test_http_cmd.bat` - 测试HTTP API /cmd 接口
- `13_test_http_result.bat` - 测试HTTP API /result/{id} 接口
- `14_test_http_cancel.bat` - 测试HTTP API /cancel/ 接口
- `15_test_http_backup.bat` - 测试HTTP API /backup/ 接口

### 功能性测试

- `16_test_cron.bat` - 测试定时任务功能
- `17_test_task_cleaner.bat` - 测试任务清理功能
- `18_test_multi_version.bat` - 测试多版本工具调用功能

### Cobra特性测试

- `19_test_cobra_structure.bat` - 验证Cobra命令结构设计
- `20_test_viper_config.bat` - 验证Viper配置集成
- `21_test_completion.sh` - 测试命令自动补全功能 (需要Git Bash)

### 兼容性测试

- `22_test_compatibility.bat` - 测试与现有脚本和工具的兼容性
- `23_test_config_compatibility.bat` - 测试与现有配置文件的兼容性
- `24_test_performance.bat` - 测试重构后的性能表现

### 运行所有测试

- `run_all_tests.bat` - 运行所有测试脚本

## 使用方法

### 运行单个测试

在命令行中直接运行对应的批处理文件即可：

```batch
01_test_port.bat
```

### 运行所有测试

在命令行中运行：

```batch
run_all_tests.bat
```

## 注意事项

1. 测试脚本假设Exporter程序的源代码位于 `../../cmd/exporter/main.go`，如果路径有变化，需要修改测试脚本中的路径。

2. 部分测试需要使用特定端口（8080-9003），请确保这些端口在测试时未被占用。

3. 测试脚本会创建临时目录和文件，测试完成后会自动清理。

4. 测试命令自动补全功能需要Git Bash环境，请确保系统已安装Git Bash并添加到PATH环境变量中。

5. 测试过程中可能会出现端口占用的情况，如果测试中断，可能需要手动关闭占用端口的进程。

6. 测试脚本中使用了一些Windows命令行工具，如curl、findstr等，请确保系统中有这些工具。

## 测试结果解读

每个测试脚本会输出测试过程和结果，成功的测试会显示 `[成功]`，失败的测试会显示 `[失败]`。

在重构过程中，可以根据测试结果来调整代码，确保所有功能正常工作。