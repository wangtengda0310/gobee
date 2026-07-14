@echo off
setlocal enabledelayedexpansion

set MCP_URL=http://127.0.0.1:8765
set TEST_ID=0
set PASS_COUNT=0
set FAIL_COUNT=0
set SKIP_COUNT=0

echo ========================================
echo  MCP Server Automated Test Script
echo  URL: %MCP_URL%
echo ========================================

REM 辅助函数：验证响应是否包含指定字符串
REM 使用方法：call :assert_response "响应内容" "期望字符串" "测试名称"

REM ========================================
REM [1] Server Health Check
REM ========================================
echo.
echo ========================================
echo [1] Server Health Check
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] Testing tools/list...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/list\",\"id\":0}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "tools" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] tools/list - 响应包含工具列表
    set /a PASS_COUNT+=1
) else (
    echo [FAIL] tools/list - 响应缺少工具列表
    type %TEMP%\mcp_test_%TEST_ID%.json
    set /a FAIL_COUNT+=1
)

REM ========================================
REM [2] Config Management Tools
REM ========================================
echo.
echo ========================================
echo [2] Config Management Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_func_config...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_func_config\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "jsons_dir" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_func_config - 返回有效配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_func_config - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_func_config - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_excel_config...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_excel_config\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "excel_resources_dir" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_excel_config - 返回有效配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_excel_config - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_excel_config - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [3] Global Settings Tools
REM ========================================
echo.
echo ========================================
echo [3] Global Settings Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_feishu_config...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_feishu_config\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "fei_shu" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_feishu_config - 返回飞书配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_feishu_config - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_feishu_config - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_mcp_config...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_mcp_config\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "enabled" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_mcp_config - 返回MCP配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_mcp_config - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_mcp_config - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_mcp_status...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_mcp_status\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "running" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_mcp_status - 返回运行状态
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_mcp_status - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_mcp_status - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [4] Game Data Tools
REM ========================================
echo.
echo ========================================
echo [4] Game Data Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_all_hero_cfg...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_all_hero_cfg\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_all_hero_cfg - 返回英雄配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_all_hero_cfg - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_all_hero_cfg - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_all_card_cfg...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_all_card_cfg\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_all_card_cfg - 返回卡牌配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_all_card_cfg - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_all_card_cfg - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_all_skill_cfg...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_all_skill_cfg\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_all_skill_cfg - 返回技能配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_all_skill_cfg - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_all_skill_cfg - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_msg_id_map...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_msg_id_map\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_msg_id_map - 返回消息ID映射
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_msg_id_map - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_msg_id_map - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_error_code_map...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_error_code_map\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_error_code_map - 返回错误码映射
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_error_code_map - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_error_code_map - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_property_type_map...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_property_type_map\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_property_type_map - 返回属性类型映射
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_property_type_map - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_property_type_map - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [5] Function Test Tools
REM ========================================
echo.
echo ========================================
echo [5] Function Test Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] is_running...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"is_running\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "running" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] is_running - 返回运行状态
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] is_running - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] is_running - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_test_logs...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_test_logs\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_test_logs - 返回测试日志
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_test_logs - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_test_logs - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_hero_list...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_hero_list\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_hero_list - 返回英雄列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_hero_list - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_hero_list - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_fight_cases...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_fight_cases\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_fight_cases - 返回战斗用例列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_fight_cases - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_fight_cases - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_fight_test_summary...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_fight_test_summary\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_fight_test_summary - 返回测试汇总
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_fight_test_summary - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_fight_test_summary - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [6] Excel Check Tools
REM ========================================
echo.
echo ========================================
echo [6] Excel Check Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] list_table_rule_types...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"list_table_rule_types\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] list_table_rule_types - 返回规则类型列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] list_table_rule_types - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] list_table_rule_types - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [7] Feishu Notify Config
REM ========================================
echo.
echo ========================================
echo [7] Feishu Notify Config
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_feishu_notify_config...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_feishu_notify_config\",\"arguments\":{}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_feishu_notify_config - 返回飞书通知配置
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_feishu_notify_config - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_feishu_notify_config - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [8] Case Management Tools
REM ========================================
echo.
echo ========================================
echo [8] Case Management Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_categories (jsons_dir=cases/fight_cases)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_categories\",\"arguments\":{\"dirPath\":\"cases/fight_cases\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_categories - 返回分类列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_categories - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_categories - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_case_list (filePath=cases/fight_cases)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_case_list\",\"arguments\":{\"filePath\":\"cases/fight_cases\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_case_list - 返回用例列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_case_list - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_case_list - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] search_cases (keyword=测试)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"search_cases\",\"arguments\":{\"filePath\":\"cases/fight_cases\",\"keyword\":\"荆轲\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] search_cases - 返回搜索结果
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] search_cases - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] search_cases - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [9] Excel Rule Tools
REM ========================================
echo.
echo ========================================
echo [9] Excel Rule Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_excel_rules (dir=cases/excel_cases)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_excel_rules\",\"arguments\":{\"dir\":\"cases/excel_cases\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_excel_rules - 返回规则列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_excel_rules - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_excel_rules - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_table_rules (sheetName=Hero)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_table_rules\",\"arguments\":{\"dir\":\"cases/excel_cases\",\"sheetName\":\"Hero\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_table_rules - 返回表级规则
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_table_rules - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_table_rules - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [10] Excel Preview Tools
REM ========================================
echo.
echo ========================================
echo [10] Excel Preview Tools
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_all_excels (dirPath=../../config/excel)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_all_excels\",\"arguments\":{\"dirPath\":\"../../config/excel\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_all_excels - 返回Excel文件列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_all_excels - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_all_excels - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_excel_sheets (filePath=../../config/excel/Hero.xlsx)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_excel_sheets\",\"arguments\":{\"filePath\":\"../../config/excel/Hero.xlsx\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_excel_sheets - 返回Sheet列表
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_excel_sheets - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_excel_sheets - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] preview_excel_sheet (Hero.xlsx/Hero)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"preview_excel_sheet\",\"arguments\":{\"filePath\":\"../../config/excel/Hero.xlsx\",\"sheetName\":\"Hero\",\"rows\":5}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] preview_excel_sheet - 返回预览数据
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] preview_excel_sheet - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] preview_excel_sheet - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] get_excel_column_info (Hero.xlsx/Hero)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_excel_column_info\",\"arguments\":{\"filePath\":\"../../config/excel/Hero.xlsx\",\"sheetName\":\"Hero\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_excel_column_info - 返回列信息
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_excel_column_info - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_excel_column_info - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [11] Wiki Check Tool
REM ========================================
echo.
echo ========================================
echo [11] Wiki Check Tool
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] check_hero_wiki (excelDir=../../config/excel)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"check_hero_wiki\",\"arguments\":{\"excelDir\":\"../../config/excel\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] check_hero_wiki - 返回Wiki检查结果
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] check_hero_wiki - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] check_hero_wiki - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [12] Fight Test Async
REM ========================================
echo.
echo ========================================
echo [12] Fight Test Async
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] run_fight_test_async (heroId=11205)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"run_fight_test_async\",\"arguments\":{\"heroId\":11205}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr "content" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] run_fight_test_async - 异步测试已启动
    set /a PASS_COUNT+=1
) else (
    findstr "error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] run_fight_test_async - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] run_fight_test_async - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM [13] Tools Requiring External Resources (必须跳过)
REM ========================================
echo.
echo ========================================
echo [13] Tools Requiring External Resources
echo ========================================
echo [SKIP] get_guild_leaders - 原因: 需要游戏服务器连接 (serverIp, serverPort)
echo [SKIP] create_guild_with_city - 原因: 需要游戏服务器连接
echo [SKIP] upgrade_guild_city - 原因: 需要游戏服务器连接
set /a SKIP_COUNT+=3

REM ========================================
REM [14] Tools with Side Effects (必须跳过)
REM ========================================
echo.
echo ========================================
echo [14] Tools with Side Effects
echo ========================================
echo [SKIP] save_func_config - 原因: 会覆盖配置文件
echo [SKIP] save_excel_config - 原因: 会覆盖配置文件
echo [SKIP] save_mcp_config - 原因: 会重启MCP服务
echo [SKIP] update_feishu_config - 原因: 会修改飞书配置
echo [SKIP] set_feishu_notify_config - 原因: 会修改通知配置
echo [SKIP] send_feishu_message - 原因: 会发送真实飞书消息
echo [SKIP] stop_robot_test - 原因: 会停止正在运行的测试
echo [SKIP] clear_test_logs - 原因: 会清除测试日志
echo [SKIP] add_table_rule - 原因: 会添加规则到配置
echo [SKIP] del_table_rule - 原因: 会删除规则
echo [SKIP] save_excel_rules - 原因: 会覆盖规则配置
echo [SKIP] save_hero_wiki - 原因: 需要复杂的data参数
set /a SKIP_COUNT+=12

REM ========================================
REM [15] Tools with Complex Parameters (必须跳过)
REM ========================================
echo.
echo ========================================
echo [15] Tools with Complex Parameters
echo ========================================
echo [SKIP] run_robot_test - 原因: 需要服务器连接和复杂参数
echo [SKIP] run_fight_test - 原因: 同步执行耗时长，用async替代
echo [SKIP] get_test_progress - 原因: 需要先有运行中的taskId
echo [SKIP] get_case_by_name - 原因: 需要知道具体用例名
echo [SKIP] check_excel_rules - 原因: 需要复杂的rules数组参数
set /a SKIP_COUNT+=5

REM ========================================
REM [16] Git Diff 相关工具测试
REM ========================================
echo.
echo ========================================
echo [16] Git Diff 相关工具
echo ========================================

set /a TEST_ID+=1
echo [%TEST_ID%] get_git_changed_excels (基础测试)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_git_changed_excels\",\"arguments\":{\"repoPath\":\"../../work/config\",\"targetDir\":\"config/excel\",\"baseCommit\":\"HEAD~1\"}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr /C:"relPath" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] get_git_changed_excels - 返回文件列表
    set /a PASS_COUNT+=1
) else (
    findstr /C:"error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] get_git_changed_excels - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] get_git_changed_excels - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

set /a TEST_ID+=1
echo [%TEST_ID%] check_table_rules_only (仅通知规则)...
curl -s -X POST %MCP_URL% -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"check_table_rules_only\",\"arguments\":{\"dir\":\"../../work/config/excel\",\"ruleTypes\":[\"NEW_ROW_NOTIFY\",\"ROW_CHANGE_NOTIFY\"]}},\"id\":%TEST_ID%}" 2>nul > %TEMP%\mcp_test_%TEST_ID%.json
findstr /C:"tableResults" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] check_table_rules_only - 返回检查结果
    set /a PASS_COUNT+=1
) else (
    findstr /C:"error" %TEMP%\mcp_test_%TEST_ID%.json >nul 2>&1
    if !errorlevel! equ 0 (
        echo [FAIL] check_table_rules_only - 返回错误
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    ) else (
        echo [FAIL] check_table_rules_only - 响应格式异常
        type %TEMP%\mcp_test_%TEST_ID%.json
        set /a FAIL_COUNT+=1
    )
)

REM ========================================
REM 测试结果汇总
REM ========================================
echo.
echo ========================================
echo  Test Summary
echo ========================================
echo  Total Tests: %TEST_ID%
echo  Passed:      %PASS_COUNT%
echo  Failed:      %FAIL_COUNT%
echo  Skipped:     %SKIP_COUNT%
echo ========================================

REM 清理临时文件
del %TEMP%\mcp_test_*.json 2>nul

REM 返回退出码：有失败则返回1，否则返回0
if %FAIL_COUNT% gtr 0 (
    echo.
    echo [RESULT] TEST FAILED - %FAIL_COUNT% test(s) failed
    exit /b 1
) else (
    echo.
    echo [RESULT] ALL TESTS PASSED
    exit /b 0
)
