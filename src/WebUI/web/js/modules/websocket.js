/**
 * WebSocket management module
 */

import { authToken } from './auth.js';
import { pendingRequests } from './api.js';
import { handleRoutingEvent, handleSyncState } from './visualization.js';
import { updateWsStatus, showToast } from './ui.js';
import { fetchBots } from './bots.js';
import { addEventLog, renderLogMessage } from './logs.js';
import { t } from './i18n.js';

let wsSubscriber = null;
let wsReconnectAttempts = 0;
const MAX_WS_RECONNECT_ATTEMPTS = 5;
const WS_RECONNECT_DELAY = 5000;

export function initWebSocket() {
    if (wsSubscriber && wsSubscriber.readyState === WebSocket.CONNECTING) {
        return;
    }
    
    if (wsReconnectAttempts >= MAX_WS_RECONNECT_ATTEMPTS) {
        console.error('WebSocket reconnection limit reached');
        updateWsStatus('danger', t('status_error') || '连接失败');
        addEventLog({ type: 'system', message: 'WebSocket 连接失败次数过多，请检查网络或服务器状态' });
        return;
    }
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsPort = window.location.port ? `:${window.location.port}` : '';
    
    // Ensure we have the latest token from localStorage if the imported one is missing
    const token = authToken || localStorage.getItem('wxbot_token');
    
    let wsUrl = `${protocol}//${window.location.hostname}${wsPort}/ws/subscriber?role=subscriber`;
    if (token && token !== 'null' && token !== 'undefined') {
        wsUrl += `&token=${encodeURIComponent(token)}`;
    } else {
        console.warn('[WS] No auth token found, connection might be rejected');
    }
    
    console.log('[WS] Connecting to:', wsUrl.split('&token=')[0] + (token ? '&token=***' : ''));
    
    try {
        wsSubscriber = new WebSocket(wsUrl);
        window.wsSubscriber = wsSubscriber; // Attach to window for cleanup visibility
        
        // Connection timeout
        const connectTimeout = setTimeout(() => {
            if (wsSubscriber && wsSubscriber.readyState === WebSocket.CONNECTING) {
                console.warn('WebSocket connection timeout');
                wsSubscriber.close();
            }
        }, 10000);
        
        wsSubscriber.onopen = () => {
            clearTimeout(connectTimeout);
            wsReconnectAttempts = 0;
            updateWsStatus('success', t('status_connected') || '已连接');
            addEventLog({ type: 'system', message: t('ws_connected') || 'WebSocket 连接成功' });
            fetchBots();
        };
        
        wsSubscriber.onerror = (error) => {
            clearTimeout(connectTimeout);
            console.error('WebSocket connection error details:', {
                url: wsUrl.split('&token=')[0] + '...',
                readyState: wsSubscriber.readyState,
                error: error
            });
            updateWsStatus('warning', t('status_error') || '连接错误');
            addEventLog({ type: 'system', message: (t('ws_error') || 'WebSocket 连接错误') + ': ' + (error.message || 'Check console') });
        };
        
        wsSubscriber.onclose = () => {
            clearTimeout(connectTimeout);
            updateWsStatus('danger', t('status_disconnected') || '已断开');
            addEventLog({ type: 'system', message: t('ws_disconnected_retry') || 'WebSocket 连接断开，正在尝试重连...' });
            
            wsSubscriber = null;
            wsReconnectAttempts++;
            setTimeout(initWebSocket, WS_RECONNECT_DELAY);
        };

        wsSubscriber.onmessage = (evt) => {
            try {
                const data = JSON.parse(evt.data);
                console.log("WS Recv:", data);
                
                if (data.type === 'routing_event') {
                    handleRoutingEvent(data);
                    return;
                }

                if (data.type === 'sync_state') {
                    handleSyncState(data);
                    return;
                }

                // Handle API Responses
                if (data.echo && pendingRequests.has(data.echo)) {
                    const req = pendingRequests.get(data.echo);
                    clearTimeout(req.timeout);
                    pendingRequests.delete(data.echo);
                    req.resolve(data);
                }

                // Auto-refresh bot list
                if (data.echo === 'internal_get_login_info' || 
                    (data.post_type === 'meta_event' && data.meta_event_type === 'lifecycle')) {
                    setTimeout(fetchBots, 500);
                }

                if (data.post_type === 'log') {
                    const log = data.data;
                    const selfId = data.self_id || '';
                    
                    if (log.message && log.message.includes('Client connected:')) {
                        setTimeout(fetchBots, 1000);
                    }

                    // Widget Log Update
                    const selector = document.getElementById('log-bot-selector');
                    const filter = selector ? selector.value : '';
                    if (filter === '' || (filter === 'system' && !selfId) || filter === selfId) {
                        const container = document.getElementById('log-container');
                        if (container) {
                            const div = document.createElement('div');
                            div.className = 'log-entry';
                            div.innerHTML = `<span class="log-time">[${log.time}]</span><span class="log-level-${log.level}">${log.level}</span>: ${renderLogMessage(log.message)}`;
                            container.appendChild(div);
                            if (container.children.length > 500) {
                                container.removeChild(container.firstChild);
                            }
                            const autoRefresh = document.getElementById('auto-refresh-logs');
                            if (autoRefresh && autoRefresh.checked) {
                                container.scrollTop = container.scrollHeight;
                            }
                        }
                    }

                    // Full Page Log Update
                    const selectorFull = document.getElementById('log-bot-selector-full');
                    const filterFull = selectorFull ? selectorFull.value : '';
                    if (filterFull === '' || (filterFull === 'system' && !selfId) || filterFull === selfId) {
                        const containerFull = document.getElementById('log-container-full');
                        if (containerFull) {
                            const div = document.createElement('div');
                            div.className = 'log-entry';
                            div.innerHTML = `<span class="log-time">[${log.time}]</span><span class="log-level-${log.level}">${log.level}</span>: ${renderLogMessage(log.message)}`;
                            containerFull.appendChild(div);
                            if (containerFull.children.length > 500) {
                                containerFull.removeChild(containerFull.firstChild);
                            }
                            const autoRefreshFull = document.getElementById('auto-refresh-logs-full');
                            if (autoRefreshFull && autoRefreshFull.checked) {
                                containerFull.scrollTop = containerFull.scrollHeight;
                            }
                        }
                    }
                }
                
                addEventLog(data);
            } catch (e) {
                console.error('WS message parse error:', e);
            }
        };
    } catch (e) {
        console.error('WebSocket initialization failed:', e);
        updateWsStatus('danger', '初始化失败');
        wsReconnectAttempts++;
        setTimeout(initWebSocket, WS_RECONNECT_DELAY);
    }
}

export function closeWebSocket() {
    if (wsSubscriber) {
        console.log('[WebSocket] Closing connection...');
        wsSubscriber.onclose = null; // Prevent reconnection
        wsSubscriber.close();
        wsSubscriber = null;
    }
}
