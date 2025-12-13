@echo off
echo Starting deployment...
python scripts/update.py
if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)
echo Deployment finished successfully.
