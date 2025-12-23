import { fetchWithAuth, callBotApi } from './api.js';
import { t } from './i18n.js';
import { latestChatStats } from './stats.js';
import { loadGroupMembers } from './members.js';

export let currentGroups = [];
export let currentGroupId = 0;
export let currentGuildId = null;
export let currentContactType = 'group';
export let groupSortBy = 'name';
export let groupSortAsc = true;

export async function refreshGroupList(botId = null) {
    const listEl = document.getElementById('group-list');
    if (!listEl) return;
    listEl.innerHTML = '<div class="text-center p-4"><div class="spinner-border text-primary"></div></div>';
    
    // Fallback to global currentBotId if not provided
    const targetBotId = botId || (document.getElementById('global-bot-selector-groups') ? document.getElementById('global-bot-selector-groups').value : currentBotId);
    if (!targetBotId) {
        listEl.innerHTML = '<div class="text-center p-4 text-muted">请先选择机器人</div>';
        return;
    }

    try {
        const url = `/api/contacts?bot_id=${encodeURIComponent(targetBotId)}&refresh=true`;
        const response = await fetchWithAuth(url);
        const text = await response.text();
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${text}`);
        }
        let contacts = [];
        try {
            contacts = JSON.parse(text);
            if (!Array.isArray(contacts)) contacts = [];
        } catch (e) {
            console.error("Failed to parse contacts JSON:", e, "Raw text:", text);
            throw new Error("JSON 解析错误: " + e.message);
        }
        
        const filtered = contacts.filter(c => {
            if (!c) return false;
            if (targetBotId && c.bot_id != targetBotId) return false;
            // Only include groups and guilds (channels), exclude private contacts (friends)
            return c.type === 'group' || c.type === 'guild';
        });
        
        currentGroups = filtered;
        renderGroups(filtered);
        window.renderGroups = () => renderGroups(currentGroups);
        const countEl = document.getElementById('group-count');
        if (countEl) countEl.innerText = `共 ${filtered.length} 个会话`;
    } catch (e) {
        listEl.innerHTML = `<div class="text-center p-4 text-danger">获取失败: ${e.message}</div>`;
    }
}

export function sortGroups(field) {
    if (groupSortBy === field) {
        groupSortAsc = !groupSortAsc;
    } else {
        groupSortBy = field;
        groupSortAsc = true;
        if (field === 'count' || field === 'msg_today') groupSortAsc = false;
    }
    
    ['name', 'count', 'id', 'msg_today'].forEach(f => {
        const btn = document.getElementById(`btn-sort-group-${f}`);
        if (btn) {
            if (f === groupSortBy) {
                btn.classList.add('active');
                let label = '';
                switch(f) {
                    case 'name': label = t('sort_name') || '名称'; break;
                    case 'count': label = t('sort_count') || '数量'; break;
                    case 'id': label = t('sort_id') || 'ID'; break;
                    case 'msg_today': label = t('sort_today') || '今日'; break;
                }
                btn.innerHTML = label + (groupSortAsc ? ' ↑' : ' ↓');
            } else {
                btn.classList.remove('active');
                let label = '';
                switch(f) {
                    case 'name': label = t('sort_name') || '名称'; break;
                    case 'count': label = t('sort_count') || '数量'; break;
                    case 'id': label = t('sort_id') || 'ID'; break;
                    case 'msg_today': label = t('sort_today') || '今日'; break;
                }
                btn.innerHTML = label;
            }
        }
    });

    filterGroups();
}

export function renderGroups(groups) {
    const listEl = document.getElementById('group-list');
    if (!listEl) return;
    if (!groups || groups.length === 0) {
        listEl.innerHTML = `<div class="text-center p-4 text-muted">${t('no_groups') || '未找到会话'}</div>`;
        return;
    }

    // Sort logic
    const sortedGroups = [...groups].sort((a, b) => {
        let res = 0;
        if (!a || !b) return 0;
        if (groupSortBy === 'name') {
            res = (a.name || '').localeCompare(b.name || '', 'zh-CN');
        } else if (groupSortBy === 'count') {
            res = 0; // Not available in session
        } else if (groupSortBy === 'id') {
            res = String(a.id || '').localeCompare(String(b.id || ''));
        } else if (groupSortBy === 'msg_today') {
            const statA = latestChatStats.group_stats_today ? (latestChatStats.group_stats_today[a.id] || 0) : 0;
            const statB = latestChatStats.group_stats_today ? (latestChatStats.group_stats_today[b.id] || 0) : 0;
            res = statA - statB;
        }
        return groupSortAsc ? res : -res;
    });
    
    listEl.innerHTML = sortedGroups.map(g => {
        const todayMsg = latestChatStats.group_stats_today ? (latestChatStats.group_stats_today[g.id] || 0) : 0;
        const totalMsg = latestChatStats.group_stats ? (latestChatStats.group_stats[g.id] || 0) : 0;
        const memberCount = g.member_count || 0;
        const memberCountHtml = memberCount > 0 ? `<div class="ms-1 flex-shrink-0 text-muted" style="font-size: 0.75rem;"><i class="bi bi-person me-1"></i>${memberCount}</div>` : '';
        
        let typeBadge = '<span class="badge bg-secondary" style="font-size: 0.6rem;">群</span>';
        if (g.type === 'private') {
            typeBadge = '<span class="badge bg-success" style="font-size: 0.6rem;">私</span>';
        } else if (g.type === 'guild') {
            typeBadge = '<span class="badge bg-info" style="font-size: 0.6rem;">频</span>';
        }

        // Avatar URL logic
        let avatarUrl = '';
        if (g.type === 'group') {
            avatarUrl = `https://p.qlogo.cn/gh/${g.id}/${g.id}/640/`;
        } else if (g.type === 'private') {
            avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${g.id}&s=100`;
        } else {
            avatarUrl = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/diagram-3.svg';
        }

        if (avatarUrl.startsWith('http')) {
            avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
        }

        const safeName = (g.name || '').replace(/'/g, "\\'").replace(/"/g, '&quot;');
        return `
        <div class="list-group-item list-group-item-action p-0 ${g.id == currentGroupId ? 'active' : ''}">
            <div class="d-flex align-items-center">
                <div class="ps-2 pe-1">
                    <input type="checkbox" class="form-check-input mass-send-checkbox-group" data-id="${g.id}" data-type="${g.type}" data-name="${safeName}" data-bot-id="${g.bot_id || ''}" data-guild="${g.guild_id || ''}" onclick="event.stopPropagation()">
                </div>
                <button class="btn btn-link text-decoration-none text-start p-2 flex-grow-1 border-0" style="color: inherit; text-align: left;" onclick="selectGroup('${g.id}', '${safeName}', '${g.type}', '${g.guild_id || ''}')">
                    <div class="d-flex w-100 align-items-center">
                        <div class="rounded-circle me-2 flex-shrink-0" style="width: 40px; height: 40px; overflow: hidden; background: var(--bg-body);">
                             <img src="${avatarUrl}" alt="${g.type}" style="width: 100%; height: 100%; object-fit: cover;" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/${g.type === 'private' ? 'person' : 'people'}.svg';this.style.padding='8px';">
                        </div>
                        <div class="flex-grow-1 overflow-hidden" style="min-width: 0;">
                            <div class="d-flex w-100 justify-content-between align-items-center">
                                <div class="flex-grow-1" style="min-width: 0; padding-right: 8px;">
                                    <h6 class="mb-0" style="font-size: 0.85rem; font-weight: 600; word-break: break-all;">${g.name || 'Unknown'}</h6>
                                </div>
                                ${memberCountHtml}
                            </div>
                            <div class="d-flex w-100 justify-content-between align-items-center mt-1">
                                <div class="d-flex align-items-center">
                                    <small class="${g.id == currentGroupId ? 'text-light' : 'text-muted'}" style="font-size: 0.7rem;">ID: ${g.id}</small>
                                    <div class="ms-2">
                                        ${typeBadge}
                                    </div>
                                </div>
                                <span class="badge ${g.id == currentGroupId ? 'bg-light text-dark' : 'bg-secondary'} rounded-pill" style="font-size: 0.65rem;" title="今日: ${todayMsg} / 总计: ${totalMsg}">${todayMsg}</span>
                            </div>
                        </div>
                    </div>
                </button>
            </div>
        </div>
    `}).join('');
}

export function filterGroups() {
    const keyword = (document.getElementById('group-search') ? document.getElementById('group-search').value : '').toLowerCase();
    const filtered = currentGroups.filter(g => 
        (g.name || '').toLowerCase().includes(keyword) || 
        String(g.id || '').includes(keyword)
    );
    renderGroups(filtered);
}

export async function selectGroup(id, name, type = 'group', guildId = '') {
    currentGroupId = id;
    currentContactType = type;
    currentGuildId = guildId;

    // Update currentBotId from the selected group
    const groupInfo = currentGroups.find(g => g.id == id);
    if (groupInfo && groupInfo.bot_id) {
        window.currentBotId = groupInfo.bot_id;
        console.log(`[Select] Group ${id} uses bot ${window.currentBotId}`);
    }

    document.getElementById('current-group-name').innerText = name;
    document.getElementById('current-group-id').innerText = id;
    
    // Update Stats
    const todayMsg = latestChatStats.group_stats_today ? (latestChatStats.group_stats_today[id] || 0) : 0;
    const totalMsg = latestChatStats.group_stats ? (latestChatStats.group_stats[id] || 0) : 0;
    
    const group = currentGroups.find(x => x.id == id);
    const memberCount = (group && group.member_count) ? group.member_count : 0;
    const memberInfo = memberCount > 0 ? ` / 成员: ${memberCount}` : '';
    
    document.getElementById('current-group-stats').innerText = `今日: ${todayMsg} / 总计: ${totalMsg}${memberInfo}`;

    // Update Avatar
    const avatarEl = document.getElementById('current-group-avatar');
    if (avatarEl) {
        let avatarUrl = '';
        if (type === 'group') {
            avatarUrl = `https://p.qlogo.cn/gh/${id}/${id}/640/`;
        } else if (type === 'private') {
            avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
        } else {
             avatarUrl = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/diagram-3.svg';
        }

        if (avatarUrl.startsWith('http')) {
            avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
        }
        
        avatarEl.src = avatarUrl;
        avatarEl.style.display = 'block';
        avatarEl.onerror = function() {
            this.src = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/' + (type === 'private' ? 'person' : 'people') + '.svg';
            this.style.padding = '4px';
        };
    }
    
    document.getElementById('group-detail-empty').style.display = 'none';
    document.getElementById('group-detail-content').style.setProperty('display', 'flex', 'important');
    
    // Highlight selection
    filterGroups();

    // Handle View switching
    const membersTabBtn = document.getElementById('members-tab');
    const actionsTabBtn = document.getElementById('actions-tab');
    
    if (type === 'group' || type === 'guild') {
        membersTabBtn.style.display = 'block';
        // Switch to members tab if not already active or if we just hid it previously
        const bsTab = new bootstrap.Tab(membersTabBtn);
        bsTab.show();
        loadGroupMembers(id);
    } else {
        membersTabBtn.style.display = 'none';
        // Switch to actions tab
        const bsTab = new bootstrap.Tab(actionsTabBtn);
        bsTab.show();
    }
}

