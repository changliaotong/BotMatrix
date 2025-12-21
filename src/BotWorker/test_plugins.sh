#!/bin/bash

echo "=== 机器人插件命令行测试工具 ==="
echo ""

echo "1. 测试积分查询命令"
go run test_cli.go -msg '/ 积分'
echo ""

echo "2. 测试签到命令"
go run test_cli.go -msg '/ 签到'
echo ""

echo "3. 测试猜拳游戏命令"
go run test_cli.go -msg '/ 猜拳 石头'
echo ""

echo "4. 测试计算命令"
go run test_cli.go -msg '/ 计算 1+2*3'
echo ""

echo "5. 测试报时命令"
go run test_cli.go -msg '/ 报时'
echo ""

echo "6. 测试抽奖命令"
go run test_cli.go -msg '/ 抽奖'
echo ""

echo "7. 测试命令解析器"
go run test_command_parser.go -msg '/  积分   '  # 测试多个空格
echo ""

echo "8. 测试命令解析器 - 无空格"
go run test_command_parser.go -msg '/积分'
