@echo off
title Deploy ALL Services
echo ===================================================
echo   Deploying ALL Services
echo   Target: /opt/BotMatrix
echo ===================================================

powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1

if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Deployment finished successfully.
pause
