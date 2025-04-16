@echo off
REM 确保在任何目录下调用都能正确定位mp.exe
REM 文件夹内自带压缩工具GnuZip\zip.exe
setlocal

REM 提示乱码解决方法
REM 若控制台出现乱码，请取消下方chcp行的注释以切换为UTF-8代码页
chcp 65001 >nul

REM 参数校验
if "%~1"=="" (
    echo Error: 缺少第一个参数（game-meta导出目录） >&2
    exit /b 1
)
if "%~2"=="" (
    echo Error: 缺少第二个参数（输出目录） >&2
    exit /b 1
)

set "SRC_DIR=%~f1"
set "DST_DIR=%~f2"

REM 清理并重建输出目录
if exist "%DST_DIR%" (
    rd /s /q "%DST_DIR%"
    if errorlevel 1 (
        echo Error: 无法删除目录 "%DST_DIR%" >&2
        exit /b %errorlevel%
    )
)
md "%DST_DIR%"
if errorlevel 1 (
    echo Error: 无法创建目录 "%DST_DIR%" >&2
    exit /b %errorlevel%
)

REM 创建标准子目录结构
for %%d in (
    "msgpackdata"
    "ui_systemui"
    "ui_gameplayui"
    "archivedconfigs"
) do (
    md "%DST_DIR%\%%~d"
    if errorlevel 1 (
        echo Error: 无法创建子目录 "%%~d" >&2
        exit /b %errorlevel%
    )
)

REM 设置工具路径
set "MP_TOOL=%~dp0mp.exe"

REM 数据转换流程（注意以下输出可能包含UTF-8字符）
REM CSV转换（可能产生乱码日志）
"%MP_TOOL%" --csvdir "%SRC_DIR%\config" "%DST_DIR%\msgpackdata"
if errorlevel 1 (
    echo Error: CSV转换失败，请检查原始数据 >&2
    exit /b %errorlevel%
)

REM JSON配置转换（可能产生乱码日志）
"%MP_TOOL%" --jsondir "%SRC_DIR%\config-json" "%DST_DIR%\msgpackdata"
if errorlevel 1 (
    echo Error: JSON配置转换失败 >&2
    exit /b %errorlevel%
)

REM 系统UI处理
xcopy /Y /E "%SRC_DIR%\ui\*" "%DST_DIR%\ui_systemui"
if errorlevel 1 (
    echo Error: 系统UI处理失败 >&2
    exit /b %errorlevel%
)

REM 玩法UI处理
xcopy /Y /E "%SRC_DIR%\config-json\ui\*" "%DST_DIR%\ui_gameplayui"
if errorlevel 1 (
    echo Error: 玩法UI处理失败 >&2
    exit /b %errorlevel%
)

set "ZIP_TOOL=%~dp0GnuZip\zip.exe"
cd "%DST_DIR%\archivedconfigs\"
REM 开始打包msgpack的zip文件
xcopy /E /Y /I "%DST_DIR%\msgpackdata" msgpack
"%ZIP_TOOL%" -9 -r msgpack_archived.bytes msgpack
if errorlevel 1 (
    echo Error: 打包msgpack的zip文件失败 >&2
    exit /b %errorlevel%
)

endlocal
echo 本地私服构建完成
exit /b 0