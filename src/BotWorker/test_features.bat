@echo off
chcp 65001 >nul

echo ====================
echo BotWorker 功能测试脚本
echo ====================
echo.
echo 此脚本用于测试群管理功能和欢迎语插件

echo.
echo 1. 启动 BotWorker 机器人
echo.  请确保已配置好 OneBot 服务端并连接到机器人

echo.
echo 2. 测试头衔设置功能（仅群主可用）
echo.  命令格式：/settitle <用户ID> <头衔>
echo.  示例：/settitle 123456 测试头衔
echo.  注意：头衔长度不能超过12个字符

echo.
echo 3. 测试欢迎语插件
echo.  当有新成员加入群聊时，机器人会自动发送欢迎消息

echo.
echo 4. 测试帮助信息
echo.  命令：help
echo.  查看所有可用命令，包括新添加的settitle命令

echo.
echo 按任意键启动机器人...
pause >nul

echo.
echo 启动机器人中...
go run cmd/main.go

echo.
echo 机器人已停止运行
pause