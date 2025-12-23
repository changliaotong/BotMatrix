/**
 * UI Management Module
 */

import { authRole, authToken } from './auth.js';
import { currentLang, setLanguage } from './i18n.js';
import { fetchUsers } from './admin.js';
import { loadBackendConfig } from './config.js';
import { serverStartTime } from './stats.js';
import { fetchRoutingRules } from './routing.js';
import { loadDockerContainers } from './docker.js';
import { fetchBots, currentBots, currentBotId } from './bots.js';
import { fetchWorkers, currentWorkers } from './workers.js';
import { refreshFriendList, currentFriends } from './friends.js';
import { refreshGroupList, currentGroups } from './groups.js';
import { fetchLogs, fetchLogsFull } from './logs.js';
import { loadTestGroups } from './test.js';
import { initVisualizer } from './visualization.js';
import { findBotInfo, formatUserAvatar, formatBytes, timeAgo } from './utils.js';

export function showTab(tabId) {
    if (!tabId) return;

    // Handle #hash
    if (tabId.startsWith('#')) tabId = tabId.substring(1);
    // Handle tab- prefix
    if (tabId.startsWith('tab-')) tabId = tabId.replace('tab-', '');
    
    try {
        // Handle visualization init (Moved below content section handling for cleaner logic)
        
        // 隐藏所有内容区域
        const contentSections = document.querySelectorAll('.main-content > .tab-content');
        contentSections.forEach(section => {
            section.classList.remove('active-tab');
        });
        
        // 显示目标内容区域
        const targetTab = document.getElementById('tab-' + tabId);
        if (targetTab) {
            targetTab.classList.add('active-tab');
            
            // Visualization resize after tab becomes visible
            if (tabId === 'visualization' && window.visualizer) {
                setTimeout(() => window.visualizer.resize(), 50);
            }
        }
        
        // 更新导航状态
        const navLinks = document.querySelectorAll('.sidebar .nav-link');
        navLinks.forEach(link => {
            link.classList.remove('active');
        });
        
        const targetNav = document.getElementById('nav-' + tabId) || 
                          document.querySelector(`.sidebar .nav-link[href="#${tabId}"]`);
        if (targetNav) {
            targetNav.classList.add('active');
        }
        
        // 更新URL hash (使用 replaceState 避免触发 hashchange 事件)
        const newHash = '#' + tabId;
        if (window.location.hash !== newHash) {
            window.history.replaceState(null, null, newHash);
        }
        
        // 根据tab触发相应的初始化函数
        switch(tabId) {
            case 'dashboard':
                // 仪表板数据在 startApp 中定时更新
                break;
            case 'bots':
                fetchBots(window.currentBots && window.currentBots.length === 0);
                fetchWorkers(window.currentWorkers && window.currentWorkers.length === 0);
                break;
            case 'groups':
                const groupsBotId = document.getElementById('global-bot-selector-groups')?.value || window.currentBotId;
                if (groupsBotId && (!window.currentGroups || window.currentGroups.length === 0 || (window.currentGroups[0] && window.currentGroups[0].bot_id !== groupsBotId))) {
                    refreshGroupList(groupsBotId);
                }
                break;
            case 'friends':
                const friendsBotId = document.getElementById('global-bot-selector-friends')?.value || window.currentBotId;
                if (friendsBotId && (!window.currentFriends || window.currentFriends.length === 0 || (window.currentFriends[0] && window.currentFriends[0].bot_id !== friendsBotId))) {
                    refreshFriendList(friendsBotId);
                }
                break;
            case 'monitor':
                const activeMonitorTab = document.querySelector('#monitor-tabs .nav-link.active');
                if (activeMonitorTab && activeMonitorTab.id === 'monitor-logs-tab') {
                    fetchLogs();
                }
                break;
            case 'docker':
                loadDockerContainers();
                break;
            case 'routing':
                fetchRoutingRules();
                break;
            case 'logs':
                fetchLogsFull();
                break;
            case 'debug':
                loadTestGroups();
                break;
            case 'users':
                fetchUsers();
                break;
            case 'settings':
                loadBackendConfig();
                break;
            case 'visualization':
                initVisualizer();
                break;
        }

        // 移动端关闭侧边栏
        if (window.innerWidth <= 768) {
            const sidebar = document.getElementById('sidebar');
            const overlay = document.getElementById('sidebar-overlay');
            if (sidebar) sidebar.classList.remove('show');
            if (overlay) overlay.classList.remove('show');
        }
        
        console.log('切换到标签页:', tabId);
    } catch (e) {
        console.error('Error in showTab:', e);
    }
}

export function toggleFullScreen() {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().catch(err => {
            console.error(`Error attempting to enable full-screen mode: ${err.message}`);
        });
    } else {
        if (document.exitFullscreen) {
            document.exitFullscreen();
        }
    }
}

/**
 * Open Overmind (Matrix-style management)
 */
export function openOvermind() {
    const lang = currentLang || 'zh-CN';
    const langParam = lang === 'zh-CN' ? 'zh' : 'en';
    const token = authToken || localStorage.getItem('wxbot_token');
    const url = `/overmind/?lang=${langParam}&token=${encodeURIComponent(token)}`;
    window.open(url, '_blank');
}

/**
 * Show toast notification
 */
