import { fetchWithAuth } from './api.js';
import { t } from './i18n.js';

/**
 * 渲染日志消息，支持长文本折叠
 * @param {string} message 日志内容
 * @returns {string} 渲染后的 HTML
 */
export function renderLogMessage(message) {
    const MAX_LEN = 300;
    if (!message || message.length <= MAX_LEN) return message;
    
    const shortMsg = message.substring(0, MAX_LEN) + '...';
    
    return `<span class="log-short" style="cursor:pointer; text-decoration:underline dotted; color: inherit;" onclick="toggleLogExpand(this)" title="${t('log_expand') || '点击展开'}">${shortMsg}</span>` + 
           `<span class="log-full" style="display:none; cursor:pointer; color: inherit;" onclick="toggleLogExpand(this)" title="${t('log_collapse') || '点击折叠'}">${message}</span>`;
}

/**
 * 切换日志显示状态（展开/折叠）
 * @param {HTMLElement} el 点击的元素
 */
export function toggleLogExpand(el) {
    const parent = el.parentElement;
    const shortSpan = parent.querySelector('.log-short');
    const fullSpan = parent.querySelector('.log-full');
    
    // 暂停日志自动刷新，方便用户阅读
    const autoRefreshFull = document.getElementById('auto-refresh-logs-full');
    if (autoRefreshFull) autoRefreshFull.checked = false;
    
    const autoRefreshWidget = document.getElementById('auto-refresh-logs');
    if (autoRefreshWidget) autoRefreshWidget.checked = false;

    if (el.classList.contains('log-short')) {
        shortSpan.style.display = 'none';
        fullSpan.style.display = 'inline';
    } else {
        shortSpan.style.display = 'inline';
        fullSpan.style.display = 'none';
    }
}

/**
 * 获取全屏日志
 */
export async function fetchLogsFull() {
    if (!window.authToken) return;
    const container = document.getElementById('log-container-full');
    if (!container) return;

    // 如果容器为空，显示加载中
    if (!container.innerHTML || container.innerHTML.trim() === '') {
        container.innerHTML = `<div class="text-center py-5 text-muted"><div class="spinner-border spinner-border-sm me-2"></div>${t('loading') || '加载中...'}</div>`;
    }

    const selector = document.getElementById('log-bot-selector-full');
    const botId = selector ? selector.value : '';

    try {
        const response = await fetchWithAuth(`/api/logs?bot_id=${botId}`);
        const data = await response.json();
        const logs = data.logs || [];
        
        container.innerHTML = logs.slice().reverse().map(log => `
            <div class="log-entry">
                <span class="log-time">[${log.time}]</span>
                <span class="log-level-${log.level}">${log.level}</span>: 
                ${renderLogMessage(log.message)}
            </div>
        `).join('');
        
        // 自动滚动到顶部
        container.scrollTop = 0;
    } catch (err) {
        console.error('获取日志失败:', err);
    }
}

/**
 * 更新全屏日志页面的机器人选择器
 * @param {Array} bots 机器人列表
 */
export function updateLogBotSelectorFull(bots) {
    const selector = document.getElementById('log-bot-selector-full');
    if (!selector) return;
    
    const current = selector.value;
    
    let html = `<option value="">${t('log_all') || '全部日志'}</option>`;
    html += `<option value="system">${t('log_system') || '系统日志'}</option>`;
    
    bots.forEach(b => {
        const name = b.nickname || b.self_id;
        const status = b.is_alive ? '' : ` (${t('status_offline')})`;
        html += `<option value="${b.self_id}">${name} (${b.platform})${status}</option>`;
    });
    
    selector.innerHTML = html;
    selector.value = current;
}

/**
 * 获取小组件日志
 */
export async function fetchLogs() {
    if (!window.authToken) return;
    const container = document.getElementById('log-container');
    if (!container) return;

    if (!container.innerHTML || container.innerHTML.trim() === '') {
        container.innerHTML = `<div class="text-center py-4 text-muted small"><div class="spinner-border spinner-border-sm me-2" style="width: 1rem; height: 1rem;"></div>${t('loading') || '加载中...'}</div>`;
    }

    const selector = document.getElementById('log-bot-selector');
    const botId = selector ? selector.value : '';

    try {
        const response = await fetchWithAuth(`/api/logs?bot_id=${botId}`);
        const data = await response.json();
        const logs = data.logs || [];
        
        container.innerHTML = logs.slice().reverse().map(log => `
            <div class="log-entry">
                <span class="log-time">[${log.time}]</span>
                <span class="log-level-${log.level}">${log.level}</span>: 
                ${renderLogMessage(log.message)}
            </div>
        `).join('');
        
        container.scrollTop = 0;
    } catch (err) {
        console.error('获取小组件日志失败:', err);
    }
}

/**
 * 添加事件日志（实时事件流）
 * @param {Object} data 事件数据
 */
