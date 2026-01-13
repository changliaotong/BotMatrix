# BotWorker - OneBot Protocol Compatible Bot Processor

[简体中文](../../zh-CN/components/BotWorker.md) | [Back to Home](../../../README.md) | [Back to Docs Center](../README.md)

BotWorker is a bot processor written in Go that is compatible with the OneBot protocol. It supports both WebSocket and HTTP communication and provides a flexible plugin system for easy functionality extension.

## Features

- ✅ OneBot v11 Protocol Support
- ✅ Dual WebSocket and HTTP Communication
- ✅ Flexible Plugin System
- ✅ Private and Group Message Handling
- ✅ Event Handling (Message, Notice, Request)
- ✅ Complete API Interface

### 🎯 Core Functionality

#### 🔍 Utilities
- **Weather Query** - Real-time weather information
- **Translation** - English-Chinese translation via Azure Translator API
- **Music** - Search and play songs
- **Time** - Current time display
- **Calculation** - Mathematical calculations
- **Manual** - Plugin usage instructions
- **System Info** - Server hardware, software, and performance stats

#### 🏆 Achievement System
- **Management** - Unlock achievements, track progress
- **List** - View all available achievements
- **My Achievements** - View earned achievements
- **Leaderboard** - Achievement rankings

#### 🎮 Entertainment
- **Sign-in** - Daily check-ins for points
- **Lottery** - Random lucky draws
- **Card Games** - Three Cards, Showdown
- **Rock Paper Scissors**
- **Dice Games** - Guess Big/Small
- **Divination** - Traditional fortune sticks
- **Daily Fortune**
- **Idiom Solitaire**
- **Jokes & Ghost Stories**

#### 🐾 Pet System
- **Adoption** - Adopt cute pets
- **Care** - Feeding, playing, cleaning
- **Leveling** - Upgrade your pet
- **Leaderboard** - Pet rankings

#### 🐎 Mount System
- **Shop** - Purchase various mounts
- **Inventory** - View owned mounts
- **Equipment** - Equip your mount
- **Upgrades** - Enhance mount attributes
- **Leaderboard** - Mount rankings

#### 💰 Points System
- **Management** - Gain, spend, and query points
- **Tipping** - Give points to other users
- **Banking** - Deposit and withdraw points
- **Leaderboard** - Points rankings
- **Trading** - Buy and sell points
- **Computing Power** - System for gaining and using compute points

#### 👥 Social Interaction
- **Greetings** - Good morning/night messages
- **Group Owner Interaction**
- **Titles** - Custom user titles
- **Transformation** - Character transformation
- **Welcome** - New member greetings

#### 🛡️ Moderation System
- **Message Control** - Recall messages
- **Member Control** - Mute, kick, ban
- **Lists** - Blacklist, whitelist, graylist management
- **Filters** - Sensitive words, ads, images, URLs
- **Group Config** - Per-group moderation settings
- **Automatic Actions** - Auto-ban on kick or leave
- **Notifications** - Group alerts for kicks and departures

#### 🧠 Intelligent Features
- **Auto Sign-in** - Automatic check-in on speaking
- **Activity Stats** - Group chatter statistics
- **Ultimate Agent** - Intelligent conversation
- **Tutorial** - Bot usage teaching
- **Group Info** - Query group details
- **Voice Reply** - AI voice messages (per-group toggle)
- **Self-Destruct** - Auto-recall replies for privacy (per-group toggle)
- **Multi-step Dialog** - Information collection and configuration flows

## Project Structure

```
BotWorker/
├── cmd/
│   └── main.go              # Main entry
├── internal/
│   ├── onebot/              # OneBot protocol definitions
│   ├── plugin/              # Plugin system
│   ├── server/              # Server implementations
│   ├── config/              # Configuration management
│   └── utils/               # Utilities
├── plugins/                 # Plugin implementations
├── go.mod                   # Go module definition
└── README.md                # Project README
```

## Quick Start

### Requirements
- Go 1.20 or higher

### Installation
```bash
go mod tidy
```

### Configuration
Copy `configs/config.yaml` and modify settings for translation, weather, and music APIs.

### Running
```bash
go run cmd/main.go
```
The program will start:
- WebSocket Server: `ws://localhost:8080/ws`
- HTTP Server: `http://localhost:8081`
