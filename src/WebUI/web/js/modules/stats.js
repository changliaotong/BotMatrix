/**
 * System Statistics and Charts module
 */

import { fetchWithAuth } from './api.js';
import { t } from './i18n.js';
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
                 document.getElementById('metric-cpu-cores').innerText = (stats.cpu_cores_physical || 0) + 'P/' + (stats.cpu_cores_logical || 0) + 'L ' + t('cores_label');
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
    console.debug('initCharts called');
    if (typeof Chart === 'undefined') {
        console.error('Chart.js is not loaded');
        return;
    }
    
    const destroyExisting = (id) => {
        const el = document.getElementById(id);
        if (el) {
            const existingChart = Chart.getChart(el);
            if (existingChart) {
                console.debug(`Destroying existing chart for ${id}`);
                existingChart.destroy();
            }
        }
    };

    destroyExisting('memChart');
    destroyExisting('cpuChart');
    destroyExisting('msgChart');
    
    const safeInit = (id, callback) => {
        const el = document.getElementById(id);
        if (el) {
            try {
                console.debug(`Initializing chart ${id}`);
                callback(el);
            } catch (e) {
                console.error(`Error initializing chart ${id}:`, e);
            }
        } else {
            console.debug(`Element ${id} not found for chart initialization`);
        }
    };

    safeInit('memChart', (el) => {
        memChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [{
                label: t('memory_used'),
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
        cpuChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [{
                label: t('stat_cpu'),
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
        msgChart = new Chart(el.getContext('2d'), {
            type: 'line',
            data: { labels: [], datasets: [
                { label: t('label_recv'), data: [], borderColor: '#198754', tension: 0.4, fill: false, borderWidth: 2 },
                { label: t('label_sent'), data: [], borderColor: '#0d6efd', tension: 0.4, fill: false, borderWidth: 2 },
                { label: t('total_messages'), data: [], borderColor: '#6c757d', tension: 0.4, fill: false, borderDash: [5, 5], borderWidth: 1 }
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

export async function updateStats(providedData = null) {
    console.debug('updateStats called', providedData ? 'with data' : 'fetching data');
    const authToken = window.authToken || localStorage.getItem('wxbot_token');
    if (!authToken && !providedData) {
        console.debug('updateStats: no authToken and no provided data');
        return;
    }
    try {
        let data;
        if (providedData) {
            data = providedData;
        } else {
            const response = await fetchWithAuth('/api/stats');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            data = await response.json();
        }
        
        window.latestStats = data;
        console.debug('updateStats: processing data', data);
            
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
            osVerEl.innerText = t('version_label') + (data.os_version || 'Unknown');
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

        // CPU Chart Update
        if (cpuChart) {
            if (data.cpu_trend && data.cpu_trend.length > 0 && (cpuChart.data.datasets[0].data.length === 0 || justInitialized)) {
                console.debug('Initializing cpuChart trend data, points:', data.cpu_trend.length);
                cpuChart.data.labels = data.cpu_trend.map(() => '');
                cpuChart.data.datasets[0].data = [...data.cpu_trend];
                cpuChart.update();
                justInitialized = true;
            } else if (data.cpu_usage !== undefined) {
                const cpuVal = typeof data.cpu_usage === 'string' ? parseFloat(data.cpu_usage) : data.cpu_usage;
                cpuChart.data.labels.push('');
                cpuChart.data.datasets[0].data.push(cpuVal);
                if (cpuChart.data.labels.length > 60) {
                    cpuChart.data.labels.shift();
                    cpuChart.data.datasets[0].data.shift();
                }
                cpuChart.update();
            }
        }

        // Memory Chart Update
        if (memChart) {
            if (data.mem_trend && data.mem_trend.length > 0 && (memChart.data.datasets[0].data.length === 0 || justInitialized)) {
                console.debug('Initializing memChart trend data, points:', data.mem_trend.length);
                memChart.data.labels = data.mem_trend.map(() => '');
                memChart.data.datasets[0].data = data.mem_trend.map(v => parseFloat((v / 1024 / 1024).toFixed(1)));
                memChart.update();
                justInitialized = true;
            } else if (data.memory_used !== undefined) {
                memChart.data.labels.push('');
                memChart.data.datasets[0].data.push(parseFloat((data.memory_used / 1024 / 1024).toFixed(1)));
                if (memChart.data.labels.length > 60) {
                    memChart.data.labels.shift();
                    memChart.data.datasets[0].data.shift();
                }
                memChart.update();
            }
        }

        // Message Chart Update
        if (msgChart) {
            if (data.msg_trend && data.msg_trend.length > 0 && (msgChart.data.datasets[0].data.length === 0 || justInitialized)) {
                console.debug('Initializing msgChart trend data, points:', data.msg_trend.length);
                msgChart.data.labels = data.msg_trend.map(() => '');
                const sentTrend = data.sent_trend || data.msg_trend;
                const recvTrend = data.recv_trend || data.msg_trend;
                msgChart.data.datasets[0].data = sentTrend.map(v => parseFloat(v));
                msgChart.data.datasets[1].data = recvTrend.map(v => parseFloat(v));
                msgChart.update();
            } else if (data.message_count !== undefined) {
                msgChart.data.labels.push('');
                // 使用传入的 msg_per_sec/sent_per_sec 或计算出的增量
                const sentVal = parseFloat(data.sent_per_sec || 0);
                const recvVal = parseFloat(data.msg_per_sec || 0);
                msgChart.data.datasets[0].data.push(sentVal);
                msgChart.data.datasets[1].data.push(recvVal);
                if (msgChart.data.labels.length > 60) {
                    msgChart.data.labels.shift();
                    msgChart.data.datasets[0].data.shift();
                    msgChart.data.datasets[1].data.shift();
                }
                msgChart.update();
            }
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
        list.innerHTML = `<li class="flex items-center justify-center py-4 text-gray-500 text-xs italic">${t('no_data')}</li>`;
        return;
    }

    const sorted = Object.entries(stats)
        .sort(([, a], [, b]) => b - a)
        .slice(0, 5);

    list.innerHTML = sorted.map(([id, count], index) => {
        const name = getDisplayName(id, names, type);
        let avatar = '';
        if (type === 'User') {
             avatar = `<img src="https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100" class="w-8 h-8 rounded-full border border-black/5 dark:border-white/10" onerror="this.src='https://ui-avatars.com/api/?name=U&background=random'">`;
        } else if (type === 'Group') {
             avatar = `<img src="https://p.qlogo.cn/gh/${id}/${id}/100/" class="w-8 h-8 rounded-full border border-black/5 dark:border-white/10" onerror="this.src='https://ui-avatars.com/api/?name=G&background=random'">`;
        }

        return `
            <li class="flex items-center justify-between p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5 hover:border-matrix/30 transition-all group">
                <div class="flex items-center gap-3 min-w-0">
                    <div class="flex-shrink-0 w-6 text-[10px] font-bold text-gray-400">#${index + 1}</div>
                    ${avatar}
                    <div class="min-w-0">
                        <div class="text-xs font-bold dark:text-white truncate" title="${id}">${name}</div>
                        <div class="text-[8px] text-gray-500 mono truncate">${id}</div>
                    </div>
                </div>
                <div class="flex-shrink-0 px-2 py-1 rounded-lg bg-matrix/10 text-matrix text-[10px] font-bold mono">
                    ${count}
                </div>
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
    
    const titleEl = document.getElementById('statsModalTitle');
    const headerEl = document.getElementById('statsModalNameHeader');
    const tbody = document.getElementById('statsModalBody');
    
    if (titleEl) {
        const titleKey = type === 'Group' ? 'top_active_groups' : 'top_active_users';
        titleEl.setAttribute('data-i18n', titleKey);
        titleEl.innerText = t(titleKey) || (type === 'Group' ? '今日活跃群组' : '今日龙王');
    }
    if (headerEl) {
        const headerKey = type === 'Group' ? 'group_name_default' : 'user_nickname';
        headerEl.setAttribute('data-i18n', headerKey);
        headerEl.innerText = t(headerKey) || (type === 'Group' ? '群组名称' : '用户昵称');
    }
    
    if (!tbody) return;
    
    if (!stats || Object.keys(stats).length === 0) {
        tbody.innerHTML = `<tr><td colspan="3" class="px-6 py-12 text-center text-gray-500 italic">${t('no_data')}</td></tr>`;
    } else {
        const sorted = Object.entries(stats).sort(([, a], [, b]) => b - a);
        tbody.innerHTML = sorted.map(([id, count], index) => {
            const name = getDisplayName(id, names, type);
            let avatar = '';
            if (type === 'User') {
                 let userAvatar = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
                 userAvatar = `/api/proxy/avatar?url=${encodeURIComponent(userAvatar)}`;
                 avatar = `<img src="${userAvatar}" class="w-10 h-10 rounded-full border border-black/5 dark:border-white/10" onerror="this.src='https://ui-avatars.com/api/?name=U&background=random'">`;
            } else {
                 let groupAvatar = `https://p.qlogo.cn/gh/${id}/${id}/100/`;
                 groupAvatar = `/api/proxy/avatar?url=${encodeURIComponent(groupAvatar)}`;
                 avatar = `<img src="${groupAvatar}" class="w-10 h-10 rounded-full border border-black/5 dark:border-white/10" onerror="this.src='https://ui-avatars.com/api/?name=G&background=random'">`;
            }
            
            return `
                <tr class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                    <td class="px-6 py-4">
                        <span class="text-xs font-bold text-gray-400">#${index + 1}</span>
                    </td>
                    <td class="px-6 py-4">
                        <div class="flex items-center gap-3">
                            ${avatar}
                            <div class="min-w-0">
                                <div class="text-sm font-bold dark:text-white truncate max-w-[200px]" title="${name}">${name}</div>
                                <div class="text-[10px] text-gray-500 mono">${id}</div>
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4 text-right">
                        <span class="px-3 py-1 rounded-lg bg-matrix/10 text-matrix font-bold mono text-sm">${count}</span>
                    </td>
                </tr>
            `;
        }).join('');
    }
    
    const modalEl = document.getElementById('statsModal');
    if (modalEl) {
        modalEl.classList.remove('hidden');
        modalEl.style.display = 'flex';
        if (typeof lucide !== 'undefined') lucide.createIcons();
    }
}

// Expose functions to window for legacy compatibility
if (typeof window !== 'undefined') {
    window.initCharts = initCharts;
    window.updateStats = updateStats;
    window.updateChatStats = updateChatStats;
    window.showAllStats = showAllStats;
}
