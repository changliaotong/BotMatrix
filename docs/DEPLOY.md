# Server Deployment Guide (Ubuntu/Linux)

This guide describes how to deploy the **BotMatrix** system (BotNexus + WxBot) to an Ubuntu server using Docker.

## 1. Prerequisites

Ensure your server has:
*   **Docker** installed (`curl -fsSL https://get.docker.com | bash`)
*   **Docker Compose** installed.
*   **Network Access**:
    *   Access to an external **Redis** server (optional but recommended for persistence).
    *   Ports `3111` (WebSocket) and `5000` (Web UI) open in the firewall.

## 2. File Upload

Upload the entire project directory to your server (e.g., `/opt/BotMatrix`).
Required files/directories:
*   `BotNexus/`
*   `WxBot/`
*   `docker-compose.yml`

## 3. Configuration

### Environment Variables
Check `docker-compose.yml` for default values. You can override them by creating a `.env` file or modifying `docker-compose.yml` directly.

Key variables:
*   `REDIS_ADDR`: Address of your Redis server (default: `192.168.0.126:6379`).
*   `REDIS_PWD`: Redis password.
*   `BOT_SELF_ID`: The QQ/Wx ID of your bot (default: `1098299491`).

### Port Mapping
By default:
*   **Web Dashboard**: `http://<server-ip>:5000`
*   **WebSocket Gateway**: `ws://<server-ip>:3111`

If these ports are occupied, modify the `ports` section in `docker-compose.yml`.

## 4. Deployment

Run the following command in the project root:

```bash
# Build and start services in background
sudo docker-compose up -d --build
```

To stop:
```bash
sudo docker-compose down
```

## 5. Verification & Login

1.  **Check Logs**:
    ```bash
    sudo docker-compose logs -f
    ```
    Ensure `bot-manager` starts successfully and `wxbot` connects to it.

2.  **Login**:
    *   Open `http://<server-ip>:5000` in your browser.
    *   You should see the dashboard.
    *   Check the logs or dashboard for the Login QR Code from `wxbot`.
    *   Scan the code with your WeChat mobile app.

3.  **Client Connection**:
    *   Connect your C# clients or other OneBot tools to `ws://<server-ip>:3111`.
    *   BotNexus will route messages between your clients and the WxBot.

## 6. Troubleshooting

*   **Connection Refused**: Check firewall settings (`ufw allow 5000`, `ufw allow 3111`).
*   **Redis Error**: Ensure the Redis address in `docker-compose.yml` is reachable from within the Docker container.
*   **WxBot Disconnected**: Check `docker-compose logs wxbot`. Ensure `MANAGER_URL` is correct (`ws://bot-manager:3001`).