export function addEventLog(data) {
    if (data.type === 'sync_state' || data.meta_event_type === 'heartbeat') {
        return; 
    }

    const container = document.getElementById('event-container');
    if (!container) {
        return;
    }

    if (container.querySelector('.animate-pulse') || container.innerText.includes('等待事件连接')) {
        container.innerHTML = '';
    }
    
    const div = document.createElement('div');
    div.className = 'log-entry';
    div.style.borderBottom = '1px solid var(--border-color)';
    div.style.padding = '8px 0';
    
    const time = new Date().toLocaleTimeString();
    
    // 格式化内容
    let content = '';
    let category = 'other';

    if (data.type === 'system') {
        category = 'system';
        content = `<span class="text-info">[SYSTEM]</span> ${data.message}`;
    } else if (data.type === 'routing_event') {
        category = 'routing';
        const direction = data.direction === 'bot_to_user' ? '→' : (data.direction === 'user_to_bot' ? '←' : '↔');
        content = `<span class="text-matrix">[FLOW]</span> <span class="text-info">${data.source}</span> ${direction} <span class="text-success">${data.target}</span> <span class="text-muted ms-2">(${data.msg_type})</span>`;
    } else if (data.post_type === 'message' || data.post_type === 'message_sent') {
        category = 'message';
        const sender = data.sender ? (data.sender.nickname || data.sender.card || data.user_id) : data.user_id;
        const group = data.group_id ? `[群:${data.group_id}] ` : '[私聊] ';
        let msg = data.message;
        if (typeof msg !== 'string') {
            msg = JSON.stringify(msg);
        }
        content = `<span class="text-success">[MSG]</span> ${group}<span class="fw-bold text-warning">${sender}</span>: ${msg}`;
    } else if (data.post_type === 'log') {
        // 日志不再显示在事件流中，因为已有专门的日志页面
        return;
    } else if (data.post_type === 'meta_event') {
        category = 'meta';
        if (data.meta_event_type === 'heartbeat') return; // 忽略心跳包
        content = `<span class="text-secondary">[META]</span> ${data.meta_event_type}`;
    } else {
        content = `<span class="text-muted">[EVENT]</span> <pre class="d-inline m-0" style="font-size:0.8em">${JSON.stringify(data)}</pre>`;
    }
    
    div.setAttribute('data-category', category);
    div.innerHTML = `<span class="log-time">${time}</span> ${content}`;
    
    // 根据当前过滤器决定是否隐藏
    const filterSelect = document.getElementById('event-filter');
    const filter = filterSelect ? filterSelect.value : 'all';
    if (filter !== 'all' && filter !== category) {
        div.style.display = 'none';
    }
    
    container.prepend(div);
    
    // 限制条目数量
    if (container.children.length > 500) {
        container.removeChild(container.lastChild);
    }

    // 更新仪表板最近日志
    const dashLog = document.getElementById('recent-logs');
    if (dashLog) {
        const clone = div.cloneNode(true);
        clone.style.display = ''; // 仪表板不应用过滤器
        dashLog.prepend(clone);
        if (dashLog.children.length > 20) {
            dashLog.removeChild(dashLog.lastChild);
        }
    }
}

/**
 * 过滤事件显示
 * @param {string} category 类别
 */
export function filterEvents(category) {
    const container = document.getElementById('event-container');
    if (!container) return;

    const entries = container.querySelectorAll('.log-entry');
    entries.forEach(entry => {
        if (category === 'all' || entry.getAttribute('data-category') === category) {
            entry.style.display = '';
        } else {
            entry.style.display = 'none';
        }
    });
}

/**
 * 清除事件日志
 */
export function clearEvents() {
    const container = document.getElementById('event-container');
    if (container) {
        container.innerHTML = `<div class="text-muted text-center mt-5">${t('waiting_for_events') || '等待事件连接...'}</div>`;
    }
}

/**
 * 检查日志选择状态并暂停自动刷新
 */
export function checkLogSelectionAndPause() {
    const selection = window.getSelection();
    if (!selection || selection.toString().length === 0) return;

    let node = selection.anchorNode;
    if (!node) return;
    if (node.nodeType === 3) node = node.parentNode;

    const container = document.getElementById('log-container');
    const containerFull = document.getElementById('log-container-full');
    
    if ((container && container.contains(node)) || 
        (containerFull && containerFull.contains(node))) {
         
         const autoRefreshFull = document.getElementById('auto-refresh-logs-full');
         if (autoRefreshFull && autoRefreshFull.checked) {
             autoRefreshFull.checked = false;
         }
         
         const autoRefreshWidget = document.getElementById('auto-refresh-logs');
         if (autoRefreshWidget && autoRefreshWidget.checked) {
             autoRefreshWidget.checked = false;
         }
    }
}

// 绑定全局事件
document.addEventListener('mouseup', checkLogSelectionAndPause);
document.addEventListener('keyup', (e) => {
     if (e.shiftKey) checkLogSelectionAndPause();
});

// 全局绑定
window.renderLogMessage = renderLogMessage;
window.toggleLogExpand = toggleLogExpand;
window.fetchLogsFull = fetchLogsFull;
window.updateLogBotSelectorFull = updateLogBotSelectorFull;
window.fetchLogs = fetchLogs;
window.addEventLog = addEventLog;
window.clearEvents = clearEvents;
window.checkLogSelectionAndPause = checkLogSelectionAndPause;
