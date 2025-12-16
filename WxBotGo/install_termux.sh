#!/bin/bash

echo "==================================="
echo "   WxBot Automatic Installer"
echo "==================================="

# 1. Request Storage Permission
echo "[*] Requesting storage access..."
termux-setup-storage
sleep 2

# 2. Find the file
SOURCE_FILE="/sdcard/Download/wxbot-android-arm64"
if [ ! -f "$SOURCE_FILE" ]; then
    echo "[!] Error: File not found at $SOURCE_FILE"
    echo "Please make sure you downloaded 'wxbot-android-arm64' to your Downloads folder."
    exit 1
fi

# 3. Copy and Permission
echo "[*] Installing..."
cp "$SOURCE_FILE" "$HOME/wxbot"
chmod +x "$HOME/wxbot"

# 4. Create Launch Script
echo "#!/bin/bash" > "$HOME/start-wxbot.sh"
echo "export MANAGER_URL='ws://192.168.0.167:3001'" >> "$HOME/start-wxbot.sh"
echo "$HOME/wxbot" >> "$HOME/start-wxbot.sh"
chmod +x "$HOME/start-wxbot.sh"

echo "==================================="
echo "   Success! "
echo "==================================="
echo "To start the bot, simply type:"
echo "  ./start-wxbot.sh"
echo ""
