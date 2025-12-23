import { callBotApi, fetchWithAuth } from './api.js';
import { t } from './i18n.js';
import { latestChatStats } from './stats.js';
import { currentContactType } from './groups.js';

export let currentMembers = [];
export let currentGroupId = null;
export let memberSortBy = 'role';
export let memberSortAsc = false;

export async function loadGroupMembers(groupId) {
    currentGroupId = groupId;
    const tbody = document.getElementById('member-list-body');
    if (!tbody) return;

    tbody.innerHTML = '<tr><td colspan="6" class="text-center"><div class="spinner-border spinner-border-sm text-primary me-2"></div>加载中...</td></tr>';
    
    try {
        let action = 'get_group_member_list';
        let params = { group_id: groupId };
        
        // window.currentContactType is global for now
        if (window.currentContactType === 'guild') {
            action = 'get_guild_member_list';
            params = { guild_id: groupId };
        }

        console.log(`[Members] Calling ${action} with`, params);
        let response;
        try {
            response = await callBotApi(action, params);
        } catch (apiError) {
            console.error(`[Members] ${action} failed:`, apiError);
            if (window.currentContactType === 'guild' && action === 'get_guild_member_list') {
                console.warn('[Members] get_guild_member_list failed, trying get_guild_members');
                action = 'get_guild_members';
                response = await callBotApi(action, params);
            } else {
                throw apiError;
            }
        }

        const isFailed = response && (response.status === 'failed' || (response.retcode !== undefined && response.retcode !== 0));
        
        if (isFailed) {
            const errMsg = response.message || response.msg || response.wording || '机器人返回获取失败';
            
            if (currentContactType === 'guild' && action === 'get_guild_member_list') {
            console.log('[Members] Trying get_guild_members as fallback...');
            action = 'get_guild_members';
            response = await callBotApi(action, params);
        } else {
                if (response.retcode === 0 && (Array.isArray(response.data) || (response.data && Array.isArray(response.data.members)))) {
                     console.log('[Members] Retcode is 0 despite failed status, proceeding with data');
                } else {
                    throw new Error(errMsg);
                }
            }
        }

        // 兼容不同机器人的返回格式
        let members = [];
        if (response.data) {
            if (Array.isArray(response.data.members)) {
                members = response.data.members;
            } else if (Array.isArray(response.data.list)) {
                members = response.data.list;
            } else if (Array.isArray(response.data)) {
                members = response.data;
            } else if (response.data.data && Array.isArray(response.data.data)) {
                // Some adapters nest data twice
                members = response.data.data;
            }
        } else if (Array.isArray(response)) {
            members = response;
        } else if (response.members && Array.isArray(response.members)) {
            members = response.members;
        }
        
        // Final mapping to ensure user_id exists (some adapters use 'id' or 'uid')
        members = members.map(m => {
            if (!m.user_id && m.id) m.user_id = m.id;
            if (!m.user_id && m.uid) m.user_id = m.uid;
            if (!m.nickname && m.name) m.nickname = m.name;
            return m;
        });
        
        console.log(`[Members] Extracted ${members.length} members`);
        currentMembers = members || [];
        
        // Final sanity check for empty list but retcode 0
        if (currentMembers.length === 0 && response.retcode === 0) {
            console.warn('[Members] API returned success but empty list. This might be a platform delay or permission issue.');
        }

        renderMembers();
    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="6" class="text-center text-danger">加载失败: ${e.message}</td></tr>`;
        console.error('loadGroupMembers error:', e);
    }
}

export function filterMembers() {
    renderMembers();
}

export function sortMembers(field) {
    if (memberSortBy === field) {
        memberSortAsc = !memberSortAsc;
    } else {
        memberSortBy = field;
        memberSortAsc = true;
        if (field === 'time' || field === 'msg_today') memberSortAsc = false;
    }

    // Update Icons
    ['card', 'id', 'role', 'time', 'msg_today'].forEach(f => {
        const icon = document.getElementById(`sort-icon-${f}`);
        if (icon) {
            if (f === memberSortBy) {
                icon.className = memberSortAsc ? 'bi bi-sort-alpha-down' : 'bi bi-sort-alpha-down-alt';
                if (f === 'id' || f === 'time' || f === 'msg_today') icon.className = memberSortAsc ? 'bi bi-sort-numeric-down' : 'bi bi-sort-numeric-down-alt';
            } else {
                icon.className = 'bi';
            }
        }
    });

    renderMembers();
}

export function renderMembers() {
    const tbody = document.getElementById('member-list-body');
    if (!tbody) return;

    const searchInput = document.getElementById('member-search');
    const keyword = searchInput ? searchInput.value.toLowerCase() : '';

    if (!currentMembers || currentMembers.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="text-center">
            <div class="p-3">
                <div class="text-muted mb-2">${t('no_members')}</div>
                <button class="btn btn-sm btn-outline-primary" onclick="loadGroupMembers('${currentGroupId}')">
                    <i class="bi bi-arrow-clockwise me-1"></i>${t('retry')}
                </button>
            </div>
        </td></tr>`;
        return;
    }

    // Filter
    const filteredMembers = currentMembers.filter(m => {
        if (!keyword) return true;
        const card = (m.card || '').toLowerCase();
        const nick = (m.nickname || '').toLowerCase();
        const uid = String(m.user_id);
        return card.includes(keyword) || nick.includes(keyword) || uid.includes(keyword);
    });
    
    if (filteredMembers.length === 0) {
         tbody.innerHTML = `<tr><td colspan="6" class="text-center">${t('no_match_members')}</td></tr>`;
         return;
    }

    // Sort
    const sortedMembers = [...filteredMembers].sort((a, b) => {
        let res = 0;
        if (memberSortBy === 'card') {
            const nameA = a.card || a.nickname || '';
            const nameB = b.card || b.nickname || '';
            res = nameA.localeCompare(nameB, 'zh-CN');
        } else if (memberSortBy === 'id') {
            res = String(a.user_id).localeCompare(String(b.user_id), undefined, {numeric: true});
        } else if (memberSortBy === 'role') {
            const roleWeight = { 'owner': 3, 'admin': 2, 'member': 1 };
            res = (roleWeight[a.role] || 0) - (roleWeight[b.role] || 0);
        } else if (memberSortBy === 'time') {
            res = (a.last_sent_time || 0) - (b.last_sent_time || 0);
        } else if (memberSortBy === 'msg_today') {
            const statA = latestChatStats.user_stats_today ? (latestChatStats.user_stats_today[a.user_id] || 0) : 0;
            const statB = latestChatStats.user_stats_today ? (latestChatStats.user_stats_today[b.user_id] || 0) : 0;
            res = statA - statB;
        }
        return memberSortAsc ? res : -res;
    });

    tbody.innerHTML = sortedMembers.map(m => {
        const todayMsg = latestChatStats.user_stats_today ? (latestChatStats.user_stats_today[m.user_id] || 0) : 0;
        const totalMsg = latestChatStats.user_stats ? (latestChatStats.user_stats[m.user_id] || 0) : 0;
        
        const roleKey = 'role_' + m.role;
        const roleText = t(roleKey) || m.role;
        const statsTooltip = (t('stats_title_tooltip') || '今日: {today}, 总计: {total}').replace('{today}', todayMsg).replace('{total}', totalMsg);
        const safeName = (m.card || m.nickname || '').replace(/'/g, "\\'").replace(/"/g, '&quot;');
        const mID = String(m.user_id);
        const gID = String(currentGroupId);

        return `
        <tr>
            <td>
                <div class="d-flex align-items-center">
                    <img src="${mID.startsWith('http') ? mID : `https://q1.qlogo.cn/g?b=qq&nk=${mID}&s=100`}" 
                         class="rounded-circle me-2" 
                         width="32" height="32" 
                         onerror="this.src='/api/proxy/avatar?url=' + encodeURIComponent(this.src); this.onerror=() => { this.src='https://ui-avatars.com/api/?name=${encodeURIComponent(m.card || m.nickname)}&background=random' }"
                         alt="Avatar">
                    <div>
                        <div>${m.card || m.nickname}</div>
                        ${m.card && m.card !== m.nickname ? `<small class="text-muted">${m.nickname}</small>` : ''}
                    </div>
                </div>
            </td>
            <td>${mID}</td>
            <td>
                <span class="badge bg-${m.role === 'owner' ? 'warning' : (m.role === 'admin' ? 'success' : 'secondary')}">
                    ${roleText}
                </span>
            </td>
            <td>
                <div class="d-flex flex-column" title="${statsTooltip}">
                    <span class="badge bg-primary rounded-pill mb-1" style="width: fit-content;">${todayMsg}</span>
                    <small class="text-muted">${t('stats_total') || '总计: '}${totalMsg}</small>
                </div>
            </td>
            <td>
                <small class="text-muted">${m.last_sent_time ? new Date(m.last_sent_time * 1000).toLocaleString() : '-'}</small>
            </td>
            <td>
                <div class="dropdown">
                    <button class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                        ${t('action_ops') || '操作'}
                    </button>
                    <ul class="dropdown-menu">
                        <li><a class="dropdown-item" href="#" onclick="setCard('${gID}', '${mID}', '${safeName}')">${t('action_set_card')}</a></li>
                        <li><a class="dropdown-item" href="#" onclick="banMember('${gID}', '${mID}')">${t('action_ban')}</a></li>
                        <li><a class="dropdown-item" href="#" onclick="unbanMember('${gID}', '${mID}')">${t('action_unban')}</a></li>
                        <li><hr class="dropdown-divider"></li>
                        <li><a class="dropdown-item text-danger" href="#" onclick="kickMember('${gID}', '${mID}')">${t('action_kick')}</a></li>
                    </ul>
                </div>
            </td>
        </tr>
    `}).join('');
}

export function banMember(groupId, userId) {
    const duration = prompt(t('prompt_ban_duration') || '请输入禁言时长（分钟），0 为解除禁言', "30");
    if (duration === null) return;
    const minutes = parseInt(duration);
    if (isNaN(minutes)) {
        alert(t.alert_invalid_number || '请输入有效的数字');
        return;
    }

    callBotApi('set_group_ban', {
        group_id: groupId,
        user_id: userId,
        duration: minutes * 60
    }).then(() => {
        alert(minutes > 0 ? (t.alert_banned || '禁言成功') : (t.alert_unbanned || '已解除禁言'));
    }).catch(e => alert((t.alert_op_failed || '操作失败: ') + e.message));
}

export function unbanMember(groupId, userId) {
    callBotApi('set_group_ban', {
        group_id: groupId,
        user_id: userId,
        duration: 0
    }).then(() => {
        alert(t('alert_unbanned') || '已解除禁言');
    }).catch(e => alert((t('alert_op_failed') || '操作失败: ') + e.message));
}

export function setCard(groupId, userId, currentCard) {
    const newCard = prompt(t('prompt_new_card') || '请输入新的名片内容', currentCard);
    if (newCard === null) return;

    callBotApi('set_group_card', {
        group_id: groupId,
        user_id: userId,
        card: newCard
    }).then(() => {
        alert(t('alert_card_set') || '名片设置成功');
        loadGroupMembers(groupId); // Reload to show change
    }).catch(e => alert((t('alert_op_failed') || '操作失败: ') + e.message));
}

export function kickMember(groupId, userId) {
    if (!confirm((t('confirm_kick_member') || '确定要将用户 {id} 移出群聊吗？').replace('{id}', userId))) return;
    callBotApi('set_group_kick', {
        group_id: groupId,
        user_id: userId
    }).then(() => {
        alert(t('alert_kicked') || '已移出群聊');
        loadGroupMembers(groupId);
    }).catch(e => alert((t('alert_op_failed') || '操作失败: ') + e.message));
}

export function leaveGroup() {
    if (!confirm((t('confirm_leave_group') || '确定要退出群聊 {id} 吗？').replace('{id}', currentGroupId))) return;
     callBotApi('set_group_leave', {
        group_id: currentGroupId
    }).then(() => {
        alert(t('alert_left_group') || '已退出群聊');
        if (window.refreshGroupList) window.refreshGroupList();
        document.getElementById('group-detail-empty').style.display = 'block';
        document.getElementById('group-detail-content').style.setProperty('display', 'none', 'important');
    }).catch(e => alert((t('alert_op_failed') || '操作失败: ') + e.message));
}

export function checkGroupMember() {
    const userId = document.getElementById('check-member-input').value.trim();
    const resultEl = document.getElementById('check-member-result');
    
    if (!userId) {
        resultEl.innerHTML = `<span class="text-danger">${t('alert_invalid_number') || '无效的 ID'}</span>`;
        return;
    }

    resultEl.innerHTML = `<span class="text-muted">${t('loading') || '正在查询...'}</span>`;
    
    callBotApi('get_group_member_info', {
        group_id: Number(currentGroupId),
        user_id: Number(userId),
        no_cache: true
    }).then(data => {
        if (data && (data.user_id || data.nickname)) {
            // Found
            let name = data.nickname || userId;
            let card = data.card || '';
            let text = (t('member_found') || '找到成员: {name} ({card})').replace('{name}', name).replace('{card}', card);
            resultEl.innerHTML = `<span class="text-success"><i class="bi bi-check-circle"></i> ${text}</span>`;
        } else {
            // Not found
            resultEl.innerHTML = `<span class="text-warning">${t('member_not_found') || '未找到该成员'}</span>`;
        }
    }).catch(e => {
        console.error(e);
        let msg = (t('check_error') || '查询出错: ') + e.message;
        if (e.message.includes('not found') || e.message.includes('不存在')) {
            msg = t('member_not_found') || '未找到该成员';
        }
        resultEl.innerHTML = `<span class="text-danger"><i class="bi bi-x-circle"></i> ${msg}</span>`;
    });
}
// End of module
