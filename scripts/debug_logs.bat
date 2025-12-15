@echo off
echo =========================================================
echo   BotMatrix Log Watcher
echo =========================================================
echo.
echo Usage: 
echo   View last 100 lines and follow:
echo     scripts\debug_logs.bat
echo.
echo   View specific container (e.g. tencent-bot):
echo     scripts\debug_logs.bat tencent-bot
echo.
echo =========================================================

set TARGET=%1
if "%TARGET%"=="" (
    set TARGET=
)

if "%TARGET%"=="" (
    echo [INFO] Following logs for ALL containers...
    docker-compose logs -f --tail=100
) else (
    echo [INFO] Following logs for container: %TARGET%...
    docker-compose logs -f --tail=100 %TARGET%
)
