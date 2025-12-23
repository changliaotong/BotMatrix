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

// Note: DOM Ready initialization is now handled in modules/app.js to avoid redundancy
// and ensure consistent behavior across different pages (index.html, legacy.html).


