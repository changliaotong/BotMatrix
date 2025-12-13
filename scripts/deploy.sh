#!/bin/bash

# Configuration
SERVER_IP="${1:-192.168.0.167}"
USERNAME="${2:-derlin}"
REMOTE_DIR="/opt/wxbot"
TEMP_ZIP="/tmp/wxbot_deploy.zip"

echo "========================================"
echo "   Deploying to ${USERNAME}@${SERVER_IP}"
echo "========================================"

# 1. Pack
echo "[Step 1/3] Packing project..."
# Force use of python packer to ensure current working directory state is captured
# (git archive HEAD would miss uncommitted changes/moves)
python3 scripts/pack_project.py

# if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
#     git archive --format=zip --output=wxbot_deploy.zip HEAD
# else
#     python3 scripts/pack_project.py
# fi

if [ ! -f "wxbot_deploy.zip" ]; then
    echo "Error: wxbot_deploy.zip not found!"
    exit 1
fi

# 2. Upload
echo "[Step 2/3] Uploading to server..."
scp wxbot_deploy.zip ${USERNAME}@${SERVER_IP}:${TEMP_ZIP}

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
    sudo docker-compose down --remove-orphans
    sudo docker-compose up -d --build
    
    echo '--> Deployment SUCCESS!'
"

echo "Done."
