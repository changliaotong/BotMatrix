# BotMatrix 国际化 (I18N) 开发指南

本指南旨在规范 BotMatrix WebUI 的国际化开发流程，防止硬编码字符进入代码库，并提供标准的问题排查方案。

## 1. 核心规范 (Core Rules)

### 1.1 零硬编码原则
- **严禁** 在 `.vue` 模板或 `.ts` 脚本中直接书写中/英/日文等可见文本。
- 所有文本必须使用 `t('key_name')` 函数获取。

### 1.2 翻译文件结构
- **单一源**：所有翻译存储在 `src/WebUI/src/utils/i18n.ts` 中。
- **必须同步**：新增一个 Key 时，必须同时在以下四个语种区块中添加：
  - `zh-CN` (简体中文)
  - `zh-TW` (繁体中文)
  - `en-US` (English)
  - `ja-JP` (日本語)

### 1.3 Key 命名规范
- 使用 **小写蛇形命名法 (lower_snake_case)**。
- 具有命名空间感：
  - 菜单类：`menu_dashboard`, `menu_settings`
  - 按钮类：`btn_save`, `btn_cancel`
  - 占位符：`placeholder_search`
  - 确认语：`confirm_delete_bot`

---

## 2. 开发流程 (Workflow)

1. **定义 Key**：在 `i18n.ts` 的四个 `translations` 对象中添加新 Key 及其翻译。
2. **在组件中使用**：
   ```vue
   <script setup>
   import { useSystemStore } from '@/stores/system';
   const systemStore = useSystemStore();
   const t = (key) => systemStore.t(key);
   </script>
   
   <template>
     <h1>{{ t('nexus_title') }}</h1>
   </template>
   ```

---

## 3. 问题排查方案 (Troubleshooting)

### 3.1 现象：界面显示 Key 名称（如 "nexus_desc"）
- **原因**：`i18n.ts` 中缺失该 Key，或当前语言区块下未定义。
- **排查**：
  - 检查 `i18n.ts` 是否包含该 Key。
  - 检查 Key 是否有拼写错误。
  - 确保 Key 在四个语种下都存在。

### 3.2 现象：界面显示占位符或空白
- **原因**：异步加载失败或 Key 对应的值为空字符串。
- **排查**：
  - 检查 `systemStore.t` 函数逻辑。
  - 检查浏览器控制台是否有 `404` 或翻译 API 调用失败。

### 3.3 自动化排查工具
运行项目根目录下的审计脚本（详见 `scripts/`）：
```bash
node scripts/audit-i18n.js
```
该脚本会：
1. 检查 `i18n.ts` 中四个语种的 Key 是否完全对齐（无遗漏）。
2. 检查所有 `.vue` 文件中使用的 `t('key')` 是否都在 `i18n.ts` 中定义。

---

## 4. 如何防止 AI 生成代码出错？

为了防止 AI（如 Copilot, Gemini, Cursor）在生成代码时引入硬编码或缺失 Key，请遵循以下策略：

1. **Prompt 注入**：在要求 AI 写 UI 代码前，先发送一段约束指令：
   > "请注意：本项目严格遵守国际化规范。所有 UI 文本必须使用 `t('key')`，且必须先在 `src/WebUI/src/utils/i18n.ts` 中定义对应的 zh-CN, zh-TW, en-US, ja-JP 翻译后再在组件中使用。严禁直接书写中文或英文。"

2. **上下文提供**：让 AI 读取 `i18n.ts` 的结构，这样它能自动复用已有的 Key 而不是胡乱发明新 Key。

3. **强制审计**：在 AI 完成代码后，立即运行 `audit-i18n.js` 脚本进行校验。AI 产生的代码如果违反规范，脚本会直接报错。
