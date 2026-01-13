import re
import os

I18N_FILE = 'src/WebUI/src/utils/i18n.ts'

# Manual mapping for Chinese keys to English identifiers
key_mapping = {
    'ai.上传文档以开始构建机': 'ai.upload_kb_desc',
    'ai.上传新文档': 'ai.upload_new_doc',
    'ai.专员': 'ai.specialist',
    'ai.专家': 'ai.expert',
    'ai.主动记忆': 'ai.active_memory',
    'ai.助理': 'ai.assistant',
    'ai.开启语音播报': 'ai.voice_enable',
    'ai.忙碌': 'ai.busy',
    'ai.技能集': 'ai.skills',
    'ai.描述该数字员工的性格': 'ai.employee_bio_desc',
    'ai.搜索文档': 'ai.search_docs',
    'ai.数字员工档案': 'ai.employee_profile',
    'ai.数据分析': 'ai.data_analysis',
    'ai.本地受限环境': 'ai.local_restricted',
    'ai.核心研发部': 'ai.core_rd',
    'ai.正在上传': 'ai.uploading',
    'ai.直属主管': 'ai.supervisor',
    'ai.知识库检索': 'ai.kb_retrieval',
    'ai.知识库管理': 'ai.kb_management',
    'ai.研发助理': 'ai.rd_assistant',
    'ai.语速': 'ai.voice_rate',
    'ai.语音角色': 'ai.voice_role',
    'ai.语音设置': 'ai.voice_settings',
    'ai.资深专家': 'ai.senior_expert',
    'ai.逗号分隔': 'ai.comma_separated',
    'ai.链接已复制到剪贴板': 'ai.link_copied',
    'ai.长期记忆': 'ai.long_term_memory',
    'ai.隔离沙箱': 'ai.isolated_sandbox',
    'ai.默认中文女声': 'ai.default_zh_female',
    'plugins.个插件吗': 'plugins.plugins_count_confirm',
    'plugins.中心端和所有': 'plugins.center_and_all',
    'plugins.全部停止': 'plugins.stop_all',
    'plugins.全部启动': 'plugins.start_all',
    'plugins.刷新列表': 'plugins.refresh_list',
    'plugins.只读': 'plugins.read_only',
    'plugins.批量停止': 'plugins.batch_stop',
    'plugins.批量启动': 'plugins.batch_start',
    'plugins.批量操作完成': 'plugins.batch_op_success',
    'plugins.插件管理': 'plugins.management',
    'plugins.搜索插件': 'plugins.search',
    'plugins.暂无描述': 'plugins.no_description',
    'plugins.未找到插件': 'plugins.not_found',
    'plugins.核心组件': 'plugins.core_components',
    'plugins.没有可': 'plugins.no_available',
    'plugins.的插件': 'plugins.plugins_suffix',
    'plugins.确定要': 'plugins.confirm_action',
    'plugins.节点的插件': 'plugins.node_plugins',
    'plugins.节点离线': 'plugins.node_offline',
    'plugins.节点离线无法重启': 'plugins.node_offline_no_restart',
    'plugins.节点离线无法重载': 'plugins.node_offline_no_reload',
    'plugins.重载配置': 'plugins.reload_config',
    'bots.收到消息': 'bots.message_received',
    'groupsetup.仅回复关键词': 'groupsetup.keyword_only',
    'groupsetup.全自动': 'groupsetup.auto_mode',
    'groupsetup.关联群组': 'groupsetup.linked_groups',
    'pricing.功能特性': 'pricing.features',
    'pricing.商业版': 'pricing.business_edition',
    'pricing.矩阵版': 'pricing.matrix_edition',
    'pricing.社区版': 'pricing.community_edition',
    'guard.一键部署': 'guard.one_click_deploy',
    'guard.位环境支持': 'guard.env_support',
    'guard.快速部署指令': 'guard.quick_deploy_cmd',
    'guard.注需要': 'guard.note_required',
    'meow.喵喵助手': 'meow.assistant',
    'meow.摸鱼达人': 'meow.slacker',
    'meow.深夜电台': 'meow.late_night_radio',
    'docs.中主动摄取并更新认知': 'docs.active_cognition_desc',
    'docs.支持从': 'docs.support_from',
    'home.机制关键决策始终由人': 'home.human_audit_desc'
}

def refactor_i18n_ts():
    if not os.path.exists(I18N_FILE):
        # Try relative path
        file_path = os.path.join(os.getcwd(), 'src/utils/i18n.ts')
        if not os.path.exists(file_path):
             print(f"Error: {I18N_FILE} not found.")
             return
    else:
        file_path = I18N_FILE

    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Replace keys in the entire file
    # We use regex to match 'old_key': or "old_key":
    for old_key, new_key in key_mapping.items():
        # Match 'old_key':
        content = re.sub(rf"(['\"]){re.escape(old_key)}\1\s*:", f"'{new_key}':", content)
        # Match t('old_key') or tt('old_key') in case they are in the file (though unlikely)
        content = re.sub(rf"t\((['\"]){re.escape(old_key)}\1", f"t('{new_key}'", content)
        content = re.sub(rf"tt\((['\"]){re.escape(old_key)}\1", f"tt('{new_key}'", content)

    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    print("i18n.ts refactored successfully.")

def refactor_vue_files():
    # Recursively find all Vue files in src/views and src/components
    views_dir = os.path.join(os.getcwd(), 'src/views')
    components_dir = os.path.join(os.getcwd(), 'src/components')
    
    vue_files = []
    for root_dir in [views_dir, components_dir]:
        if not os.path.exists(root_dir): continue
        for root, dirs, files in os.walk(root_dir):
            for file in files:
                if file.endswith('.vue'):
                    vue_files.append(os.path.join(root, file))

    for vue_file in vue_files:
        with open(vue_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        for old_key, new_key in key_mapping.items():
            # Match t('old_key') or tt('old_key')
            content = re.sub(rf"t\((['\"]){re.escape(old_key)}\1", f"t('{new_key}'", content)
            content = re.sub(rf"tt\((['\"]){re.escape(old_key)}\1", f"tt('{new_key}'", content)
        
        if content != original_content:
            with open(vue_file, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"Refactored: {os.path.relpath(vue_file)}")

if __name__ == "__main__":
    refactor_i18n_ts()
    refactor_vue_files()
