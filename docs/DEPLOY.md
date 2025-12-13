# Linux 服务器部署指南

本指南将帮助你将微信机器人部署到 Linux 服务器上，并使用 Docker 进行管理。

## 1. 准备工作

在开始之前，请确保你的 Linux 服务器满足以下条件：
- 已安装 Docker
- 已安装 Docker Compose
- **重要**：服务器能够访问到你的 SQL Server 数据库（默认配置 IP 为 `192.168.0.114`）。如果数据库在内网，请确保服务器与数据库在同一网络，或修改配置连接到可访问的数据库地址。

## 2. 上传文件

将项目目录下的所有文件上传到 Linux 服务器的一个目录中，例如 `/opt/BotMatrix`。

必须包含的文件：
- `Dockerfile`
- `docker-compose.yml`
- `requirements.txt`
- `config.json`
- `*.py` (所有 Python 源代码文件)

## 3. 修改配置

### 修改数据库连接
打开 `docker-compose.yml`，在 `environment` 部分修改数据库连接信息：

```yaml
    environment:
      - DB_SERVER=你的数据库IP  # 如果是 192.168.0.114 请确保服务器能访问
      - DB_NAME=sz84_robot
      - DB_USER=derlin
      - DB_PASSWORD=fkueiqiq461686
```

### 修改机器人端口配置
如果需要调整端口，请修改 `config.json` 和 `docker-compose.yml`。

## 4. 自动化部署 (推荐)

在本地 Windows 终端直接运行：
```powershell
python deploy.py
```
该脚本会自动打包、上传并重启远程服务。

## 5. 登录机器人

部署成功后，服务会自动启动。

1.  **扫码登录**: 
    访问管理后台：`http://192.168.0.167:5000`
    如果页面未显示二维码，可访问 API 直接获取：`http://192.168.0.167:5000/api/qr_code?bot_id=YOUR_BOT_ID`

2.  **连接机器人**:
    机器人 WebSocket 服务已映射到以下端口：
    - **Port 3111** (对应容器内 3001)
    - **Port 3112** (对应容器内 3002)
    - **Port 3113** (对应容器内 3003)
    
    客户端连接地址示例：`ws://192.168.0.167:3111`

## 6. 常用命令

登录服务器 (`ssh derlin@192.168.0.167`) 后执行：

- **查看实时日志**: 
  ```bash
  cd /opt/BotMatrix
  sudo docker-compose logs -f --tail=100
  ```

## 7. GitHub Actions 自动部署 (可选)

如果你希望在每次 `git push` 到 `main` 分支时自动部署，可以使用项目自带的 Workflow。

### 前置条件：配置 Self-hosted Runner
由于目标服务器 IP (`192.168.0.167`) 是内网地址，GitHub 无法直接访问，因此需要将目标服务器配置为 GitHub 的 Self-hosted Runner。

1.  进入 GitHub 仓库页面 -> **Settings** -> **Actions** -> **Runners**。
2.  点击 **New self-hosted runner**。
3.  选择 **Linux**，按照页面提示在服务器 (`192.168.0.167`) 上执行安装命令。
4.  安装完成后，确保 Runner 处于 `Idle` (空闲) 状态。

### 配置 Secrets
在 GitHub 仓库 **Settings** -> **Secrets and variables** -> **Actions** 中添加以下 Secrets (用于生成 `config.json`)：

- `DB_SERVER`: 数据库 IP (如 192.168.0.114)
- `DB_NAME`: 数据库名 (如 sz84_robot)
- `DB_USER`: 数据库用户名
- `DB_PASSWORD`: 数据库密码

### 启用
配置完成后，每次推送代码到 `main` 分支，服务器上的 Runner 会自动拉取代码、生成配置并重启 Docker 服务。

## 常见问题

**Q: 无法连接数据库？**
A: 请检查服务器是否能 ping 通数据库 IP。如果数据库在本地 Windows 电脑上，Linux 服务器在云端，你需要配置 VPN 或将数据库暴露到公网（不推荐）。

**Q: 扫码后没有反应？**
A: 查看日志 `docker-compose logs -f`，确认是否有报错信息。
