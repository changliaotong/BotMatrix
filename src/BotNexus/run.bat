@echo off
SETLOCAL EnableDelayedExpansion

echo ========================================
echo   BotNexus Build ^& Run Script
echo ========================================

:: 1. Cleanup
if exist botnexus.exe (
    echo [1/3] Cleaning old binary...
    del botnexus.exe
)

:: 2. Build
echo [2/3] Building BotNexus...
go build -o botnexus.exe ./cmd/botnexus
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed! Please check your code.
    pause
    exit /b %ERRORLEVEL%
)

:: 3. Run
echo [3/3] Starting BotNexus...
echo ----------------------------------------
botnexus.exe
echo ----------------------------------------

ENDLOCAL
