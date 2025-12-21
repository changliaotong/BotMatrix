# BotMatrix 小程序

BotMatrix 小程序是一个与 Overmind 后端服务配套的移动端管理应用，提供机器人管理、系统监控、实时通信等功能。

## 功能特性

### 🏠 首页
- 系统状态概览
- 机器人状态统计
- 实时告警信息
- 快速操作入口

### 🤖 机器人管理
- 机器人列表展示
- 状态实时监控
- 批量操作支持
- 搜索和筛选功能

### 📊 系统监控
- CPU、内存、磁盘使用率
- 网络状态监控
- 性能指标展示
- 历史数据图表

### 📋 日志管理
- 实时日志查看
- 日志级别筛选
- 关键词搜索
- 日志导出功能

### ⚙️ 系统设置
- 系统配置管理
- 用户权限设置
- 通知配置
- 主题切换

## 技术架构

### 前端技术
- **框架**: 微信小程序原生开发
- **样式**: WXSS + CSS3
- **数据管理**: 小程序原生数据绑定
- **网络通信**: WebSocket + HTTPS

### 后端集成
- **API服务**: Overmind REST API
- **实时通信**: WebSocket 服务
- **数据格式**: JSON
- **认证方式**: Token 认证

## 项目结构

```
miniprogram/
├── app.js                 # 小程序入口文件
├── app.json              # 全局配置
├── app.wxss              # 全局样式
├── project.config.json   # 项目配置
├── sitemap.json         # 站点地图配置
├── pages/               # 页面目录
│   ├── index/          # 首页
│   ├── bots/           # 机器人管理
│   ├── bot-detail/     # 机器人详情
│   ├── monitoring/     # 系统监控
│   ├── logs/           # 日志管理
│   └── settings/       # 系统设置
├── components/         # 自定义组件
├── utils/              # 工具函数
│   ├── miniprogram_adapter.js  # 统一适配器
│   └── miniprogram_api.js      # API 封装
└── images/             # 图片资源
```

## 快速开始

### 环境要求
- 微信开发者工具
- 小程序 AppID
- Node.js 环境（可选，用于构建工具）

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/your-repo/botmatrix-miniprogram.git
```

2. **导入项目**
- 打开微信开发者工具
- 选择"导入项目"
- 选择项目根目录
- 填写 AppID 或选择测试号

3. **配置后端服务**
- 修改 `utils/miniprogram_api.js` 中的 API_BASE_URL
- 配置 WebSocket 连接地址
- 设置认证 Token

4. **运行项目**
- 点击"编译"按钮
- 预览小程序效果

## API 接口

### 系统相关
- `GET /api/system/status` - 获取系统状态
- `GET /api/system/monitoring` - 获取监控数据
- `GET /api/system/performance` - 获取性能数据

### 机器人相关
- `GET /api/bots` - 获取机器人列表
- `GET /api/bots/:id` - 获取机器人详情
- `POST /api/bots/:id/control` - 控制机器人
- `DELETE /api/bots/:id` - 删除机器人

### 日志相关
- `GET /api/logs` - 获取日志列表
- `GET /api/logs/:id` - 获取日志详情
- `POST /api/logs/export` - 导出日志

### WebSocket 事件
- `system_status` - 系统状态更新
- `bot_status_change` - 机器人状态变化
- `system_alert` - 系统告警
- `system_metrics` - 系统指标更新

## 配置说明

### app.json 配置
```json
{
  "pages": [
    "pages/index/index",
    "pages/bots/bots",
    "pages/bot-detail/bot-detail",
    "pages/monitoring/monitoring",
    "pages/logs/logs",
    "pages/settings/settings"
  ],
  "tabBar": {
    "list": [
      {
        "pagePath": "pages/index/index",
        "text": "首页"
      }
      // ... 其他 tab 配置
    ]
  }
}
```

### 网络配置
在 `utils/miniprogram_api.js` 中配置：
```javascript
const API_BASE_URL = 'https://your-overmind-server.com';
const WS_BASE_URL = 'wss://your-overmind-server.com/ws';
```

## 开发规范

### 代码风格
- 使用 ES6+ 语法
- 遵循小程序开发规范
- 统一使用 async/await 处理异步
- 错误处理使用 try/catch

### 文件命名
- 页面文件：使用小写和中划线，如 `bot-detail.js`
- 组件文件：使用小写和中划线，如 `status-card.js`
- 工具文件：使用小写和下划线，如 `miniprogram_api.js`

### 样式规范
- 使用 WXSS 语法
- 统一使用 rpx 单位
- 遵循 BEM 命名规范
- 支持深色模式

## 功能对比

| 功能 | Overmind Web | 小程序 | 状态 |
|------|-------------|--------|------|
| 系统状态监控 | ✅ | ✅ | 已完成 |
| 机器人管理 | ✅ | ✅ | 已完成 |
| 实时通信 | ✅ | ✅ | 已完成 |
| 系统监控 | ✅ | ✅ | 已完成 |
| 日志查看 | ✅ | ✅ | 已完成 |
| 系统设置 | ✅ | ✅ | 已完成 |
| 主题切换 | ✅ | ✅ | 已完成 |
| 深色模式 | ✅ | ✅ | 已完成 |
| 响应式布局 | ✅ | ✅ | 已完成 |

## 更新日志

### v1.1.69 (2025-12-18)
- ✅ 修复API地址配置错误
- ✅ 完善数据可视化功能
- ✅ 优化WebSocket连接配置
- ✅ 实现系统监控图表展示

### v1.0.0 (2024-01-01)
- ✨ 初始版本发布
- ✨ 实现首页功能
- ✨ 添加机器人管理
- ✨ 集成系统监控
- ✨ 支持 WebSocket 实时通信

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 注意事项

### 图片资源临时解决方案
目前项目中使用了 emoji 替代部分图片资源，主要包括：
- tabBar 图标
- 首页功能按钮图标
- 加载状态和错误提示图标
- 日志页面空状态图标

**后续优化建议**：替换为正式的 SVG 或 PNG 图标文件，提升用户体验。

## 支持

如遇到问题，请通过以下方式联系我们：
- 🐛 Issues：GitHub Issues

---

**BotMatrix 小程序** - 让机器人管理更简单 🚀