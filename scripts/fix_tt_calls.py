import re
import os

def fix_tt_calls(file_path):
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Matches tt('key', tt('other_key')) or tt('key', 'fallback')
    # The regex looks for tt('key', followed by a comma, optional spaces, and either tt(...) or a string
    new_content = re.sub(r"tt\((['\"][^'\"]+['\"])\s*,\s*(?:['\"][^'\"]+['\"]|tt\([^)]+\))\)", r"tt(\1)", content)
    
    if new_content != content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Fixed tt() calls in {file_path}")
    else:
        print(f"No redundant tt() calls found in {file_path}")

if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1:
        fix_tt_calls(sys.argv[1])
