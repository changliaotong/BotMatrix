import { fetchWithAuth } from './api.js';
import { t } from './i18n.js';
import { showToast } from './ui.js';

export let massSendType = 'group';

export function showMassSendModal(type) {
    massSendType = type;
    const checkboxes = document.querySelectorAll(`.mass-send-checkbox-${type}:checked`);
    if (checkboxes.length === 0) {
        showToast(t('mass_send_select_first') || '请先选择要群发的目标', 'warning');
        return;
    }
    
    const modalEl = document.getElementById('massSendModal');
    if (!modalEl) return;
    
    const modal = new bootstrap.Modal(modalEl);
    document.getElementById('mass-send-content').value = '';
    document.getElementById('mass-send-progress').classList.add('d-none');
    document.getElementById('mass-send-status').innerText = t('mass_send_selected', { count: checkboxes.length });
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
        showToast(t('mass_send_input_content') || '请输入消息内容', 'warning');
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
    if (statusText) statusText.innerText = t('mass_send_preparing', { count: targets.length });

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
            showToast(t('mass_send_started'), 'success');
            
            if (statusText) statusText.innerText = t('mass_send_started_detail', { count: targets.length });
            if (startBtn) {
                startBtn.disabled = false;
                startBtn.innerText = t('mass_send_complete');
                startBtn.onclick = () => {
                    const modalEl = document.getElementById('massSendModal');
                    const modal = bootstrap.Modal.getInstance(modalEl);
                    if (modal) modal.hide();
                    
                    setTimeout(() => {
                        startBtn.innerText = t('mass_send_start');
                        startBtn.onclick = executeMassSend;
                    }, 500);
                };
            }
        } else {
            showToast(t('mass_send_failed') + (result.message || t('unknown')), 'danger');
            if (startBtn) startBtn.disabled = false;
        }
    } catch (e) {
        console.error('Mass send error:', e);
        showToast(t('mass_send_exception') + e.message, 'danger');
        if (startBtn) startBtn.disabled = false;
    }
}

// Global exposure
window.showMassSendModal = showMassSendModal;
window.toggleSelectAll = toggleSelectAll;
window.executeMassSend = executeMassSend;
