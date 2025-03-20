echo %~1 %~2 %~3
curl "%~1?libCode=test&libVersionCode=test_default&status=1"
@REM copy "%~dp0\ToolsPack_GForgeData2MsgPack_20250304_1824.zip" .\ToolsPack_GForgeData2MsgPack_20250304_1824.zip
@REM %~dp0game-meta.exe serialize -p %~1 -b %~2 --meta.branch dev
@REM powershell -Command "Expand-Archive -Path 'D:\wangtengda\gobee\lvan\exporter\cmd\cpworkdir\ToolsPack_GForgeData2MsgPack_20250304_1824.zip' -DestinationPath '.'"
@REM xcopy ".\output\client\config\*" "ToolsPack_GForgeData2MsgPack\zzDesingerDataTransform\gforge_output_data\" /E /Y /I
@REM xcopy ".\output\client\config-json" "ToolsPack_GForgeData2MsgPack\zzDesingerDataTransform\gforge_output_data\config-json\" /E /Y /I
@REM xcopy ".\output\client\script" "ToolsPack_GForgeData2MsgPack\zzDesingerDataTransform\gforge_output_data\config-json\" /E /Y /I
@REM ToolsPack_GForgeData2MsgPack\zzDesingerDataTransform\_Run_AIO_DesignerDataConvert.bat