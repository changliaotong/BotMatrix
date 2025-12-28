# BotMatrix 插件市场规范 (BMPK Standard)

为了实现“伟大”的插件生态，我们需要一套标准化的分发与安装流程。本规范定义了 BotMatrix 插件包 (**BMPK**) 的格式与市场接入标准。

## 1. 插件包格式 (.bmpk)

`.bmpk` 文件本质上是一个经过签名和加密（可选）的 ZIP 压缩包，包含以下核心文件：

- `plugin.json`: 插件元数据与权限声明（必须）。
- `icon.png`: 插件图标（建议 256x256）。
- `README.md`: 插件详细说明与配置指南。
- `bin/` 或 `src/`: 插件执行主体。
- `scripts/`: 安装/卸载/更新钩子脚本。

## 2. 插件包元数据 (plugin.json) 扩展

除了基础字段外，市场插件还需包含：

```json
{
  "id": "com.botmatrix.market.weather",
  "category": "Tools",
  "tags": ["weather", "realtime", "api"],
  "price": 0.0,
  "support_url": "https://github.com/example/weather/issues",
  "dependencies": {
    "python": ">=3.8",
    "botmatrix_sdk": ">=2.1.0"
  }
}
```

## 3. 分发与安装流程

### A. 开发者侧
1. 使用 `bm-cli pack` 命令将代码打包成 `.bmpk`。
2. 在开发者控制台上传插件包，并进行版本签名。
3. 提交审核（安全扫描、权限合规性检查）。

### B. 用户侧 (Core/WebUI)
1. 在插件市场预览、搜索插件。
2. 点击“安装”，WebUI 通知 BotNexus。
3. BotNexus 执行：
   - 下载 `.bmpk`。
   - 校验签名与完整性。
   - 解析 `plugin.json` 并动态创建沙箱环境。
   - 运行安装钩子脚本。

## 4. 安全策略

- **权限动态申请**：用户安装时需手动确认插件请求的权限（如：读取群成员、发送图片）。
- **资源限制 (Quotas)**：市场插件默认受到 CPU、内存、网络连接数限制。
- **自动更新**：支持热更新，无缝切换插件版本而不丢失上下文。

## 5. 收益模型

- **免费插件**：开源或引流性质。
- **付费插件**：一次性买断。
- **订阅插件**：按月/年计费。
- **API 计费**：根据插件调用的外部 API 次数计费（由平台代扣）。
