import { fetchWithAuth } from './api.js';
import { currentLang, translations } from './i18n.js';
import { showToast } from './ui.js';
import { timeAgo } from './utils.js';
import { authToken } from './auth.js';

export let currentBots = [];
export let currentBotId = null;
export let botSortBy = 'name'; // name, id, count, time, platform, msg
export let botSortAsc = true;
export let botFilterText = '';
export let botViewMode = localStorage.getItem('bot_view_mode') || 'detail';

export async function fetchBots(showLoading = false) {
    if (!window.authToken && !localStorage.getItem('wxbot_token')) return;
    const t = translations[currentLang] || translations['zh-CN'];
    
    if (showLoading) {
        const container = document.getElementById('bot-list-container');
        if (container) {
            container.innerHTML = `
                <div class="col-12 text-center py-5">
                    <div class="spinner-border text-primary" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                    <div class="mt-2 text-muted">${t.loading || '加载中...'}</div>
                </div>
            `;
        }
    }

    try {
        const response = await fetchWithAuth('/api/bots');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        
        const data = await response.json();
        // Handle both {bots:[]} and [] formats
        let bots = [];
        if (data && data.bots && Array.isArray(data.bots)) {
            bots = data.bots;
        } else if (Array.isArray(data)) {
            bots = data;
        }
        
        currentBots = bots;
        window.currentBots = bots; // Legacy compatibility
        
        // Update Global Selectors
        updateGlobalBotSelectors(bots);
        
        const onlineCount = bots.filter(b => b.is_alive).length;
        
        // Update Badge
        const badge = document.getElementById('badge-bots');
        if (badge) badge.innerText = onlineCount + '/' + bots.length;

        // Update Metrics
        const metricBots = document.getElementById('metric-bots');
        if (metricBots) metricBots.innerText = onlineCount;

        renderBots();
        window.renderBots = renderBots;
        
        // Update other selectors if functions exist
        if (typeof window.updateLogBotSelector === 'function') window.updateLogBotSelector(bots);
        if (typeof window.updateLogBotSelectorFull === 'function') window.updateLogBotSelectorFull(bots);

    } catch (err) {
        console.error('获取机器人列表失败:', err);
        currentBots = [];
        renderBots();
    }
}

export function filterBots(text) {
    botFilterText = text.toLowerCase();
    renderBots();
}

export function sortBots(field) {
    if (botSortBy === field) {
        botSortAsc = !botSortAsc;
    } else {
        botSortBy = field;
        botSortAsc = true;
        if (field === 'count' || field === 'time' || field === 'msg') botSortAsc = false;
    }

    const t = translations[currentLang] || translations['zh-CN'];

    // Update UI buttons
    ['name', 'id', 'platform', 'count', 'time', 'msg'].forEach(f => {
        const btn = document.getElementById(`btn-sort-bot-${f}`);
        if (!btn) return;
        
        if (f === botSortBy) {
            btn.classList.add('active');
            let label = '';
            switch(f) {
                case 'name': label = t.sort_name; break;
                case 'id': label = t.sort_id; break;
                case 'platform': label = t.sort_platform; break;
                case 'count': label = t.sort_group_count; break;
                case 'time': label = t.sort_time; break;
                case 'msg': label = t.sort_msg_count; break;
            }
            btn.innerHTML = label + (botSortAsc ? ' ↑' : ' ↓');
        } else {
            btn.classList.remove('active');
            let label = '';
            switch(f) {
                case 'name': label = t.sort_name; break;
                case 'id': label = t.sort_id; break;
                case 'platform': label = t.sort_platform; break;
                case 'count': label = t.sort_group_count; break;
                case 'time': label = t.sort_time; break;
                case 'msg': label = t.sort_msg_count; break;
            }
            btn.innerHTML = label;
        }
    });

    renderBots();
}

export function setBotViewMode(mode) {
    botViewMode = mode;
    window._botViewModeManuallySet = true;
    localStorage.setItem('bot_view_mode', mode);
    updateViewModeUI('bot', mode);
    renderBots();
}

