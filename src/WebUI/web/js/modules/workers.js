import { fetchWithAuth } from './api.js';
import { timeAgo } from './utils.js';
import { t } from './i18n.js';
import { authToken } from './auth.js';

export let currentWorkers = [];
export let workerFilterText = '';
export let workerSortBy = 'addr';
export let workerSortAsc = true;
export let workerViewMode = localStorage.getItem('worker_view_mode') || 'detail';

export async function fetchWorkers(showLoading = false) {
    if (!window.authToken) return;

    if (showLoading) {
        const container = document.getElementById('worker-list-container');
        if (container) {
            container.innerHTML = `
                <div class="col-12 text-center py-5">
                    <div class="spinner-border text-success" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                    <div class="mt-2 text-muted">${t('loading')}</div>
                </div>
            `;
        }
    }

    try {
        const response = await fetchWithAuth('/api/workers');
        const result = await response.json();
        
        // Ensure data is array - Support both direct array and wrapped object
        let workers = [];
        if (Array.isArray(result)) {
            workers = result;
        } else if (result && Array.isArray(result.workers)) {
            workers = result.workers;
        } else if (result && result.data && Array.isArray(result.data)) {
            workers = result.data;
        }
        
        currentWorkers = workers;
        
        // Update Badge
        const badge = document.getElementById('badge-workers');
        if (badge) badge.innerText = workers.length;

        // Update Metrics
        if (document.getElementById('metric-workers')) {
            document.getElementById('metric-workers').innerText = workers.length;
        }

        renderWorkers();
    window.renderWorkers = renderWorkers;
    } catch (err) {
        console.error('获取处理端列表失败:', err);
        currentWorkers = [];
        window.currentWorkers = [];
        renderWorkers();
    }
}

export function filterWorkers(text) {
    workerFilterText = text.toLowerCase();
    renderWorkers();
}

export function sortWorkers(field) {
    if (workerSortBy === field) {
        workerSortAsc = !workerSortAsc;
    } else {
        workerSortBy = field;
        workerSortAsc = true;
        if (field === 'time' || field === 'msg') workerSortAsc = false;
    }

    // Update UI
    ['addr', 'time', 'status', 'msg'].forEach(f => {
        const btn = document.getElementById(`btn-sort-worker-${f}`);
        if (!btn) return;
        
        if (f === workerSortBy) {
            btn.classList.add('active');
            let label = '';
            switch(f) {
                case 'addr': label = t('sort_address'); break;
                case 'time': label = t('sort_time'); break;
                case 'status': label = t('sort_status'); break;
                case 'msg': label = t('sort_processed'); break;
            }
            btn.innerHTML = label + (workerSortAsc ? ' ↑' : ' ↓');
        } else {
            btn.classList.remove('active');
            let label = '';
            switch(f) {
                case 'addr': label = t('sort_address'); break;
                case 'time': label = t('sort_time'); break;
                case 'status': label = t('sort_status'); break;
                case 'msg': label = t('sort_processed'); break;
            }
            btn.innerHTML = label;
        }
    });

    renderWorkers();
}

