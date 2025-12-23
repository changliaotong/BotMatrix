import { fetchWithAuth, callBotApi } from './api.js';
import { t } from './i18n.js';
import { serverStartTime, memChart, cpuChart, msgChart } from './stats.js';
import { addEventLog, clearEvents } from './logs.js';
import { showTab } from './ui.js';

export let lastSystemStats = null;
export let detailChartInstance = null;
export let currentDetailType = 'cpu';
let currentDetailTimeRange = '1h';

/**
 * 获取系统统计信息
 */
export async function fetchSystemStats(updateDetails = false) {
    try {
        const res = await fetchWithAuth('/api/system/stats');
        if (!res.ok) throw new Error(res.statusText);
        const data = await res.json();
        
        window.lastSystemStats = data;

        const procBody = document.getElementById('process-list');
        if (procBody && data.processes) {
            procBody.innerHTML = data.processes.length === 0 ? 
                `<tr><td colspan="4" class="text-center">${t('no_data')}</td></tr>` : 
                data.processes.map(p => `<tr><td>${p.pid}</td><td class="text-truncate" style="max-width: 150px;" title="${p.name}">${p.name || 'unknown'}</td><td>${(p.cpu || 0).toFixed(1)}%</td><td>${((p.memory || 0) / 1024 / 1024).toFixed(1)}</td></tr>`).join('');
        }
        
        // 如果在详情页或显式要求更新详情
        const isDetailTabVisible = document.getElementById('tab-view-system-details') && document.getElementById('tab-view-system-details').style.display !== 'none';
        if (updateDetails || isDetailTabVisible) {
            updateDetailChart();
            renderHardwareGrid();
        }
    } catch (e) {
        console.error("Fetch system stats error:", e);
    }
}

/**
 * 更新系统统计
 */
export function updateSystemStats() {
    fetchSystemStats();
}

/**
 * 显示系统详情 (切换到标签页)
 */
export function showSystemDetails(type) {
    if (window.showTab) window.showTab('view-system-details');
    currentDetailType = type || 'cpu';
    
    // 更新标签激活状态
    document.querySelectorAll('#system-detail-tabs .nav-link').forEach(el => el.classList.remove('active'));
    const activeTab = document.querySelector(`#system-detail-tabs .nav-link[onclick*="'${currentDetailType}'"]`);
    if (activeTab) activeTab.classList.add('active');
    
    updateDetailChart();
    fetchSystemStats(true);
}

/**
 * 切换详情标签
 */
export function switchDetailTab(type, event) {
    if (event) event.preventDefault();
    currentDetailType = type;
    document.querySelectorAll('#system-detail-tabs .nav-link').forEach(el => el.classList.remove('active'));
    if (event && event.target) event.target.classList.add('active');
    updateDetailChart();
}

/**
 * 更新详情时间范围
 */
export function updateDetailTimeRange(range) {
    currentDetailTimeRange = range;
    document.querySelectorAll('#tab-view-system-details .btn-group .btn').forEach(btn => btn.classList.remove('active'));
    const activeBtn = document.querySelector(`#tab-view-system-details .btn-group .btn[onclick*="'${range}'"]`);
    if (activeBtn) activeBtn.classList.add('active');
    updateDetailChart();
}

/**
 * Update time and uptime display
 */
