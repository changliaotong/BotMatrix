@echo off
echo ===============================================
echo    BotNexus 实时日志查看器
echo ===============================================
echo.
echo 正在查看BotNexus日志，按 Ctrl+C 停止...
echo.
echo 日志中会显示：
echo - [INFO]  Bot连接/断开信息
echo - [WARN]  警告信息（如心跳超时）
echo - [ERROR] 错误信息
echo - [DEBUG] 心跳调试信息
echo.
echo 关键词：
echo - "heartbeat" - 心跳相关
echo - "timeout" - 连接超时
echo - "disconnected" - 连接断开
echo - "connected" - 连接建立
echo.

:loop
cd /d "%~dp0BotNexus"
timeout /t 1 >nul
cls
echo ===============================================
echo    BotNexus 实时日志查看器
echo ===============================================
echo 最后更新时间: %date% %time%
echo.
echo 最近日志：
echo.

rem 显示最后50行日志，过滤关键信息
powershell -Command "Get-Content botnexus.log -ErrorAction SilentlyContinue | Select-Object -Last 50 | Where-Object { \$_ -match '(INFO|WARN|ERROR|DEBUG|heartbeat|timeout|connected|disconnected)' }"

goto loop