export function setWorkerViewMode(mode) {
    workerViewMode = mode;
    window._workerViewModeManuallySet = true;
    localStorage.setItem('worker_view_mode', mode);
    updateViewModeUI('worker', mode);
    renderWorkers();
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

/**
 * 更新 Worker 选择下拉菜单（用于路由配置等）
 */
export function updateWorkerDropdown() {
    const dropdown = document.getElementById('worker-select-dropdown');
    if (!dropdown) return;

    const workers = Array.isArray(currentWorkers) ? currentWorkers : [];
    if (workers.length === 0) {
        dropdown.innerHTML = '<li><a class="dropdown-item disabled" href="#">无在线节点</a></li>';
    } else {
        dropdown.innerHTML = workers.map(w => 
            `<li><a class="dropdown-item" href="javascript:void(0)" onclick="document.getElementById('routing-target').value='${w.id}'">${w.id} (${w.addr || w.remote_addr || 'N/A'})</a></li>`
        ).join('');
    }
}

export function renderWorkers() {
    const container = document.getElementById('worker-list-container');
    if (!container) return;

    // Ensure currentWorkers is an array
    if (!Array.isArray(currentWorkers)) {
        console.warn('currentWorkers is not an array, resetting to []');
        currentWorkers = [];
    }

    let workers = currentWorkers.filter(w => {
        if (!workerFilterText) return true;
        return (w.remote_addr || '').includes(workerFilterText) || 
               (w.status || '').toLowerCase().includes(workerFilterText);
    });

    // Auto switch to compact mode if many workers
    if (workers.length > 8 && workerViewMode === 'detail' && !window._workerViewModeManuallySet) {
        workerViewMode = 'compact';
        const btnDetail = document.getElementById('btn-worker-view-detail');
        const btnCompact = document.getElementById('btn-worker-view-compact');
        if (btnDetail) btnDetail.classList.remove('active');
        if (btnCompact) btnCompact.classList.add('active');
    }

    // Sort
    workers.sort((a, b) => {
        let res = 0;
        if (workerSortBy === 'addr') {
            res = (a.remote_addr || '').localeCompare(b.remote_addr || '');
        } else if (workerSortBy === 'time') {
            res = new Date(a.connected) - new Date(b.connected);
        } else if (workerSortBy === 'status') {
            res = (a.status || '').localeCompare(b.status || '');
        } else if (workerSortBy === 'msg') {
            res = (a.handled_count || 0) - (b.handled_count || 0);
        }
        return workerSortAsc ? res : -res;
    });

    if (workers.length === 0) {
        container.innerHTML = `<div class="col-12 text-center text-muted">${t('no_workers')}</div>`;
        return;
    }
    
    container.innerHTML = workers.map(w => {
        const connDate = new Date(w.connected);
        const timeStr = timeAgo(w.connected);
        const addr = w.remote_addr || w.id || 'Unknown';
        const status = w.status || 'Online';
        const addrDisplay = addr.includes(':') ? addr.split(':')[0] : addr;
        
        if (workerViewMode === 'compact') {
            return `
            <div class="col-sm-6 col-md-6 col-lg-4 col-xl-4 col-xxl-3 mb-3">
                <div class="card p-2 h-100 shadow-sm">
                    <div class="d-flex align-items-center mb-2 cursor-pointer" onmouseover="this.style.opacity=0.8" onmouseout="this.style.opacity=1">
                        <div class="bg-success text-white rounded-circle d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 32px; height: 32px;">
                            <i class="bi bi-gear-wide-connected fs-6"></i>
                        </div>
                        <div class="overflow-hidden w-100">
                            <div class="fw-bold text-truncate" style="font-size: 0.85rem;">${t('worker_node_title')}</div>
                            <div class="text-muted text-truncate" style="font-size: 0.7rem;">${addrDisplay}</div>
                        </div>
                    </div>
                    <div class="rounded p-1 mb-2 flex-grow-1" style="background-color: var(--bg-list-item); font-size: 0.75rem;">
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_status')}:</span>
                            <span class="text-success">${status}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_latency')}:</span>
                            <span class="text-info fw-bold">${w.avg_rtt || '0s'}</span>
                        </div>
                        <div class="d-flex justify-content-between">
                            <span class="text-muted">${t('card_label_processed')}:</span>
                            <span class="text-primary fw-bold">${w.handled_count || 0}</span>
                        </div>
                    </div>
                </div>
            </div>`;
        }

        return `
            <div class="col-sm-12 col-md-6 col-lg-6 col-xl-6 col-xxl-4 mb-3">
                <div class="card p-2 h-100 shadow-sm">
                    <div class="d-flex align-items-center mb-2 cursor-pointer" onmouseover="this.style.opacity=0.8" onmouseout="this.style.opacity=1">
                        <div class="bg-success text-white rounded-circle d-flex align-items-center justify-content-center me-2 flex-shrink-0" style="width: 36px; height: 36px;">
                            <i class="bi bi-gear-wide-connected fs-5"></i>
                        </div>
                        <div class="overflow-hidden w-100">
                            <div class="fw-bold text-truncate" style="font-size: 0.9rem;">${t('worker_node')}</div>
                            <div class="text-muted text-truncate" style="font-size: 0.7rem;">${t('worker_node_title')}</div>
                        </div>
                    </div>
                    <div class="rounded p-2 mb-2 flex-grow-1" style="background-color: var(--bg-list-item); font-size: 0.75rem;">
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_ip')}:</span>
                            <span class="text-truncate" style="max-width: 80px;" title="${addr}">${addrDisplay}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_connected')}:</span>
                            <span title="${connDate.toLocaleString()}">${timeStr}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_status')}:</span>
                            <span class="text-success">${status}</span>
                        </div>
                        <div class="d-flex justify-content-between mb-1">
                            <span class="text-muted">${t('card_label_latency')}:</span>
                            <span class="text-info fw-bold">${w.avg_rtt || '0s'}</span>
                        </div>
                        <div class="d-flex justify-content-between">
                            <span class="text-muted">${t('card_label_processed_msgs')}:</span>
                            <span class="text-primary fw-bold">${w.handled_count || 0}</span>
                        </div>
                    </div>
                </div>
            </div>
        `}).join('');
}
// End of module
