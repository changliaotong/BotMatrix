/**
 * System Statistics and Charts module
 */

import { fetchWithAuth } from './api.js';
import { currentLang, translations } from './i18n.js';
import { currentGroups, renderGroups } from './groups.js';
import { renderMembers, currentMembers } from './members.js';
import { formatBytes } from './utils.js';
import { showToast } from './ui.js?v=1.1.88';

export let latestChatStats = {
    group_stats: {},
    user_stats: {},
    group_names: {},
    user_names: {}
};
export let serverStartTime = null;

export let lastMsgCount = 0;
export let lastSentCount = 0;
export let msgHistory = [];
export let sentHistory = [];
export let recvHistory = [];
export const MSG_HISTORY_SIZE = 30; // 30 * 2s = 60s
export const MAX_CHART_POINTS = 1800; // 1800 * 2s = 60 minutes

export let memChart = null;
export let cpuChart = null;
export let msgChart = null;

/**
 * 保存统计到缓存
 */
export function saveStatsToCache(stats) {
    localStorage.setItem('bot_stats_cache', JSON.stringify(stats));
}

export function loadStatsFromCache() {
    const cached = localStorage.getItem('bot_stats_cache');
    if (cached) {
        try {
            const stats = JSON.parse(cached);
            const setVal = (id, val) => {
                const el = document.getElementById(id);
                if (el) el.innerText = val;
            };

            setVal('metric-bots', stats.bot_count || 0);
            setVal('metric-bots-offline', stats.bot_count_offline || 0);
            setVal('metric-bots-total', stats.bot_count_total || 0);
            setVal('metric-workers', stats.worker_count || 0);
            
            if (document.getElementById('metric-mem-total') && stats.memory_total) {
                document.getElementById('metric-mem-total').innerText = formatBytes(stats.memory_total);
            }

            if (document.getElementById('metric-cpu-model') && stats.cpu_model) {
                 let model = stats.cpu_model;
                 if (model.length > 25) model = model.substring(0, 25) + '...';
                 document.getElementById('metric-cpu-model').innerText = model;
                 const card = document.getElementById('metric-cpu-model').closest('.stat-card');
                 if (card) card.title = stats.cpu_model;
            }
            if (document.getElementById('metric-cpu-cores') && stats.cpu_cores_physical) {
                 document.getElementById('metric-cpu-cores').innerText = (stats.cpu_cores_physical || 0) + 'P/' + (stats.cpu_cores_logical || 0) + 'L 核';
            }
            if (document.getElementById('metric-cpu-freq') && stats.cpu_freq) {
                 document.getElementById('metric-cpu-freq').innerText = (stats.cpu_freq || 0).toFixed(0) + ' MHz';
            }
        } catch (e) {
            console.error("Failed to load cached stats", e);
        }
    }
}

export function initCharts() {
    if (typeof Chart === 'undefined') {
        console.error('Chart.js is not loaded');
        return;
    }
    
    const destroyExisting = (id) => {
        const el = document.getElementById(id);
        if (el) {
            const existingChart = Chart.getChart(el);
            if (existingChart) existingChart.destroy();
        }
    };

    destroyExisting('memChart');
    destroyExisting('cpuChart');
    destroyExisting('msgChart');
    
    const safeInit = (id, callback) => {
        const el = document.getElementById(id);
        if (el) {
            try {
                callback(el);
            } catch (e) {
                console.error(`Error initializing chart ${id}:`, e);
            }
        }
    };

    safeInit('memChart', (el) => {
        const t = translations[currentLang] || translations['zh-CN'];
        memChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [{
                label: t.mem_alloc || 'Memory Alloc (MB)',
                data: [],
                borderColor: '#0d6efd',
                tension: 0.4,
                fill: true,
                backgroundColor: 'rgba(13, 110, 253, 0.1)'
            }]},
            options: {
                responsive: true,
                animation: false,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: { 
                    y: { beginAtZero: true, grid: { display: true, drawBorder: false } },
                    x: { grid: { display: false } }
                }
            }
        });
    });

    safeInit('cpuChart', (el) => {
        const t = translations[currentLang] || translations['zh-CN'];
        cpuChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [{
                label: t.cpu_usage || 'CPU Usage (%)',
                data: [],
                borderColor: '#ffc107',
                tension: 0.4,
                fill: true,
                backgroundColor: 'rgba(255, 193, 7, 0.1)'
            }]},
            options: {
                responsive: true,
                animation: false,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: { 
                    y: { beginAtZero: true, max: 100, grid: { display: true, drawBorder: false } },
                    x: { grid: { display: false } }
                }
            }
        });
    });

    safeInit('msgChart', (el) => {
        const t = translations[currentLang] || translations['zh-CN'];
        msgChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [
                { label: t.received || '接收', data: [], borderColor: '#198754', tension: 0.4, fill: false, borderWidth: 2 },
                { label: t.sent || '发送', data: [], borderColor: '#0d6efd', tension: 0.4, fill: false, borderWidth: 2 },
                { label: t.total || '总量', data: [], borderColor: '#6c757d', tension: 0.4, fill: false, borderDash: [5, 5], borderWidth: 1 }
            ]},
            options: {
                responsive: true,
                animation: false,
                maintainAspectRatio: false,
                interaction: { mode: 'index', intersect: false },
                plugins: { 
                    legend: { display: true, position: 'top', labels: { boxWidth: 10, usePointStyle: true } },
                    tooltip: { mode: 'index', intersect: false }
                },
                scales: { 
                    y: { beginAtZero: true, grid: { display: true, drawBorder: false } },
                    x: { grid: { display: false } }
                }
            }
        });
    });
}

