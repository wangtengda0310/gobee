@echo off
echo 传入的环境变量 a %exporter_cmd_now_resource%
for /l %%i in (1,1,10) do @(echo %time:~0,8% & ping -n 2 127.0.0.1 >nul)