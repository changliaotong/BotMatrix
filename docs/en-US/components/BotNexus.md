# BotNexus System Documentation

[简体中文](../../zh-CN/components/BotNexus.md) | [Back to Home](../../../README.md) | [Back to Docs Center](../README.md)

BotNexus is a unified bot matrix management system featuring a Go backend and a modern Web frontend. It supports large-scale bot topology visualization, real-time monitoring, Docker container management, and intelligent message routing.

## 🏗️ System Architecture

BotNexus serves as a central hub, connecting and managing multiple bot instances (Bots) and processing nodes (Workers).

### Key Features
- **3D Topology Visualization (Matrix 3D)**: Real-time universe topology based on Three.js. Supports node clustering, real-time message particle effects, automatic connection optimization, and **Full Path Routing Trace (User-Group-Nexus-Worker)**.
- **Intelligent Routing**: Dynamic routing algorithm with RTT awareness. Supports precise ID and wildcard (*) matching, rule persistence, and automatic failover for offline nodes. Features **Intelligent Cache Enhancement** to auto-complete message metadata.
- **Internationalization (i18n)**: Full support for Chinese/English interface switching.
- **System Log Management**: Real-time streaming log display with keyword filtering, one-click clear, and log history export.
- **Core Plugins**: Security interceptors integrated at the message routing layer. Supports global toggles, black/white lists, sensitive word filtering, URL filtering, and admin command control.
- **User Management**: Comprehensive RBAC permissions model. Supports `session_version` for forced token invalidation.
- **Data Persistence**: Core caches (contacts/stats/config) are persisted in **PostgreSQL**, ensuring sub-second synchronization after service restarts.

### Technology Stack
- **Backend**: Go 1.20+, PostgreSQL, Redis (Cluster), JWT (Auth), Docker SDK
- **Frontend**: Vue 3 (Composition API), Three.js (3D), Tailwind CSS, Lucide Icons
- **Mobile**: Flutter (Overmind)

## 🚀 Quick Start

### Requirements
- **Docker**: Must be installed and running for container management features.
- **Go**: 1.19+ (for local compilation).
- **PostgreSQL**: Core business database.
- **Redis**: For message queuing and state synchronization.

### Startup Steps
1. **Clone the code**:
   ```bash
   git clone <repository_url>
   cd BotNexus
   ```
2. **Run the service**:
   ```bash
   go run .
   ```
3. **Access the dashboard**:
   - URL: `http://localhost:5000`
   - Default Username: `admin`
   - Default Password: `admin123`

## 📡 API Overview

### Authentication
- `POST /api/login` - Get JWT Token
- `GET /api/me` - Get user profile

### Docker Management
- `GET /api/docker/list` - Get container list
- `POST /api/docker/action` - Execute container action (start/stop/restart/delete)
- `POST /api/docker/add-bot` - Deploy a bot
- `POST /api/docker/add-worker` - Deploy a worker node

### User Management
- `GET /api/admin/users` - Get all users
- `POST /api/admin/users` - Manage users (create/delete/reset_password/toggle_active)

## 🎯 Core Logic

### 3D Optimization (Performance)
- **Material Caching**: Reuses GPU textures and materials to reduce memory footprint.
- **Light Source Limiting**: Dynamically limits real-time point lights to maintain 60 FPS during message spikes.
- **Auto-Sync**: WebSocket `sync_state` ensures frontend node states are strictly consistent with the backend.

### Security Model
- **JWT Middleware**: Global API permission validation.
- **Admin Middleware**: Secondary verification for core management operations.
- **Password Hashing**: High-strength encryption using bcrypt.

## 🤝 Contribution & Feedback
Please submit suggestions or report bugs via GitHub Issues.

---
*BotNexus - Powering your bot matrix with elegance.*
