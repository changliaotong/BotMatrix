# BotMatrix Docker 容器管理脚本
# 提供简化的容器操作命令

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('start', 'stop', 'restart', 'status', 'logs', 'shell', 'up', 'down', 'clean')]
    [string]$Action = 'status',
    
    [Parameter(Mandatory=$false)]
    [string]$Service = '',
    
    [Parameter(Mandatory=$false)]
    [switch]$All
)

# 容器名称映射表
$ContainerMap = @{
    'manager' = 'btmgr'
    'system' = 'btsys'
    'qq' = 'btqq'
    'wechat' = 'btwc'
    'wechat-go' = 'btwg'
    'dingtalk' = 'btdt'
    'feishu' = 'btfs'
    'telegram' = 'bttg'
    'discord' = 'btdc'
    'slack' = 'btsl'
    'kook' = 'btkk'
    'email' = 'btem'
    'wecom' = 'btwm'
}

# 反向映射（容器名到服务名）
$ReverseMap = @{
    'btmgr' = 'manager'
    'btsys' = 'system'
    'btqq' = 'qq'
    'btwc' = 'wechat'
    'btwg' = 'wechat-go'
    'btdt' = 'dingtalk'
    'btfs' = 'feishu'
    'bttg' = 'telegram'
    'btdc' = 'discord'
    'btsl' = 'slack'
    'btkk' = 'kook'
    'btem' = 'email'
    'btwm' = 'wecom'
}

function Show-Help {
    Write-Host @"
BotMatrix Docker 管理脚本

用法: .\docker-manager.ps1 -Action <操作> [-Service <服务名>] [-All]

操作:
  start    - 启动容器
  stop     - 停止容器  
  restart  - 重启容器
  status   - 查看状态（默认）
  logs     - 查看日志
  shell    - 进入容器shell
  up       - 启动所有服务
  down     - 停止所有服务
  clean    - 清理所有容器和镜像

服务名:
  manager     - 管理平台 (btmgr)
  system      - 系统服务 (btsys)
  qq          - QQ机器人 (btqq)
  wechat      - 微信机器人 (btwc)
  wechat-go   - 微信Go版 (btwg)
  dingtalk    - 钉钉机器人 (btdt)
  feishu      - 飞书机器人 (btfs)
  telegram    - Telegram机器人 (bttg)
  discord     - Discord机器人 (btdc)
  slack       - Slack机器人 (btsl)
  kook        - Kook机器人 (btkk)
  email       - 邮件机器人 (btem)
  wecom       - 企业微信机器人 (btwm)

示例:
  .\docker-manager.ps1 -Action status                    # 查看所有容器状态
  .\docker-manager.ps1 -Action restart -Service manager # 重启管理平台
  .\docker-manager.ps1 -Action logs -Service wechat    # 查看微信机器人日志
  .\docker-manager.ps1 -Action shell -Service manager   # 进入管理平台shell
  .\docker-manager.ps1 -Action up -All                  # 启动所有服务
  .\docker-manager.ps1 -Action stop -Service qq          # 停止QQ机器人

快捷命令:
  查看状态: docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  快速重启: docker restart <容器名>
  查看日志: docker logs -f <容器名>
"@
}

function Get-ContainerStatus {
    Write-Host "容器状态一览:" -ForegroundColor Green
    Write-Host "=============" -ForegroundColor Green
    
    # 获取所有相关容器
    $containers = docker ps -a --format "{{.Names}}" | Where-Object { $ReverseMap.ContainsKey($_) }
    
    if ($containers.Count -eq 0) {
        Write-Host "未找到BotMatrix容器" -ForegroundColor Yellow
        return
    }
    
    foreach ($container in $containers) {
        $serviceName = $ReverseMap[$container]
        $status = docker inspect -f '{{.State.Status}}' $container 2>$null
        $health = docker inspect -f '{{.State.Health.Status}}' $container 2>$null
        $ports = docker inspect -f '{{range $p, $conf := .NetworkSettings.Ports}}{{$p}} -> {{(index $conf 0).HostPort}} {{end}}' $container 2>$null
        
        $statusColor = if ($status -eq 'running') { 'Green' } else { 'Red' }
        Write-Host "$($serviceName.PadRight(12)) ($container): " -NoNewline
        Write-Host $status -ForegroundColor $statusColor -NoNewline
        if ($health -and $health -ne '') {
            Write-Host " (健康: $health)" -ForegroundColor $(if ($health -eq 'healthy') { 'Green' } else { 'Red' })
        } else {
            Write-Host ""
        }
        if ($ports -and $ports -ne '') {
            Write-Host "  端口: $ports" -ForegroundColor Gray
        }
    }
}

