/**
 * API communication module
 */
import { authToken } from './auth.js';
import { currentBotId } from './bots.js';

export const pendingRequests = new Map(); // echo -> {resolve, reject, timeout}

export async function fetchWithAuth(url, options = {}) {
    const token = authToken || localStorage.getItem('wxbot_token');
    if (!token) {
        console.error('No auth token found');
        // If we're not on the login page already, we might want to show it
        const lp = document.getElementById('loginPage');
        if (lp && window.getComputedStyle(lp).display === 'none') {
            lp.classList.remove('hidden');
            lp.style.display = 'flex';
            const mainApp = document.querySelector('.main-app');
            if (mainApp) mainApp.style.display = 'none';
        }
        throw new Error('未登录');
    }

    const defaultOptions = {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    };
    
    // 合并 options
    const mergedOptions = {
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...(options.headers || {})
        }
    };
    
    const res = await fetch(url, mergedOptions);
    
    // Handle 401 Unauthorized
    if (res.status === 401) {
        console.warn('Auth token expired or invalid, redirecting to login');
        localStorage.removeItem('wxbot_token');
        localStorage.removeItem('wxbot_role');
        window.authToken = null;
        
        const lp = document.getElementById('loginPage');
        if (lp) {
            lp.classList.remove('hidden');
            lp.style.display = 'flex';
            // Hide main app container if needed
            const mainApp = document.querySelector('.main-app');
            if (mainApp) mainApp.style.display = 'none';
        }
        throw new Error('认证失效，请重新登录');
    }
    
    return res;
}

// Global exposure for legacy compatibility
window.fetchWithAuth = fetchWithAuth;
window.callBotApi = callBotApi;
window.sendAction = sendAction;

export async function callBotApi(action, params = {}, botId = null) {
    const targetBotId = botId || window.currentBotId;
    if (!targetBotId) {
        const selector = document.getElementById('global-bot-selector');
        if (selector && selector.value) {
            window.currentBotId = selector.value;
        } else {
            throw new Error('No bot selected');
        }
    }

    const response = await fetchWithAuth('/api/action', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            bot_id: targetBotId || window.currentBotId,
            action: action,
            params: params
        })
    });

    if (response.status === 401) {
        localStorage.removeItem('wxbot_token');
        location.reload();
        throw new Error("Unauthorized");
    }

    const result = await response.json();
    if (result.error) throw new Error(result.error);
    
    // If there's an echo, it's an async request that will be resolved via WebSocket
    if (result.echo) {
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                pendingRequests.delete(result.echo);
                reject(new Error("Request timed out (30s)"));
            }, 30000);
            
            pendingRequests.set(result.echo, { resolve, reject, timeout });
        });
    }

    return result;
}

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
        const { currentLang, translations } = window;
        const t = translations && translations[currentLang] ? translations[currentLang] : {};
        alert(t.alert_json_error || 'JSON format error');
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

window.callBotApi = callBotApi;
window.sendAction = sendAction;
window.fetchWithAuth = fetchWithAuth;
