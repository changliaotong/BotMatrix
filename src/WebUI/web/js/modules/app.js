import { initCharts, updateStats, updateChatStats, rotateCombinedStats } from './stats.js?v=1.1.87';
import { fetchBots, selectBotFromDropdown } from './bots.js?v=1.1.87';
import { fetchWorkers } from './workers.js?v=1.1.87';
import { fetchRoutingRules } from './routing.js?v=1.1.87';
import { initVisualizer } from './visualization.js?v=1.1.87';
import { loadDockerContainers } from './docker.js?v=1.1.87';
import { fetchUsers } from './admin.js?v=1.1.87';
import { fetchLogs, fetchLogsFull } from './logs.js?v=1.1.87';
import { initLanguage, t } from './i18n.js?v=1.1.87';
import { initWebSocket, closeWebSocket } from './websocket.js?v=1.1.87';
import { updateSystemStats, updateTimeDisplay } from './system.js?v=1.1.87';
import { refreshGroupList } from './groups.js?v=1.1.87';
import { refreshFriendList } from './friends.js?v=1.1.87';
import { initAuth, handleMagicToken } from './auth.js?v=1.1.87';
import { loadBackendConfig, updateRefreshRate } from './config.js?v=1.1.87';
import { applyRoleUI, showToast, showTab } from './ui.js?v=1.1.87';
import { initTheme } from './theme.js?v=1.1.87';
import './debug.js?v=1.1.87'; // Ensure console history and error handlers are initialized

/**
 * 页面刷新时的清理逻辑
 */
export function cleanupBeforeUnload() {
    // 清理WebSocket连接
    if (window.wsSubscriber) {
        try {
            window.wsSubscriber.close();
            window.wsSubscriber = null;
        } catch (e) {
            console.error('清理WebSocket失败:', e);
        }
    }
    
    // 清理所有定时器
    const intervals = [
        'updateInterval', 'systemStatsInterval', 'chatStatsInterval', 
        'botsInterval', 'workersInterval', 'timeUpdateInterval', 'combinedStatsInterval'
    ];
    
    intervals.forEach(intervalName => {
        if (window[intervalName]) {
            clearInterval(window[intervalName]);
            window[intervalName] = null;
        }
    });
    
    // 清理所有待处理的请求
    if (window.pendingRequests) {
        window.pendingRequests.clear();
    }
    
    console.log('页面清理完成');
}

let isAppStarted = false;

/**
 * 启动应用程序
 */
