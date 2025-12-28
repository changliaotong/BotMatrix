# SuperGroupManager Plugin Documentation

English | [简体中文](./super_group_manager.md)

## 🌟 Introduction
`SuperGroupManager` is an advanced group management plugin developed using the BotMatrix C# SDK (v3.0). It provides more than just basic moderation; it features a smart middleware interception system and an asynchronous interactive configuration panel, making it a powerful tool for automated group management.

## ✨ Core Features
- **Smart Middleware Interception**: Real-time message scanning to automatically detect and intercept sensitive words, executing message deletion and auto-mute penalties.
- **Welcome System**: Automatically recognizes new members and sends personalized welcome messages.
- **Interactive Config Panel**: Administrators can invoke an interactive menu via the `!config` command, allowing settings to be modified through conversational interaction.
- **Permission Management**: Strictly follows the BotMatrix permission system, ensuring only authorized administrators can perform sensitive actions.

## 🛠️ Installation & Usage

### Installation
1. Download the `SuperGroupManager.bmpk` package.
2. Load the plugin via BotMatrix Dashboard or `bm-cli`:
   ```bash
   ./bm-cli load ./SuperGroupManager.bmpk
   ```

### Command Reference
| Command | Description | Permission |
| :--- | :--- | :--- |
| `!config` | Opens the interactive configuration panel | Group Admin / Plugin Admin |

### Interactive Config Options
After invoking `!config`, you can choose from the following options:
1. **Enable/Disable Welcome Message**
2. **Edit Sensitive Word List**
3. **Set Auto-Mute Duration**
4. **Exit Configuration**

## 🛡️ Security
This plugin requires the following BotMatrix permissions:
- `mute_user`: For auto-muting violating users.
- `delete_message`: For deleting sensitive messages.
- `send_message`: For sending welcome messages and config feedback.

## 📝 Developer Info
- **Version**: 1.0.0
- **SDK Version**: BotMatrix C# SDK v3.0
- **Language**: C# (net6.0)
