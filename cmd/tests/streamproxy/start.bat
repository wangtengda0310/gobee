@echo off
setlocal

set HTTP_LOCAL=20144
set HTTP_REMOTE=10.254.114.204:20144
set TCP_LOCAL=18000
set TCP_REMOTE=10.254.114.204:18000

echo ============================================
echo   streamproxy
echo   HTTP: :%HTTP_LOCAL% ^<-^> %HTTP_REMOTE%
echo   TCP:  :%TCP_LOCAL%  ^<-^> %TCP_REMOTE%
echo ============================================
echo.

cd /d "%~dp0"
echo Working dir: %cd%
echo.

set GOOS=windows
set GOARCH=amd64
set GOTMPDIR=C:\temp
go build -o streamproxy.exe .
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build success. Starting...
.\streamproxy.exe -http :%HTTP_LOCAL%:%HTTP_REMOTE% -tcp :%TCP_LOCAL%:%TCP_REMOTE%
pause