export async function updateStats() {
    const authToken = window.authToken || localStorage.getItem('wxbot_token');
    if (!authToken) return;
    try {
        const response = await fetchWithAuth('/api/stats');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        window.latestStats = data;
            
        const setVal = (id, val) => {
            const el = document.getElementById(id);
            if (el) el.innerText = val !== undefined && val !== null ? val : '0';
        };

        setVal('metric-goroutines', data.goroutines);
        
        const memEl = document.getElementById('metric-mem');
        if (memEl) {
            const used = data.memory_used !== undefined ? data.memory_used : (data.memory_alloc !== undefined ? data.memory_alloc : 0);
            memEl.innerText = formatBytes(used);
            const memBar = document.getElementById('metric-mem-bar');
            if (data.memory_total && memBar) {
                const pct = (used / data.memory_total) * 100;
                memBar.style.width = pct.toFixed(1) + '%';
            }
        }

        setVal('metric-mem-total', data.memory_total !== undefined ? formatBytes(data.memory_total) : '0 B');
        setVal('metric-bots', data.bot_count);
        setVal('metric-bots-offline', data.bot_count_offline);
        setVal('metric-bots-total', data.bot_count_total);
        setVal('metric-workers', data.worker_count);

        const cpuModelEl = document.getElementById('metric-cpu-model');
        if (cpuModelEl && data.cpu_model) {
             let model = data.cpu_model;
             if (model.length > 25) model = model.substring(0, 25) + '...';
             cpuModelEl.innerText = model;
             const card = cpuModelEl.closest('.stat-card');
             if (card) card.title = data.cpu_model;
        }

        setVal('metric-cpu-cores', (data.cpu_cores_physical || 0) + 'P/' + (data.cpu_cores_logical || 0) + 'L 核');
        setVal('metric-cpu-freq', (data.cpu_freq || 0).toFixed(0) + ' MHz');
        setVal('metric-os-plat', data.os_platform || 'Unknown');

        const osVerEl = document.getElementById('metric-os-ver');
        if (osVerEl) {
            const t = translations[currentLang] || translations['zh-CN'] || {};
            osVerEl.innerText = (t.version_label || '版本: ') + (data.os_version || 'Unknown');
        }

        if (data.os_arch || (data.host_info && data.host_info.kernelArch)) {
            const archEl = document.getElementById('metric-os-arch');
            if (archEl) archEl.innerText = data.os_arch || data.host_info.kernelArch;
        }

        if (data.start_time) serverStartTime = data.start_time;

        setVal('metric-groups-today', data.active_groups_today);
        setVal('metric-groups-total', data.active_groups);
        setVal('metric-users-today', data.active_users_today);
        setVal('metric-users-total', data.active_users);
        setVal('metric-msgs-total', data.message_count);
        setVal('metric-sent-total', data.sent_message_count);

        const now = new Date().toLocaleTimeString();
        let justInitialized = false;

        // Init History
        if (data.cpu_trend && cpuChart && cpuChart.data.labels.length === 0) {
            data.cpu_trend.forEach(() => cpuChart.data.labels.push(''));
            cpuChart.data.datasets[0].data = data.cpu_trend;
            cpuChart.update();
            justInitialized = true;
        }
        if (data.mem_trend && memChart && memChart.data.labels.length === 0) {
            data.mem_trend.forEach(() => memChart.data.labels.push(''));
            memChart.data.datasets[0].data = data.mem_trend.map(v => v / 1024 / 1024);
            memChart.update();
            justInitialized = true;
        }
        if (data.msg_trend && msgChart && msgChart.data.labels.length === 0) {
            const historyWindowSize = 12; 
            msgHistory = [...data.msg_trend];
            sentHistory = data.sent_trend ? [...data.sent_trend] : new Array(msgHistory.length).fill(0);
            recvHistory = data.recv_trend ? [...data.recv_trend] : new Array(msgHistory.length).fill(0);

            const chartDataRecv = [];
            const chartDataSent = [];
            const chartDataTotal = [];
            const labels = [];
            
            for (let i = 0; i < msgHistory.length; i++) {
                let start = Math.max(0, i - historyWindowSize + 1);
                let sumTotal = 0, sumSent = 0, sumRecv = 0;
                for (let j = start; j <= i; j++) {
                    sumTotal += msgHistory[j] || 0;
                    sumSent += sentHistory[j] || 0;
                    sumRecv += recvHistory[j] || 0;
                }
                chartDataRecv.push(sumRecv);
                chartDataSent.push(sumSent);
                chartDataTotal.push(sumTotal);
                labels.push('');
            }
            
            if (msgHistory.length > MSG_HISTORY_SIZE) {
                msgHistory = msgHistory.slice(-MSG_HISTORY_SIZE);
                sentHistory = sentHistory.slice(-MSG_HISTORY_SIZE);
                recvHistory = recvHistory.slice(-MSG_HISTORY_SIZE);
            }
            msgChart.data.labels = labels;
            msgChart.data.datasets[0].data = chartDataRecv;
            msgChart.data.datasets[1].data = chartDataSent;
            msgChart.data.datasets[2].data = chartDataTotal;
            msgChart.update();
            
            lastMsgCount = data.message_count || 0;
            lastSentCount = data.sent_message_count || 0;
            justInitialized = true;
        }

        // Real-time Updates
        if (memChart && !justInitialized) {
            if (memChart.data.labels.length > MAX_CHART_POINTS) {
                memChart.data.labels.shift();
                memChart.data.datasets[0].data.shift();
            }
            memChart.data.labels.push(now);
            memChart.data.datasets[0].data.push(data.memory_alloc / 1024 / 1024);
            memChart.update();
        }
        
        if (msgChart && !justInitialized) {
            const currentTotal = data.message_count || 0;
            const currentSent = data.sent_message_count || 0;
            const diffTotal = currentTotal - lastMsgCount;
            const diffSent = currentSent - lastSentCount;
            
            const valTotal = lastMsgCount === 0 ? 0 : (diffTotal < 0 ? 0 : diffTotal);
            const valSent = lastSentCount === 0 ? 0 : (diffSent < 0 ? 0 : diffSent);
            const valRecv = Math.max(0, valTotal - valSent);

            msgHistory.push(valTotal);
            sentHistory.push(valSent);
            recvHistory.push(valRecv);

            if (msgHistory.length > MSG_HISTORY_SIZE) {
                msgHistory.shift(); sentHistory.shift(); recvHistory.shift();
            }

            const sumTotal = msgHistory.reduce((a, b) => a + b, 0);
            const sumSent = sentHistory.reduce((a, b) => a + b, 0);
            const sumRecv = recvHistory.reduce((a, b) => a + b, 0);

            if (msgChart.data.labels.length > MAX_CHART_POINTS) {
                msgChart.data.labels.shift();
                msgChart.data.datasets[0].data.shift();
                msgChart.data.datasets[1].data.shift();
                msgChart.data.datasets[2].data.shift();
            }
            msgChart.data.labels.push(now);
            msgChart.data.datasets[0].data.push(sumRecv);
            msgChart.data.datasets[1].data.push(sumSent);
            msgChart.data.datasets[2].data.push(sumTotal);
            msgChart.update();

            lastMsgCount = currentTotal;
            lastSentCount = currentSent;
        }

        if (cpuChart && !justInitialized) {
            if (cpuChart.data.labels.length > MAX_CHART_POINTS) {
                cpuChart.data.labels.shift();
                cpuChart.data.datasets[0].data.shift();
            }
            cpuChart.data.labels.push(now);
            cpuChart.data.datasets[0].data.push(data.cpu_percent || 0);
            cpuChart.update();
        }

        saveStatsToCache({
            bot_count: data.bot_count,
            bot_count_offline: data.bot_count_offline || 0,
            bot_count_total: data.bot_count_total || 0,
            memory_total: data.memory_total,
            cpu_model: data.cpu_model,
            cpu_cores_physical: data.cpu_cores_physical,
            cpu_cores_logical: data.cpu_cores_logical,
            cpu_freq: data.cpu_freq,
            worker_count: data.worker_count
        });

    } catch (e) {
        console.error('Failed to update stats:', e);
    }
}

