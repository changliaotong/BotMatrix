/**
 * Auth and Login management
 */

import { applyRoleUI } from './ui.js';
import { translations, currentLang } from './i18n.js';
import { fetchWithAuth } from './api.js';

export { applyRoleUI };

export let authToken = localStorage.getItem('wxbot_token');
if (authToken === 'undefined' || authToken === 'null') {
    authToken = null;
    localStorage.removeItem('wxbot_token');
}

export let authRole = localStorage.getItem('wxbot_role') || 'user';
if (authRole === 'undefined' || authRole === 'null') {
    authRole = 'user';
    localStorage.removeItem('wxbot_role');
}

/**
 * Handle Magic Link Login
 */
export async function handleMagicToken(magicToken) {
    console.log('[Auth] Processing magic token...');
    try {
        const res = await fetch('/api/login/magic', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token: magicToken })
        });
        
        if (!res.ok) throw new Error('Magic login failed');
        
        const data = await res.json();
        console.log('[Auth] Magic login success');
        
        authToken = data.token;
        authRole = data.role;
        localStorage.setItem('wxbot_token', authToken);
        localStorage.setItem('wxbot_role', authRole);
        
        // Update globals
        window.authToken = authToken;
        window.authRole = authRole;
        
        // Clear URL params
        window.history.replaceState({}, document.title, window.location.pathname);
        
        return { success: true, role: authRole, data: data };
    } catch (err) {
        console.error('[Auth] Magic login error:', err);
        const t = translations[currentLang] || translations['zh-CN'] || {};
        alert(t.magic_login_fail || '免密码登录失败，链接可能已过期。请重新获取或使用密码登录。');
        window.history.replaceState({}, document.title, window.location.pathname);
        return { success: false, error: err.message };
    }
}

export function handleLogin() {
    initAuth();
}

export function handleLogout() {
    logout();
}

/**
 * Initialize Auth logic and listeners
 */
export function initAuth() {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('loginUser').value;
            const password = document.getElementById('loginPass').value;
            const errorBox = document.getElementById('loginError');
            const btnText = document.getElementById('loginBtnText');
            const btnLoading = document.getElementById('loginBtnLoading');
            
            if (errorBox) errorBox.classList.add('d-none');
            if (btnText) btnText.classList.add('d-none');
            if (btnLoading) btnLoading.classList.remove('d-none');
            
            try {
                const res = await fetch('/api/login', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({username, password})
                });
                
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.message || '登录失败：用户名或密码错误');
                }
                
                const data = await res.json();
                authToken = data.token;
                authRole = data.role;
                localStorage.setItem('wxbot_token', authToken);
                localStorage.setItem('wxbot_role', authRole);
                
                // Update globals
                window.authToken = authToken;
                window.authRole = authRole;
                
                applyRoleUI(authRole);

                // Hide Login Page
                const loginPage = document.getElementById('loginPage');
                if (loginPage) {
                    loginPage.classList.add('hidden');
                    setTimeout(() => {
                        loginPage.style.display = 'none';
                    }, 500);
                }
                
                // Init App
                if (typeof window.startApp === 'function') {
                    window.startApp();
                }
                
            } catch (err) {
                if (errorBox) {
                    errorBox.innerText = err.message;
                    errorBox.classList.remove('d-none');
                }
            } finally {
                if (btnText) btnText.classList.remove('d-none');
                if (btnLoading) btnLoading.classList.add('d-none');
            }
        });
    }
}

export function logout() {
    // Clear WebSocket connection if exists
    if (window.wsSubscriber) {
        try {
            window.wsSubscriber.close();
        } catch (e) {
            console.error('WebSocket close error:', e);
        }
        window.wsSubscriber = null;
    }
    
    // Clear all intervals
    if (window.updateInterval) clearInterval(window.updateInterval);
    if (window.systemStatsInterval) clearInterval(window.systemStatsInterval);
    if (window.chatStatsInterval) clearInterval(window.chatStatsInterval);
    if (window.botsInterval) clearInterval(window.botsInterval);
    if (window.workersInterval) clearInterval(window.workersInterval);
    if (window.timeUpdateInterval) clearInterval(window.timeUpdateInterval);
    if (window.combinedStatsInterval) clearInterval(window.combinedStatsInterval);
    
    // Clear localStorage
    localStorage.removeItem('wxbot_token');
    localStorage.removeItem('wxbot_role');
    
    // Clear session variables
    window.authToken = null;
    window.authRole = null;
    window.currentBotId = '';
    
    // Force reload with cache clear
    location.href = location.origin + location.pathname + '?logout=1';
}

/**
 * Update user password
 */
export async function updatePassword() {
    const oldPwd = document.getElementById('pwd-current').value;
    const newPwd = document.getElementById('pwd-new').value;
    const confirmPwd = document.getElementById('pwd-confirm').value;

    const t = translations[currentLang] || translations['zh-CN'] || {};

    if (!newPwd) return alert(t.enter_new_pwd || '请输入新密码');
    if (newPwd !== confirmPwd) return alert(t.pwd_mismatch || '两次输入的密码不一致');
    if (!oldPwd) return alert(t.enter_current_pwd || '请输入当前密码');

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

        alert(t.password_change_success || '密码修改成功');
        document.getElementById('pwd-current').value = '';
        document.getElementById('pwd-new').value = '';
        document.getElementById('pwd-confirm').value = '';
    } catch (e) {
        alert((t.password_change_fail || '修改失败: ') + e.message);
    }
}

/**
 * Set Auth Token
 */
export function setAuthToken(token) {
    authToken = token;
    localStorage.setItem('wxbot_token', token);
    window.authToken = token;
}

export function setAuthRole(role) {
    authRole = role;
    localStorage.setItem('wxbot_role', role);
    window.authRole = role;
}

export function setCurrentBotId(id) {
    currentBotId = id;
}

// Global exposure for legacy compatibility
window.authToken = authToken;
window.authRole = authRole;
Object.defineProperty(window, 'currentBotId', {
    get: () => currentBotId,
    set: (v) => { currentBotId = v; },
    configurable: true
});

window.setAuthToken = setAuthToken;
window.setAuthRole = setAuthRole;
window.setCurrentBotId = setCurrentBotId;
window.initAuth = initAuth;
window.handleMagicToken = handleMagicToken;
window.logout = logout;
window.updatePassword = updatePassword;
