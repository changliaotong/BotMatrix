import { fetchWithAuth } from './api.js';
import { t } from './i18n.js';

export let routingRules = [];

export async function fetchRoutingRules() {
    const container = document.getElementById('routing-list-container');
    if (!container) return;
    
    container.innerHTML = `
        <div class="col-12 text-center text-muted">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            <p class="mt-2" data-i18n="loading_routing_rules">正在加载路由规则...</p>
        </div>
    `;
    
    try {
        const response = await fetchWithAuth('/api/admin/routing');
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const data = await response.json();
        // Convert object format to array format for easier rendering
        routingRules = [];
        if (data.rules && typeof data.rules === 'object') {
            for (const [key, workerID] of Object.entries(data.rules)) {
                routingRules.push({
                    key: key,
                    worker_id: workerID
                });
            }
        }
        renderRoutingRules();
        
    } catch (error) {
        console.error('Failed to fetch routing rules:', error);
        container.innerHTML = `
            <div class="col-12 text-center text-danger">
                <i class="bi bi-exclamation-triangle fs-4"></i>
                <p class="mt-2">${t('loading_failed') || '加载失败'}: ${error.message}</p>
                <button class="btn btn-sm btn-outline-primary" onclick="fetchRoutingRules()">
                    <i class="bi bi-arrow-clockwise"></i> ${t('retry') || '重试'}
                </button>
            </div>
        `;
    }
}

export function toggleRoutingHelp() {
    const helpSection = document.getElementById('routing-help-section');
    if (helpSection) {
        if (helpSection.style.display === 'none') {
            helpSection.style.display = 'block';
        } else {
            helpSection.style.display = 'none';
        }
    }
}

export function renderRoutingRules() {
    const container = document.getElementById('routing-list-container');
    if (!container) return;
    
    if (!routingRules || routingRules.length === 0) {
        container.innerHTML = `
            <div class="col-12 text-center text-muted">
                <i class="bi bi-diagram-3 fs-1"></i>
                <p class="mt-3" data-i18n="no_routing_rules">暂无路由规则</p>
                <button class="btn btn-primary" onclick="showAddRoutingRuleDialog()">
                    <i class="bi bi-plus-lg"></i> <span data-i18n="routing_add_rule">添加路由规则</span>
                </button>
            </div>
        `;
        return;
    }
    
    container.innerHTML = routingRules.map(rule => `
        <div class="col-md-6 col-lg-4 mb-3">
            <div class="card h-100">
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-start mb-2">
                        <h6 class="card-title mb-0 text-truncate" title="${rule.key || ''}">${rule.key || '未命名规则'}</h6>
                        <div class="dropdown">
                            <button class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                                <i class="bi bi-three-dots-vertical"></i>
                            </button>
                            <ul class="dropdown-menu">
                                <li><a class="dropdown-item" href="#" onclick="editRoutingRule('${rule.key}')">
                                    <i class="bi bi-pencil"></i> ${t.edit || '编辑'}
                                </a></li>
                                <li><a class="dropdown-item text-danger" href="#" onclick="deleteRoutingRule('${rule.key}')">
                                    <i class="bi bi-trash"></i> ${t.delete || '删除'}
                                </a></li>
                            </ul>
                        </div>
                    </div>
                    <p class="card-text small text-muted mb-2">
                        <i class="bi bi-arrow-right"></i> ${rule.worker_id || '未设置目标'}
                    </p>
                    <div class="d-flex justify-content-between align-items-center">
                        <small class="text-muted">
                            <i class="bi bi-diagram-3"></i> ${t.target_worker || '目标工作节点'}: ${rule.worker_id || 'N/A'}
                        </small>
                    </div>
                </div>
            </div>
        </div>
    `).join('');
}

