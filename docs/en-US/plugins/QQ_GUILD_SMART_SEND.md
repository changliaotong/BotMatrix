# QQ Guild/Group Smart Collaboration Logic

> [🌐 English](QQ_GUILD_SMART_SEND.md) | [简体中文](../zh-CN/QQ_GUILD_SMART_SEND.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

This document details the special restriction handling solutions for Tencent QQ Guild Bots in group scenarios within the BotNexus system, namely the "Smart Send" and "WakeUp" mechanisms.

## 1. Background & Restrictions

Tencent's official QQ Guild Bots face strict security restrictions when accessing QQ groups:

1.  **Cannot Send Proactively**: Bots cannot speak proactively in groups like regular QQ users or legacy bots.
2.  **Reply-Only**: Bots must "reply" to an existing user message.
3.  **Message ID (MsgID) Validity**:
    *   The user's `message_id` must be included when replying.
    *   The `message_id` is only valid for **5 minutes**.
    *   After 5 minutes, or if the message has been replied to more than 5 times (in some scenarios), the ID expires.

**Pain Point**: If users haven't spoken in the group for a long time, or if the bot needs to push timed/proactive messages, the sending will fail due to the lack of a valid `message_id`.

## 2. Solution: Smart Collaboration (Smart Send)

To overcome these restrictions, we designed a mechanism based on **BotNexus session state management** and **multi-bot collaboration**.

### Core Principle

Use **other regular bots** in the group (e.g., regular QQ account bots using NapCat/OneBot11 protocol) as "auxiliary wakers."

1.  **Cache Valid Credentials**: The system continuously records the latest message ID for each group.
2.  **Check Validity**: Before sending, it checks if the latest message ID is within the 5-minute safety window.
3.  **Collaborative WakeUp**: If the ID is expired, a regular bot is instructed to send a message: `@GuildBot [WakeUp]`.
4.  **Capture New Credential**: When the Guild Bot receives the mentioned message, the system automatically updates the group's `LastMsgID`.
5.  **Retry Sending**: Complete the message sending using the freshly acquired `message_id`.

## 3. Detailed Logic Flow

### 3.1 State Management

The `ContactSession` struct in BotNexus maintains key information:
*   `LastMsgID`: The latest message ID received in the group.
*   `LastMsgTime`: The timestamp when the message was received.
*   `ActiveBots`: A list of other bots that have been active in the group (used to find an auxiliary waker).

The system updates these states whenever a new message is generated in the group (whether by a user or another bot).

### 3.2 Smart Send Flow

When calling the `/api/smart_action` interface to send a group message, the system performs the following checks:

1.  **Check Cache**:
    *   Retrieve the `LastMsgTime` for the target group.
    *   Calculate `Current Time - LastMsgTime`.

2.  **Branch A: Valid Credential (Time < 290s)**
    *   Use the cached `LastMsgID` as the `message_id` parameter directly.
    *   Call the underlying sending interface.
    *   **Result**: Message is sent immediately.

3.  **Branch B: Expired Credential (Time >= 290s)**
    *   **Find Waker**: Look for a bot in the `ActiveBots` list that is not the current sender (a regular QQ bot).
    *   **Execute WakeUp**:
        *   Send instruction to the waker bot: `send_group_msg(group_id, message="@GuildBot [WakeUp]")`.
    *   **Wait for Propagation**:
        *   The system pauses (default 2 seconds) to wait for Tencent servers to push this mention message to the Guild Bot.
        *   Guild Bot receives message -> BotNexus updates `LastMsgID`.
    *   **Reload Credential**:
        *   Reread the session state.
        *   Check if `LastMsgID` has been updated.
    *   **Final Send**:
        *   Send the original message using the (hopefully) updated `LastMsgID`.

## 4. How to Use

### Web Console

1.  Enter the **Actions** panel.
2.  Select the target **Guild Bot** and **Group**.
3.  Enter the message content.
4.  Click the blue button **"Smart Send (WakeUp)"**.

### Important Notes

*   **Prerequisite**: There must be at least one **regular QQ bot** (e.g., NapCat) under BotNexus management in the target group, and this bot must have been active in the group (sent a message or received an event).
*   **Latency**: Triggering the WakeUp mechanism results in a sending delay of approximately 2-3 seconds.
*   **Frequency**: While "infinite" sending is achieved, please follow Tencent's risk control rules to avoid high-frequency spamming.
