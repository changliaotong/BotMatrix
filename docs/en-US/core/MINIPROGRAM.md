# BotMatrix Miniprogram

[简体中文](../zh-CN/MINIPROGRAM.md) | [Back to Home](../../README.md) | [Back to Docs Center](README.md)

The BotMatrix Miniprogram is a mobile management application designed to work with the Overmind backend service, providing robot management, system monitoring, real-time communication, and more.

## Features

### 🏠 Home
- System status overview
- Robot status statistics
- Real-time alert information
- Quick access entry

### 🤖 Robot Management
- Robot list display
- Real-time status monitoring
- Batch operation support
- Search and filtering features

### 📊 System Monitoring
- CPU, memory, and disk usage
- Network status monitoring
- Performance metrics display
- Historical data charts

### 📋 Log Management
- Real-time log viewing
- Log level filtering
- Keyword search
- Log export functionality

### ⚙️ System Settings
- System configuration management
- User permission settings
- Notification configuration
- Theme switching

## Technical Architecture

### Frontend Technology
- **Framework**: WeChat Miniprogram Native Development
- **Styling**: WXSS + CSS3
- **Data Management**: Native data binding
- **Communication**: WebSocket + HTTPS

### Backend Integration
- **API Service**: Overmind REST API
- **Real-time Communication**: WebSocket service
- **Data Format**: JSON
- **Authentication**: Token-based

## Project Structure

```
miniprogram/
├── app.js                 # Miniprogram entry
├── app.json              # Global configuration
├── app.wxss              # Global styles
├── project.config.json   # Project configuration
├── sitemap.json         # Sitemap configuration
├── pages/               # Page directory
│   ├── index/          # Home
│   ├── bots/           # Robot management
│   ├── bot-detail/     # Robot details
│   ├── monitoring/     # System monitoring
│   ├── logs/           # Log management
│   └── settings/       # System settings
├── components/         # Custom components
├── utils/              # Utility functions
│   ├── miniprogram_adapter.js  # Unified adapter
│   └── miniprogram_api.js      # API wrapper
└── images/             # Image resources
```

## Quick Start

### Requirements
- WeChat Developer Tools
- Miniprogram AppID
- Node.js environment (optional, for build tools)

### Installation Steps

1. **Clone the project**
```bash
git clone https://github.com/your-repo/botmatrix-miniprogram.git
```

2. **Import the project**
- Open WeChat Developer Tools
- Select "Import Project"
- Select the project root directory
- Fill in the AppID or select the test account

3. **Configure Backend Service**
- Modify `API_BASE_URL` in `utils/miniprogram_api.js`
- Configure the WebSocket connection address
- Set the authentication Token

4. **Run the Project**
- Click the "Compile" button
- Preview the miniprogram effect

## API Endpoints

### System Related
- `GET /api/system/status` - Get system status
- `GET /api/system/monitoring` - Get monitoring data
- `GET /api/system/performance` - Get performance data

### Robot Related
- `GET /api/bots` - Get robot list
- `GET /api/bots/:id` - Get robot details
- `POST /api/bots/:id/control` - Control robot
- `DELETE /api/bots/:id` - Delete robot

### Log Related
- `GET /api/logs` - Get log list
- `GET /api/logs/:id` - Get log details
- `POST /api/logs/export` - Export logs

### WebSocket Events
- `system_status` - System status update
- `bot_status_change` - Robot status change
- `system_alert` - System alert
- `system_metrics` - System metrics update

## Configuration

### app.json
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
        "text": "Home"
      }
      // ... other tab configurations
    ]
  }
}
```

### Network Configuration
Configure in `utils/miniprogram_api.js`:
```javascript
const API_BASE_URL = 'https://your-overmind-server.com';
const WS_BASE_URL = 'wss://your-overmind-server.com/ws';
```

## Development Guidelines

### Coding Style
- Use ES6+ syntax
- Follow miniprogram development guidelines
- Use async/await for asynchronous operations
- Use try/catch for error handling

### File Naming
- Page files: use lowercase and hyphens, e.g., `bot-detail.js`
- Component files: use lowercase and hyphens, e.g., `status-card.js`
- Utility files: use lowercase and underscores, e.g., `miniprogram_api.js`

### Styling
- Use WXSS syntax
- Use `rpx` units
- Follow BEM naming convention
- Support dark mode

## Feature Comparison

| Feature | Overmind Web | Miniprogram | Status |
|------|-------------|--------|------|
| System Monitoring | ✅ | ✅ | Done |
| Robot Management | ✅ | ✅ | Done |
| Real-time Communication | ✅ | ✅ | Done |
| Performance Monitoring | ✅ | ✅ | Done |
| Log Viewing | ✅ | ✅ | Done |
| System Settings | ✅ | ✅ | Done |
| Theme Switching | ✅ | ✅ | Done |
| Dark Mode | ✅ | ✅ | Done |
| Responsive Layout | ✅ | ✅ | Done |

## Changelog

### v1.1.69 (2025-12-18)
- ✅ Fixed API address configuration error
- ✅ Improved data visualization features
- ✅ Optimized WebSocket connection configuration
- ✅ Implemented system monitoring charts

### v1.0.0 (2024-01-01)
- Initial release
