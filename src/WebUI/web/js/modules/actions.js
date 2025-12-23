/**
 * Bot Actions module
 */

import { fetchWithAuth } from './api.js';
import { t } from './i18n.js';
import { currentBotId } from './bots.js';

/**
 * 发送通用机器人动作 (Debug Tab)
 */
export function sendAction() {
    let botId = document.getElementById('action-bot-id').value;
    if (!botId && window.currentBotId) {
        botId = window.currentBotId;
    }
    
    const action = document.getElementById('action-name').value;
    let params = {};
    try {
        const paramsValue = document.getElementById('action-params').value;
        params = JSON.parse(paramsValue || '{}');
    } catch (e) {
        alert(t('alert_json_error') || 'JSON 格式错误');
        return;
    }

    const resultEl = document.getElementById('action-result');
    if (resultEl) resultEl.innerText = 'Sending...';

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
        if (resultEl) resultEl.innerText = JSON.stringify(res, null, 2);
    })
    .catch(err => {
        if (resultEl) resultEl.innerText = 'Error: ' + err;
    });
}

// Global exposure
window.sendAction = sendAction;
