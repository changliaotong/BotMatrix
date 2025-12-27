# BotMatrix Fission System

> [🌐 English](FISSION_SYSTEM.md) | [简体中文](../zh-CN/FISSION_SYSTEM.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

The Fission System is the core growth engine of BotMatrix, guiding users to spontaneously invite friends through incentive mechanisms to achieve viral growth of the robot.

## 1. Promotional Copy

The following copy can be used for robot automatic replies or promotional posters:

### 1.1 Invitation Pitch
> 🎁 **Invite friends, win big prizes!**
> Your exclusive invitation code has been generated: `【{CODE}】`
> 
> **How to participate:**
> 1. Send this invitation code to your friends.
> 2. Friends add the robot and send "`bind {CODE}`".
> 3. Both parties can receive point rewards. The more you invite, the richer the rewards!
> 
> 🔗 **Exclusive Link:** {LINK}
> (Click the link to directly view the invitation progress)

### 1.2 Task Reward Copy
> 🎯 **Today's Fission Tasks:**
> - **Group Entry Reward**: Join the official exchange group and get 50 points immediately!
> - **Activity Reward**: Use the robot more than 5 times daily for an extra 10 points.
> - **Promotion Expert**: Accumulate 10 friend invitations to unlock the "Promotion Ambassador" exclusive badge.

---

## 2. Technical Documentation

### 2.1 Architecture Design
The fission system adopts a "**Centralized Management + Marginal Execution**" architecture:
- **BotNexus (Forwarding Center)**: Responsible for global fission rule configuration, full data statistics, and the back-end management panel.
- **BotWorker (Plugin Side)**: Responsible for instruction interaction with users, task progress collection, and local reward distribution logic.
- **Service Layer**: Core business logic is implemented in `BotWorker/internal/fission`, ensuring plugins remain lightweight.

### 2.2 Robot Commands

| Command | Description |
| :--- | :--- |
| `invite` / `invite code` | Get personal exclusive invitation code and promotion link |
| `bind [code]` | New users bind to an inviter to receive entry rewards |
| `reward` / `progress` | View personal invitation count, point earnings, and level |
| `fission rank` | View Top 10 inviters across the server |
| `task` | View currently participating fission tasks and completion status |

### 2.3 Task System
The system has several built-in task types and supports dynamic extension in the back-end:
- **register**: User successfully binds an invitation code for the first time.
- **usage**: User interacts with the robot (e.g., chatting, using features).
- **group_join**: Listens for OneBot `group_increase` events to identify user group entry.

### 2.4 Anti-Fraud Mechanism
To ensure growth quality, the system integrates multiple validations:
- **Self-binding Restriction**: Users are strictly prohibited from binding their own invitation codes.
- **Daily Limit**: Limits the number of valid invitations a single user can obtain per day.
- **Device/IP Validation**: Based on the `CheckInvitationFraud` interface to identify fake accounts from the same device or IP.

### 2.5 Data Models
Core data tables:
- `fission_configs`: Stores global switches, reward values, and promotion templates.
- `user_fission_records`: Stores cumulative invitation volume, total points, and promotion level.
- `invitations`: Records every real invitation binding relationship.
- `fission_tasks`: Stores task definitions.
- `fission_reward_logs`: Complete reward distribution audit logs.

---

## 3. Development & Maintenance

- **Back-end API**: Located in `BotNexus/handlers_fission.go`, prefixed with `/api/admin/fission/`.
- **Plugin Entry**: `BotWorker/plugins/fission.go`.
- **Core Logic**: `BotWorker/internal/fission/service.go`.

---
*BotMatrix Fission System - Helping every bot reach the top*
