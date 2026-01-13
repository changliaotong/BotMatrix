import re
import os

LOCALES_DIR = 'src/locales'
PROJECT_ROOT = os.getcwd()

def parse_i18n_ts():
    locales_path = os.path.join(PROJECT_ROOT, LOCALES_DIR)
    if not os.path.exists(locales_path):
        locales_path = os.path.join(PROJECT_ROOT, 'src/WebUI', LOCALES_DIR)
        if not os.path.exists(locales_path):
            print(f"Error: {LOCALES_DIR} not found.")
            return None

    languages = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP']
    results = {}

    for lang in languages:
        file_path = os.path.join(locales_path, f"{lang}.ts")
        if os.path.exists(file_path):
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            # Match 'key': 'value' using a more robust regex that handles escaped quotes
            entries = re.findall(r"^\s*'([^']+)':\s*(['\"])(.*?)(?<!\\)\2,", content, re.MULTILINE)
            results[lang] = {k: v.replace("\\'", "'") for k, _, v in entries}
        else:
            results[lang] = {}

    return results

def check_missing_translations(data):
    if not data: return

    zh_keys = set(data['zh-CN'].keys())
    other_langs = ['zh-TW', 'en-US', 'ja-JP']

    report = {}
    for lang in other_langs:
        lang_keys = set(data[lang].keys())
        missing = zh_keys - lang_keys
        
        # Also check if the value is just the same as the key (often a placeholder)
        # or if it's the same as zh-CN but shouldn't be (e.g. en-US shouldn't have Chinese)
        identical_to_zh = []
        for key in (zh_keys & lang_keys):
            if lang == 'en-US' and any('\u4e00' <= c <= '\u9fff' for c in data[lang][key]):
                identical_to_zh.append(key)
            elif data[lang][key] == key and len(key) > 3: # Simple heuristic for placeholder
                identical_to_zh.append(key)

        report[lang] = {
            'missing': sorted(list(missing)),
            'potentially_untranslated': sorted(identical_to_zh)
        }

    return report

def main():
    data = parse_i18n_ts()
    if not data: return

    report = check_missing_translations(data)
    
    total_zh = len(data['zh-CN'])
    print(f"Total keys in zh-CN: {total_zh}")
    print("-" * 30)

    for lang, info in report.items():
        missing_count = len(info['missing'])
        potential_count = len(info['potentially_untranslated'])
        print(f"Language: {lang}")
        print(f"  Missing: {missing_count}")
        print(f"  Potentially untranslated (contains Chinese or key=val): {potential_count}")
        
        if missing_count > 0:
            print(f"  Sample missing keys: {info['missing'][:10]}")
        if potential_count > 0:
            print(f"  Sample potential keys: {info['potentially_untranslated'][:10]}")
        print("-" * 30)

    # Save to JSON for further processing
    import json
    with open('i18n_report.json', 'w', encoding='utf-8') as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
    print("Report saved to i18n_report.json")

if __name__ == "__main__":
    main()
