import os
import re
import json

# Configuration
SEARCH_DIRS = ['src/views', 'src/components', 'src/router', 'src/stores', 'src/utils', 'src/api']
EXCLUDE_FILES = ['i18n.ts', 'zh-CN.ts', 'en-US.ts', 'ja-JP.ts', 'zh-TW.ts', 'index.ts']
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
OUTPUT_FILE = os.path.join(os.path.dirname(__file__), 'findings.json')

# Regex to match Chinese characters (including punctuation)
CHINESE_PATTERN = re.compile(r'[\u4e00-\u9fa5]+[，。？！；：、“”‘’（）\u4e00-\u9fa5]*')

def scan_files():
    results = {}
    print(f"Scanning project root: {PROJECT_ROOT}")
    
    for search_dir in SEARCH_DIRS:
        full_path = os.path.join(PROJECT_ROOT, search_dir)
        if not os.path.exists(full_path):
            print(f"Directory not found: {full_path}")
            continue
            
        for root, dirs, files in os.walk(full_path):
            for file in files:
                if file in EXCLUDE_FILES:
                    continue
                if not (file.endswith('.vue') or file.endswith('.ts') or file.endswith('.js')):
                    continue
                    
                file_path = os.path.join(root, file)
                rel_path = os.path.relpath(file_path, PROJECT_ROOT)
                
                with open(file_path, 'r', encoding='utf-8') as f:
                    try:
                        content = f.read()
                    except UnicodeDecodeError:
                        continue
                        
                    # Remove comments
                    content_no_comments = re.sub(r'<!--.*?-->', '', content, flags=re.DOTALL)
                    content_no_comments = re.sub(r'//.*?\n', '\n', content_no_comments)
                    content_no_comments = re.sub(r'/\*.*?\*/', '', content_no_comments, flags=re.DOTALL)
                    
                    # Find all Chinese strings
                    # 1. Vue text nodes: >...中文...<
                    # Improved to avoid matching across tags and script blocks
                    if '.vue' in file:
                        template_match = re.search(r'<template>(.*)</template>', content_no_comments, re.DOTALL)
                        if template_match:
                            template_content = template_match.group(1)
                            vue_text_nodes_raw = re.findall(r'>([^<]*?[\u4e00-\u9fa5]+[^<]*?)<', template_content)
                        else:
                            vue_text_nodes_raw = []
                    else:
                        vue_text_nodes_raw = []
                    
                    # 2. JS/TS Strings: '...中文...', "...中文...", `...中文...`
                    # Improved to avoid capturing entire code blocks
                    js_strings_raw = re.findall(r"(['\"`][^'\"`\n]{1,100}?[\u4e00-\u9fa5]+[^'\"`\n]{0,100}?['\"`])", content_no_comments)
                    # Clean JS strings (remove quotes)
                    js_strings = []
                    for s in js_strings_raw:
                        cleaned = s[1:-1]
                        if cleaned and len(cleaned) < 200: # Sanity check on length
                            js_strings.append(cleaned)

                    # 3. Attributes: attr="...中文..."
                    attr_strings_raw = re.findall(r'\b[a-z-]+=(["\'])([^"\']{1,100}?[\u4e00-\u9fa5]+[^"\']{0,100}?)\1', content_no_comments)
                    attr_strings = [m[1] for m in attr_strings_raw]

                    # 4. Existing i18n keys: t('...'), tt('...')
                    # Try to capture both key and optional default value
                    i18n_calls = re.findall(r"(?:t|tt)\(\s*['\"]([^'\"]+?)['\"](?:\s*,\s*(?:['\"]([^'\"]*?)['\"]|tt\(\s*['\"]([^'\"]+?)['\"]\s*\)))?\s*\)", content_no_comments)
                    
                    # We'll store keys as they are, and if they have a default value, we'll store that too
                    # To keep findings.json simple, we'll use a special format for key+default: "KEY|DEFAULT"
                    i18n_matches = []
                    for key, default1, default2 in i18n_calls:
                        # Use a sentinel to detect if default was actually provided, even if empty
                        default = default1 if default1 is not None else default2
                        if default is not None:
                            i18n_matches.append(f"{key}|{default}")
                        else:
                            i18n_matches.append(key)
                    
                    all_matches_raw = vue_text_nodes_raw + js_strings + attr_strings + i18n_matches
                    
                    # Filter and clean
                    all_matches = []
                    for m in all_matches_raw:
                        cleaned = m.strip()
                        # If it's a key (starts with portal., common., etc.), we check if it's missing later
                        # For now, we collect all potential strings and keys
                        if cleaned:
                            # If it has Chinese, it's a hardcoded string
                            if any(c >= '\u4e00' and c <= '\u9fa5' for c in cleaned):
                                # Check if it looks like a translation call already or a variable
                                if re.search(r'{{|}}|t\\\'', cleaned):
                                    continue
                                # If it's inside t() or tt(), it might be a nested call or a default value
                                # We'll handle those in apply.py or by not filtering them here
                                all_matches.append(cleaned)
                            elif '.' in cleaned: # Likely an i18n key
                                all_matches.append(cleaned)
                    
                    if all_matches:
                        unique_matches = sorted(list(set(all_matches)))
                        if unique_matches:
                            results[rel_path] = unique_matches
                            
    return results

def main():
    print("Starting i18n scan...")
    findings = scan_files()
    
    if not findings:
        print("No hardcoded Chinese strings found.")
        if os.path.exists(OUTPUT_FILE):
            os.remove(OUTPUT_FILE)
        return

    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        json.dump(findings, f, ensure_ascii=False, indent=2)
    
    total_files = len(findings)
    total_strings = sum(len(texts) for texts in findings.values())
    print(f"Scan complete. Found {total_strings} strings in {total_files} files.")
    print(f"Findings saved to {OUTPUT_FILE}")

if __name__ == "__main__":
    main()
