import re
import os

LOCALES_DIR = 'src/locales'
PROJECT_ROOT = os.getcwd()

def clean_i18n():
    locales_path = os.path.join(PROJECT_ROOT, LOCALES_DIR)
    if not os.path.exists(locales_path):
        locales_path = os.path.join(PROJECT_ROOT, 'src/WebUI', LOCALES_DIR)
        if not os.path.exists(locales_path):
            print(f"Error: {LOCALES_DIR} not found.")
            return

    langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP']
    
    for lang in langs:
        file_path = os.path.join(locales_path, f"{lang}.ts")
        if not os.path.exists(file_path):
            continue

        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Find all key-value pairs: 'key': 'value', or 'key': "value",
        kv_pairs = re.findall(r"'([^']+)':\s*(['\"])(.*?)\2,", content)
        
        # Use a dict to deduplicate keys, keeping the last one
        lang_dict = {}
        for k, _, v in kv_pairs:
            lang_dict[k] = v

        # Reconstruct the file
        new_content = "export default {\n"
        for k in sorted(lang_dict.keys()):
            v = lang_dict[k]
            # Escape single quotes in value
            v_escaped = v.replace("'", "\\'")
            new_content += f"  '{k}': '{v_escaped}',\n"
        new_content += "};\n"

        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        print(f"Cleaned {lang}.ts: {len(lang_dict)} keys")

if __name__ == "__main__":
    clean_i18n()
