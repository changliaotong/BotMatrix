import { fetchWithAuth } from './api.js';
import { currentLang, translations } from './i18n.js';
import { showToast } from './ui.js';

export let massSendType = 'group';

export function showMassSendModal(type) {
    massSendType = type;
    const checkboxes = document.querySelectorAll(`.mass-send-checkbox-${type}:checked`);
    if (checkboxes.length === 0) {
        const t = translations[currentLang] || translations['zh-CN'];
        showToast(t.mass_send_select_first || '请先选择要群发的目标', 'warning');
        return;
    }
    
    const modalEl = document.getElementById('massSendModal');
    if (!modalEl) return;
    
    const modal = new bootstrap.Modal(modalEl);
    document.getElementById('mass-send-content').value = '';
    document.getElementById('mass-send-progress').classList.add('d-none');
    document.getElementById('mass-send-status').innerText = `已选择 ${checkboxes.length} 个目标`;
    modal.show();
}

export function toggleSelectAll(type) {
    const master = document.getElementById(`select-all-${type}`);
    if (!master) return;
    const checkboxes = document.querySelectorAll(`.mass-send-checkbox-${type}`);
    checkboxes.forEach(cb => cb.checked = master.checked);
}

export async function executeMassSend() {
    const content = document.getElementById('mass-send-content').value.trim();
    if (!content) {
        const t = translations[currentLang] || translations['zh-CN'];
        showToast(t.mass_send_input_content || '请输入消息内容', 'warning');
        return;
    }

    const checkboxes = document.querySelectorAll(`.mass-send-checkbox-${massSendType}:checked`);
    const targets = Array.from(checkboxes).map(cb => ({
        id: cb.getAttribute('data-id'),
        type: cb.getAttribute('data-type'),
        name: cb.getAttribute('data-name'),
        bot_id: cb.getAttribute('data-bot-id'),
        guild_id: cb.getAttribute('data-guild') || ''
    }));

    if (targets.length === 0) return;

    const progressContainer = document.getElementById('mass-send-progress');
    const statusText = document.getElementById('mass-send-status');
    const startBtn = document.querySelector('#massSendModal .modal-footer .btn-danger');

    if (progressContainer) progressContainer.classList.remove('d-none');
    if (startBtn) startBtn.disabled = true;
    if (statusText) statusText.innerText = `正在准备发送 ${targets.length} 个目标...`;

    try {
        const response = await fetchWithAuth('/api/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                action: 'batch_send_msg',
                params: {
                    targets: targets,
                    message: content
                }
            })
        });

        const result = await response.json();
        if (result.success || result.status === 'ok') {
            showToast(`群发任务已启动，请在日志中查看进度`, 'success');
            
            if (statusText) statusText.innerText = `已启动批量发送任务，共 ${targets.length} 个目标。请在日志中查看进度。`;
            if (startBtn) {
                startBtn.disabled = false;
                startBtn.innerText = '完成';
                startBtn.onclick = () => {
                    const modalEl = document.getElementById('massSendModal');
                    const modal = bootstrap.Modal.getInstance(modalEl);
                    if (modal) modal.hide();
                    
                    setTimeout(() => {
                        startBtn.innerText = '开始群发';
                        startBtn.onclick = executeMassSend;
                    }, 500);
                };
            }
        } else {
            showToast('群发失败: ' + (result.message || '未知错误'), 'danger');
            if (startBtn) startBtn.disabled = false;
        }
    } catch (e) {
        console.error('Mass send error:', e);
        showToast('群发请求异常: ' + e.message, 'danger');
        if (startBtn) startBtn.disabled = false;
    }
}

// Global exposure
window.showMassSendModal = showMassSendModal;
window.toggleSelectAll = toggleSelectAll;
window.executeMassSend = executeMassSend;
