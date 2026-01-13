import os
import re
import json

# Configuration
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LOCALES_DIR = os.path.join(PROJECT_ROOT, 'src/locales')
FINDINGS_FILE = os.path.join(os.path.dirname(__file__), 'findings.json')
MAPPING_FILE = os.path.join(os.path.dirname(__file__), 'mapping.json')

COMMON_MAP = {
    "搜索": "common.search",
    "确认": "common.confirm",
    "取消": "common.cancel",
    "确定": "common.ok",
    "保存": "common.save",
    "成功": "common.success",
    "失败": "common.failed",
    "错误": "common.error",
    "设置": "common.settings",
    "关闭": "common.close",
    "开启": "common.open",
    "在线": "common.online",
    "离线": "common.offline",
    "管理": "common.manage",
    "停止": "common.stop",
    "启动": "common.start",
    "重启": "common.restart",
    "重试": "common.retry",
    "全部": "common.all",
    "翻译": "common.translate",
    "高级": "common.advanced",
    "无": "common.none",
    "早喵": "common.earlymeow",
    "添加": "common.add",
    "删除": "common.delete",
    "操作": "common.actions",
    "状态": "common.status",
    "正在加载": "common.loading",
    "加载中": "common.loading",
    "详情": "common.details",
    "返回": "common.back",
    "提交": "common.submit",
    "刷新": "common.refresh",
    "控制台": "common.console",
    "机器人": "common.bot",
    "好友": "common.friend",
    "群组": "common.group",
    "成员": "common.member",
    "消息": "common.message",
    "发送": "common.send",
    "输入": "common.input",
    "昵称": "common.nickname",
    "等级": "common.level",
    "创建": "common.create",
    "编辑": "common.edit",
    "退出": "common.logout",
    "登录": "common.login",
    "个人中心": "common.profile",
    "系统设置": "common.system_settings",
}

def get_key_prefix(rel_path):
    filename = os.path.basename(rel_path).split('.')[0].lower()
    if 'nexusai' in filename: return 'ai'
    if 'plugins' in filename: return 'plugins'
    if 'pricing' in filename: return 'pricing'
    if 'nexusguard' in filename: return 'guard'
    if 'portal' in rel_path.lower(): return 'portal'
    if 'admin' in rel_path.lower(): return 'admin'
    return filename

def load_existing_locales():
    key_to_val = {}
    val_to_key = {}
    file_path = os.path.join(LOCALES_DIR, 'zh-CN.ts')
    if os.path.exists(file_path):
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            # Matches 'key': 'value', handling escaped quotes and potential multi-line values
            matches = re.findall(r"^\s*'([^']+)':\s*'(.*?)',", content, re.MULTILINE | re.DOTALL)
            for key, val in matches:
                # Clean up the value
                val = val.replace("\\'", "'").replace("\\\\", "\\")
                key_to_val[key] = val
                if val and val not in val_to_key:
                    val_to_key[val] = key
    return key_to_val, val_to_key

def generate_mapping(findings):
    mapping = {} # Chinese -> Key
    key_to_val, val_to_key = load_existing_locales()
    existing_keys = set(key_to_val.keys())
    
    # Load existing mapping if available
    if os.path.exists(MAPPING_FILE):
        with open(MAPPING_FILE, 'r', encoding='utf-8') as f:
            mapping = json.load(f)
            existing_keys.update(mapping.values())

    # Track all keys we want to ensure have values
    all_entries = {}
    new_entries = {}
    
    for rel_path, texts in findings.items():
        prefix = get_key_prefix(rel_path)
        for text in texts:
            # 1. Handle KEY|DEFAULT format from scan.py
            if '|' in text and not any(c >= '\u4e00' and c <= '\u9fa5' for c in text.split('|')[0]):
                parts = text.split('|', 1)
                key = parts[0]
                default_val = parts[1]
                
                # Resolve nested default keys if any
                if default_val in key_to_val:
                    resolved_val = key_to_val[default_val]
                    all_entries[key] = resolved_val
                elif default_val:
                    all_entries[key] = default_val
                
                if key not in existing_keys:
                    new_entries[key] = all_entries.get(key, default_val)
                    existing_keys.add(key)
                continue
            
            # 2. If it's just a key (no Chinese, contains dots)
            if '.' in text and not any(c >= '\u4e00' and c <= '\u9fa5' for c in text):
                # We don't have a default value here, so we can't do much unless it's in mapping
                continue

            # 3. Check if it's already in mapping.json
            if text in mapping:
                all_entries[mapping[text]] = text
                continue
                
            # 4. Check if it's in our COMMON_MAP
            if text in COMMON_MAP:
                mapping[text] = COMMON_MAP[text]
                all_entries[mapping[text]] = text
                continue
            
            # 5. Check if it already exists in zh-CN.ts (by value)
            if text in val_to_key:
                mapping[text] = val_to_key[text]
                all_entries[mapping[text]] = text
                continue
            
            # 6. Generate a new key for hardcoded Chinese
            # Skip if the text itself looks like an i18n key
            if text.startswith(f"{prefix}.") or text.startswith("common.") or text.startswith("ai."):
                continue

            safe_text = re.sub(r'[^\u4e00-\u9fa5a-zA-Z0-9]', '', text)
            key = f"{prefix}.{safe_text[:15]}"
            
            base_key = key
            counter = 1
            while key in existing_keys:
                key = f"{base_key}{counter}"
                counter += 1
                
            mapping[text] = key
            new_entries[key] = text
            all_entries[key] = text
            existing_keys.add(key)
            
    # Save updated mapping
    with open(MAPPING_FILE, 'w', encoding='utf-8') as f:
        json.dump(mapping, f, ensure_ascii=False, indent=2)
        
    return mapping, all_entries

