#!/bin/bash

# Configuration
SERVER_IP="${1:-192.168.0.167}"
USERNAME="${2:-derlin}"
REMOTE_DIR="/opt/BotMatrix"
TEMP_ZIP="/tmp/botmatrix_deploy.zip"

echo "========================================"
echo "   Deploying to ${USERNAME}@${SERVER_IP}"
echo "========================================"

# 1. Pack
echo "[Step 1/3] Packing project..."
# Force use of python packer to ensure current working directory state is captured
# (git archive HEAD would miss uncommitted changes/moves)
python3 scripts/pack_project.py

# if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
#     git archive --format=zip --output=botmatrix_deploy.zip HEAD
# else
#     python3 scripts/pack_project.py
# fi

if [ ! -f "botmatrix_deploy.zip" ]; then
    echo "Error: botmatrix_deploy.zip not found!"
    exit 1
fi

# 2. Upload
echo "[Step 2/3] Uploading to server..."
scp botmatrix_deploy.zip ${USERNAME}@${SERVER_IP}:${TEMP_ZIP}

# 3. Deploy
echo "[Step 3/3] Executing remote commands..."
ssh -t ${USERNAME}@${SERVER_IP} "
    echo '--> Creating directory...'
    sudo mkdir -p ${REMOTE_DIR}
    
    echo '--> Unzipping...'
    sudo unzip -o ${TEMP_ZIP} -d ${REMOTE_DIR}
    sudo rm ${TEMP_ZIP}
    
    echo '--> Switching directory...'
    cd ${REMOTE_DIR}
    
    echo '--> Restarting services...'
    # If partial deploy logic were here, we'd add cleanup commands.
    # But this script seems to do a full down/up every time.
    sudo docker-compose down --remove-orphans
    # Add aggressive cleanup just in case down misses something (unlikely but safe)
    # sudo docker rm -f tencent-bot wxbot botmatrix-manager botmatrix-system-worker || true
    sudo docker-compose up -d --build
    
    echo '--> Deployment SUCCESS!'
"

echo "Done."