function Start-Container {
    param([string]$container)
    Write-Host "启动容器: $container" -ForegroundColor Green
    docker start $container
}

function Stop-Container {
    param([string]$container)
    Write-Host "停止容器: $container" -ForegroundColor Yellow
    docker stop $container
}

function Restart-Container {
    param([string]$container)
    Write-Host "重启容器: $container" -ForegroundColor Cyan
    docker restart $container
}

function Show-Logs {
    param([string]$container)
    Write-Host "查看日志: $container (按 Ctrl+C 退出)" -ForegroundColor Green
    docker logs -f --tail 100 $container
}

function Enter-Shell {
    param([string]$container)
    Write-Host "进入容器shell: $container" -ForegroundColor Green
    docker exec -it $container /bin/sh
}

function Start-AllServices {
    Write-Host "启动所有BotMatrix服务..." -ForegroundColor Green
    docker-compose up -d
}

function Stop-AllServices {
    Write-Host "停止所有BotMatrix服务..." -ForegroundColor Yellow
    docker-compose down
}

function Clean-All {
    Write-Host "清理所有BotMatrix容器和镜像..." -ForegroundColor Red
    Write-Host "警告: 这将删除所有相关容器和镜像！" -ForegroundColor Red
    $confirm = Read-Host "确认清理? (输入 'yes' 确认)"
    if ($confirm -eq 'yes') {
        docker-compose down --rmi all --volumes
        Write-Host "清理完成" -ForegroundColor Green
    } else {
        Write-Host "操作已取消" -ForegroundColor Yellow
    }
}

# 主逻辑
if ($Action -eq 'help') {
    Show-Help
    exit
}

# 处理服务名到容器名的转换
$containerName = $null
if ($Service) {
    if ($ContainerMap.ContainsKey($Service)) {
        $containerName = $ContainerMap[$Service]
    } else {
        # 可能是直接的容器名
        $containerName = $Service
    }
}

# 执行操作
switch ($Action) {
    'status' {
        if ($Service) {
            # 检查特定容器状态
            $status = docker inspect -f '{{.State.Status}}' $containerName 2>$null
            if ($status) {
                Write-Host "$containerName 状态: $status" -ForegroundColor $(if ($status -eq 'running') { 'Green' } else { 'Red' })
            } else {
                Write-Host "容器 $containerName 不存在" -ForegroundColor Red
            }
        } else {
            Get-ContainerStatus
        }
    }
    'start' {
        if ($All) {
            Start-AllServices
        } elseif ($Service) {
            Start-Container $containerName
        } else {
            Write-Host "请指定服务名或使用 -All 参数" -ForegroundColor Red
        }
    }
    'stop' {
        if ($All) {
            Stop-AllServices
        } elseif ($Service) {
            Stop-Container $containerName
        } else {
            Write-Host "请指定服务名或使用 -All 参数" -ForegroundColor Red
        }
    }
    'restart' {
        if ($Service) {
            Restart-Container $containerName
        } else {
            Write-Host "请指定服务名" -ForegroundColor Red
        }
    }
    'logs' {
        if ($Service) {
            Show-Logs $containerName
        } else {
            Write-Host "请指定服务名" -ForegroundColor Red
        }
    }
    'shell' {
        if ($Service) {
            Enter-Shell $containerName
        } else {
            Write-Host "请指定服务名" -ForegroundColor Red
        }
    }
    'up' {
        Start-AllServices
    }
    'down' {
        Stop-AllServices
    }
    'clean' {
        Clean-All
    }
    default {
        Show-Help
    }
}