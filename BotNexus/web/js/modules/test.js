/**
 * Message Test module
 */

import { callBotApi } from './api.js';
import { currentLang, translations } from './i18n.js';

export async function loadTestGroups() {
    const select = document.getElementById('mt-group-select');
    if (!select) return;
    
    select.innerHTML = '<option value="">加载中...</option>';
    
    try {
        const response = await callBotApi('get_group_list');
        const groups = response.data || [];
        
        // Sort by name
        groups.sort((a, b) => a.group_name.localeCompare(b.group_name, 'zh-CN'));
        
        let html = '<option value="">-- 选择群组 (或手动输入 UID) --</option>';
        groups.forEach(g => {
            html += `<option value="${g.group_id}">[${g.group_id}] ${g.group_name}</option>`;
        });
        select.innerHTML = html;
    } catch (e) {
        select.innerHTML = '<option value="">加载失败: ' + e.message + '</option>';
    }
}

export function onTestGroupSelectChange() {
    const select = document.getElementById('mt-group-select');
    const targetInput = document.getElementById('mt-target');
    if (select && targetInput && select.value) {
        targetInput.value = select.value;
    }
}

export function pasteTargetUid() {
    const select = document.getElementById('mt-group-select');
    const targetInput = document.getElementById('mt-target');
    if (select && targetInput && select.value) {
        targetInput.value = select.value;
    }
}

export function updateMsgForm() {
    const t = translations[currentLang] || translations['zh-CN'];
    const typeEl = document.getElementById('mt-type');
    if (!typeEl) return;
    
    const type = typeEl.value;
    
    // Hide all specific inputs
    const groups = ['mt-group-text', 'mt-group-file', 'mt-group-music', 'mt-group-share', 'mt-group-raw'];
    groups.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.style.display = 'none';
    });
    
    if (type === 'text' || type === 'poke') {
        const textGroup = document.getElementById('mt-group-text');
        const contentInput = document.getElementById('mt-content');
        if (textGroup) textGroup.style.display = 'block';
        if (contentInput) contentInput.placeholder = type === 'poke' ? t.placeholder_poke : t.placeholder_text_cq;
    } else if (type === 'image' || type === 'record' || type === 'video') {
        const fileGroup = document.getElementById('mt-group-file');
        const label = document.getElementById('mt-file-label');
        if (fileGroup) fileGroup.style.display = 'block';
        if (label) {
            let labelText = t.label_file_url;
            if (type === 'image') labelText = t.label_image_url;
            else if (type === 'record') labelText = t.label_record_url;
            else if (type === 'video') labelText = t.label_video_url;
            label.textContent = labelText;
        }
    } else if (type === 'music') {
        const musicGroup = document.getElementById('mt-group-music');
        if (musicGroup) musicGroup.style.display = 'block';
    } else if (type === 'share') {
        const shareGroup = document.getElementById('mt-group-share');
        if (shareGroup) shareGroup.style.display = 'block';
    } else if (type === 'json' || type === 'xml') {
        const rawGroup = document.getElementById('mt-group-raw');
        const label = document.getElementById('mt-raw-label');
        if (rawGroup) rawGroup.style.display = 'block';
        if (label) label.textContent = type.toUpperCase() + t.label_raw_content_suffix;
    }
}