def update_locale_files(all_entries):
    langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP']
    for lang in langs:
        file_path = os.path.join(LOCALES_DIR, f"{lang}.ts")
        if not os.path.exists(file_path):
            print(f"Locale file not found: {file_path}")
            continue
            
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            
        # Parse existing keys and their values
        existing_entries = {}
        # Match 'key': 'value' or "key": "value"
        matches = re.findall(r"^\s*['\"]([^'\"]+)['\"]\s*:\s*['\"](.*?)['\"]\s*,?", content, re.MULTILINE)
        for key, val in matches:
            existing_entries[key] = val
              
        added_count = 0
        updated_count = 0
        new_lines = []
        
        # We need to preserve the order and existing content, but update values
        updated_content = content
        
        for key, val in all_entries.items():
            escaped_val = val.replace("\\", "\\\\").replace("'", "\\'")
            if key in existing_entries:
                # If existing value is empty or looks like a placeholder, update it
                existing_val = existing_entries[key]
                # Force update for portal.about keys if they are empty or shorter than the new value
                is_about_key = key.startswith('portal.about.')
                if not existing_val or existing_val == key or (is_about_key and len(val) > len(existing_val)):
                    # Update the value in content
                    # More robust pattern: match the key and the value quotes
                    pattern = rf"(['\"]{re.escape(key)}['\"]\s*:\s*['\"])(.*?)(['\"])"
                    new_updated_content = re.sub(pattern, rf"\1{escaped_val}\3", updated_content)
                    if new_updated_content != updated_content:
                        updated_content = new_updated_content
                        updated_count += 1
            else:
                new_lines.append(f"  '{key}': '{escaped_val}',")
                added_count += 1
        
        if new_lines:
            # Safer insertion: look for the LAST closing brace of the export default object
            match = re.search(r'};\s*$', updated_content)
            if match:
                last_brace_idx = match.start()
                updated_content = updated_content[:last_brace_idx] + "\n".join(new_lines) + "\n" + updated_content[last_brace_idx:]
            else:
                print(f"Could not find valid end of object in {lang}.ts")
                continue

        if added_count > 0 or updated_count > 0:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(updated_content)
            print(f"Updated {lang}.ts: added {added_count} keys, updated {updated_count} values.")

def replace_in_files(findings, mapping):
    for rel_path, texts in findings.items():
        # NEVER process locale files or the scripts themselves
        if 'src/locales' in rel_path.replace('\\', '/') or 'scripts/i18n' in rel_path.replace('\\', '/'):
            continue
            
        full_path = os.path.join(PROJECT_ROOT, rel_path)
        with open(full_path, 'r', encoding='utf-8') as f:
            content = f.read()
            
        new_content = content
        
        # Sort texts by length descending to avoid partial replacements
        sorted_texts = sorted(texts, key=len, reverse=True)
        
        # Determine t function (t or tt)
        t_func = 't'
        if 'const tt =' in content or '{{ tt(' in content:
            t_func = 'tt'
        elif 'useI18n' in content:
            t_func = 't'
            
        for text in sorted_texts:
            # Skip if it's already a key (contains dots or matches KEY|DEFAULT)
            if '.' in text or '|' in text:
                continue
                
            if text not in mapping:
                continue
                
            key = mapping[text]
            escaped_text = re.escape(text)
            
            # 1. Vue Templates: > 中文 < -> > {{ t('key') }} <
            # Match > followed by optional whitespace, then the text, then optional whitespace and <
            vue_text_pattern = rf'>(\s*){escaped_text}(\s*)<'
            new_content = re.sub(vue_text_pattern, rf'>\1{{{{ {t_func}("{key}") }}}}\2<', new_content)
            
            # 2. Vue Attributes: attr="中文" -> :attr="t('key')"
            attrs = ['placeholder', 'title', 'label', 'text', 'ok-text', 'cancel-text', 'confirm-text', 'tip', 'message']
            for attr in attrs:
                # Match attr="text" or attr='text'
                attr_pattern = rf'\b{attr}=([\'"]){escaped_text}\1'
                new_content = re.sub(attr_pattern, rf':{attr}="{t_func}(\'{key}\')"', new_content)
                
            # 3. JS/TS Strings: '中文' -> t('key')
            # Only if not already part of a t() call or tt() call
            # We look for 'text' or "text" or `text`
            js_str_pattern = rf"(?<!{t_func}\()(['\"`]){escaped_text}\1"
            new_content = re.sub(js_str_pattern, rf"{t_func}('{key}')", new_content)
            
            # 4. Router Title: title: '中文' -> title: t('key')
            router_title_pattern = rf"title:\s*([\'\"]){escaped_text}\1"
            new_content = re.sub(router_title_pattern, rf"title: {t_func}('{key}')", new_content)

        if new_content != content:
            with open(full_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"Applied changes to {rel_path}")

def main():
    if not os.path.exists(FINDINGS_FILE):
        print(f"No findings file found at {FINDINGS_FILE}. Run scan.py first.")
        return
        
    with open(FINDINGS_FILE, 'r', encoding='utf-8') as f:
        findings = json.load(f)
        
    mapping, all_entries = generate_mapping(findings)
    update_locale_files(all_entries)
    replace_in_files(findings, mapping)
    print("Application complete.")

if __name__ == "__main__":
    main()