export function rotateCombinedStats() {
    const container = document.getElementById('combined-stats-container');
    if (!container) return;
    const slides = container.querySelectorAll('.stats-slide');
    if (slides.length <= 1) return;
    
    let activeIndex = Array.from(slides).findIndex(s => s.classList.contains('active'));
    if (activeIndex === -1) {
        slides[0].classList.add('active', 'fade-in');
        slides[0].style.display = 'block';
        return;
    }
    
    const currentSlide = slides[activeIndex];
    currentSlide.classList.remove('fade-in');
    currentSlide.classList.add('fade-out');
    
    setTimeout(() => {
        currentSlide.classList.remove('active', 'fade-out');
        currentSlide.style.display = 'none';
        
        const nextIndex = (activeIndex + 1) % slides.length;
        const nextSlide = slides[nextIndex];
        nextSlide.style.display = 'block';
        nextSlide.classList.add('active', 'fade-in');
        
        const iconElement = document.querySelector('#combined-stats-icon i');
        if (iconElement) {
            const icons = ['bi-activity', 'bi-people', 'bi-chat-text'];
            iconElement.className = `bi ${icons[nextIndex]}`;
        }
    }, 500);
}


/**
 * 获取聊天统计数据
 */
export async function updateChatStats() {
    const authToken = window.authToken || localStorage.getItem('wxbot_token');
    if (!authToken) return;
    try {
        const response = await fetchWithAuth('/api/stats/chat');
        const data = await response.json();
        latestChatStats = data;
        window.latestChatStats = data;
        
        // 渲染今日排行榜
        renderTopList('top-groups', data.group_stats_today, data.group_names, 'Group');
        renderTopList('top-users', data.user_stats_today, data.user_names, 'User');
        
        // 如果当前在群组标签页，刷新列表
        const activeTab = document.querySelector('.nav-link.active')?.getAttribute('href');
        if (activeTab === '#groups') {
            if (currentGroups && currentGroups.length > 0) {
                renderGroups(currentGroups);
            }
            if (typeof window.renderMembers === 'function') {
                window.renderMembers();
            }
        }
    } catch (e) {
        console.error("Update chat stats error:", e);
    }
}

