@echo off
SETLOCAL EnableDelayedExpansion

echo ========================================
echo   BotWorker Build ^& Run Script
echo ========================================

:: 1. Cleanup
if exist botworker.exe (
    echo [1/3] Cleaning old binary...
    del botworker.exe
)

:: 2. Build
echo [2/3] Building BotWorker...
go build -o botworker.exe ./cmd/botworker
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed! Please check your code.
    pause
    exit /b %ERRORLEVEL%
)

:: 3. Run
echo [3/3] Starting BotWorker...
echo ----------------------------------------
botworker.exe
echo ----------------------------------------

ENDLOCAL
