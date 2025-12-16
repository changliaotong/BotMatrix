# Overmind

The central control unit for BotMatrix.

## Setup

Since the project structure was generated manually, please run the following command in this directory to generate platform-specific files (Android/iOS):

```bash
flutter create .
```

## Features

- **Nexus Dashboard**: View all connected bots and their status.
- **Log Console**: Real-time log streaming from BotNexus.
- **Sci-Fi UI**: Dark mode interface with "Overmind" aesthetic.

## Connection

By default, the app attempts to connect to:
- `ws://10.0.2.2:3005` (Android Emulator loopback to host)
- `ws://localhost:3005` (Web/Desktop)

Ensure BotNexus is running and port 3005 is exposed.