function updateViewModeUI(type, mode) {
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

export function renderBots() {
    const container = document.getElementById('bot-list-container');
    if (!container) return;

    let bots = currentBots.filter(b => {
        if (!botFilterText) return true;
        return (b.nickname || '').toLowerCase().includes(botFilterText) || 
               (b.self_id || b.id || '').includes(botFilterText) ||
               (b.platform || '').toLowerCase().includes(botFilterText);
    });

    // Auto switch to compact mode if many bots
    if (bots.length > 12 && botViewMode === 'detail' && !window._botViewModeManuallySet) {
        botViewMode = 'compact';
        const btnDetail = document.getElementById('btn-bot-view-detail');
        const btnCompact = document.getElementById('btn-bot-view-compact');
        if (btnDetail) btnDetail.classList.remove('active');
        if (btnCompact) btnCompact.classList.add('active');
    }

    // Sort
    bots.sort((a, b) => {
        // Always put online bots first
        const aAlive = a.is_alive !== undefined ? a.is_alive : a.online;
        const bAlive = b.is_alive !== undefined ? b.is_alive : b.online;
        if (aAlive !== bAlive) {
            return aAlive ? -1 : 1;
        }

        let res = 0;
        if (botSortBy === 'name') {
            res = (a.nickname || '').localeCompare(b.nickname || '', 'zh-CN');
        } else if (botSortBy === 'id') {
            res = (a.self_id || a.id || '').localeCompare(b.self_id || b.id || '');
        } else if (botSortBy === 'count') {
            res = (a.group_count || 0) - (b.group_count || 0);
        } else if (botSortBy === 'time') {
            res = new Date(a.connected) - new Date(b.connected);
        } else if (botSortBy === 'platform') {
            res = (a.platform || 'QQ').localeCompare(b.platform || 'QQ', 'zh-CN');
        } else if (botSortBy === 'msg') {
            res = (a.msg_count || 0) - (b.msg_count || 0);
        }
        
        if (res !== 0) {
            return botSortAsc ? res : -res;
        }
        
        return (a.self_id || a.id || '').localeCompare(b.self_id || b.id || '');
    });

    const t = translations[currentLang] || translations['zh-CN'];
    if (bots.length === 0) {
        container.innerHTML = `<div class="col-12 text-center text-muted">${t.no_bot_data}</div>`;
        return;
    }

    container.innerHTML = bots.map(bot => {
        const id = bot.self_id || bot.id;
        const connDate = new Date(bot.connected);
        const timeStr = timeAgo(bot.connected);
        const isAlive = bot.is_alive !== undefined ? bot.is_alive : bot.online;
        const isOffline = !isAlive;
        const cardStyle = isOffline ? 'filter: grayscale(100%); opacity: 0.8;' : '';
        const statusBadge = isOffline ? `<span class="badge bg-secondary ms-2">${t.bot_status_offline}</span>` : `<span class="badge bg-success ms-2">${t.bot_status_online}</span>`;
        const timeLabel = isOffline ? t.card_status_offline_since : t.card_status_connected_at;
        const platform = bot.platform || 'QQ';
        let avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg`;
        let avatarBg = 'var(--bg-list-item)';
        
        if (platform.toUpperCase().includes('QQ')) {
            avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=640`;
            avatarBg = '#12B7F522';
        } else if (platform.toUpperCase().includes('TELEGRAM') || platform.toUpperCase() === 'TG') {
            avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/telegram.svg`;
            avatarBg = '#0088cc22';
        } else if (platform.toUpperCase().includes('DISCORD')) {
            avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/discord.svg`;
            avatarBg = '#5865F222';
        } else if (platform.toUpperCase().includes('WECHAT') || platform.toUpperCase().includes('WX')) {
            avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/wechat.svg`;
            avatarBg = '#07C16022';
        }
        
        if (avatarUrl.startsWith('http')) {
            avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
        }
        
        const platformBadge = `<span class="badge bg-info ms-1">${platform}</span>`;

        if (botViewMode === 'compact') {
            return `
            <div class="col-sm-6 col-md-4 col-lg-3 col-xl-2 mb-3">
                <div class="card p-2 h-100 shadow-sm" style="${cardStyle}">
                    <div class="d-flex align-items-center mb-2 ${isOffline ? '' : 'cursor-pointer'}" 
                         ${isOffline ? '' : `onclick="setGlobalBot('${id}')"`}>
                        <div class="rounded-circle d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 32px; height: 32px; overflow: hidden; background: ${avatarBg};">
                            <img src="${avatarUrl}" alt="Avatar" style="width: 100%; height: 100%; object-fit: cover;" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';">
                        </div>
                        <div class="w-100">
                                    <div class="fw-bold d-flex align-items-center flex-wrap" title="${bot.nickname || id}" style="font-size: 0.85rem; word-break: break-all;">
                                        <span>${bot.nickname || id}</span>
                                        ${isOffline ? '<span class="badge bg-secondary ms-1" style="font-size: 0.6em;">OFF</span>' : '<span class="badge bg-success ms-1" style="font-size: 0.6em;">ON</span>'}
                                        <span class="badge bg-info ms-1" style="font-size: 0.6em;">${platform}</span>
                                    </div>
                                    <div class="text-muted" style="font-size: 0.7rem;">
                                        <span>${id}</span>
                                    </div>
                                </div>
                    </div>
                    <div class="rounded p-1 mb-2 flex-grow-1" style="background-color: var(--bg-list-item); font-size: 0.75rem;">
                         <div class="d-flex justify-content-between">
                            <span class="text-muted">${t.card_label_groups}: ${bot.group_count || 0}</span>
                            <span class="text-muted">${t.card_label_friends}: ${bot.friend_count || 0}</span>
                        </div>
                        <div class="d-flex justify-content-between">
                            <span class="text-muted">${t.card_label_msgs}:</span>
                            <span class="text-primary fw-bold">${bot.msg_count || 0} <small class="text-muted">(${bot.msg_count_today || 0})</small></span>
                        </div>
                    </div>
                </div>
            </div>`;
        }

        return `
            <div class="col-sm-12 col-md-6 col-lg-4 col-xl-4 col-xxl-3 mb-3">
                <div class="card p-2 h-100 shadow-sm" style="${cardStyle}">
                    <div class="d-flex align-items-center mb-2 ${isOffline ? '' : 'cursor-pointer'}"
                         ${isOffline ? '' : `onclick="setGlobalBot('${id}')"`}>
                        <div class="rounded-circle d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 48px; height: 48px; overflow: hidden; background: ${avatarBg};">
                            <img src="${avatarUrl}" alt="Avatar" style="width: 100%; height: 100%; object-fit: cover;" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';">
                        </div>
                        <div class="w-100">
                                    <div class="fw-bold d-flex align-items-center flex-wrap" title="${bot.nickname || id}" style="font-size: 0.9rem; word-break: break-all;">
                                        <span>${bot.nickname || id}</span>
                                        ${statusBadge}
                                        ${platformBadge}
                                    </div>
                                    <div class="text-muted mt-1" style="font-size: 0.7rem;">
                                        <span>ID: ${id}</span>
                                    </div>
                                </div>
                    </div>
                    <div class="rounded p-2 mb-2 flex-grow-1" style="background-color: var(--bg-list-item); font-size: 0.75rem;">
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t.card_label_groups}:</span>
                            <span class="fw-bold text-primary">${bot.group_count || 0}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t.card_label_friends}:</span>
                            <span class="fw-bold text-success">${bot.friend_count || 0}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t.card_label_msgs}:</span>
                            <span><span class="fw-bold text-primary">${bot.msg_count || 0}</span> <span class="text-muted" style="font-size: 0.7em;">(${t.sort_today}:${bot.msg_count_today || 0})</span></span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t.card_label_ip}:</span>
                            <span class="text-truncate" style="max-width: 80px;" title="${bot.remote_addr}">${bot.remote_addr ? bot.remote_addr.split(':')[0] : 'N/A'}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${timeLabel}</span>
                            <span title="${connDate.toLocaleString()}">${timeStr}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t.card_label_protocol}:</span>
                            <span class="fw-bold text-info">${platform}</span>
                        </div>
                    </div>
                </div>
            </div>
        `}).join('');
}

export function updateGlobalBotSelectors(bots = null) {
    if (!bots) bots = currentBots;
    if (!bots || !Array.isArray(bots)) return;
    
    const t = translations[currentLang] || translations['zh-CN'];
    const selectorConfigs = [
        { inputId: 'global-bot-selector', listId: 'global-bot-list', contentId: 'global-bot-selected-content' },
        { inputId: 'global-bot-selector-groups', listId: 'global-bot-list-groups', contentId: 'global-bot-selected-content-groups' },
        { inputId: 'global-bot-selector-friends', listId: 'global-bot-list-friends', contentId: 'global-bot-selected-content-friends' }
    ];

    selectorConfigs.forEach(config => {
        const selectorInput = document.getElementById(config.inputId);
        const dropdownList = document.getElementById(config.listId);
        const contentEl = document.getElementById(config.contentId);
        
        if (!selectorInput || !dropdownList || !contentEl) return;
        
        const currentVal = selectorInput.value;
        
        if (bots.length === 0) {
            dropdownList.innerHTML = `<li><span class="dropdown-item disabled">${t.no_bot_data_simple || '无机器人数据'}</span></li>`;
            contentEl.innerHTML = t.no_bot_data_simple || '无机器人数据';
        } else {
            dropdownList.innerHTML = bots.map(bot => {
                const isOffline = !bot.is_alive;
                const statusClass = isOffline ? 'text-muted' : 'text-success fw-bold';
                const badge = isOffline ? `<span class="badge bg-secondary ms-auto">${t.bot_status_offline || '离线'}</span>` : `<span class="badge bg-success ms-auto">${t.bot_status_online || '在线'}</span>`;
                const platform = bot.platform || 'QQ';
                let avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg`;
                
                if (platform.toUpperCase() === 'QQ') {
                    avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${bot.self_id}&s=100`;
                }
                
                if (avatarUrl.startsWith('http')) {
                    avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
                }

                return `
                    <li>
                        <a class="dropdown-item d-flex align-items-center p-2 ${isOffline ? 'disabled' : ''}" href="#" onclick="${isOffline ? 'return false;' : `selectBotFromDropdown('${bot.self_id}', '${config.inputId}')`}">
                            <div class="rounded-circle d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 32px; height: 32px; overflow: hidden; background: var(--bg-list-item);">
                                <img src="${avatarUrl}" alt="" style="width: 100%; height: 100%; object-fit: cover;" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';">
                            </div>
                            <div class="flex-grow-1 overflow-hidden">
                                <div class="d-flex align-items-center">
                                    <div class="fw-bold text-truncate ${statusClass}" style="max-width: 120px;">${bot.nickname || bot.self_id}</div>
                                    ${badge}
                                </div>
                                <div class="text-muted small">${bot.self_id}</div>
                            </div>
                        </a>
                    </li>
                `;
            }).join('');
            
            // Restore selection or select first online
            let targetId = currentVal;
            if (!targetId || !bots.find(b => b.self_id === targetId && b.is_alive)) {
                const firstOnline = bots.find(b => b.is_alive);
                if (firstOnline) {
                    targetId = firstOnline.self_id;
                }
            }
            
            if (targetId) {
                selectorInput.value = targetId;
                updateBotSelectorUI(targetId, config.inputId);
                
                // If it's the main selector and currentBotId is not set, set it
                if (config.inputId === 'global-bot-selector' && !currentBotId) {
                    currentBotId = targetId;
                }
            }
        }
    });
}

export function selectBotFromDropdown(botId, selectorInputId = 'global-bot-selector') {
    const input = document.getElementById(selectorInputId);
    if (input) {
        // Update Global State
        currentBotId = botId;

        // Sync all selectors
        ['global-bot-selector', 'global-bot-selector-groups', 'global-bot-selector-friends'].forEach(id => {
            const el = document.getElementById(id);
            if (el) {
                el.value = botId;
                updateBotSelectorUI(botId, id);
            }
        });
        
        const t = translations[currentLang] || translations['zh-CN'];
        const loadingText = `<div class="text-center p-4 text-muted">${t.switching_bot || '切换机器人中...'}</div>`;

        // Clear views that depend on global bot
        const groupList = document.getElementById('group-list');
        if (groupList) groupList.innerHTML = loadingText;
        const groupDetail = document.getElementById('group-detail-content');
        if (groupDetail) groupDetail.style.setProperty('display', 'none', 'important');
        const groupEmpty = document.getElementById('group-detail-empty');
        if (groupEmpty) groupEmpty.style.display = 'block';

        const friendList = document.getElementById('friend-list');
        if (friendList) friendList.innerHTML = loadingText;
        const friendDetail = document.getElementById('friend-detail-content');
        if (friendDetail) friendDetail.style.setProperty('display', 'none', 'important');
        const friendEmpty = document.getElementById('friend-detail-empty');
        if (friendEmpty) friendEmpty.style.display = 'block';

        if (window.currentGroups) window.currentGroups = [];
        if (window.currentFriends) window.currentFriends = [];

        if (typeof window.fetchStats === 'function') window.fetchStats();
        
        // Auto refresh current tab
        const activeTab = window.activeTab || 'dashboard';
        if (activeTab === 'groups') {
            if (typeof window.refreshGroupList === 'function') window.refreshGroupList(botId);
        } else if (activeTab === 'friends') {
            if (typeof window.refreshFriendList === 'function') window.refreshFriendList(botId);
        } else {
            if (typeof window.refreshCurrentTabList === 'function') window.refreshCurrentTabList();
        }
    }
}

export function fetchContacts(type, botId) {
    if (type === 'group') {
        if (typeof window.refreshGroupList === 'function') window.refreshGroupList(botId);
    } else if (type === 'private') {
        if (typeof window.refreshFriendList === 'function') window.refreshFriendList(botId);
    }
}

export function updateBotSelectorUI(botId, selectorInputId = 'global-bot-selector') {
    const bot = currentBots.find(b => b.self_id === botId);
    if (!bot) return;
    
    let contentId = 'global-bot-selected-content';
    if (selectorInputId === 'global-bot-selector-groups') contentId = 'global-bot-selected-content-groups';
    if (selectorInputId === 'global-bot-selector-friends') contentId = 'global-bot-selected-content-friends';
    
    const contentEl = document.getElementById(contentId);
    if (!contentEl) return;
    
    const platform = bot.platform || 'QQ';
    let avatarUrl = `https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg`;
    if (platform.toUpperCase() === 'QQ') {
        avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${bot.self_id}&s=100`;
    }
    if (avatarUrl.startsWith('http')) {
        avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
    }

    contentEl.innerHTML = `
        <div class="d-flex align-items-center">
            <img src="${avatarUrl}" class="rounded-circle me-2" width="24" height="24" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';">
            <div class="text-truncate fw-bold" style="max-width: 100px;">${bot.nickname || bot.self_id}</div>
        </div>
    `;
    
    const input = document.getElementById(selectorInputId);
    if (input) input.value = botId;
}