export function onContactClick(id, name, type, botId, guildId) {
    selectGroup(id, name, type, guildId);
}

export function toggleAutoRecallInput() {
    const checked = document.getElementById('auto-recall-check').checked;
    document.getElementById('auto-recall-input-group').style.display = checked ? 'flex' : 'none';
}

export async function sendGroupMsg() {
    const input = document.getElementById('group-msg-input');
    const msg = input.value.trim();
    if (!msg) return;
    
    let autoRecall = 0;
    if (document.getElementById('auto-recall-check').checked) {
        const delayStr = document.getElementById('auto-recall-delay').value;
        autoRecall = parseInt(delayStr);
        if (isNaN(autoRecall) || autoRecall < 0) autoRecall = 0;
    }

    try {
        let action = 'send_group_msg';
        let params = { message: msg };

        if (currentContactType === 'group') {
            action = 'send_group_msg';
            params.group_id = currentGroupId;
        } else if (currentContactType === 'private') {
            action = 'send_private_msg';
            params.user_id = currentGroupId;
        } else if (currentContactType === 'guild') {
            action = 'send_msg';
            params.message_type = 'guild';
            params.channel_id = currentGroupId;
            params.guild_id = currentGuildId;
        }

        const body = {
            bot_id: window.currentBotId,
            action: action,
            params: params,
            auto_recall: autoRecall
        };

        const res = await fetchWithAuth('/api/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });
        
        const result = await res.json();
        if (result.error) throw new Error(result.error);

        alert(t('alert_send_success') || '发送成功');
        input.value = '';
        if (window.addEventLog) {
            window.addEventLog({type: 'system', message: `发送消息到 [${currentGroupId}]: ${msg}`});
        }
    } catch (e) {
        alert((t('alert_op_failed') || '操作失败: ') + e.message);
    }
}

export async function sendSmartGroupMsg() {
    if (!currentGroupId || !window.currentBotId) return;
    const content = document.getElementById('group-msg-input').value.trim();
    if (!content) return;
    
    let autoRecall = 0;
    if (document.getElementById('auto-recall-check').checked) {
        const delayStr = document.getElementById('auto-recall-delay').value;
        autoRecall = parseInt(delayStr);
        if (isNaN(autoRecall) || autoRecall < 0) autoRecall = 0;
    }
    
    const btn = event.target.closest('button');
    const originalText = btn.innerHTML;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> 发送中...';
    btn.disabled = true;

    try {
        const res = await fetchWithAuth('/api/smart_action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                action: 'send_group_msg',
                params: {
                    group_id: currentGroupId,
                    message: content
                },
                self_id: window.currentBotId,
                auto_recall: autoRecall
            })
        });

        if (!res.ok) throw new Error('请求失败');

        const data = await res.json();
        alert('智能发送请求已提交: ' + (data.detail || 'Success'));
        document.getElementById('group-msg-input').value = '';
    } catch (e) {
        alert('智能发送失败: ' + e.message);
    } finally {
        btn.innerHTML = originalText;
        btn.disabled = false;
    }
}
// End of module
