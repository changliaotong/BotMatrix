/**
 * Configuration management
 */

import { fetchWithAuth } from './api.js';
import { authRole } from './auth.js';
import { t } from './i18n.js';
import { showToast } from './ui.js?v=1.1.88';
import { updateStats, updateChatStats } from './stats.js';
import { updateSystemStats } from './system.js';

export function updateRefreshRate(rate) {
    rate = parseInt(rate);
    localStorage.setItem('refresh_rate', rate);
    const ms = rate * 1000;
    
    // Stop existing timers
    if (window.updateInterval) clearInterval(window.updateInterval);
    if (window.systemStatsInterval) clearInterval(window.systemStatsInterval);
    if (window.chatStatsInterval) clearInterval(window.chatStatsInterval);
    
    const safeSetInterval = (fn, interval, name) => {
        return setInterval(() => {
            try {
                if (typeof fn === 'function') {
                    fn();
                } else if (typeof window[fn] === 'function') {
                    window[fn]();
                }
            } catch (error) {
                console.error(`${name}定时器执行失败:`, error);
            }
        }, interval);
    };
    
    // Re-start with new rate if not 0
    if (ms > 0) {
        window.updateInterval = safeSetInterval(updateStats, ms, 'updateStats');
        window.systemStatsInterval = safeSetInterval(updateSystemStats, ms, 'updateSystemStats');
        window.chatStatsInterval = safeSetInterval(updateChatStats, ms * 3, 'updateChatStats');
    }
    
    showToast('刷新频率已更新', 'info');
}

export async function updatePassword() {
    const oldPwd = document.getElementById('pwd-current').value;
    const newPwd = document.getElementById('pwd-new').value;
    const confirmPwd = document.getElementById('pwd-confirm').value;

    if (!newPwd) return alert(t('enter_new_pwd') || '请输入新密码');
    if (newPwd !== confirmPwd) return alert(t('pwd_mismatch') || '两次输入的密码不一致');
    if (!oldPwd) return alert(t('enter_current_pwd') || '请输入当前密码');

    try {
        const res = await fetchWithAuth('/api/user/password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ old_password: oldPwd, new_password: newPwd })
        });

        if (!res.ok) {
            const txt = await res.text();
            throw new Error(txt);
        }

        alert(t('password_change_success') || '密码修改成功');
        document.getElementById('pwd-current').value = '';
        document.getElementById('pwd-new').value = '';
        document.getElementById('pwd-confirm').value = '';
    } catch (e) {
        alert((t('password_change_fail') || '修改失败: ') + e.message);
    }
}

export async function loadBackendConfig() {
    if (authRole !== 'admin' && authRole !== 'super') return;
    
    try {
        const res = await fetchWithAuth('/api/admin/config');
        if (!res.ok) throw new Error('Failed to load config');
        
        const config = await res.json();
        document.getElementById('cfg-ws-port').value = config.ws_port || '';
        document.getElementById('cfg-webui-port').value = config.webui_port || '';
        document.getElementById('cfg-redis-addr').value = config.redis_addr || '';
        document.getElementById('cfg-redis-pwd').value = config.redis_pwd || '';
        document.getElementById('cfg-jwt-secret').value = config.jwt_secret || '';
        document.getElementById('cfg-admin-pwd').value = config.default_admin_password || '';
        document.getElementById('cfg-stats-file').value = config.stats_file || '';
    } catch (e) {
        console.error('Error loading backend config:', e);
    }
}

export async function updateBackendConfig() {
    const config = {
        ws_port: document.getElementById('cfg-ws-port').value,
        webui_port: document.getElementById('cfg-webui-port').value,
        redis_addr: document.getElementById('cfg-redis-addr').value,
        redis_pwd: document.getElementById('cfg-redis-pwd').value,
        jwt_secret: document.getElementById('cfg-jwt-secret').value,
        default_admin_password: document.getElementById('cfg-admin-pwd').value,
        stats_file: document.getElementById('cfg-stats-file').value
    };

    if (!confirm('确定要保存后端配置吗？部分修改（如端口）需要重启服务后生效。')) return;

    try {
        const res = await fetchWithAuth('/api/admin/config', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });

        const data = await res.json();
        if (data.success) {
            alert(data.message || '配置已更新');
        } else {
            alert('更新失败: ' + data.message);
        }
    } catch (e) {
        alert('更新出错: ' + e.message);
    }
}

// Expose to window for legacy compatibility
window.updatePassword = updatePassword;
window.loadBackendConfig = loadBackendConfig;
window.updateBackendConfig = updateBackendConfig;
window.updateRefreshRate = updateRefreshRate;
