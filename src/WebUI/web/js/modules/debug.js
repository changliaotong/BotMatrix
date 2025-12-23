/**
 * Debug and Test Actions Module
 */
import { fetchWithAuth, callBotApi } from './api.js';
import { t } from './i18n.js';
import { showToast } from './ui.js';

// --- Console History & Error Handling ---
console.history = console.history || [];
(function() {
    const originalLog = console.log;
    const originalError = console.error;
    const originalWarn = console.warn;
    
    console.log = function() {
        console.history.push('[LOG] ' + Array.from(arguments).join(' '));
        originalLog.apply(console, arguments);
    };
    console.error = function() {
        console.history.push('[ERR] ' + Array.from(arguments).join(' '));
        originalError.apply(console, arguments);
    };
    console.warn = function() {
        console.history.push('[WRN] ' + Array.from(arguments).join(' '));
        originalWarn.apply(console, arguments);
    };
})();

// Global Error Handler for better debugging
window.onerror = function(message, source, lineno, colno, error) {
    console.error('Global Error Caught:', { message, source, lineno, colno, error });
    // Don't show alert for known external library errors that don't break the app
    if (source && (source.includes('chart.js') || source.includes('three.js'))) return;
    return false;
};

export function sendAction() {
    let botId = document.getElementById('action-bot-id').value;
    if (!botId && window.currentBotId) {
        botId = window.currentBotId;
    }
    
    const action = document.getElementById('action-name').value;
    let params = {};
    try {
        params = JSON.parse(document.getElementById('action-params').value);
    } catch (e) {
        showToast(t('alert_json_error') || 'JSON 格式错误', 'danger');
        return;
    }

    fetchWithAuth('/api/action', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            bot_id: botId,
            action: action,
            params: params
        })
    })
    .then(r => r.json())
    .then(res => {
        const resultEl = document.getElementById('action-result');
        if (resultEl) resultEl.innerText = JSON.stringify(res, null, 2);
    })
    .catch(err => {
        const resultEl = document.getElementById('action-result');
        if (resultEl) resultEl.innerText = 'Error: ' + err;
    });
}

export async function loadTestGroups() {
    const select = document.getElementById('mt-group-select');
    if (!select) return;
    select.innerHTML = `<option value="">${t('loading') || '加载中...'}</option>`;
    
    try {
        const response = await callBotApi('get_group_list');
        const groups = response.data || [];
        
        // Sort by name
        groups.sort((a, b) => (a.group_name || '').localeCompare(b.group_name || '', 'zh-CN'));
        
        let html = `<option value="">-- ${t('select_group_placeholder') || '选择群组 (或手动输入 UID)'} --</option>`;
        groups.forEach(g => {
            html += `<option value="${g.group_id}">[${g.group_id}] ${g.group_name}</option>`;
        });
        select.innerHTML = html;
    } catch (e) {
        select.innerHTML = `<option value="">${t('load_failed') || '加载失败'}: ` + e.message + '</option>';
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
        const groupText = document.getElementById('mt-group-text');
        const contentEl = document.getElementById('mt-content');
        if (groupText) groupText.style.display = 'block';
        if (contentEl) contentEl.placeholder = type === 'poke' ? (t('placeholder_poke') || '请输入QQ号') : (t('placeholder_text_cq') || '请输入文字或CQ码');
    } else if (type === 'image' || type === 'record' || type === 'video') {
        const groupFile = document.getElementById('mt-group-file');
        const fileLabel = document.getElementById('mt-file-label');
        if (groupFile) groupFile.style.display = 'block';
        if (fileLabel) {
            let label = t('label_file_url') || '文件 URL';
            if (type === 'image') label = t('label_image_url') || '图片 URL';
            else if (type === 'record') label = t('label_record_url') || '语音 URL';
            else if (type === 'video') label = t('label_video_url') || '视频 URL';
            fileLabel.textContent = label;
        }
    } else if (type === 'music') {
        const groupMusic = document.getElementById('mt-group-music');
        if (groupMusic) groupMusic.style.display = 'block';
    } else if (type === 'share') {
        const groupShare = document.getElementById('mt-group-share');
        if (groupShare) groupShare.style.display = 'block';
    } else if (type === 'json' || type === 'xml') {
        const groupRaw = document.getElementById('mt-group-raw');
        const rawLabel = document.getElementById('mt-raw-label');
        if (groupRaw) groupRaw.style.display = 'block';
        if (rawLabel) rawLabel.textContent = type.toUpperCase() + (t('label_raw_content_suffix') || ' 内容');
    }
}

export async function submitTestMsg() {
    const resultEl = document.getElementById('mt-result');
    if (!resultEl) return;
    
    resultEl.style.display = 'none';
    resultEl.className = 'alert alert-secondary';
    resultEl.textContent = t('msg_sending') || '正在发送...';
    resultEl.style.display = 'block';
    
    const targetEl = document.getElementById('mt-target');
    const typeEl = document.getElementById('mt-type');
    if (!targetEl || !typeEl) return;
    
    const target = targetEl.value;
    const type = typeEl.value;
    
    if (!target) {
        resultEl.className = 'alert alert-danger';
        resultEl.textContent = t('alert_enter_target_uid') || '请输入目标 UID';
        return;
    }

    let action = 'send_group_msg';
    let params = {
        group_id: parseInt(target)
    };
    
    if (type === 'text') {
        const content = document.getElementById('mt-content').value;
        if (!content) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t('alert_enter_content') || '请输入内容';
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
            resultEl.textContent = t('alert_enter_file_url') || '请输入文件 URL';
            return;
        }
        params.message = [{ type: type, data: { file: file } }];
    } else if (type === 'music') {
        const mType = document.getElementById('mt-music-type').value;
        const mId = document.getElementById('mt-music-id').value;
        if (!mId) {
            resultEl.className = 'alert alert-danger';
            resultEl.textContent = t('msg_input_song_id') || '请输入歌曲 ID';
            return;
        }
        params.message = [{ type: 'music', data: { type: mType, id: mId } }];
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
            resultEl.textContent = t('msg_input_content') || '请输入内容';
            return;
        }
        params.message = [{ type: type, data: { data: raw } }];
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
    const targetEl = document.getElementById('mt-target');
    const typeEl = document.getElementById('mt-type');
    if (!targetEl || !typeEl) return;
    
    const target = targetEl.value;
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
    const previewEl = document.getElementById('code-preview');
    if (!previewEl) return;
    
    const content = previewEl.textContent;
    navigator.clipboard.writeText(content).then(() => {
        const btn = document.querySelector('button[onclick="copyCodePreview()"]');
        if(btn) {
            const original = btn.textContent;
            btn.textContent = t('btn_copied') || '已复制';
            setTimeout(() => btn.textContent = original, 2000);
        }
    });
}

// Global exposure
window.sendAction = sendAction;
window.loadTestGroups = loadTestGroups;
window.onTestGroupSelectChange = onTestGroupSelectChange;
window.pasteTargetUid = pasteTargetUid;
window.updateMsgForm = updateMsgForm;
window.submitTestMsg = submitTestMsg;
window.updateCodePreview = updateCodePreview;
window.copyCodePreview = copyCodePreview;
