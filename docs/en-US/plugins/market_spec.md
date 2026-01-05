# BotMatrix Plugin Market Specification (BMPK Standard)

> [简体中文](../../zh-CN/plugins/market_spec.md) | [🌐 English](market_spec.md)
> [⬅️ Back to Docs Center](../README.md) | [🏠 Back to Project Home](../../../README.md)

To achieve a "great" plugin ecosystem, we need a standardized distribution and installation process. This specification defines the format of the BotMatrix Plugin Package (**BMPK**) and market entry standards.

---

## 1. Plugin Package Format (.bmpk)

A `.bmpk` file is essentially a ZIP archive that has been signed and optionally encrypted, containing the following core files:

- `plugin.json`: Plugin metadata and permission declarations (Required).
- `icon.png`: Plugin icon (Recommended 256x256).
- `README.md`: Detailed plugin description and configuration guide.
- `bin/` or `src/`: The main executable body of the plugin.
- `scripts/`: Hooks for installation, uninstallation, and updates.

---

## 2. Plugin Metadata (plugin.json) Extensions

In addition to the base fields, market plugins must include:

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

---

## 3. Distribution & Installation Flow

### A. Developer Side
1. Use the `bm-cli pack` command to package the code into a `.bmpk`.
2. Upload the package to the developer console and perform version signing.
3. Submit for review (security scanning, permission compliance check).

### B. User Side (Core/WebUI)
1. Preview and search for plugins in the Plugin Market.
2. Click "Install", and the WebUI notifies BotNexus.
3. BotNexus performs the following:
   - Downloads the `.bmpk`.
   - Verifies the signature and integrity.
   - Parses `plugin.json` and dynamically creates a sandbox environment.
   - Runs the installation hook scripts.

---

## 4. Security Policy

- **Dynamic Permission Request**: Users must manually confirm the permissions requested by the plugin (e.g., reading group members, sending images) during installation.
- **Resource Quotas**: Market plugins are subject to CPU, memory, and network connection limits by default.
- **Automatic Updates**: Support for hot updates, seamlessly switching plugin versions without losing context.

---

## 5. Revenue Model

- **Free Plugins**: Open source or for traffic generation.
- **Paid Plugins**: One-time purchase.
- **Subscription Plugins**: Monthly/yearly billing.
- **API Billing**: Billed based on the number of external API calls made by the plugin (deducted by the platform).

---
*Last Updated: 2025-12-28*
