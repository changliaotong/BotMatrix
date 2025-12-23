/**
 * Main Application Entry Point
 */

import './modules/utils.js';
import { checkAuth, authRole, logout, updatePassword, initAuth } from './modules/auth.js';
import { initWebSocket, closeWebSocket } from './modules/websocket.js';
import { fetchBots, renderBots, updateGlobalBotSelectors } from './modules/bots.js';
import { fetchWorkers } from './modules/workers.js';
import { updateStats, initCharts, updateSystemStats, updateChatStats, rotateCombinedStats, fetchSystemStats } from './modules/stats.js';
import { showTab, applyRoleUI, toggleSidebar, showToast } from './modules/ui.js';
import { loadDockerContainers } from './modules/docker.js';
import { fetchRoutingRules } from './modules/routing.js';
import { fetchUsers } from './modules/admin.js';
import { initLanguage, currentLang, t } from './modules/i18n.js';
import { initVisualizer, handleRoutingEvent, handleSyncState, clearVisualization, toggleFullScreen, openOvermind } from './modules/visualization.js';

let isAppStarted = false;
let serverStartTime = null;

async function startApp() {
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

        console.log('1. Checking Authentication...');
        const isAuthenticated = await checkAuth();
        if (!isAuthenticated) {
            console.log('[App] Not authenticated, redirecting to login...');
            isAppStarted = false; // Reset so it can be retried after login
            return;
        }

        console.log('2. Applying role UI, role:', authRole);
        applyRoleUI(authRole);
        
        console.log('3. Calculating initial tab');
        let initialTab = 'dashboard';
        const hash = window.location.hash;
        if (hash) {
            const tabId = hash.substring(1);
            console.log('Found hash tabId:', tabId);
            const validTabs = ['dashboard', 'bots', 'groups', 'friends', 'monitor', 'docker', 'routing', 'logs', 'debug', 'users', 'settings', 'visualization', 'overmind'];
            const isValidTab = validTabs.includes(tabId) && document.getElementById('tab-' + tabId);
            
            if (isValidTab) {
                const adminOnlyTabs = ['debug', 'users', 'settings', 'routing', 'docker', 'overmind'];
                const isAdmin = (authRole === 'super' || authRole === 'admin');
                
                if (adminOnlyTabs.includes(tabId) && !isAdmin) {
                    console.log('Restricting access to admin tab:', tabId);
                    initialTab = 'dashboard';
                } else {
                    initialTab = tabId;
                }
            }
        }
        
        console.log('4. Showing initial tab:', initialTab);
        try {
            showTab(initialTab);
        } catch (showTabError) {
            console.error('Initial showTab failed:', showTabError);
            if (initialTab !== 'dashboard') {
                showTab('dashboard');
            }
        }

        console.log('5. Scheduling component initializations');
        initCharts();
        
        setTimeout(() => {
            try {
                initWebSocket();
            } catch (wsError) {
                console.error('WebSocket初始化失败:', wsError);
            }
        }, 100);
        
        // Fetch basic data immediately
        updateStats();
        fetchBots();

        // Sub-tasks with delays
        setTimeout(() => updateSystemStats(), 200);
        setTimeout(() => fetchWorkers(), 400);
        setTimeout(() => updateChatStats(), 600);
        
        // Admin-only initial data
        if (authRole === 'admin' || authRole === 'super') {
            setTimeout(() => {
                loadDockerContainers(true);
                fetchRoutingRules();
                fetchUsers();
            }, 800);
        }

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
        
        console.log('App initialization completed successfully.');
    } catch (error) {
        console.error('startApp初始化失败:', error);
        isAppStarted = false;
        const lp = document.getElementById('loginPage');
        if (lp) {
            lp.style.display = 'flex';
            lp.classList.remove('hidden');
        }
        alert(t('startup_error') || '应用初始化失败，请刷新页面重试: ' + error.message);
    }
}

function updateTimeDisplay() {
    if (document.getElementById('metric-current-time')) {
        const now = new Date();
        const lang = currentLang || 'zh-CN';
        const dateStr = now.toLocaleDateString(lang, { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-');
        const timeStr = now.toLocaleTimeString(lang, { hour12: false });
        document.getElementById('metric-current-time').innerText = `${dateStr} ${timeStr}`;
    }
    if (document.getElementById('metric-uptime') && serverStartTime) {
        const uptimeSeconds = Math.floor(Date.now() / 1000 - serverStartTime);
        const d = Math.floor(uptimeSeconds / 86400);
        const h = Math.floor((uptimeSeconds % 86400) / 3600);
        const m = Math.floor((uptimeSeconds % 3600) / 60);
        const s = uptimeSeconds % 60;
        
        const dStr = d > 0 ? `${d}${t('time_days') || 'd'} ` : '';
        const hStr = h.toString().padStart(2, '0');
        const mStr = m.toString().padStart(2, '0');
        const sStr = s.toString().padStart(2, '0');
        
        document.getElementById('metric-uptime').innerHTML = 
            `${dStr}${hStr}<span class="time-sep">:</span>${mStr}<span class="time-sep">:</span>${sStr}`;
    }
}

function cleanupBeforeUnload() {
    console.log('[App] Cleaning up...');
    closeWebSocket();
    const intervals = [
        'updateInterval', 'systemStatsInterval', 'timeUpdateInterval', 
        'chatStatsInterval', 'botsInterval', 'workersInterval', 'combinedStatsInterval'
    ];
    intervals.forEach(intervalName => {
        if (window[intervalName]) {
            clearInterval(window[intervalName]);
            window[intervalName] = null;
        }
    });
}

// Global exposure for legacy compatibility
window.startApp = startApp;
window.updateTimeDisplay = updateTimeDisplay;
window.cleanupBeforeUnload = cleanupBeforeUnload;
window.showTab = showTab;
window.logout = logout;
window.updatePassword = updatePassword;
window.showToast = showToast;
window.toggleSidebar = toggleSidebar;
window.initAuth = initAuth;
window.initVisualizer = initVisualizer;
window.handleRoutingEvent = handleRoutingEvent;
window.handleSyncState = handleSyncState;
window.clearVisualization = clearVisualization;
window.toggleFullScreen = toggleFullScreen;
window.openOvermind = openOvermind;

// Start initialization
document.addEventListener('DOMContentLoaded', () => {
    initLanguage();
    initAuth();
    // startApp will be called inside initAuth or manually if needed
});

window.addEventListener('beforeunload', cleanupBeforeUnload);
window.addEventListener('pagehide', cleanupBeforeUnload);
