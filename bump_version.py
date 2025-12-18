import os
import re

def bump_version(part='patch'):
    # root_dir is current directory since the script is in root
    root_dir = os.path.dirname(os.path.abspath(__file__))
    version_file = os.path.join(root_dir, 'VERSION')
    html_file = os.path.join(root_dir, 'BotNexus', 'index.html')

    # 1. Read current version
    if not os.path.exists(version_file):
        print(f"Error: {version_file} not found")
        return None
    
    with open(version_file, 'r') as f:
        current_version = f.read().strip()
    
    match = re.match(r'(\d+)\.(\d+)\.(\d+)', current_version)
    if not match:
        print(f"Error: Invalid version format {current_version}")
        return None
    
    major, minor, patch = map(int, match.groups())

    # 2. Increment version
    if part == 'major':
        major += 1
        minor = 0
        patch = 0
    elif part == 'minor':
        minor += 1
        patch = 0
    else: # patch
        patch += 1
    
    new_version = f"{major}.{minor}.{patch}"
    print(f"Bumping version: {current_version} -> {new_version}")

    # 3. Write new version
    with open(version_file, 'w') as f:
        f.write(new_version)
    
    # 4. Update index.html
    if os.path.exists(html_file):
        with open(html_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Replace vX.X.X with vNew.New.New
        new_content = re.sub(r'>v\d+\.\d+\.\d+<', f'>v{new_version}<', content)
        
        with open(html_file, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Updated {html_file}")
    else:
        print(f"Warning: {html_file} not found, skipping HTML update")

    return new_version

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--part', choices=['major', 'minor', 'patch'], default='patch')
    args = parser.parse_args()
    
    bump_version(args.part)
