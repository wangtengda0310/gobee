
set MCP_URL=http://127.0.0.1:8765
set TEST_ID=0
set PASS_COUNT=0
set FAIL_COUNT=0
set SKIP_COUNT=0

set /a TEST_ID+=1
echo [%TEST_ID%] get_hero_cfg_by_name...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_hero_cfg_by_name\",\"arguments\":{\"hero_name\":\"赵云\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_hero_cfg_by_name - 返回英雄配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_hero_cfg_by_name - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_hero_cfg_by_name - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)