export function startApp() {
    if (isAppStarted) {
        console.log('App already started, skipping redundant startApp()');
        return;
    }
    isAppStarted = true;
    console.log('--- startApp Start ---');
    
    try {
        // Ensure the main UI is visible and not hidden by login page remnants
        const lp = document.getElementById('loginPage');
        if (lp) {
            lp.classList.add('hidden');
            lp.style.display = 'none';
        }

        // Show main app container if it exists (for legacy.html)
        const mainApp = document.querySelector('.main-app');
        if (mainApp) {
            mainApp.style.display = 'block';
        }
        
        // Ensure overlay is hidden
        const overlay = document.getElementById('sidebar-overlay');
        if (overlay) {
            overlay.classList.remove('show');
            overlay.style.display = 'none';
        }
        
        // Ensure sidebar is in correct state for desktop
        const sidebar = document.getElementById('sidebar');
        if (sidebar && window.innerWidth > 768) {
            sidebar.classList.remove('show');
            sidebar.style.transform = '';
        }

        const role = window.authRole || localStorage.getItem('wxbot_role');
        console.log('1. Applying role UI, role:', role);
        applyRoleUI(role);
        
        console.log('2. Calculating initial tab');
        let initialTab = 'dashboard';
        const hash = window.location.hash;
        if (hash) {
            const tabId = hash.substring(1);
            console.log('Found hash tabId:', tabId);
            const validTabs = ['dashboard', 'bots', 'groups', 'friends', 'monitor', 'docker', 'routing', 'logs', 'debug', 'users', 'settings', 'visualization', 'overmind'];
            const isValidTab = validTabs.includes(tabId) && document.getElementById('tab-' + tabId);
            
            if (isValidTab) {
                const adminOnlyTabs = ['debug', 'users', 'settings', 'routing', 'docker', 'overmind'];
                const isAdmin = (role === 'super' || role === 'admin');
                
                if (adminOnlyTabs.includes(tabId) && !isAdmin) {
                    console.log('Restricting access to admin tab:', tabId);
                    initialTab = 'dashboard';
                } else {
                    initialTab = tabId;
                }
            } else {
                console.warn('Invalid tab or missing element for tabId:', tabId);
            }
        }
        
        console.log('3. Showing initial tab:', initialTab);
        try {
            showTab(initialTab);
        } catch (showTabError) {
            console.error('Initial showTab failed:', showTabError);
            if (initialTab !== 'dashboard') {
                showTab('dashboard');
            }
        }

        console.log('4. Scheduling component initializations');
        initCharts();
        
        setTimeout(() => {
            try {
                console.log('Initializing WebSocket...');
                initWebSocket();
            } catch (wsError) {
                console.error('WebSocket初始化失败:', wsError);
            }
        }, 100);
        
        // Fetch basic data immediately
        refreshData();

        // 启动定时器
        console.log('Starting intervals...');
        const rate = parseInt(localStorage.getItem('refresh_rate')) || 2000;
        const rateSelect = document.getElementById('setting-refresh-rate');
        if (rateSelect) rateSelect.value = rate.toString();

        const safeSetInterval = (fn, interval, name) => {
            return setInterval(() => {
                try {
                    if (typeof fn === 'function') {
                        fn();
                    } else {
                        console.warn(`${name} 不是一个有效的函数`);
                    }
                } catch (error) {
                    console.error(`${name}定时器执行失败:`, error);
                }
            }, interval);
        };
        
        // Clear existing intervals if any
        cleanupBeforeUnload();

        window.updateInterval = safeSetInterval(updateStats, rate, 'updateStats');
        window.systemStatsInterval = safeSetInterval(updateSystemStats, rate, 'updateSystemStats');
        window.timeUpdateInterval = safeSetInterval(updateTimeDisplay, 1000, 'updateTimeDisplay');
        window.chatStatsInterval = safeSetInterval(updateChatStats, 5000, 'updateChatStats');
        window.botsInterval = safeSetInterval(fetchBots, 5000, 'fetchBots');
        window.workersInterval = safeSetInterval(fetchWorkers, 5000, 'fetchWorkers');
        window.combinedStatsInterval = safeSetInterval(rotateCombinedStats, 5000, 'rotateCombinedStats');
        
        // 7. Add event listeners for cleanup
        window.addEventListener('beforeunload', cleanupBeforeUnload);
        window.addEventListener('pagehide', cleanupBeforeUnload);
        
        // 8. Add visibility change listener
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                console.log('[App] Page hidden, some background tasks could be paused');
            } else {
                console.log('[App] Page visible, resuming background tasks');
                refreshData(); // Refresh data when page becomes visible again
            }
        });

        console.log('[App] BotMatrix Overmind initialized successfully.');
    } catch (error) {
        console.error('startApp初始化失败:', error);
        const lp = document.getElementById('loginPage');
        if (lp) {
            lp.style.display = 'flex';
            lp.classList.remove('hidden');
        }
        alert((t('startup_error') || '启动失败') + ': ' + error.message);
    }
}



function showLoginPage() {
    const lp = document.getElementById('loginPage');
    if (lp) {
        lp.style.display = 'flex';
        lp.classList.remove('hidden');
    }
    const mainApp = document.querySelector('.main-app');
    if (mainApp) mainApp.style.display = 'none';
}

/**
 * 刷新所有数据
 */
export async function refreshData() {
    console.log('Refreshing data...');
    try {
        await Promise.allSettled([
            updateStats(),
            updateChatStats(),
            fetchBots(true),
            fetchWorkers(true),
            fetchRoutingRules(),
            loadDockerContainers(true),
            fetchUsers(),
            fetchLogs(),
            updateSystemStats()
        ]);
    } catch (err) {
        console.error('Data refresh error:', err);
    }
}

/**
 * 刷新当前标签页列表
 */
