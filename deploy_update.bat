@echo off
echo ===================================================
echo   Deploying UPDATED Services to Server
echo   Services: tencent-bot system-worker bot-manager wxbot
echo   Target: /opt/BotMatrix
echo ===================================================

powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1 -Service "tencent-bot system-worker bot-manager wxbot"

if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Deployment finished successfully.
pause