export function showAddRoutingRuleDialog() {
    // Reset form for new rule
    const patternInput = document.getElementById('routing-pattern');
    const targetInput = document.getElementById('routing-target');
    const enabledInput = document.getElementById('routing-enabled');
    
    if (patternInput) patternInput.value = '';
    if (targetInput) targetInput.value = '';
    if (enabledInput) enabledInput.checked = true;
    
    // 确保获取最新的节点列表
    if (window.fetchWorkers) {
        window.fetchWorkers();
    }
    
    // Populate worker dropdown
    const dropdown = document.getElementById('worker-select-dropdown');
    if (dropdown) {
        const updateDropdown = () => {
            const workers = Array.isArray(window.currentWorkers) ? window.currentWorkers : [];
            if (workers.length === 0) {
                dropdown.innerHTML = '<li><a class="dropdown-item disabled" href="#">无在线节点</a></li>';
            } else {
                dropdown.innerHTML = workers.map(w => 
                    `<li><a class="dropdown-item" href="javascript:void(0)" onclick="document.getElementById('routing-target').value='${w.id}'">${w.id} (${w.addr || w.remote_addr || 'N/A'})</a></li>`
                ).join('');
            }
        };
        updateDropdown();
        setTimeout(updateDropdown, 500); 
    }

    // Update modal title for adding new rule
    const modalTitle = document.querySelector('#routingRuleModal .modal-title');
    if (modalTitle) {
        modalTitle.textContent = '添加路由规则';
        modalTitle.setAttribute('data-i18n', 'routing_add_rule');
    }
    
    // Update save button for adding new rule
    const saveButton = document.querySelector('#routingRuleModal .modal-footer .btn-primary');
    if (saveButton) {
        saveButton.textContent = '保存';
        saveButton.setAttribute('data-i18n', 'save');
        saveButton.onclick = () => saveRoutingRule();
    }
    
    const modalElement = document.getElementById('routingRuleModal');
    if (modalElement) {
        const modal = new bootstrap.Modal(modalElement);
        modal.show();
    }
}

export async function saveRoutingRule() {
    const pattern = document.getElementById('routing-pattern')?.value.trim();
    const target = document.getElementById('routing-target')?.value.trim();
    
    if (!pattern || !target) {
        alert('请填写完整信息');
        return;
    }
    
    try {
        const response = await fetchWithAuth('/api/admin/routing', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                key: pattern,
                worker_id: target
            })
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || '保存失败');
        }
        
        const modalElement = document.getElementById('routingRuleModal');
        if (modalElement) {
            const modal = bootstrap.Modal.getInstance(modalElement);
            if (modal) modal.hide();
        }
        
        fetchRoutingRules(); // Refresh the list
        
    } catch (error) {
        console.error('Failed to save routing rule:', error);
        alert('保存失败: ' + error.message);
    }
}

export async function deleteRoutingRule(ruleId) {
    if (!confirm(t('confirm_delete') || '确定要删除这条路由规则吗？')) {
        return;
    }
    
    try {
        const response = await fetchWithAuth(`/api/admin/routing?key=${encodeURIComponent(ruleId)}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || `HTTP ${response.status}`);
        }
        
        fetchRoutingRules(); // Refresh the list
        
    } catch (error) {
        console.error('Failed to delete routing rule:', error);
        alert('删除失败: ' + error.message);
    }
}

export function editRoutingRule(ruleId) {
    const rule = routingRules.find(r => r.key === ruleId);
    if (!rule) {
        alert('规则不存在');
        return;
    }
    
    showAddRoutingRuleDialog();
    
    // Pre-fill the form after modal is shown
    setTimeout(() => {
        const patternInput = document.getElementById('routing-pattern');
        const targetInput = document.getElementById('routing-target');
        const enabledInput = document.getElementById('routing-enabled');
        
        if (patternInput) patternInput.value = rule.key || '';
        if (targetInput) targetInput.value = rule.worker_id || '';
        if (enabledInput) enabledInput.checked = true;
        
        // Change modal title
        const modalTitle = document.querySelector('#routingRuleModal .modal-title');
        if (modalTitle) {
            modalTitle.textContent = '编辑路由规则';
            modalTitle.setAttribute('data-i18n', 'routing_edit_rule');
        }
        
        // Change save button to call saveRoutingRule (backend POST handles both)
        const saveButton = document.querySelector('#routingRuleModal .modal-footer .btn-primary');
        if (saveButton) {
            saveButton.textContent = '保存';
            saveButton.setAttribute('data-i18n', 'save');
            saveButton.onclick = () => saveRoutingRule();
        }
    }, 200);
}

export async function updateRoutingRule(ruleId) {
    return saveRoutingRule();
}
// End of module