export function updateTimeDisplay() {
    // Current Time & Date
    if (document.getElementById('metric-current-time')) {
        const now = new Date();
        const lang = currentLang || 'zh-CN';
        const dateStr = now.toLocaleDateString(lang, { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-');
        const timeStr = now.toLocaleTimeString(lang, { hour12: false });
        
        document.getElementById('metric-current-time').innerText = `${dateStr} ${timeStr}`;
    }
    // Up Time
    if (document.getElementById('metric-uptime') && serverStartTime) {
        const uptimeSeconds = Math.floor(Date.now() / 1000 - serverStartTime);
        const d = Math.floor(uptimeSeconds / 86400);
        const h = Math.floor((uptimeSeconds % 86400) / 3600);
        const m = Math.floor((uptimeSeconds % 3600) / 60);
        const s = uptimeSeconds % 60;
        
        const dStr = d > 0 ? `${d}${t('time_days') || 'd'} ` : '';
        const hStr = h.toString().padStart(2, '0');
        const mStr = m.toString().padStart(2, '0');
        const sStr = s.toString().padStart(2, '0');
        
        document.getElementById('metric-uptime').innerHTML = 
            `${dStr}${hStr}<span class="time-sep">:</span>${mStr}<span class="time-sep">:</span>${sStr}`;
    }
}

export function showSystemDetail(type) {
    currentDetailType = type;
    const modal = new bootstrap.Modal(document.getElementById('systemDetailModal'));
    modal.show();
    
    // Reset canvas to avoid chart.js reuse issues
    const container = document.getElementById('detailChartContainer');
    container.innerHTML = '<canvas id="detailChart"></canvas>';
    
    const titleMap = {
        'cpu': 'CPU 使用率详情',
        'mem': '内存使用详情',
        'msg': '消息趋势详情',
        'disk': '磁盘使用详情',
        'net': '网络流量详情'
    };
    document.getElementById('systemDetailModalLabel').innerText = titleMap[type] || '系统详情';
    
    updateDetailChart();
}

export function updateDetailChart() {
    const canvas = document.getElementById('detailChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    
    if (detailChartInstance) {
        detailChartInstance.destroy();
    }

    let cpuDataBuffer = [], memDataBuffer = [], msgDataBuffer = [], sentDataBuffer = [], recvDataBuffer = [];
    let netSentTrend = [], netRecvTrend = [];
    let sourceIsRaw = false;

    // Use global stats buffers if available
    if (window.latestStats) {
        cpuDataBuffer = window.latestStats.cpu_trend || [];
        memDataBuffer = window.latestStats.mem_trend || [];
        msgDataBuffer = window.latestStats.msg_trend || [];
        sentDataBuffer = window.latestStats.sent_trend || [];
        recvDataBuffer = window.latestStats.recv_trend || [];
        netSentTrend = window.latestStats.net_sent_trend || [];
        netRecvTrend = window.latestStats.net_recv_trend || [];
        sourceIsRaw = true;
    } else {
        // Fallback to existing charts if available
        if (cpuChart) cpuDataBuffer = cpuChart.data.datasets[0].data;
        if (memChart) memDataBuffer = memChart.data.datasets[0].data;
        if (msgChart) {
            recvDataBuffer = msgChart.data.datasets[0].data;
            sentDataBuffer = msgChart.data.datasets[1].data;
            msgDataBuffer = msgChart.data.datasets[2].data;
        }
    }

    let labels = [];
    let datasets = [];
    let type = 'line';
    let options = {
        responsive: true,
        maintainAspectRatio: false,
        animation: false,
        plugins: {
            legend: { display: true },
            tooltip: { mode: 'index', intersect: false }
        },
        scales: {
            x: { display: true, grid: { display: false } },
            y: { display: true, beginAtZero: true }
        }
    };

    if (currentDetailType === 'cpu') {
        labels = Array.from({length: cpuDataBuffer.length}, (_, i) => i);
        datasets = [{
            label: t('cpu_usage_label') || 'CPU Usage (%)',
            data: cpuDataBuffer,
            borderColor: '#0d6efd',
            backgroundColor: 'rgba(13, 110, 253, 0.1)',
            fill: true,
            tension: 0.4
        }];
        options.scales.y.max = 100;
    } else if (currentDetailType === 'mem') {
        labels = Array.from({length: memDataBuffer.length}, (_, i) => i);
        const dataGB = memDataBuffer.map(v => sourceIsRaw ? (v / 1024 / 1024 / 1024) : (v / 1024));
        datasets = [{
            label: t('mem_usage_label') || 'Memory Usage (GB)',
            data: dataGB,
            borderColor: '#dc3545',
            backgroundColor: 'rgba(220, 53, 69, 0.1)',
            fill: true,
            tension: 0.4
        }];
    } else if (currentDetailType === 'msg') {
        labels = Array.from({length: msgDataBuffer.length}, (_, i) => i);
        datasets = [
            {
                label: t('total_messages_label') || 'Total Messages',
                data: msgDataBuffer,
                borderColor: '#198754',
                backgroundColor: 'rgba(25, 135, 84, 0.1)',
                fill: true
            },
            {
                label: t('sent_messages_label') || 'Sent Messages',
                data: sentDataBuffer,
                borderColor: '#0dcaf0',
                borderDash: [5, 5],
                fill: false
            },
            {
                label: t('recv_messages_label') || 'Recv Messages',
                data: recvDataBuffer,
                borderColor: '#ffc107',
                borderDash: [2, 2],
                fill: false
            }
        ];
    } else if (currentDetailType === 'disk') {
        type = 'bar';
        if (window.lastSystemStats && window.lastSystemStats.disk_usage) {
            labels = window.lastSystemStats.disk_usage.map(d => d.path);
            const dataUsed = window.lastSystemStats.disk_usage.map(d => (d.used / 1024 / 1024 / 1024).toFixed(2));
            const dataFree = window.lastSystemStats.disk_usage.map(d => (d.free / 1024 / 1024 / 1024).toFixed(2));
            
            datasets = [
                {
                    label: t('used_label') || 'Used (GB)',
                    data: dataUsed,
                    backgroundColor: '#dc3545'
                },
                {
                    label: t('free_label') || 'Free (GB)',
                    data: dataFree,
                    backgroundColor: '#198754'
                }
            ];
            options.scales.x = { stacked: true };
            options.scales.y = { stacked: true, beginAtZero: true };
        } else {
            labels = [t('loading_dots') || 'Loading...'];
            datasets = [{label: t('no_data') || 'No Data', data: []}];
        }
    } else if (currentDetailType === 'net') {
         if (netSentTrend.length > 1) {
             labels = Array.from({length: netSentTrend.length - 1}, (_, i) => i);
             const sentThroughput = [];
             const recvThroughput = [];
             
             for (let i = 1; i < netSentTrend.length; i++) {
                 const s = (netSentTrend[i] - netSentTrend[i-1]) / 1024 / 5;
                 const r = (netRecvTrend[i] - netRecvTrend[i-1]) / 1024 / 5;
                 sentThroughput.push(s.toFixed(2));
                 recvThroughput.push(r.toFixed(2));
             }

             datasets = [
                 {
                     label: t('sent_kb_s') || 'Sent (KB/s)',
                     data: sentThroughput,
                     borderColor: '#0dcaf0',
                     backgroundColor: 'rgba(13, 202, 240, 0.1)',
                     fill: true,
                     tension: 0.4
                 },
                 {
                     label: t('recv_kb_s') || 'Recv (KB/s)',
                     data: recvThroughput,
                     borderColor: '#ffc107',
                     backgroundColor: 'rgba(255, 193, 7, 0.1)',
                     fill: true,
                     tension: 0.4
                 }
             ];
         } else if (window.lastSystemStats && window.lastSystemStats.net_io && window.lastSystemStats.net_io.length > 0) {
             type = 'bar';
             const io = window.lastSystemStats.net_io[0];
             labels = [t('total_sent_label') || 'Total Sent', t('total_recv_label') || 'Total Recv'];
             datasets = [{
                 label: t('bytes_mb_label') || 'Bytes (MB)',
                 data: [
                     (io.bytesSent / 1024 / 1024).toFixed(2), 
                     (io.bytesRecv / 1024 / 1024).toFixed(2)
                 ],
                 backgroundColor: ['#0dcaf0', '#ffc107']
             }];
         } else {
            labels = [t('loading_dots') || 'Loading...'];
            datasets = [{label: t('no_data') || 'No Data', data: []}];
         }
    }

    detailChartInstance = new Chart(ctx, {
        type: type,
        data: { labels, datasets },
        options: options
    });
}

/**
 * 调用系统操作
 * @param {string} action 操作名称
 */
export async function callSystemAction(action) {
    if (action === 'clean_logs') {
        const dashLog = document.getElementById('recent-logs');
        if (dashLog) dashLog.innerHTML = '';
        
        clearEvents();
        
        addEventLog({type: 'system', message: t('log_cleaned')});
        return;
    }
    
    try {
        const res = await callBotApi(action);
        alert(t('alert_op_success') + JSON.stringify(res.data));
    } catch (e) {
        alert(t('alert_op_failed') + e.message);
    }
}

/**
 * 渲染硬件网格
 */
export function renderHardwareGrid() {
    const container = document.getElementById('hardware-info-grid');
    if (!container || !window.lastSystemStats) return;
    
    const s = window.lastSystemStats;
    let html = '';
    
    if (s.host_info) {
         html += `
            <div class="col-12 mb-3">
                <div class="card h-100">
                    <div class="card-header fw-bold">${t('host_info_title') || '主机信息'}</div>
                    <div class="card-body small">
                        <div class="row">
                            <div class="col-md-6">
                                <div><span class="text-muted">${t('hostname_label') || 'Hostname'}:</span> ${s.host_info.hostname}</div>
                                <div><span class="text-muted">${t('os_label') || 'OS'}:</span> ${s.host_info.platform} ${s.host_info.platformVersion}</div>
                                <div><span class="text-muted">${t('kernel_label') || 'Kernel'}:</span> ${s.host_info.kernelVersion}</div>
                            </div>
                            <div class="col-md-6">
                                <div><span class="text-muted">${t('arch_label') || 'Arch'}:</span> ${s.host_info.kernelArch}</div>
                                <div><span class="text-muted">${t('uptime_label') || 'Uptime'}:</span> ${(s.host_info.uptime / 3600).toFixed(1)} Hours</div>
                                <div><span class="text-muted">${t('boot_time_label') || 'Boot Time'}:</span> ${new Date(s.host_info.bootTime * 1000).toLocaleString()}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>`;
    }
    
    if (s.disk_usage) {
        html += `
            <div class="col-md-6 mb-3">
                <div class="card h-100">
                    <div class="card-header fw-bold">${t('disk_usage_title') || '磁盘使用'}</div>
                    <div class="card-body small" style="overflow-y: auto; max-height: 200px;">
                        ${s.disk_usage.map(d => `
                            <div class="mb-2">
                                <div class="d-flex justify-content-between">
                                    <span>${d.path}</span>
                                    <span>${d.usedPercent.toFixed(1)}%</span>
                                </div>
                                <div class="progress" style="height: 6px;">
                                    <div class="progress-bar bg-${d.usedPercent > 90 ? 'danger' : (d.usedPercent > 70 ? 'warning' : 'success')}" 
                                         role="progressbar" style="width: ${d.usedPercent}%"></div>
                                </div>
                                <div class="text-muted" style="font-size: 0.8em;">
                                    ${(d.used/1024/1024/1024).toFixed(1)} GB / ${(d.total/1024/1024/1024).toFixed(1)} GB
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>`;
    }
    
    if (s.net_interfaces) {
         html += `
            <div class="col-md-6 mb-3">
                <div class="card h-100">
                    <div class="card-header fw-bold">${t('net_interfaces_title') || '网络接口'}</div>
                    <div class="card-body small" style="overflow-y: auto; max-height: 200px;">
                        ${s.net_interfaces.map(i => `
                            <div class="mb-2 border-bottom pb-1">
                                <div class="fw-bold text-primary">${i.name}</div>
                                ${i.addrs.map(a => `<div>${a.addr}</div>`).join('')}
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>`;
    }
    
    container.innerHTML = html;
}

