#!/usr/bin/env python3
"""
Unified pack_project.py
Ensures all necessary files are properly included in deployment package
"""

import os
import zipfile
import json
import subprocess
import sys
import glob
from datetime import datetime

def pack_project():
    """Pack the entire project for deployment"""
    
    # Get project root directory (current directory)
    root_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Create deployment package
    deploy_zip = os.path.join(root_dir, "botmatrix_deploy.zip")
    
    print(f"Creating deployment package: {deploy_zip}")
    
    # Files and directories to include
    include_patterns = [
        # Core files
        "go.mod", "go.sum", "docker-compose.yml", "docker-compose.prod.yml",
        "README.md", "CHANGELOG.md", ".env.example", "VERSION",
        
        # Root scripts (new location)
        "*.py", "*.sh", "*.ps1",
        
        # BotNexus core
        "BotNexus/*.go", "BotNexus/Dockerfile*", "BotNexus/web/**",
        "BotNexus/overmind/**",
        
        # Bot directories (config files only, exclude node_modules)
        "DingTalkBot/config.sample.json", "DingTalkBot/Dockerfile",
        "DiscordBot/config.sample.json", "DiscordBot/Dockerfile", 
        "EmailBot/config.sample.json", "EmailBot/Dockerfile",
        "FeishuBot/config.sample.json", "FeishuBot/Dockerfile",
        "KookBot/config.sample.json", "KookBot/Dockerfile",
        "SlackBot/config.sample.json", "SlackBot/Dockerfile", 
        "TelegramBot/config.sample.json", "TelegramBot/Dockerfile",
        "TencentBot/config.sample.json", "TencentBot/Dockerfile",
        "WeComBot/config.sample.json", "WeComBot/Dockerfile",
        "WxBot/config.sample.json", "WxBot/Dockerfile",
        
        # Documentation
        "docs/**", "*.md",
        
        # Architecture docs
        "hybrid-architecture-strategy.md", "zaomiao-shop/**",
    ]
    
    # Files to exclude
    exclude_patterns = [
        "**/node_modules/**", "**/dist/**", "**/build/**",
        "**/.git/**", "**/__pycache__/**", "**/*.pyc",
        "**/logs/**", "**/tmp/**", "**/.DS_Store",
        "**/stats.json", "**/botmatrix.db", "**/botmatrix.db-journal",
        "**/*.log", "**/*.tmp", "botmatrix_deploy.zip"
    ]
    
    # Create ZIP file
    with zipfile.ZipFile(deploy_zip, 'w', zipfile.ZIP_DEFLATED) as zipf:
        # Add files matching patterns
        for pattern in include_patterns:
            if '**' in pattern:
                # Handle recursive patterns
                base_dir = pattern.split('/**')[0]
                full_base_dir = os.path.join(root_dir, base_dir)
                if os.path.exists(full_base_dir):
                    for root, dirs, files in os.walk(full_base_dir):
                        # Remove excluded directories
                        dirs[:] = [d for d in dirs if not any(excl in os.path.join(root, d) for excl in exclude_patterns)]
                        
                        for file in files:
                            file_path = os.path.join(root, file)
                            arc_path = os.path.relpath(file_path, root_dir)
                            
                            # Check exclusions
                            if not any(excl in arc_path for excl in exclude_patterns):
                                zipf.write(file_path, arc_path)
            else:
                # Handle single file or wildcard patterns
                full_pattern = os.path.join(root_dir, pattern)
                for file_path in glob.glob(full_pattern):
                    if os.path.isfile(file_path):
                        arc_path = os.path.relpath(file_path, root_dir)
                        
                        # Check exclusions
                        if not any(excl in arc_path for excl in exclude_patterns):
                            zipf.write(file_path, arc_path)
        
        # Add deployment metadata
        metadata = {
            "packaged_at": datetime.now().isoformat(),
            "files_count": len(zipf.namelist()),
            "total_size": sum(info.file_size for info in zipf.infolist())
        }
        
        zipf.writestr("deployment_metadata.json", json.dumps(metadata, indent=2))
    
    # Verify the package
    print(f"\n📦 Deployment package created: {deploy_zip}")
    print(f"   Size: {os.path.getsize(deploy_zip) / 1024 / 1024:.2f} MB")
    
    # List contents
    with zipfile.ZipFile(deploy_zip, 'r') as zipf:
        files = zipf.namelist()
        print(f"   Files: {len(files)}")
        
        # Check for key files
        key_files = ["docker-compose.yml", "BotNexus/Dockerfile", "BotNexus/main.go"]
        for key_file in key_files:
            if key_file in files:
                print(f"   ✅ {key_file}")
            else:
                print(f"   ❌ {key_file} - Missing!")
    
    print(f"\n✅ Package ready for deployment!")
    return deploy_zip

if __name__ == "__main__":
    pack_project()
