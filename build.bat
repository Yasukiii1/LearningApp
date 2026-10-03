@echo off
setlocal
cd /d "%~dp0"

if not exist builds mkdir builds

echo Building LearningApp...
go build -ldflags="-H=windowsgui" -o "builds\LearningApp.exe" .

if errorlevel 1 (
    echo.
    echo Build failed.
    exit /b 1
)

echo.
echo Build complete:
echo %CD%\builds\LearningApp.exe
endlocal