export function refreshCurrentTabList() {
    const groupsTab = document.getElementById('tab-groups');
    const friendsTab = document.getElementById('tab-friends');
    
    if (friendsTab && friendsTab.classList.contains('active-tab')) {
        refreshFriendList();
    } else if (groupsTab && groupsTab.classList.contains('active-tab')) {
        refreshGroupList();
    }
}

/**
 * 设置全局机器人并切换到群组
 */
export function setGlobalBot(id) {
    selectBotFromDropdown(id);
    showTab('groups');
}

/**
 * 页面加载完成后的初始化
 */
document.addEventListener('DOMContentLoaded', () => {
    console.log('--- DOMContentLoaded Triggered ---');
    
    // Helper function for initialization that requires loginPage
    const initializeWithLp = (lp) => {
        console.log('--- DOMContentLoaded Initializing with loginPage ---');
        try {
            console.log('1. Initializing theme and language...');
            try {
                initTheme();
            } catch (themeErr) {
                console.error('Theme init error:', themeErr);
            }
            
            try {
                initLanguage();
            } catch (langErr) {
                console.error('Language init error:', langErr);
            }

            // Resource checks
            if (typeof bootstrap === 'undefined') console.warn('Bootstrap JS failed to load');
            if (typeof Chart === 'undefined') console.warn('Chart.js failed to load');
            if (typeof THREE === 'undefined') console.warn('Three.js failed to load');

            // Initialize UI
            try {
                initUI();
            } catch (uiErr) {
                console.error('UI init error:', uiErr);
            }

            // Force logout if requested via URL parameter
            const urlParams = new URLSearchParams(window.location.search);
            if (urlParams.get('logout') === '1') {
                console.log('Logout parameter detected, performing logout...');
                localStorage.removeItem('wxbot_token');
                localStorage.removeItem('wxbot_role');
                window.authToken = null;
                window.authRole = 'user';
                
                if (lp) {
                    lp.style.display = 'flex';
                    lp.classList.remove('hidden');
                }
                
                // Remove the logout parameter from URL
                window.history.replaceState({}, document.title, window.location.pathname);
                
                clearTimeout(emergencyTimeout);
                return;
            }

            // Magic Token Login (Support 'magic', 'token', and 'magic_token' params)
            const magicToken = urlParams.get('magic') || urlParams.get('token') || urlParams.get('magic_token');
            if (magicToken) {
                console.log('Magic token detected, attempting login...');
                handleMagicToken(magicToken).then(res => {
                    if (res.success) {
                        startApp();
                    } else {
                        lp.style.display = 'flex';
                        lp.classList.remove('hidden');
                    }
                });
                clearTimeout(emergencyTimeout);
                return;
            }

            // Check for existing token
            const token = window.authToken || localStorage.getItem('wxbot_token');
            if (token && token !== 'undefined' && token !== 'null') {
                console.log('Existing token found, starting app...');
                startApp();
            } else {
                console.log('No valid token found, showing login page');
                lp.style.display = 'flex';
                lp.classList.remove('hidden');
            }
            
            // Initialize Auth listeners
            initAuth();
            
            // If app started, clear timeout immediately
            if (isAppStarted) {
                clearTimeout(emergencyTimeout);
            }
        } catch (err) {
            console.error('DOMContentLoaded initialization error:', err);
            if (lp) {
                lp.style.display = 'flex';
                lp.classList.remove('hidden');
            }
            clearTimeout(emergencyTimeout);
        }
    };

    const lp = document.getElementById('loginPage');
    
    // Set an emergency timeout: if after 5 seconds the page is still blank (neither main UI nor login page shown), show the login page
    const emergencyTimeout = setTimeout(() => {
        // If app already started, do nothing
        if (isAppStarted) return;

        const currentLp = document.getElementById('loginPage');
        const dashboard = document.getElementById('tab-dashboard');
        
        // Use getComputedStyle to check actual visibility as style.display might be empty if set via CSS
        const isLpVisible = currentLp && window.getComputedStyle(currentLp).display !== 'none' && window.getComputedStyle(currentLp).visibility !== 'hidden' && window.getComputedStyle(currentLp).opacity !== '0';
        const isDashboardVisible = dashboard && window.getComputedStyle(dashboard).display !== 'none' && window.getComputedStyle(dashboard).visibility !== 'hidden' && window.getComputedStyle(dashboard).opacity !== '0';
        
        if (!isLpVisible && !isDashboardVisible) {
            console.error('Emergency Timeout: Page remains blank after 5s. Force showing login page.');
            if (currentLp) {
                currentLp.style.display = 'flex';
                currentLp.classList.remove('hidden');
                
                // Check if an emergency alert already exists
                if (!document.getElementById('emergency-alert')) {
                    const alertDiv = document.createElement('div');
                    alertDiv.id = 'emergency-alert';
                    alertDiv.className = 'alert alert-warning m-3';
                    alertDiv.style.cssText = 'position:absolute;top:0;left:0;right:0;z-index:10001;box-shadow:0 5px 15px rgba(0,0,0,0.2)';
                    
                    const text = document.createElement('div');
                    if (document.documentElement.classList.contains('resource-error')) {
                        text.innerText = '检测到部分核心资源(如 Chart.js/Three.js)加载失败，已强制显示登录页。';
                    } else {
                        text.innerText = '系统加载超时 (5s)，已强制显示登录页。';
                    }
                    alertDiv.appendChild(text);
                    
                    const subtext = document.createElement('small');
                    subtext.className = 'd-block mt-1 text-muted';
                    if (document.documentElement.classList.contains('resource-error')) {
                        subtext.innerText = '请检查网络连接、防火墙或代理设置，确保 cdn.staticfile.org 可访问。';
                    } else {
                        subtext.innerText = '如果问题持续，请检查网络连接或刷新页面。';
                    }
                    alertDiv.appendChild(subtext);

                    const btnContainer = document.createElement('div');
                    btnContainer.className = 'mt-2';
                    
                    const retryBtn = document.createElement('button');
                    retryBtn.className = 'btn btn-sm btn-primary me-2';
                    retryBtn.innerText = '刷新页面';
                    retryBtn.onclick = () => window.location.reload();
                    btnContainer.appendChild(retryBtn);
                    
                    const diagBtn = document.createElement('button');
                    diagBtn.className = 'btn btn-sm btn-outline-secondary';
                    diagBtn.innerText = '查看诊断日志';
                    diagBtn.onclick = () => {
                        const logs = window.console.history || ['No logs captured. Check browser console (F12).'];
                        alert('诊断日志 (前10条):\n' + logs.slice(-10).join('\n'));
                    };
                    btnContainer.appendChild(diagBtn);
                    
                    alertDiv.appendChild(btnContainer);
                    currentLp.appendChild(alertDiv);
                }
            }
        }
    }, 5000);

    if (!lp) {
        console.log('loginPage not found in DOM initially. Waiting for potential async rendering...');
        
        // If we have a token, we might still be able to start the app immediately
        const token = window.authToken || localStorage.getItem('wxbot_token');
        if (token && token !== 'undefined' && token !== 'null') {
            console.log('Existing token found, starting app without waiting for loginPage...');
            startApp();
            clearTimeout(emergencyTimeout);
            return;
        }

        // Use MutationObserver to wait for loginPage (handles Vue race condition in index.html)
        const observer = new MutationObserver((mutations, obs) => {
            const foundLp = document.getElementById('loginPage');
            if (foundLp) {
                obs.disconnect();
                console.log('loginPage detected via MutationObserver.');
                initializeWithLp(foundLp);
            }
        });
        observer.observe(document.body, { childList: true, subtree: true });
        
        // Also check if we are in a legacy page that doesn't use Vue
        // If it's not index.html and no loginPage after 1s, it's likely missing entirely
        setTimeout(() => {
            if (!document.getElementById('loginPage') && !isAppStarted) {
                console.warn('loginPage still not found after 1s. This page might be missing required UI elements.');
                // For legacy.html, we should still try to init theme/lang
                try { initTheme(); initLanguage(); } catch(e) {}
            }
        }, 1000);
    } else {
        initializeWithLp(lp);
    }
});

// Global exposure
window.cleanupBeforeUnload = cleanupBeforeUnload;
window.startApp = startApp;
window.refreshData = refreshData;
window.refreshCurrentTabList = refreshCurrentTabList;
window.setGlobalBot = setGlobalBot;
