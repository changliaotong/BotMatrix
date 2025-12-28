# BotMatrix 容器化部署与插件管理最佳实践

本文档旨在为开发者和运维人员提供 BotWorker 及插件系统在容器化环境（Docker/K8s）下的最佳部署实践指南。

## 1. 核心理念：Stateless Worker + Stateful Plugins

为了实现高可用和弹性扩缩容，BotWorker 保持 **无状态 (Stateless)**，而插件和用户会话状态通过外部系统和持久化存储进行管理。

### 1.1 架构图示
```text
[ User ] -> [ Reverse Proxy (Nginx) ] 
                 |
        [ BotWorker Cluster (Replicas: N) ]
          /            |            \
 [ Redis Session ] [ Shared Storage ] [ BotNexus Center ]
    (State)         (Plugins)          (Control)
```

## 2. 插件管理最佳实践

在容器化环境中，插件升级的痛点在于“更新不重启”和“多实例同步”。

### 2.1 目录结构
推荐使用版本化目录结构，由 `PluginManager` 自动扫描：
```text
/app/plugins/
  ├── weather_plugin/
  │   ├── v1.0.0/ (manifest.json, main.py)
  │   └── v1.1.0/ (manifest.json, main.py)
  └── translator_plugin/
      └── v2.0.1/
```

### 2.2 灰度发布 (Canary Release)
利用 `PluginConfig` 中的 `canary_weight`：
1.  **上传新版本**: 将插件新版本放入 `/app/plugins/id/new_version/`。
2.  **设置权重**: 在 `plugin.json` 中设置 `canary_weight: 10`。
3.  **Core 自动路由**: `PluginManager` 会根据 `CorrelationId` (Session 粘滞) 将 10% 的流量导向新版本。
4.  **全量切换**: 验证无误后，将旧版本停用，新版本权重设为 0 (或移除旧版本目录)。

### 2.3 动态同步 (Market Sync)
容器启动时，可以通过环境变量或 API 触发插件同步：
-   **方法 A**: 在 `docker-compose` 中挂载共享卷 (NFS/Ceph/HostPath)。
-   **方法 B**: 调用 `SyncFromMarket(url)` API，让每个 Worker 自动下载并热加载插件。

## 3. 部署指南

### 3.1 Dockerfile 优化
-   **多阶段构建**: 减小镜像体积，提高启动速度。
-   **预装运行环境**: 镜像中应包含 Python, .NET Runtime 等插件依赖。
-   **健康检查**: 配置 `HEALTHCHECK` 确保故障实例能被自动重启。

### 3.2 Docker Compose 示例
参考项目根目录下的 [docker-compose.yml](../../docker-compose.yml)。关键点：
-   使用 `deploy.replicas` 实现水平扩展。
-   使用 `update_config.order: start-first` 实现零停机滚动更新。

## 4. 常见问题 (FAQ)

**Q: 插件升级需要重启 Worker 容器吗？**
A: 不需要。`PluginManager` 支持热加载和热更新。只需将新插件包解压到挂载的目录，并调用热更新接口即可。

**Q: 多个 Worker 实例如何保证插件一致？**
A: 
1.  **中心化分发**: 由 BotNexus 统一下发指令，Worker 接收到指令后执行 `SyncFromMarket`。
2.  **共享存储**: 所有 Worker 挂载同一个分布式文件系统卷。

**Q: 插件产生的临时文件放在哪里？**
A: 严禁放在插件目录下。应使用 `/tmp` 或配置专门的持久化 `data` 目录。

---
*Last Updated: 2025-12-28*
