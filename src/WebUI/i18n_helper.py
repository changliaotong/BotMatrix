import os
import re
import json

# 配置
SEARCH_DIRS = ['src/views', 'src/components']
I18N_FILE = 'src/locales/zh-CN.ts'
PROJECT_ROOT = os.getcwd()

# 匹配中文字符的正则 (包含标点)
CHINESE_PATTERN = re.compile(r'[\u4e00-\u9fa5]+[，。？！；：、“”‘’（）\u4e00-\u9fa5]*')

def get_key_prefix(rel_path):
    # 根据文件路径生成 key 前缀
    # e.g. src/views/admin/NexusAI.vue -> ai
    # e.g. src/views/portal/Pricing.vue -> pricing
    filename = os.path.basename(rel_path).replace('.vue', '').lower()
    if 'nexusai' in filename: return 'ai'
    if 'plugins' in filename: return 'plugins'
    if 'pricing' in filename: return 'pricing'
    if 'nexusguard' in filename: return 'guard'
    if 'earlymeow' in rel_path.lower(): return 'meow'
    return filename

def scan_chinese_text():
    results = {}
    for search_dir in SEARCH_DIRS:
        full_path = os.path.join(PROJECT_ROOT, search_dir)
        if not os.path.exists(full_path):
            continue
            
        for root, dirs, files in os.walk(full_path):
            for file in files:
                if file.endswith('.vue'):
                    file_path = os.path.join(root, file)
                    rel_path = os.path.relpath(file_path, PROJECT_ROOT)
                    
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                        
                        # 排除注释
                        content_no_comments = re.sub(r'<!--.*?-->', '', content, flags=re.DOTALL)
                        content_no_comments = re.sub(r'//.*?\n', '\n', content_no_comments)
                        content_no_comments = re.sub(r'/\*.*?\*/', '', content_no_comments, flags=re.DOTALL)
                        
                        # 排除已经在 tt() 或 t() 中的
                        # 使用非贪婪匹配，处理多参数情况
                        content_no_tt = re.sub(r"(tt|t)\(.*?\)", '', content_no_comments, flags=re.DOTALL)
                        
                        matches = CHINESE_PATTERN.findall(content_no_tt)
                        if matches:
                            unique_matches = sorted(list(set([m.strip() for m in matches if len(m.strip()) > 0])))
                            if unique_matches:
                                results[rel_path] = unique_matches
                                
    return results

def generate_keys(findings):
    mapping = {} # Chinese -> Key
    i18n_updates = {} # Key -> Chinese
    
    # 预定义的通用词汇映射，避免重复生成不同前缀的 key
    common_map = {
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
        "等级": "common.level"
    }

    for rel_path, texts in findings.items():
        prefix = get_key_prefix(rel_path)
        for text in texts:
            if text in common_map:
                mapping[text] = common_map[text]
                continue
            
            if text in mapping:
                continue
            
            # 生成简单的 key (基于拼音首字母或序号)
            # 这里简单处理，后续可以人工微调
            safe_text = re.sub(r'[^\u4e00-\u9fa5]', '', text)
            key = f"{prefix}.{safe_text[:10]}" # 暂时用中文片段作为 key 的一部分，方便识别
            mapping[text] = key
            i18n_updates[key] = text
            
    return mapping, i18n_updates