export function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        const container = document.createElement('div');
        container.id = 'toast-container';
        container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 10000;';
        document.body.appendChild(container);
    }
    
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} fade show`;
    toast.style.cssText = 'min-width: 200px; margin-bottom: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);';
    toast.innerHTML = message;
    
    document.getElementById('toast-container').appendChild(toast);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 500);
    }, 3000);
}

/**
 * Initialize Sidebar
 */

export function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('sidebar-overlay');
    if (sidebar && overlay) {
        sidebar.classList.toggle('show');
        overlay.classList.toggle('show');
        if (overlay.classList.contains('show')) {
            overlay.style.display = 'block';
        } else {
            setTimeout(() => {
                if (!overlay.classList.contains('show')) overlay.style.display = 'none';
            }, 300);
        }
    }
}

/**
 * Apply role-based UI visibility
 */
export function applyRoleUI(role) {
    console.log('[UI] Applying role UI:', role);
    const isAdmin = (role === 'super' || role === 'admin');
    
    // Helper to toggle visibility
    const toggle = (id, show) => {
        const el = document.getElementById(id);
        if(el) {
            el.style.display = show ? '' : 'none';
        }
    };

    // Sidebar items
    toggle('nav-debug', isAdmin);
    toggle('nav-users', isAdmin);
    toggle('nav-settings', isAdmin);
    toggle('nav-routing', isAdmin);
    toggle('nav-docker', isAdmin);
    toggle('nav-overmind', isAdmin);
    
    // Admin only sections by class
    document.querySelectorAll('.admin-only').forEach(el => {
        el.style.display = isAdmin ? '' : 'none';
    });
    
    // Monitor Tabs
    toggle('monitor-logs-tab', isAdmin);
    
    // Dashboard metrics
    toggle('dash-system-stats', true);
    toggle('card-metric-bots', true);
    toggle('card-metric-workers', true);
    toggle('card-metric-combined', true);
    toggle('card-metric-time', true);
    toggle('card-metric-system', isAdmin);
    
    // Dashboard cards
    toggle('dash-process-card', isAdmin);
    toggle('dash-actions-card', isAdmin);
    toggle('dash-logs-card', isAdmin);
    
    // If user is not admin and is currently on a hidden tab, switch to dashboard
    if (!isAdmin) {
        const adminOnlyTabs = ['debug', 'users', 'settings', 'routing', 'docker', 'overmind'];
        const hash = window.location.hash;
        if (hash && adminOnlyTabs.includes(hash.substring(1))) {
            showTab('dashboard');
        }
    }

    // Toggle role labels
    const roleLabel = document.getElementById('auth-role-label');
    if (roleLabel) {
        roleLabel.innerText = role.toUpperCase();
        roleLabel.className = 'badge ' + (isAdmin ? 'bg-danger' : 'bg-primary');
    }
}

export function showOvermind(e) {
    if (e) e.preventDefault();
    showTab('overmind');
    openOvermind();
}

export function setBotViewMode(mode) {
    window.botViewMode = mode;
    window._botViewModeManuallySet = true;
    localStorage.setItem('bot_view_mode', mode);
    updateViewModeUI('bot', mode);
    if (typeof window.renderBots === 'function') window.renderBots();
}

export function setWorkerViewMode(mode) {
    window.workerViewMode = mode;
    window._workerViewModeManuallySet = true;
    localStorage.setItem('worker_view_mode', mode);
    updateViewModeUI('worker', mode);
    if (typeof window.renderWorkers === 'function') window.renderWorkers();
}

export function updateViewModeUI(type, mode) {
    const btnDetail = document.getElementById(`btn-${type}-view-detail`);
    const btnCompact = document.getElementById(`btn-${type}-view-compact`);
    if (btnDetail && btnCompact) {
        if (mode === 'detail') {
            btnDetail.classList.add('active');
            btnCompact.classList.remove('active');
        } else {
            btnDetail.classList.remove('active');
            btnCompact.classList.add('active');
        }
    }
}

export function safeInit(id, callback) {
    const el = document.getElementById(id);
    if (el) {
        try {
            callback(el);
        } catch (e) {
            console.error(`Error initializing component ${id}:`, e);
        }
    }
}

export function initUI() {
    // Mobile sidebar handling
    const sidebarToggle = document.getElementById('sidebarToggle');
    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', () => {
            const sidebar = document.getElementById('sidebar');
            if (sidebar) {
                const bsOffcanvas = new bootstrap.Offcanvas(sidebar);
                bsOffcanvas.toggle();
            }
        });
    }

    // Handle resize events
    window.addEventListener('resize', () => {
        if (window.innerWidth > 991.98) {
            const sidebar = document.getElementById('sidebar');
            if (sidebar) {
                const bsOffcanvas = bootstrap.Offcanvas.getInstance(sidebar);
                if (bsOffcanvas) bsOffcanvas.hide();
            }
        }
    });

    // Visibility change handling
    document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
            console.log('[UI] Page hidden, pausing heavy updates');
        } else {
            console.log('[UI] Page visible, resuming updates');
        }
    });
}

export function updateWsStatus(status, text) {
    const el = document.getElementById('ws-status');
    if (!el) return;
    
    let className = 'badge ms-2 ';
    switch (status) {
        case 'success': className += 'bg-success'; break;
        case 'warning': className += 'bg-warning'; break;
        case 'danger': className += 'bg-danger'; break;
        default: className += 'bg-secondary';
    }
    
    el.className = className;
    el.innerText = text;
}