export async function submitTestMsg() {
    const t = translations[currentLang] || translations['zh-CN'];
    const resultEl = document.getElementById('mt-result');
    if (!resultEl) return;
    
    resultEl.style.display = 'none';
    resultEl.className = 'alert alert-secondary';
    resultEl.textContent = t.msg_sending;
    resultEl.style.display = 'block';
    
    const targetInput = document.getElementById('mt-target');
    const typeEl = document.getElementById('mt-type');
    if (!targetInput || !typeEl) return;
    
    const target = targetInput.value;
    const type = typeEl.value;
    
    if (!target) {
        resultEl.className = 'alert alert-danger';
        resultEl.textContent = t.alert_enter_target_uid;
        return;
    }

    // Default to group message
    let action = 'send_group_msg';
    let params = {
        group_id: parseInt(target)
    };
    
    // Construct Message
    if (type === 'text') {
        const content = document.getElementById('mt-content').value;
        if (!content) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t.alert_enter_content;
            return;
        }
        params.message = content;
    } else if (type === 'poke') {
        const qq = document.getElementById('mt-content').value;
        params.message = `[CQ:poke,qq=${qq || ''}]`;
    } else if (type === 'image' || type === 'record' || type === 'video') {
        const file = document.getElementById('mt-file-path').value;
        if (!file) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t.alert_enter_file_url;
            return;
        }
        params.message = [
            {
                type: type,
                data: { file: file }
            }
        ];
    } else if (type === 'music') {
        const mType = document.getElementById('mt-music-type').value;
        const mId = document.getElementById('mt-music-id').value;
        if (!mId) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t.msg_input_song_id;
            return;
        }
        params.message = [{
            type: 'music',
            data: {
                type: mType,
                id: mId
            }
        }];
    } else if (type === 'share') {
        params.message = [{
            type: 'share',
            data: {
                url: document.getElementById('mt-share-url').value,
                title: document.getElementById('mt-share-title').value,
                content: document.getElementById('mt-share-content').value,
                image: document.getElementById('mt-share-image').value
            }
        }];
    } else if (type === 'json' || type === 'xml') {
        const raw = document.getElementById('mt-raw-content').value;
        if (!raw) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t.msg_input_content;
            return;
        }
        params.message = [{
            type: type,
            data: {
                data: raw
            }
        }];
    }

    try {
        const res = await callBotApi(action, params);
        resultEl.className = 'alert alert-success';
        resultEl.textContent = '发送成功! MSG ID: ' + (res.data ? res.data.message_id : 'Unknown');
    } catch (e) {
        resultEl.className = 'alert alert-danger';
        resultEl.textContent = '发送失败: ' + e.message;
    }
}

export function updateCodePreview() {
    const targetInput = document.getElementById('mt-target');
    const typeEl = document.getElementById('mt-type');
    if (!targetInput || !typeEl) return;
    
    const target = targetInput.value;
    const type = typeEl.value;
    
    let params = {
        group_id: target ? parseInt(target) : 0
    };
    
    if (type === 'text') {
        params.message = document.getElementById('mt-content').value;
    } else if (type === 'poke') {
        const qq = document.getElementById('mt-content').value;
        params.message = `[CQ:poke,qq=${qq || ''}]`;
    } else if (type === 'image' || type === 'record' || type === 'video') {
        params.message = [{
            type: type,
            data: { file: document.getElementById('mt-file-path').value }
        }];
    } else if (type === 'music') {
        params.message = [{
            type: 'music',
            data: {
                type: document.getElementById('mt-music-type').value,
                id: document.getElementById('mt-music-id').value
            }
        }];
    } else if (type === 'share') {
        params.message = [{
            type: 'share',
            data: {
                url: document.getElementById('mt-share-url').value,
                title: document.getElementById('mt-share-title').value,
                content: document.getElementById('mt-share-content').value,
                image: document.getElementById('mt-share-image').value
            }
        }];
    } else if (type === 'json' || type === 'xml') {
        const raw = document.getElementById('mt-raw-content').value;
        params.message = [{
            type: type,
            data: { data: raw }
        }];
    }

    const preview = {
        action: 'send_group_msg',
        params: params
    };
    
    const previewEl = document.getElementById('code-preview');
    if (previewEl) previewEl.textContent = JSON.stringify(preview, null, 2);
}

export function copyCodePreview() {
    const t = translations[currentLang] || translations['zh-CN'];
    const previewEl = document.getElementById('code-preview');
    if (!previewEl) return;
    
    const content = previewEl.textContent;
    navigator.clipboard.writeText(content).then(() => {
        const btn = document.querySelector('button[onclick="copyCodePreview()"]');
        if(btn) {
            const original = btn.textContent;
            btn.textContent = t.btn_copied;
            setTimeout(() => {
                btn.textContent = original;
            }, 2000);
        }
    });
}
// End of module
