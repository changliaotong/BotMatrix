import { fetchWithAuth, callBotApi } from './api.js';
import { t } from './i18n.js';
import { currentBotId } from './bots.js';
import { addEventLog } from './logs.js';
import { showToast } from './ui.js';

export let currentFriends = [];
export let currentFriendId = null;
export let friendSortBy = 'name';
export let friendSortAsc = true;

export async function refreshFriendList(botId = null) {
    const listEl = document.getElementById('friend-list');
    if (!listEl) return;
    
    listEl.innerHTML = '<div class="text-center p-4"><div class="spinner-border text-primary"></div></div>';
    
    const targetBotId = botId || (document.getElementById('global-bot-selector-friends') ? document.getElementById('global-bot-selector-friends').value : null) || currentBotId;
    if (!targetBotId) {
        listEl.innerHTML = '<div class="text-center p-4 text-muted">请先选择机器人</div>';
        return;
    }

    try {
        const url = `/api/contacts?bot_id=${encodeURIComponent(targetBotId)}&refresh=true`;
        const response = await fetchWithAuth(url);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        
        const contacts = await response.json();
        const friends = (contacts || []).filter(c => c && c.bot_id == targetBotId && c.type === 'private');
        
        currentFriends = friends;
        renderFriends(friends);
        window.renderFriends = () => renderFriends(currentFriends);
        
        const countEl = document.getElementById('friend-count');
        if (countEl) countEl.innerText = `共 ${friends.length} 个好友`;
    } catch (e) {
        console.error('Failed to fetch friends:', e);
        listEl.innerHTML = `<div class="text-center p-4 text-danger">加载失败: ${e.message}</div>`;
    }
}

export function renderFriends(friends = null) {
    if (!friends) friends = currentFriends;
    
    const listEl = document.getElementById('friend-list');
    if (!listEl) return;

    // Filter
    const searchInput = document.getElementById('friend-search');
    const keyword = (searchInput ? searchInput.value : '').toLowerCase();
    let filtered = (friends || []).filter(f => {
        if (!f) return false;
        const name = (f.name || '').toLowerCase();
        const id = (f.id || '').toString();
        return name.includes(keyword) || id.includes(keyword);
    });
    
    // Sort
    filtered.sort((a, b) => {
        let res = 0;
        if (friendSortBy === 'name') {
            const nameA = a ? (a.name || '') : '';
            const nameB = b ? (b.name || '') : '';
            res = nameA.localeCompare(nameB, 'zh');
        } else if (friendSortBy === 'id') {
            res = String(a ? a.id : '').localeCompare(String(b ? b.id : ''));
        }
        return friendSortAsc ? res : -res;
    });
    
    if (filtered.length === 0) {
        listEl.innerHTML = '<div class="text-center p-4 text-muted">未找到好友</div>';
        return;
    }

    listEl.innerHTML = filtered.map(f => {
        const name = f.name || 'Unknown';
        const id = f.id;
        let avatar = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
        // Use proxy for avatar reliability
        avatar = `/api/proxy/avatar?url=${encodeURIComponent(avatar)}`;
        const safeName = (name || '').replace(/'/g, "\\'").replace(/"/g, '&quot;');
        
        return `
            <div class="list-group-item list-group-item-action p-0 ${id == currentFriendId ? 'active' : ''}">
                <div class="d-flex align-items-center">
                    <div class="ps-2 pe-1">
                        <input type="checkbox" class="form-check-input mass-send-checkbox-private" data-id="${id}" data-type="private" data-name="${safeName}" data-bot-id="${f.bot_id || ''}" onclick="event.stopPropagation()">
                    </div>
                    <button class="btn btn-link text-decoration-none text-start p-2 flex-grow-1 border-0" style="color: inherit; text-align: left;" onclick="selectFriend('${id}', '${safeName}')">
                        <div class="d-flex w-100 align-items-center">
                             <div class="rounded-circle me-2" style="width: 32px; height: 32px; overflow: hidden; background: #f0f0f0;">
                                <img src="${avatar}" style="width: 100%; height: 100%; object-fit: cover;" onerror="this.src='https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/person.svg'">
                            </div>
                            <div style="overflow: hidden;">
                                <h6 class="mb-0 text-truncate" style="max-width: 150px;">${name}</h6>
                                <small class="${id == currentFriendId ? 'text-light' : 'text-muted'}">${id}</small>
                            </div>
                        </div>
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

export function filterFriends() {
    renderFriends();
}

export function sortFriends(by) {
    if (friendSortBy === by) {
        friendSortAsc = !friendSortAsc;
    } else {
        friendSortBy = by;
        friendSortAsc = true;
    }
    
    // Update buttons
    document.querySelectorAll('[id^="btn-sort-friend-"]').forEach(btn => btn.classList.remove('active'));
    const activeBtn = document.getElementById(`btn-sort-friend-${by}`);
    if (activeBtn) activeBtn.classList.add('active');
    
    renderFriends();
}

export function selectFriend(id, name) {
    currentFriendId = id;
    const nameEl = document.getElementById('current-friend-name');
    const idEl = document.getElementById('current-friend-id');
    const avatarEl = document.getElementById('current-friend-avatar');
    
    if (nameEl) nameEl.innerText = name;
    if (idEl) idEl.innerText = id;
    
    if (avatarEl) {
        let avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
        avatarEl.src = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
        avatarEl.style.display = 'block';
        avatarEl.onerror = function() {
            this.src = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/person.svg';
        };
    }
    
    // Show details
    const emptyEl = document.getElementById('friend-detail-empty');
    const contentEl = document.getElementById('friend-detail-content');
    
    if (emptyEl) emptyEl.style.display = 'none';
    if (contentEl) contentEl.style.setProperty('display', 'flex', 'important');
    
    // Re-render list to highlight active
    filterFriends();
}

export async function sendFriendMsg() {
    const input = document.getElementById('friend-msg-input');
    if (!input) return;
    
    const text = input.value;
    if (!text.trim()) return;
    
    try {
        await callBotApi('send_private_msg', {
            user_id: currentFriendId,
            message: text
        });
        showToast(t('alert_send_success') || '发送成功', 'success');
        input.value = '';
        addEventLog({type: 'message', message: `发送给好友(${currentFriendId}): ${text}`});
    } catch (e) {
        showToast((t('alert_op_failed') || '操作失败: ') + e.message, 'danger');
    }
}

export function ensureFriendListLoaded() {
    if (currentFriends.length === 0) {
        refreshFriendList();
    }
}
// End of module
