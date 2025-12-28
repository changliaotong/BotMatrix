# BotMatrix 自动化测试规划方案

[English](../../en-US/development/TESTING_PLAN.md) | [返回项目主页](../../../README.md) | [返回文档中心](../README.md)

本文档旨在为 BotMatrix 项目建立一套完善的自动化测试体系，以确保代码质量、提高开发效率并降低回归风险。

## 1. 测试层级结构

我们采用经典的测试金字塔模型，分为三个层级：

### 1.1 单元测试 (Unit Tests) - 基础层
*   **目标**：验证单个函数、方法或独立模块的逻辑正确性。
*   **范围**：
    *   `src/Common`: 工具函数（加密、令牌生成、路径处理）、配置解析、数据模型转换。
    *   `src/BotNexus`: 路由匹配算法、任务分发逻辑。
    *   `src/BotWorker`: 插件加载逻辑、消息过滤规则。
    *   各 Bot 实现：协议解析、事件转换逻辑。
*   **工具**：Go 标准库 `testing`, `testify/assert`, `golang/mock` (用于 Mock 接口)。
*   **要求**：核心逻辑单元测试覆盖率应达到 80% 以上。

### 1.2 集成测试 (Integration Tests) - 中间层
*   **目标**：验证多个组件协同工作时的交互逻辑。
*   **范围**：
    *   **Nexus-Worker 通信**：模拟 Worker 连接到 Nexus，发送心跳和接收任务。
    *   **数据库交互**：验证 `Manager` 通过 GORM 或原生 SQL 对 PostgreSQL 的 CRUD 操作。
    *   **Redis 交互**：验证幂等性检查、缓存同步等功能。
*   **工具**：`httptest` (模拟 WebSocket/HTTP 服务器), `testcontainers-go` (启动临时数据库容器)。
*   **策略**：使用专门的测试配置文件，避免对开发环境数据产生影响。

### 1.3 端到端测试 (E2E Tests) - 顶层
*   **目标**：验证完整业务流程从输入到输出的正确性。
*   **范围**：
    *   **完整消息流**：模拟外部 OneBot 消息 -> Bot 接收 -> Nexus 路由 -> Worker 处理 -> 结果返回。
    *   **WebUI 操作**：登录、查看统计、配置路由、管理容器。
*   **工具**：
    *   后端：自定义 Python/Go 脚本模拟 OneBot 客户端。
    *   前端：Playwright 或 Selenium (针对 Overmind WebUI)。
*   **环境**：使用 `docker-compose.test.yml` 启动整套环境。

## 2. 自动化流程集成 (CI/CD)

利用 GitHub Actions 实现自动化测试流水线：

1.  **静态检查 (Linting)**：
    *   使用 `golangci-lint` 进行代码规范和潜在错误检查。
2.  **单元测试自动化**：
    *   每次 Pull Request 或 Push 到 main 分支时自动运行所有单元测试。
3.  **集成测试自动化**：
    *   在特定阶段（如每日构建或发布前）运行耗时较长的集成测试。
4.  **覆盖率报告**：
    *   自动生成覆盖率报告并集成到 PR 评论中。

### GitHub Actions 示例配置 (`.github/workflows/test.yml`)

```yaml
name: Go Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: botmatrix_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Install dependencies
      run: go mod download

    - name: Run Linter
      uses: golangci/golangci-lint-action@v3

    - name: Run Unit Tests
      run: go test -v ./src/Common/...

    - name: Run Integration Tests
      env:
        DB_URL: postgres://postgres:password@localhost:5432/botmatrix_test?sslmode=disable
      run: go test -v ./src/BotWorker/...
```

## 3. 实施计划

| 阶段 | 任务内容 | 优先级 |
| :--- | :--- | :--- |
| **阶段 1** | 完善 `src/Common` 的单元测试，修复现有代码中的编译错误 | 高 |
| **阶段 2** | 为 `BotNexus` 和 `BotWorker` 的核心分发逻辑编写单元测试 | 高 |
| **阶段 3** | 搭建基于 Docker 的集成测试环境，测试数据库操作 | 中 |
| **阶段 4** | 编写端到端测试脚本，模拟典型机器人交互场景 | 低 |

## 4. 测试编写规范

1.  **文件名**：统一使用 `_test.go` 后缀。
2.  **包名**：单元测试通常位于被测试代码相同的包内。
3.  **独立性**：测试用例之间不应有依赖，能够独立运行。
4.  **Mock 使用**：对于网络、文件系统等外部依赖，优先使用 Mock 或 Stub。
