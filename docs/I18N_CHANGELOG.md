# Internationalization (i18n) Update Changelog - 2026-01-13

## Overview
This update focuses on synchronizing and completing the localization for the **Portal Pricing** page and the **Nexus Guard** product page across all supported languages (Simplified Chinese, Traditional Chinese, Japanese, and English).

## Changes by Page

### 0. Global Contact Information
- **Change**: Updated WeChat contact ID from `Matrix_Revolution_AI` and `BotMatrix_Global` to `Kuang-HuiPeng` across all pages and locale files.

### 1. Portal Pricing Page
- **Issue**: Missing price tags and inconsistent key structures across locales.
- **Fixes**:
  - Synchronized `price_free`, `price_custom`, and `price_tag` keys.
  - Aligned "FREE" and "CUSTOM" tags for Japanese and English.
  - Ensured `zh-CN`, `zh-TW`, `ja-JP`, and `en-US` follow the same key pairing pattern to prevent UI rendering issues.

### 2. EarlyMeow Navigation & Tags
- **Issue**: Missing Chinese translations for navigation items and the "Tech Driven" (技术驱动) tag.
- **Fixes**:
  - Added 11 missing `earlymeow.nav.*` keys to `zh-CN.ts`.
  - Verified the implementation of the `earlymeow.tag.tech_driven` key in `Tech.vue`.
  - Confirmed sub-navigation in `earlymeow/Layout.vue` uses the correct i18n keys.

### 3. Nexus Guard Page (`/bots/nexus-guard`)
- **Issue**: Incomplete translations and placeholder prefixes (`guard.`) in Traditional Chinese and Japanese locales.
- **Fixes**:
  - **zh-TW (Traditional Chinese)**: 
    - Converted all 64 `nexus_guard.*` keys from Simplified to Traditional Chinese.
    - Adjusted terminology (e.g., 記憶體, 非同步).
    - Removed all `guard.` prefixes.
  - **ja-JP (Japanese)**:
    - Fully translated all 64 `nexus_guard.*` keys into native Japanese.
    - Removed all `guard.` prefixes.
  - **zh-CN (Simplified Chinese)**: Verified as the source of truth for all 64 keys.

## Technical Details
- **Component**: [NexusGuard.vue](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/src/views/portal/bots/NexusGuard.vue)
- **Locale Files**:
  - [zh-CN.ts](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/src/locales/zh-CN.ts)
  - [zh-TW.ts](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/src/locales/zh-TW.ts)
  - [ja-JP.ts](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/src/locales/ja-JP.ts)
  - [en-US.ts](file:///c:/Users/彭光辉/projects/BotMatrix/src/WebUI/src/locales/en-US.ts)

## Verification
- Verified all `tt()` calls in `NexusGuard.vue` match the keys in the locale files.
- Confirmed total key count (64) is consistent across `zh-CN`, `zh-TW`, and `ja-JP`.
- Manual check of terminology consistency across all locales.
