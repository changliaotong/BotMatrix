import os
import re
import json

# Configuration
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LOCALES_DIR = os.path.join(PROJECT_ROOT, 'src/locales')
SOURCE_DIRS = ['src/views', 'src/components', 'src/layouts', 'src/router', 'src/stores', 'src/utils']

def parse_locale_file(file_path):
    if not os.path.exists(file_path):
        return {}
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    # Match 'key': 'value',
    entries = re.findall(r"^\s*'([^']+)':\s*(['\"])(.*?)(?<!\\)\2,", content, re.MULTILINE)
    return {k: v.replace("\\'", "'") for k, _, v in entries}

def check_translations():
    langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP']
    locales = {lang: parse_locale_file(os.path.join(LOCALES_DIR, f"{lang}.ts")) for lang in langs}
    
    zh_keys = set(locales['zh-CN'].keys())
    
    report = {}
    for lang in langs[1:]: # Skip zh-CN
        lang_keys = set(locales[lang].keys())
        missing = zh_keys - lang_keys
        extra = lang_keys - zh_keys
        
        untranslated = []
        for key in (zh_keys & lang_keys):
            val = locales[lang][key]
            # Check if it still contains Chinese characters in non-Chinese locales
            if lang not in ['zh-CN', 'zh-TW'] and any('\u4e00' <= c <= '\u9fff' for c in val):
                untranslated.append(key)
                
        report[lang] = {
            'missing': sorted(list(missing)),
            'extra': sorted(list(extra)),
            'untranslated': sorted(untranslated)
        }
    return report

def find_unused_keys():
    langs = ['zh-CN'] # Only check against zh-CN for usage
    locales = {lang: parse_locale_file(os.path.join(LOCALES_DIR, f"{lang}.ts")) for lang in langs}
    zh_keys = set(locales['zh-CN'].keys())
    
    used_keys = set()
    for source_dir in SOURCE_DIRS:
        full_path = os.path.join(PROJECT_ROOT, source_dir)
        if not os.path.exists(full_path): continue
        
        for root, dirs, files in os.walk(full_path):
            for file in files:
                if not (file.endswith('.vue') or file.endswith('.ts') or file.endswith('.js')):
                    continue
                file_path = os.path.join(root, file)
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    # Find t('key') or tt('key')
                    matches = re.findall(r"(?:t|tt)\(['\"]([^'\"]+)['\"]\)", content)
                    for m in matches:
                        used_keys.add(m)
                        
    unused = zh_keys - used_keys
    return sorted(list(unused))

def main():
    print("Running i18n check...")
    
    trans_report = check_translations()
    for lang, info in trans_report.items():
        print(f"\nLanguage: {lang}")
        print(f"  Missing keys: {len(info['missing'])}")
        if info['missing']:
            print(f"    Sample: {info['missing'][:5]}")
        print(f"  Extra keys: {len(info['extra'])}")
        print(f"  Untranslated (contains Chinese): {len(info['untranslated'])}")
        if info['untranslated']:
            print(f"    Sample: {info['untranslated'][:5]}")
            
    unused = find_unused_keys()
    print(f"\nUnused keys in zh-CN.ts: {len(unused)}")
    if unused:
        print(f"  Sample: {unused[:10]}")
        
    # Save report
    with open(os.path.join(os.path.dirname(__file__), 'report.json'), 'w', encoding='utf-8') as f:
        json.dump({'translations': trans_report, 'unused': unused}, f, ensure_ascii=False, indent=2)
    print(f"\nFull report saved to scripts/i18n/report.json")

if __name__ == "__main__":
    main()
