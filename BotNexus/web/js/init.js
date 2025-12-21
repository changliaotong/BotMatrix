/**
 * Application Initialization Entry Point
 */

import { initShield } from './modules/shield.js?v=1.1.87';

// 0. Debug & Error Handling (Run immediately)
initShield();

// 1. Core Utilities
import { initTheme, toggleTheme } from './modules/theme.js?v=1.1.87';
import { initLanguage, setLanguage, currentLang, translations } from './modules/i18n.js?v=1.1.87';
import { handleLogin, handleLogout, handleMagicToken, authToken as auth_token_val, authRole as auth_role_val } from './modules/auth.js?v=1.1.87';
import { callBotApi, sendAction } from './modules/api.js?v=1.1.87';

// 2. Business Modules
import { fetchBots, toggleBotState, deleteBot, openBotConfig, saveBotConfig, filterBots, sortBots, setBotViewMode } from './modules/bots.js?v=1.1.87';
import { fetchWorkers, filterWorkers, sortWorkers, setWorkerViewMode } from './modules/workers.js?v=1.1.87';
import { refreshGroupList, selectGroup, filterGroups, sortGroups, toggleAutoRecallInput, sendGroupMsg, sendSmartGroupMsg, currentContactType as current_contact_type_val } from './modules/groups.js?v=1.1.87';
import { refreshFriendList, filterFriends, sortFriends, sendFriendMsg, selectFriend } from './modules/friends.js?v=1.1.87';
import { loadGroupMembers, renderMembers, sortMembers, banMember, unbanMember, setCard, kickMember, leaveGroup, checkGroupMember } from './modules/members.js?v=1.1.87';
import { updateStats, updateChatStats, showAllStats, latestChatStats, rotateCombinedStats, saveStatsToCache, loadStatsFromCache, initCharts } from './modules/stats.js?v=1.1.87';
import { formatBytes } from './modules/utils.js?v=1.1.87';
import { addEventLog, clearEvents, checkLogSelectionAndPause, fetchLogs, fetchLogsFull } from './modules/logs.js?v=1.1.87';
import { loadDockerContainers, controlContainer, addBotContainer, addWorkerContainer, filterDockerContainers } from './modules/docker.js?v=1.1.87';
import { fetchRoutingRules, showAddRoutingRuleDialog, saveRoutingRule, deleteRoutingRule, editRoutingRule, toggleRoutingHelp } from './modules/routing.js?v=1.1.87';
import { showMassSendModal, toggleSelectAll, executeMassSend } from './modules/massMsg.js?v=1.1.87';
import { fetchUsers, showCreateUserModal, createUser, deleteUser, resetUserPassword } from './modules/admin.js?v=1.1.87';
import { initVisualizer, handleRoutingEvent, handleSyncState, clearVisualization, toggleVisualizerFullScreen } from './modules/visualization.js?v=1.1.87';
import { initWebSocket, closeWebSocket } from './modules/websocket.js?v=1.1.87';
import { showTab, toggleFullScreen, initUI, applyRoleUI, toggleSidebar, openOvermind, showToast } from './modules/ui.js?v=1.1.87';
import { callSystemAction, showSystemDetail, showSystemDetails, updateDetailTimeRange, switchDetailTab, updateTimeDisplay, fetchSystemStats, updateSystemStats } from './modules/system.js?v=1.1.87';
import { cleanupBeforeUnload } from './modules/app.js?v=1.1.87';
import { loadBackendConfig, updateBackendConfig, updatePassword } from './modules/config.js?v=1.1.87';
import { loadTestGroups, onTestGroupSelectChange, updateCodePreview, pasteTargetUid, submitTestMsg, copyCodePreview } from './modules/test.js?v=1.1.87';
import { currentBotId as current_bot_id_val, currentBots as current_bots_val, selectBotFromDropdown } from './modules/bots.js?v=1.1.87';

// 3. Main App
import { startApp } from './modules/app.js?v=1.1.87';

// Global Exposure for legacy HTML handlers
window.showToast = showToast;
window.initTheme = initTheme;
window.toggleTheme = toggleTheme;
window.initLanguage = initLanguage;
window.setLanguage = setLanguage;
window.changeLanguage = setLanguage; // Alias
window.translations = translations;
Object.defineProperty(window, 'currentLang', {
    get: () => currentLang,
    set: (v) => setLanguage(v)
});
window.handleLogin = handleLogin;
window.handleLogout = handleLogout;
window.logout = handleLogout; // Alias
window.authToken = auth_token_val;
window.authRole = auth_role_val;

