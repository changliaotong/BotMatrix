import os
import re

LOCALES_DIR = 'src/locales'
PROJECT_ROOT = os.getcwd()

def update_locales(translations_to_add):
    locales_path = os.path.join(PROJECT_ROOT, LOCALES_DIR)
    if not os.path.exists(locales_path):
        locales_path = os.path.join(PROJECT_ROOT, 'src/WebUI', LOCALES_DIR)
        if not os.path.exists(locales_path):
            print(f"Error: {LOCALES_DIR} not found.")
            return

    for lang, new_entries in translations_to_add.items():
        file_path = os.path.join(locales_path, f"{lang}.ts")
        if not os.path.exists(file_path):
            print(f"Creating new locale file: {file_path}")
            content = "export default {\n};\n"
        else:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

        # Parse existing entries using a more robust regex that handles escaped quotes
        # It looks for 'key': 'value', where value can contain \'
        kv_pairs = re.findall(r"^\s*'([^']+)':\s*(['\"])(.*?)(?<!\\)\2,", content, re.MULTILINE)
        data = {k: v.replace("\\'", "'") for k, _, v in kv_pairs}
        
        # Add new entries
        data.update(new_entries)

        # Reconstruct and sort
        new_content = "export default {\n"
        for k in sorted(data.keys()):
            v = data[k]
            # Escape single quotes for the TS file
            v_escaped = v.replace("'", "\\'")
            new_content += f"  '{k}': '{v_escaped}',\n"
        new_content += "};\n"

        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Updated {lang}.ts with {len(new_entries)} entries.")

if __name__ == "__main__":
    # Example usage (can be modified as needed)
    translations = {
        'zh-TW': {
            'about': '关于早喵', # Just an example
        },
        'en-US': {
            'about': 'About EarlyMeow',
        },
        'ja-JP': {
            'about': '早喵について',
        }
    }
    # update_locales(translations)
    pass
