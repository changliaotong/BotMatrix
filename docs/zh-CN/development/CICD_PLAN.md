# BotMatrix CI/CD 规划方案

本方案旨在为 BotMatrix 项目建立一套从代码提交到自动化部署的完整流水线。

## 1. 总体架构

流水线分为三个核心阶段：**持续集成 (CI)**、**持续交付 (CD)** 和 **监控反馈**。

### 流程概览
`代码提交` -> `静态检查 (Lint)` -> `单元测试` -> `构建镜像 (Docker)` -> `集成测试` -> `自动部署 (Staging/Prod)` -> `健康检查`

---

## 2. 持续集成 (CI) - 质量保证

### 策略
- **触发时机**：每次 Pull Request 和 Push 到 `main`/`master` 分支。
- **核心任务**：
  1. **Linting**：使用 `golangci-lint` 检查代码规范、潜在 Bug。
  2. **单元测试**：运行 `go test ./src/Common/...` 等，确保基础逻辑正确。
  3. **安全性扫描**：使用 `gosec` 扫描代码中的安全隐患。
  4. **构建校验**：确保所有模块（Nexus, Worker, Overmind）都能成功编译。

---

## 3. 持续交付 (CD) - 自动化部署

### 镜像管理
- 使用 **GitHub Packages (ghcr.io)** 或 **Docker Hub** 存储镜像。
- 镜像标签 (Tag) 策略：
  - `edge`: `master` 分支的最新构建。
  - `vX.Y.Z`: 语义化版本发布的正式版本。
  - `sha-xxxx`: 每个 Commit 对应的唯一标识。

### 部署方案
#### 方案 A：云端自动部署 (推荐用于 Web 控制台/Nexus)
1. **构建镜像**：GitHub Actions 构建 Docker 镜像。
2. **推送镜像**：推送到私有/公有镜像仓库。
3. **远程触发**：通过 SSH 或 Webhook 触发目标服务器执行 `docker-compose pull && docker-compose up -d`。

#### 方案 B：本地/私有化部署 (推荐用于 Worker 节点)
1. **分发二进制**：CI 编译出跨平台二进制文件 (Windows/Linux/ARM)。
2. **发布 Release**：自动创建 GitHub Release 并上传附件。
3. **手动/自动更新**：节点检测到新版本后拉取更新。

---

## 4. 环境规划

| 环境 | 目的 | 触发条件 | 部署方式 |
| :--- | :--- | :--- | :--- |
| **Development** | 开发者本地测试 | 手动运行 `scripts/test.ps1` | 本地运行 |
| **Staging** | 预发布测试，模拟真实环境 | 合并到 `master` 分支 | Docker Compose (自动) |
| **Production** | 正式运行环境 | 发布新的 Tag (v*.*.*) | Docker Swarm / K8s / 手动确认 |

---

## 5. 监控与告警

- **流水线监控**：GitHub Actions 失败后通过 Webhook 发送消息到 Bot (飞书/钉钉/微信)。
- **运行监控**：部署后通过 Prometheus + Grafana 监控容器健康状态和资源占用。

---

## 6. 本地化替代方案 (针对无公网环境)

如果无法使用 GitHub Actions：
1. **Drone CI / Gitea Actions**：在内网搭建轻量级 CI 平台。
2. **Jenkins**：传统的强大 CI 工具。
3. **Makefile + Git Hooks**：通过 `make deploy` 脚本手动触发本地构建与推送到服务器。