def update_i18n_ts(i18n_updates):
    if not i18n_updates: return
    
    with open(I18N_FILE, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 分别处理四个语言区块
    languages = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP']
    
    for lang in languages:
        # 寻找该语言区块的末尾
        # 假设格式是 'lang': { ... }
        pattern = rf"'{lang}': \{{(.*?)\}}"
        match = re.search(pattern, content, re.DOTALL)
        if match:
            block_content = match.group(1)
            new_entries = []
            for key, val in i18n_updates.items():
                if f"'{key}':" not in block_content:
                    new_entries.append(f"    '{key}': '{val}',")
            
            if new_entries:
                # 在最后一个逗号后插入
                last_comma_idx = block_content.rfind(',')
                if last_comma_idx != -1:
                    updated_block = block_content[:last_comma_idx+1] + "\n" + "\n".join(new_entries) + block_content[last_comma_idx+1:]
                    content = content.replace(block_content, updated_block)
    
    with open(I18N_FILE, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"已更新 {I18N_FILE}")

def replace_in_vue_files(findings, mapping):
    for rel_path in findings.keys():
        full_path = os.path.join(PROJECT_ROOT, rel_path)
        with open(full_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # 检查该文件使用的是 t 还是 tt
        if "const { t } = useI18n()" in content:
            t_func = 't'
        elif "const t = (" in content and "systemStore.t" in content:
            t_func = 't'
        elif "const tt = (" in content:
            t_func = 'tt'
        else:
            t_func = 'tt'
        
        # 排序：先替换长的，避免子串冲突
        sorted_texts = sorted(findings[rel_path], key=len, reverse=True)
        
        new_content = content
        for text in sorted_texts:
            key = mapping[text]
            
            # 1. 替换标签内容: >中文< -> >{{ t('key') }}<
            new_content = new_content.replace(f">{text}<", f">{{ {t_func}('{key}') }}<")
            
            # 2. 替换常见属性值: attr="中文" -> :attr="t('key')"
            attrs = ['placeholder', 'title', 'label', 'confirm-text', 'cancel-text', 'ok-text', 'message', 'description', 'hint']
            for attr in attrs:
                new_content = new_content.replace(f'{attr}="{text}"', f':{attr}="{t_func}(\'{key}\')"')
                new_content = new_content.replace(f"{attr}='{text}'", f":{attr}=\"{t_func}('{key}')\"")
            
            # 3. 替换 JS 字符串: '中文' -> t('key')
            # 匹配引号包围的中文
            # 必须确保它不是 t() 或 tt() 的一部分
            # 这里的 (?<!{t_func}\() 只检查了紧邻的前面，如果是 tt('key', '中文') 则无法避开
            # 改进：如果前面有 t( 或 tt( 且没有闭括号，则跳过
            # 但正则很难完美处理嵌套。这里用一个简单的方案：
            # 如果匹配到的字符串前后已经是 t('key') 这种形式，则不替换。
            
            # 先检查是否已经在 t('') 中
            if f"{t_func}('{key}')" in new_content or f"{t_func}(\"{key}\")" in new_content:
                # 如果这个 key 已经对应的文本被替换过了，
                # 我们要避免把 '中文' 替换成 t('key') 如果它已经是 tt('orig_key', '中文') 的一部分
                pass

            js_str_pattern = rf"(?<!{t_func}\()(['\"]){re.escape(text)}\1"
            # 额外的保护：避免替换已经带参数的 tt('key', '中文')
            # 我们可以匹配整个 tt('...', '中文') 并保持不变，或者匹配时要求前面不是 , 
            js_str_pattern = rf"(?<!{t_func}\()(?<!, )(['\"]){re.escape(text)}\1"
            new_content = re.sub(js_str_pattern, f"{t_func}('{key}')", new_content)
            
            # 4. 替换模板字符串中的中文: `...中文...` -> `...${t('key')}...`
            # 这种情况比较复杂，暂时只处理完全匹配或简单包含的
            # 如果整个模板字符串就是中文，则直接替换为 t('key')
            template_str_pattern = rf"(?<!{t_func}\()(`){re.escape(text)}\1"
            new_content = re.sub(template_str_pattern, f"{t_func}('{key}')", new_content)

        if new_content != content:
            with open(full_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"已更新 {rel_path}")

def main():
    print("开始扫描...")
    findings = scan_chinese_text()
    if not findings:
        print("未发现。")
        return

    mapping, i18n_updates = generate_keys(findings)
    
    # 1. 更新 i18n.ts
    update_i18n_ts(i18n_updates)
    
    # 2. 替换 Vue 文件
    replace_in_vue_files(findings, mapping)
    
    print("处理完成！")

if __name__ == "__main__":
    main()