// Expose currentBotId and currentContactType with getters/setters for legacy compatibility
Object.defineProperty(window, 'currentBotId', {
    get: () => current_bot_id_val,
    set: (v) => { /* Only read-only or handled via selectBotFromDropdown */ }
});
Object.defineProperty(window, 'currentContactType', {
    get: () => current_contact_type_val,
    set: (v) => { /* Read-only or handled via module logic */ }
});

window.callBotApi = callBotApi;
window.sendAction = sendAction;
window.fetchBots = fetchBots;
window.refreshBotList = fetchBots; // Alias
window.selectBotFromDropdown = selectBotFromDropdown;
window.toggleBotState = toggleBotState;
window.deleteBot = deleteBot;
window.openBotConfig = openBotConfig;
window.saveBotConfig = saveBotConfig;
window.filterBots = filterBots;
window.sortBots = sortBots;
window.setBotViewMode = setBotViewMode;
window.fetchWorkers = fetchWorkers;
window.refreshWorkerList = fetchWorkers; // Alias
window.filterWorkers = filterWorkers;
window.sortWorkers = sortWorkers;
window.setWorkerViewMode = setWorkerViewMode;
window.refreshGroupList = refreshGroupList;
window.selectGroup = selectGroup;
window.filterGroups = filterGroups;
window.sortGroups = sortGroups;
window.toggleAutoRecallInput = toggleAutoRecallInput;
window.sendGroupMsg = sendGroupMsg;
window.sendSmartGroupMsg = sendSmartGroupMsg;
window.refreshFriendList = refreshFriendList;
window.selectFriend = selectFriend;
window.filterFriends = filterFriends;
window.sortFriends = sortFriends;
window.sendFriendMsg = sendFriendMsg;
window.loadGroupMembers = loadGroupMembers;
window.renderMembers = renderMembers;
window.sortMembers = sortMembers;
window.banMember = banMember;
window.unbanMember = unbanMember;
window.setCard = setCard;
window.kickMember = kickMember;
window.leaveGroup = leaveGroup;
window.checkGroupMember = checkGroupMember;
window.updateStats = updateStats;
window.updateChatStats = updateChatStats;
window.showAllStats = showAllStats;
window.latestChatStats = latestChatStats;
window.rotateCombinedStats = rotateCombinedStats;
window.formatBytes = formatBytes;
window.saveStatsToCache = saveStatsToCache;
window.loadStatsFromCache = loadStatsFromCache;
window.initCharts = initCharts;
window.addEventLog = addEventLog;
window.clearEvents = clearEvents;
window.checkLogSelectionAndPause = checkLogSelectionAndPause;
window.fetchLogs = fetchLogs;
window.fetchLogsFull = fetchLogsFull;
window.loadDockerContainers = loadDockerContainers;
window.controlContainer = controlContainer;
window.callDockerAction = controlContainer; // Alias for legacy
window.addBotContainer = addBotContainer;
window.addWorkerContainer = addWorkerContainer;
window.filterDockerContainers = filterDockerContainers;
window.fetchRoutingRules = fetchRoutingRules;
window.showAddRoutingRuleDialog = showAddRoutingRuleDialog;
window.saveRoutingRule = saveRoutingRule;
window.deleteRoutingRule = deleteRoutingRule;
window.editRoutingRule = editRoutingRule;
window.toggleRoutingHelp = toggleRoutingHelp;
window.showMassSendModal = showMassSendModal;
window.toggleSelectAll = toggleSelectAll;
window.executeMassSend = executeMassSend;
window.fetchUsers = fetchUsers;
window.showCreateUserModal = showCreateUserModal;
window.createUser = createUser;
window.deleteUser = deleteUser;
window.resetUserPassword = resetUserPassword;
window.initVisualizer = initVisualizer;
window.handleRoutingEvent = handleRoutingEvent;
window.handleSyncState = handleSyncState;
window.clearVisualization = clearVisualization;
window.toggleVisualizerFullScreen = toggleVisualizerFullScreen;
window.initWebSocket = initWebSocket;
window.closeWebSocket = closeWebSocket;
window.showTab = showTab;
window.toggleFullScreen = toggleFullScreen;
window.toggleSidebar = toggleSidebar;
window.openOvermind = openOvermind;
window.initUI = initUI;
window.applyRoleUI = applyRoleUI;
window.callSystemAction = callSystemAction;
window.cleanupBeforeUnload = cleanupBeforeUnload;
window.showSystemDetail = showSystemDetail;
window.showSystemDetails = showSystemDetails;
window.updateDetailTimeRange = updateDetailTimeRange;
window.switchDetailTab = switchDetailTab;
window.updateTimeDisplay = updateTimeDisplay;
window.fetchSystemStats = fetchSystemStats;
window.updateSystemStats = updateSystemStats;
window.loadBackendConfig = loadBackendConfig;
window.updateBackendConfig = updateBackendConfig;
window.updatePassword = updatePassword;
window.loadTestGroups = loadTestGroups;
window.onTestGroupSelectChange = onTestGroupSelectChange;
window.updateCodePreview = updateCodePreview;
window.pasteTargetUid = pasteTargetUid;
window.submitTestMsg = submitTestMsg;
window.copyCodePreview = copyCodePreview;
window.startApp = startApp;

