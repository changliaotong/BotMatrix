import os
import zipfile
import fnmatch

def pack_project():
    output_filename = 'botmatrix_deploy.zip'
    
    # 包含的文件模式
    includes = [
        '*.py',
        'requirements.txt',
        'Dockerfile',
        'docker-compose*.yml',
        'config.json',
        'stats.json',
        'DEPLOY.md',
        'README.md',
        'SERVER_MANUAL.md',
        '*.sh',
        # Go / BotNexus files
        '*.go',
        '*.html',
        '*.js',
        '*.css',
        '*.json',
        '*.wasm',
        '*.png',
        '*.jpg',
        '*.jpeg',
        '*.svg',
        '*.ico',
        '*.ttf',
        '*.otf',
        '*.woff',
        '*.woff2',
        'go.mod',
        'go.sum',
        'VERSION'
    ]
    
    # 排除的目录和文件
    excludes = [
        'temp',
        '__pycache__',
        '.git',
        '.idea',
        'venv',
        'env',
        '*.zip',
        '*.pyc',
        '.DS_Store'
    ]
    
    # 获取项目根目录
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    output_path = os.path.join(root_dir, output_filename)
    
    with zipfile.ZipFile(output_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        print(f"Packing files to {output_path}...")
        
        for root, dirs, files in os.walk(root_dir):
            # 过滤目录
            dirs[:] = [d for d in dirs if d not in excludes]
            
            for file in files:
                file_path = os.path.join(root, file)
                rel_path = os.path.relpath(file_path, root_dir)
                
                # 检查是否应该包含
                should_include = False
                for pattern in includes:
                    if fnmatch.fnmatch(file, pattern):
                        should_include = True
                        break
                
                # 检查是否应该排除
                for pattern in excludes:
                    if fnmatch.fnmatch(file, pattern) or fnmatch.fnmatch(rel_path, pattern):
                        should_include = False
                        break
                
                if should_include:
                    print(f"  Adding: {rel_path}")
                    zipf.write(file_path, rel_path)
                    
    print(f"\nDone! Created {output_filename}")

if __name__ == "__main__":
    pack_project()
