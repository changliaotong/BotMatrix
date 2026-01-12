# i18n 国际化修复技术报告

本报告详细记录了针对 BotMatrix 项目中出现的 i18n 占位符失效、翻译异常以及脚本逻辑错误的修复过程。

## 1. 核心问题分析

### 1.1 占位符失效 (Nested tt Calls)
在 Vue 组件（如 `About.vue`, `NexusGuard.vue`）的 `computed` 属性或模板中，存在嵌套的 `tt()` 调用（例如 `tt('key', tt('default_key'))`）。由于 `tt` 函数在运行时无法正确解析嵌套的默认值，导致页面显示为占位符。

### 1.2 翻译异常与文本污染
- **占位符被翻译**：部分 i18n 键的默认值被错误地当作待翻译文本处理。
- **JSON 损坏**：语言文件（如 `ja-JP.ts`, `zh-TW.ts`）中出现了未闭合的字符串字面量和注入的 Vue 代码段，导致 Vite 构建失败。

### 1.3 脚本逻辑缺陷
- `apply.py` 中引用了不存在的 `load_findings()` 函数。
- 扫描逻辑未能有效区分 Vue 模板文本与脚本代码。

---

## 2. 修复方案与实施

### 2.1 自动化清理嵌套调用
创建并运行了 [fix_nested.py](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/scripts/i18n/fix_nested.py)，对全量 Vue 文件进行了正则替换。
- **修改前**：`tt('portal.about.title_prefix', tt('about.title_prefix'))`
- **修改后**：`tt('portal.about.title_prefix', 'about.title_prefix')`
涉及文件包括 `About.vue`, `NexusGuard.vue`, `Pricing.vue` 等 11 个核心视图文件。

### 2.2 语言文件修复与同步
- **清理损坏内容**：通过 `SearchReplace` 移除了各语言文件中注入的非法代码（如 `portal.portal*` 键值对）。
- **同步缺失条目**：修复了 [apply.py](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/scripts/i18n/apply.py) 的 `main` 函数，使其正确加载 `findings.json`。
- **补全关键文本**：为 `portal.about.title_suffix`（智联矩阵）等关键键位在各语言包中添加了正确的默认翻译。

### 2.3 脚本工具优化
- **[scan.py](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/scripts/i18n/scan.py)**：优化了正则匹配逻辑，支持 `KEY|DEFAULT` 格式的提取。
- **[apply.py](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/scripts/i18n/apply.py)**：增强了 `generate_mapping` 功能，确保新发现的键位能自动同步到 `mapping.json` 而不破坏现有翻译。

---

## 3. 修复结果验证

- **构建测试**：Vite 编译错误（Unterminated string literal）已解决。
- **页面表现**：
  - 关于页面的标题（`title_prefix`, `title_suffix`）现在能正确显示文本而非占位符。
  - 嵌套 `tt()` 调用已全部替换为直接默认值，消除了潜在的渲染异常。
- **数据一致性**：`zh-CN.ts`, `en-US.ts`, `ja-JP.ts`, `zh-TW.ts` 四个语言包的 JSON 结构均已校验通过。

---

## 4. 后续维护建议
- 在 Vue 组件中使用 `tt()` 时，**严禁嵌套调用**。
- 运行 `scan.py` 后，请务必检查 `findings.json` 是否包含异常的提取项，再执行 `apply.py`。
