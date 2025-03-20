@echo off
setlocal enabledelayedexpansion

for /l %%i in (1,1,10) do (
    set "current_time=!time:~0,8!"
    echo %exporter_cmd_now1_resource% 现在：!current_time!
)

echo 已完成20次时间显示