@echo off
title BotMatrix Deployment Menu
cls
echo ===================================================
echo           BotMatrix Deployment Menu
echo ===================================================
echo 1. Deploy Manager Only (BotNexus) - Fast, No Overmind Build
echo 2. Deploy Overmind (Build + Deploy Manager)
echo 3. Deploy ALL Services
echo 4. Deploy System Worker
echo 5. Deploy WxBot
echo.
set /p choice=Enter your choice (1-5): 

if "%choice%"=="1" goto deploy_manager
if "%choice%"=="2" goto deploy_overmind
if "%choice%"=="3" goto deploy_all
if "%choice%"=="4" goto deploy_worker
if "%choice%"=="5" goto deploy_wxbot
echo Invalid choice.
pause
goto end

:deploy_manager
echo Starting deployment (Manager Only)...
python scripts/update.py BotNexus --services bot-manager
goto check_error

:deploy_overmind
echo [Step 1/3] Building Overmind (Flutter Web)...
pushd Overmind
call flutter build web --release --base-href /overmind/
if %ERRORLEVEL% NEQ 0 (
    echo Flutter build failed!
    popd
    pause
    exit /b 1
)
popd

echo [Step 2/3] Copying artifacts to BotNexus/overmind...
if not exist "BotNexus\overmind" mkdir "BotNexus\overmind"
xcopy /s /e /y "Overmind\build\web\*" "BotNexus\overmind\"

echo [Step 3/3] Deploying Manager...
python scripts/update.py BotNexus --services bot-manager
goto check_error

:deploy_all
call deploy_all.bat
goto end

:deploy_worker
call deploy_system_worker.bat
goto end

:deploy_wxbot
call deploy_wxbot.bat
goto end

:check_error
if %ERRORLEVEL% NEQ 0 (
    echo Deployment failed!
    pause
    exit /b %ERRORLEVEL%
)
echo Deployment finished successfully.
pause

:end

