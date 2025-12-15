
@echo off
echo Starting SystemWorker locally for debugging...
echo Connecting to ws://localhost:3001
set BOT_MANAGER_URL=ws://localhost:3001
python SystemWorker/main.py
pause
