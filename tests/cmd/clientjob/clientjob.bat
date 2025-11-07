echo %~1 %~2 %~3
copy "%~dp0\anothercmd.zip" .\anothercmd.zip
powershell -Command "Expand-Archive -Path '.\anothercmd.zip' -DestinationPath '.'"
anothercmd.bat