/**
 * 获取显示名称
 */
export function getDisplayName(id, names, type) {
    if (names && names[id]) {
        const name = names[id];
        // 检查是否是 Go 语言 fmt.Sprintf 错误格式化的字符串: %!d(string=...)
        if (typeof name === 'string' && name.includes('%!d(string=')) {
            const match = name.match(/string=([^)]+)\)/);
            if (match && match[1]) return match[1];
        }
        return name;
    }
    if (type === 'Group') {
        const g = currentGroups.find(x => x.group_id == id);
        if (g) return g.group_name;
    }
    return `${type} ${id}`;
}

/**
 * 渲染排行榜
 */
export function renderTopList(elementId, stats, names, type) {
    const list = document.getElementById(elementId);
    if (!list) return;
    
    if (!stats || Object.keys(stats).length === 0) {
        list.innerHTML = '<li class="list-group-item text-muted text-center">暂无数据</li>';
        return;
    }

    const sorted = Object.entries(stats)
        .sort(([, a], [, b]) => b - a)
        .slice(0, 5);

    list.innerHTML = sorted.map(([id, count], index) => {
        const name = getDisplayName(id, names, type);
        let avatar = '';
        if (type === 'User') {
             avatar = `<img src="https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100" class="rounded-circle me-2" width="20" height="20" onerror="this.style.display='none'">`;
        } else if (type === 'Group') {
             avatar = `<img src="https://p.qlogo.cn/gh/${id}/${id}/100/" class="rounded-circle me-2" width="20" height="20" onerror="this.style.display='none'">`;
        }

        return `
            <li class="list-group-item d-flex justify-content-between align-items-center">
                <div class="text-truncate" style="max-width: 70%;">
                    <span class="badge bg-secondary bg-opacity-10 text-secondary me-2">#${index + 1}</span>
                    ${avatar}
                    <span title="${id}">
                        ${name}
                        ${names && names[id] ? `<small class="text-muted ms-1" style="font-size:0.8em">(${id})</small>` : ''}
                    </span>
                </div>
                <span class="badge bg-primary rounded-pill">${count}</span>
            </li>
        `;
    }).join('');
}

