@echo off
title Deploy TelegramBot
echo ===================================================
echo   Deploying Service: telegram-bot
echo   Target: /opt/BotMatrix
echo ===================================================

powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1 -Service "telegram-bot"

if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Deployment finished successfully.
pause