export function setGlobalBot(id) {
    selectBotFromDropdown(id);
    if (typeof window.showTab === 'function') {
        window.showTab('groups'); // Jump to groups tab
    }
}

export function refreshCurrentTabList() {
    const groupsTab = document.getElementById('tab-groups');
    const friendsTab = document.getElementById('tab-friends');
    
    if (friendsTab && friendsTab.classList.contains('active-tab')) {
        if (typeof window.refreshFriendList === 'function') window.refreshFriendList();
    } else if (groupsTab && groupsTab.classList.contains('active-tab')) {
        if (typeof window.refreshGroupList === 'function') window.refreshGroupList();
    }
}
/**
 * 切换机器人状态 (启动/停止)
 * @param {string} id 机器人ID
 * @param {string} action 操作类型 ('start' | 'stop')
 */
export async function toggleBot(id, action) {
    const t = translations[currentLang] || translations['zh-CN'];
    try {
        const response = await fetchWithAuth(`/api/bot/${action}?id=${id}`, { method: 'POST' });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        showToast(t.alert_op_success || '操作成功', 'success');
        fetchBots(true);
    } catch (e) {
        console.error(`Failed to ${action} bot:`, e);
        showToast((t.alert_op_failed || '操作失败: ') + e.message, 'danger');
    }
}

/**
 * 切换机器人状态 (Alias for init.js)
 */
export async function toggleBotState(id, action) {
    return toggleBot(id, action);
}

/**
 * 删除机器人
 */
export async function deleteBot(id) {
    const t = translations[currentLang] || translations['zh-CN'];
    if (!confirm(t.confirm_delete_bot || '确定要删除该机器人吗？')) return;
    
    try {
        const response = await fetchWithAuth(`/api/bot/delete?id=${id}`, { method: 'POST' });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        showToast(t.alert_op_success || '操作成功', 'success');
        fetchBots(true);
    } catch (e) {
        console.error(`Failed to delete bot:`, e);
        showToast((t.alert_op_failed || '操作失败: ') + e.message, 'danger');
    }
}

/**
 * 打开机器人配置
 */
export function openBotConfig(id) {
    console.log('Open bot config:', id);
    // TODO: Implement configuration modal logic if needed
}

/**
 * 保存机器人配置
 */
export async function saveBotConfig(id, config) {
    console.log('Save bot config:', id, config);
    // TODO: Implement save logic if needed
}

// End of module
