@echo off
echo 测试HTTP API /cmd 接口
echo.

set EXPORTER_PATH=..\..\cmd\exporter\main.go
set TEST_PORT=8892

echo 启动服务...
start /b cmd /c "go run %EXPORTER_PATH% -p %TEST_PORT%"
timeout /t 3 /nobreak > nul

echo 测试 GET 请求 /cmd 接口...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=version" > cmd_get.txt
type cmd_get.txt
findstr /i "id" cmd_get.txt > nul
if %errorlevel% equ 0 (
    echo [成功] GET 请求 /cmd 接口返回了任务ID
) else (
    echo [失败] GET 请求 /cmd 接口未返回任务ID
)

echo.
echo 测试 POST 请求 /cmd 接口（JSON格式）...
echo {"cmd":"version"} > cmd_post.json
curl -s -X POST -H "Content-Type: application/json" -d @cmd_post.json "http://localhost:%TEST_PORT%/cmd" > cmd_post_result.txt
type cmd_post_result.txt
findstr /i "id" cmd_post_result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] POST 请求 /cmd 接口（JSON格式）返回了任务ID
) else (
    echo [失败] POST 请求 /cmd 接口（JSON格式）未返回任务ID
)

echo.
echo 测试 POST 请求 /cmd 接口（YAML格式）...
echo cmd: version > cmd_post.yaml
curl -s -X POST -H "Content-Type: application/yaml" -d @cmd_post.yaml "http://localhost:%TEST_PORT%/cmd" > cmd_post_yaml_result.txt
type cmd_post_yaml_result.txt
findstr /i "id" cmd_post_yaml_result.txt > nul
if %errorlevel% equ 0 (
    echo [成功] POST 请求 /cmd 接口（YAML格式）返回了任务ID
) else (
    echo [失败] POST 请求 /cmd 接口（YAML格式）未返回任务ID
)

echo.
echo 测试 /cmd 接口的 onlyid 参数...
curl -s "http://localhost:%TEST_PORT%/cmd?cmd=version&onlyid=true" > cmd_onlyid.txt
type cmd_onlyid.txt
findstr /i "-" cmd_onlyid.txt > nul
if %errorlevel% equ 0 (
    echo [成功] /cmd 接口的 onlyid 参数正确返回了纯ID
) else (
    echo [失败] /cmd 接口的 onlyid 参数未正确返回纯ID
)

echo 关闭服务...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%TEST_PORT% "') do (
    taskkill /f /pid %%a > nul 2>&1
)

del cmd_get.txt cmd_post.json cmd_post_result.txt cmd_post.yaml cmd_post_yaml_result.txt cmd_onlyid.txt

echo.
echo 测试完成