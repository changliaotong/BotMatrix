# BotMatrix WebUI Modernization Plan

> [🌐 English](WEBUI_UPGRADE_PLAN.md) | [简体中文](../zh-CN/WEBUI_UPGRADE_PLAN.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

To handle the continuous expansion of the project, the current WebUI architecture (traditional script inclusion + global variable pollution) has reached a maintenance bottleneck. The following is the proposed modernization plan.

## 1. Core Architecture: Vite + Vue 3 (SFC) + TypeScript

### Advantages
- **Modularity (SFC)**: Split the 2000-line `index.html` into single-file components (.vue) with clear responsibilities.
- **Type Safety**: Introduce TypeScript to catch potential API calls and state management errors during the compilation phase.
- **Performance Optimization**: Vite provides extremely fast HMR (Hot Module Replacement), and builds automatically perform tree-shaking and code compression.
- **Engineering**: Use ESLint and Prettier for consistent code style, and introduce unit testing (Vitest).

## 2. Directory Structure Planning

It is recommended to restructure `src/WebUI/web` into the following:

```text
src/WebUI/
├── src/
│   ├── api/            # Unified API request encapsulation (replaces modules/api.js)
│   ├── assets/         # Static assets (images, global styles)
│   ├── components/     # Common UI components
│   ├── composables/    # Composable functions (replaces logic in modules/)
│   ├── layouts/        # Page layout components
│   ├── locales/        # Multi-language configuration files
│   ├── stores/         # State management (Pinia recommended)
│   ├── views/          # Page-level components
│   ├── App.vue         # Root component
│   └── main.ts         # Entry file
├── index.html          # Template file
├── vite.config.ts      # Vite configuration
└── package.json        # Dependency management
```

## 3. State Management Migration

Currently, `matrix.js` heavily uses `window` global variables to store state.
- **Solution**: Introduce **Pinia**.
- Split logic for `auth`, `bots`, `stats`, etc., into independent Stores to ensure clear and traceable state flow.

## 4. Styling Solution

- **Current State**: A mix of Tailwind CDN and custom CSS.
- **Recommendation**: Formally introduce **Tailwind CSS** as a build plugin, remove CDN inclusion, and reduce initial bundle size.

## 5. Migration Path (Step-by-Step)

1.  **Environment Initialization**: Initialize the Vite + Vue 3 environment in `src/WebUI`.
2.  **Logic Extraction**: Migrate JS logic under `modules/` to Vue 3 Composables (`useBots`, `useStats`, etc.).
3.  **Component Refactoring**: Start with simple components (e.g., Header, Sidebar) and gradually replace templates in `index.html`.
4.  **Routing Introduction**: Introduce `vue-router` to manage different Tab pages, replacing the current `v-show` switching.

---
*This plan is suggested by a Senior Architect AI, aiming to improve the long-term maintainability of BotMatrix.*
