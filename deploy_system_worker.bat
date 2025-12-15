@echo off
title Deploy SystemWorker
echo ===================================================
echo   Deploying Service: system-worker
echo   Target: /opt/BotMatrix
echo ===================================================

powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1 -Service "system-worker"

if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Deployment finished successfully.
pause