// DOM Ready initialization
document.addEventListener('DOMContentLoaded', () => {
    console.log('--- System Entry Point Initializing ---');
    
    const lp = document.getElementById('loginPage');
    
    // 1. Emergency Timeout (from index.html)
    const emergencyTimeout = setTimeout(() => {
        const lp = document.getElementById('loginPage');
        const dashboard = document.getElementById('tab-dashboard');
        const isLpVisible = lp && window.getComputedStyle(lp).display !== 'none' && window.getComputedStyle(lp).visibility !== 'hidden' && window.getComputedStyle(lp).opacity !== '0';
        const isDashboardVisible = dashboard && window.getComputedStyle(dashboard).display !== 'none' && window.getComputedStyle(dashboard).visibility !== 'hidden' && window.getComputedStyle(dashboard).opacity !== '0';
        
        if (!isLpVisible && !isDashboardVisible) {
            console.error('Emergency Timeout: Page remains blank after 5s. Force showing login page.');
            if (lp) {
                lp.style.display = 'flex';
                lp.classList.remove('hidden');
                
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
                    subtext.innerText = '如果问题持续，请检查网络连接或刷新页面。';
                    alertDiv.appendChild(subtext);

                    const btnContainer = document.createElement('div');
                    btnContainer.className = 'mt-2';
                    
                    const retryBtn = document.createElement('button');
                    retryBtn.className = 'btn btn-sm btn-primary me-2';
                    retryBtn.innerText = '刷新页面';
                    retryBtn.onclick = () => window.location.reload();
                    btnContainer.appendChild(retryBtn);
                    
                    alertDiv.appendChild(btnContainer);
                    lp.appendChild(alertDiv);
                }
            }
        }
    }, 5000);

    // 2. Resource checks
    if (typeof bootstrap === 'undefined') console.warn('Bootstrap JS failed to load');
    if (typeof Chart === 'undefined') console.warn('Chart.js failed to load');
    if (typeof THREE === 'undefined') console.warn('Three.js failed to load');

    // 3. Initialize core systems
    try {
        initLanguage();
        initTheme();
        initUI();
    } catch (e) {
        console.error('Core system init error:', e);
    }

    // 4. Auth & Magic Link Handling (from index.html)
    try {
        const urlParams = new URLSearchParams(window.location.search);
        
        // Force logout if requested
        if (urlParams.get('logout') === '1') {
            localStorage.removeItem('wxbot_token');
            localStorage.removeItem('wxbot_role');
            window.authToken = null;
            window.authRole = 'user';
            window.history.replaceState({}, document.title, window.location.pathname);
            alert('已强制退出登录');
        }

        const magicToken = urlParams.get('magic_token');
        if (magicToken) {
            handleMagicToken(magicToken)
                .then(data => {
                    clearTimeout(emergencyTimeout);
                    if (lp) {
                        lp.classList.add('hidden');
                        setTimeout(() => lp.style.display = 'none', 500);
                    }
                    startApp();
                    window.history.replaceState({}, document.title, window.location.pathname);
                })
                .catch(err => {
                    clearTimeout(emergencyTimeout);
                    alert('免密码登录失败，链接可能已过期。');
                    window.history.replaceState({}, document.title, window.location.pathname);
                    if (lp) {
                        lp.classList.remove('hidden');
                        lp.style.display = 'flex';
                    }
                });
            return;
        }

        const token = localStorage.getItem('wxbot_token');
        if (!token || token === 'null' || token === 'undefined') {
            clearTimeout(emergencyTimeout);
            if (lp) {
                lp.classList.remove('hidden');
                lp.style.display = 'flex';
            }
        } else {
            clearTimeout(emergencyTimeout);
            if (lp) {
                lp.classList.add('hidden');
                setTimeout(() => lp.style.display = 'none', 500);
            }
            startApp();
        }
    } catch (e) {
        clearTimeout(emergencyTimeout);
        console.error('Auth initialization fatal error:', e);
        if (lp) {
            lp.style.display = 'flex';
            lp.classList.remove('hidden');
        }
    }
});

