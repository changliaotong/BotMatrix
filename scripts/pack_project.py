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
import fnmatch
from datetime import datetime

def pack_project():
    """Pack the entire project for deployment"""
    
    # Check for no-source flag
    no_source = "--no-source" in sys.argv
    if no_source:
        print("Mode: No-Source (Excluding Go source files)")

    # Get project root directory (current directory)
    root_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Create deployment package
    deploy_zip = os.path.join(root_dir, "botmatrix_deploy.zip")
    
    print(f"Creating deployment package: {deploy_zip}")
    
    # Files and directories to include
    include_patterns = [
        # Core files
        "src/BotNexus/go.mod", "src/BotNexus/go.sum", "docker-compose.yml", "docker-compose.prod.yml",
        "README.md", "CHANGELOG.md", ".env.example", "VERSION",
        
        # Root scripts
        "*.py", "*.sh", "*.ps1",
        
        # BotNexus core
        "src/BotNexus/**",
        
        # Bot directories (Include all source files as they are built in Docker)
        "src/BotWorker/**",
        "src/DingTalkBot/**",
        "src/DiscordBot/**",
        "src/EmailBot/**",
        "src/FeishuBot/**",
        "src/KookBot/**",
        "src/SlackBot/**",
        "src/TelegramBot/**",
        "src/TencentBot/**",
        "src/WeComBot/**",
        "src/WxBot/**",
        "src/WxBotGo/**",
        
        # Common library
        "src/Common/**",
        
        # Documentation
        "docs/**", "*.md",
        
        # Architecture docs
        "hybrid-architecture-strategy.md", "zaomiao-shop/**",
    ]
    
    # Files to exclude (using fnmatch patterns)
    exclude_patterns = [
        "*/node_modules/*", "*/dist/*", "*/build/*",
        "*/.git/*", "*/__pycache__/*", "*.pyc",
        "*/logs/*", "*/tmp/*", "*.DS_Store",
        "*/stats.json", "*/botmatrix.db", "*/botmatrix.db-journal",
        "*.log", "*.tmp", "botmatrix_deploy.zip",
        
        # Configuration and data files (should not be overwritten on server)
        "config.json", "config.yaml", "storage.json", "token.json",
        "*/data/*.db", "*/data/*.json"
    ]

    if no_source:
        exclude_patterns.extend([
            "*.go",
            "go.mod",
            "go.sum",
            "src/Common/*"
        ])
    
    def is_excluded(path):
        """Check if a path matches any exclusion pattern"""
        path = path.replace('\\', '/')
        for pattern in exclude_patterns:
            if fnmatch.fnmatch(path, pattern) or \
               fnmatch.fnmatch(os.path.basename(path), pattern) or \
               any(fnmatch.fnmatch(part, pattern) for part in path.split('/')):
                return True
        return False

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
                        # Filter directories
                        dirs[:] = [d for d in dirs if not is_excluded(os.path.join(root, d))]
                        
                        for file in files:
                            file_path = os.path.join(root, file)
                            arc_path = os.path.relpath(file_path, root_dir)
                            
                            # Check exclusions
                            if not is_excluded(arc_path):
                                zipf.write(file_path, arc_path)
            else:
                # Handle single file or wildcard patterns
                full_pattern = os.path.join(root_dir, pattern)
                for file_path in glob.glob(full_pattern):
                    if os.path.isfile(file_path):
                        arc_path = os.path.relpath(file_path, root_dir)
                        
                        # Check exclusions
                        if not is_excluded(arc_path):
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
        key_files = ["docker-compose.yml", "src/BotNexus/Dockerfile", "src/BotNexus/Dockerfile.prod"]
        if not no_source:
            key_files.extend(["src/BotNexus/main.go", "src/TencentBot/main.go"])
        else:
            # In no-source mode, check for binaries instead
            key_files.extend(["src/BotNexus/BotNexus", "src/TencentBot/TencentBot", "src/BotWorker/bot-worker"])

        for key_file in key_files:
            if key_file in files:
                print(f"   ✅ {key_file}")
            else:
                print(f"   ❌ {key_file} - Missing!")
    
    print(f"\n✅ Package ready for deployment!")
    return deploy_zip

if __name__ == "__main__":
    pack_project()
