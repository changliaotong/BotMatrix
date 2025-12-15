@echo off
title Deploy BotNexus (Manager)
echo ===================================================
echo   Deploying Service: bot-manager
echo   Target: /opt/BotMatrix
echo ===================================================

powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1 -Service "bot-manager"

if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Deployment finished successfully.
pause