/**
 * 显示完整统计模态框
 */
export function showAllStats(type) {
    const stats = type === 'Group' ? latestChatStats.group_stats_today : latestChatStats.user_stats_today;
    const names = type === 'Group' ? latestChatStats.group_names : latestChatStats.user_names;
    
    const t = translations[currentLang] || translations['zh-CN'];
    const titleEl = document.getElementById('statsModalTitle');
    const headerEl = document.getElementById('statsModalNameHeader');
    const tbody = document.getElementById('statsModalBody');
    
    if (titleEl) {
        const titleKey = type === 'Group' ? 'top_active_groups' : 'top_active_users';
        titleEl.setAttribute('data-i18n', titleKey);
        titleEl.innerText = t[titleKey] || (type === 'Group' ? '今日活跃群组' : '今日龙王');
    }
    if (headerEl) {
        const headerKey = type === 'Group' ? 'group_name_default' : 'user_nickname';
        headerEl.setAttribute('data-i18n', headerKey);
        headerEl.innerText = t[headerKey] || (type === 'Group' ? '群组名称' : '用户昵称');
    }
    
    // Add translation for the "Action" column if needed
    const actionHeader = document.querySelector('#statsModal thead th:last-child');
    if (actionHeader) {
        actionHeader.setAttribute('data-i18n', 'count');
        actionHeader.innerText = t['count'] || '消息数';
    }
    
    if (!tbody) return;
    
    if (!stats || Object.keys(stats).length === 0) {
        tbody.innerHTML = `<tr><td colspan="3" class="text-center p-4">${t.no_data || '暂无数据'}</td></tr>`;
    } else {
        const sorted = Object.entries(stats).sort(([, a], [, b]) => b - a);
        tbody.innerHTML = sorted.map(([id, count], index) => {
            const name = getDisplayName(id, names, type);
            let avatar = '';
            if (type === 'User') {
                 let userAvatar = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
                 userAvatar = `/api/proxy/avatar?url=${encodeURIComponent(userAvatar)}`;
                 avatar = `<img src="${userAvatar}" class="rounded-circle me-2 border border-secondary border-opacity-25" width="32" height="32" onerror="this.src='https://ui-avatars.com/api/?name=U&background=random'">`;
            } else {
                 let groupAvatar = `https://p.qlogo.cn/gh/${id}/${id}/100/`;
                 groupAvatar = `/api/proxy/avatar?url=${encodeURIComponent(groupAvatar)}`;
                 avatar = `<img src="${groupAvatar}" class="rounded-circle me-2 border border-secondary border-opacity-25" width="32" height="32" onerror="this.src='https://ui-avatars.com/api/?name=G&background=random'">`;
            }
            
            return `
                <tr>
                    <td class="align-middle">${index + 1}</td>
                    <td class="text-truncate align-middle" style="max-width: 300px;" title="${id}">
                        <div class="d-flex align-items-center">
                            ${avatar}
                            <div class="overflow-hidden">
                                <div class="text-truncate">${name}</div>
                                <div class="text-muted small" style="font-size: 0.75rem;">${id}</div>
                            </div>
                        </div>
                    </td>
                    <td class="text-end align-middle fw-bold text-primary">${count}</td>
                </tr>
            `;
        }).join('');
    }
    
    const modalEl = document.getElementById('statsModal');
    if (modalEl) {
        const modal = bootstrap.Modal.getOrCreateInstance(modalEl);
        modal.show();
    }
}
