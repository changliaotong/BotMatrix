# BotMatrix System Architecture

> [🌐 English](ARCHITECTURE.md) | [简体中文](../zh-CN/ARCHITECTURE.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

BotMatrix is a robot matrix management system with a distributed and decoupled design. It achieves high concurrency and scalability by collaborating a central message distribution hub with multiple execution nodes.

## 🏗️ Core Components

### 1. BotNexus (Central Control Node)
BotNexus is the "brain" and "router" of the system.
- **Responsibilities**:
    - Maintain WebSocket connections with clients (e.g., WxBot, QQBot).
    - Receive raw message events (Events).
    - Determine message distribution based on **Routing Rules**.
    - Manage registration and heartbeats of Worker nodes.
    - Provide a Web Management Interface (WebUI).
- **Tech Stack**: Go, Gin, WebSocket, Redis (Pub/Sub).

### 2. BotWorker (Task Execution Node)
BotWorker is the "limbs" that handle actual business logic.
- **Responsibilities**:
    - Listen to Redis task queues.
    - Execute time-consuming tasks (e.g., AI text generation, image processing).
    - Run Plugins.
    - Return results to BotNexus or send directly.
- **Tech Stack**: Go, Python, .NET (Multi-language support).

### 3. Redis (Middleware)
Redis plays a crucial role as the core communication bus.
- **Responsibilities**:
    - **Message Distribution**: Real-time communication between Nexus and Worker using Pub/Sub.
    - **Task Queue**: Store asynchronous tasks waiting to be processed.
    - **State Storage**: Store bot online status, rate limiting policies, and dynamic configurations.
    - **Session Cache**: Maintain User Session Context.

### 4. PostgreSQL (Persistence Database)
- **Responsibilities**:
    - Store user data, routing rules, persistent configurations, and operation logs.
    - Store complex business logic data (e.g., Baby system, Marriage system data).

## 🔄 Message Flow

1.  **Receive**: External bot clients send messages to **BotNexus** via WebSocket.
2.  **Decision**: BotNexus filters via `CorePlugin` and matches target Workers based on `RoutingRules`.
3.  **Dispatch**: BotNexus publishes the message to a specific **Redis** channel.
4.  **Execution**: **BotWorker** subscribed to the channel receives the message and runs plugin logic.
5.  **Feedback**: After processing, BotWorker sends response commands back to BotNexus or calls API interfaces directly.

## 🎯 Message Routing & Skill System

### Routing Logic
BotNexus provides intelligent message routing supporting two modes:
1. **Exact Match**: Direct mapping for specific users (`user_123`), groups (`group_456`), or bots.
2. **Wildcard Match**: Supports patterns like `*_test` or `123*`.
3. **RTT-based Load Balancing**: Automatically chooses the optimal node based on average response time and health status.

### Skill System Compatibility
To ensure compatibility between different versions of Workers:
- **Feature Toggle**: Controlled via `ENABLE_SKILL=true` in `config.json` or environment variables.
- **Dynamic Discovery**: Workers report their "Skills" (capabilities) upon registration.
- **Skill-Aware Distribution**: BotNexus routes tasks only to Workers that have reported the required skill.
- **Graceful Downgrade**: If the skill system is disabled, the system falls back to traditional OneBot message forwarding.

## 📈 Scalability & High Availability

- **Horizontal Scaling**: Multiple BotWorker nodes can be started to share the load.
- **High Availability**: BotNexus supports cluster deployment (with a load balancer).
- **Pluginization**: Supports dynamic loading of plugins without downtime.
