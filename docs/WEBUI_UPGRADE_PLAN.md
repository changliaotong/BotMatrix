# BotMatrix WebUI 现代化升级方案

为了应对项目规模的持续扩大，目前的 WebUI 架构（传统脚本引入 + 全局变量污染）已达到维护瓶颈。以下是建议的现代化升级方案。

## 1. 核心架构建议：Vite + Vue 3 (SFC) + TypeScript

### 优势
- **模块化 (SFC)**：将 2000 行的 `index.html` 拆分为职责明确的 `.vue` 单文件组件。
- **类型安全**：引入 TypeScript，在编译阶段发现潜在的 API 调用和状态管理错误。
- **性能优化**：Vite 提供极快的 HMR（热更新），且构建时会自动进行 Tree-shaking 和代码压缩。
- **工程化**：可以使用 ESLint 和 Prettier 统一代码风格，引入单元测试（Vitest）。

## 2. 目录结构规划

建议将 `src/WebUI/web` 重构为如下结构：

```text
src/WebUI/
├── src/
│   ├── api/            # 统一的 API 请求封装 (取代 modules/api.js)
│   ├── assets/         # 静态资源（图片、全局样式）
│   ├── components/     # 通用 UI 组件
│   ├── composables/    # 组合式函数 (取代 modules/ 里的逻辑)
│   ├── layouts/        # 页面布局组件
│   ├── locales/        # 多语言配置文件
│   ├── stores/         # 状态管理 (建议使用 Pinia)
│   ├── views/          # 页面级组件
│   ├── App.vue         # 根组件
│   └── main.ts         # 入口文件
├── index.html          # 模板文件
├── vite.config.ts      # Vite 配置
└── package.json        # 依赖管理
```

## 3. 状态管理迁移

目前的 `matrix.js` 大量使用 `window` 全局变量来存储状态。
- **方案**：引入 **Pinia**。
- 将 `auth`, `bots`, `stats` 等逻辑拆分为独立的 Store，确保状态流向清晰可追溯。

## 4. 样式方案

- **现状**：混合了 Tailwind CDN 和自定义 CSS。
- **建议**：正式引入 **Tailwind CSS** 作为构建插件，移除 CDN 引入，减少首屏加载体积。

## 5. 迁移路径 (分步走)

1. **环境初始化**：在 `src/WebUI` 初始化 Vite + Vue 3 环境。
2. **逻辑抽离**：将 `modules/` 下的 JS 逻辑迁移为 Vue 3 的 Composables (`useBots`, `useStats` 等)。
3. **组件重构**：从简单的组件（如 Header, Sidebar）开始，逐步替换 `index.html` 中的模板。
4. **路由引入**：引入 `vue-router` 管理不同的 Tab 页面，取代目前的 `v-show` 切换。

---
*该方案由资深架构师 AI 建议，旨在提升 BotMatrix 的长期可维护性。*
