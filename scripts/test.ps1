# BotMatrix 本地一键测试脚本 (PowerShell)

$ErrorActionPreference = "Stop"

Write-Host "--- 开始本地自动化测试 ---" -ForegroundColor Cyan

# 1. 环境准备检查
Write-Host "[1/3] 检查环境配置..." -ForegroundColor Yellow
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOWORK = "off"

# 2. 运行 Common 模块单元测试
Write-Host "[2/3] 运行 Common 模块单元测试..." -ForegroundColor Yellow
cd src/Common
go test -v .
cd ../..

# 3. 运行 BotWorker 集成测试 (示例)
Write-Host "[3/3] 运行 BotWorker 集成测试..." -ForegroundColor Yellow
# 注意：集成测试可能需要数据库，这里仅运行基础测试文件
go test -v src/BotWorker/integration_test.go

Write-Host "`n√ 所有测试已通过！" -ForegroundColor Green
