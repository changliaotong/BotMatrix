@echo off
chcp 65001 >nul

echo ==============================
echo BotWorker 用户输入功能测试脚本
echo ==============================
echo.
echo 此脚本用于测试需要用户输入的功能
echo.
echo 按任意键启动机器人...
pause >nul

echo.
echo 启动机器人中...
go run cmd/main.go