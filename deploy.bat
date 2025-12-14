@echo off
echo Starting deployment (Default: Manager Only)...
python scripts/update.py BotNexus --services bot-manager
if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)
echo Deployment finished successfully.
