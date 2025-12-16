@echo off
echo ==========================================
echo Building WxBotGo for Android (ARM64/Termux)
echo ==========================================

setlocal

:: Set Environment Variables for Cross-Compilation
set GOOS=linux
set GOARCH=arm64

echo OS: %GOOS%
echo ARCH: %GOARCH%

:: Build
go build -ldflags "-s -w" -o wxbot-android-arm64 ./cmd/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [SUCCESS] Build complete!
    echo Output file: wxbot-android-arm64
    echo.
    echo Transfer this file to your Android phone and run it in Termux.
) else (
    echo.
    echo [FAILED] Build failed.
)

endlocal
pause
