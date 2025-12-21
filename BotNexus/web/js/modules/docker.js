import { fetchWithAuth } from './api.js';
import { currentLang, translations } from './i18n.js';

export function renderDockerContainers(containers) {
    const tbody = document.getElementById('docker-containers-tbody');
    if (!tbody) return;

    const t = translations[currentLang] || translations['zh-CN'];
    const searchTerm = (document.getElementById('docker-search')?.value || '').toLowerCase();

    const filtered = containers.filter(c => {
        const names = (c.Names || []).join(', ').toLowerCase();
        const image = (c.Image || '').toLowerCase();
        const id = (c.Id || '').toLowerCase();
        return names.includes(searchTerm) || image.includes(searchTerm) || id.includes(searchTerm);
    });

    if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">没有匹配的结果</td></tr>';
        return;
    }

    tbody.innerHTML = filtered.map(c => {
        const isRunning = c.State === 'running';
        const shortId = c.Id.substring(0, 12);
        return `
        <tr>
            <td><span class="font-monospace">${shortId}</span></td>
            <td>${(c.Names || []).join(', ')}</td>
            <td>${c.Image}</td>
            <td><span class="badge bg-${isRunning ? 'success' : 'secondary'}">${c.State}</span></td>
            <td>${c.Status}</td>
            <td>
                ${isRunning ? 
                    `<button class="btn btn-sm btn-outline-danger me-1" onclick="controlContainer('${c.Id}', 'stop')"><i class="bi bi-stop-fill"></i> ${t.docker_stop}</button>
                     <button class="btn btn-sm btn-outline-warning" onclick="controlContainer('${c.Id}', 'restart')"><i class="bi bi-arrow-repeat"></i> ${t.docker_restart}</button>` 
                    : 
                    `<button class="btn btn-sm btn-outline-success" onclick="controlContainer('${c.Id}', 'start')"><i class="bi bi-play-fill"></i> ${t.docker_start}</button>`
                }
            </td>
        </tr>
     `}).join('');
}

export function filterDockerContainers() {
    if (window.dockerContainers) {
        renderDockerContainers(window.dockerContainers);
    }
}

export function controlContainer(id, action) {
    const t = translations[currentLang] || translations['zh-CN'];
    const actionText = t[`docker_${action}`] || action;
    const confirmMsg = t.docker_confirm_action
        .replace('{id}', id.substring(0, 12))
        .replace('{action}', actionText);
        
    if (!confirm(confirmMsg)) return;
    
    fetchWithAuth('/api/docker/action', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            container_id: id,
            action: action
        })
    })
    .then(r => r.json())
    .then(res => {
        if (res.status === 'ok' || (res.id && !res.message)) { 
            // Use silent refresh to avoid full table flicker
            if (window.loadDockerContainers) {
                window.loadDockerContainers(true);
            }
        } else {
            alert(`${t.docker_action_failed}: ${res.message || JSON.stringify(res)}`);
        }
    })
    .catch(err => {
        alert('Error: ' + err);
    });
}

export function addBotContainer() {
    if (!confirm("确定要部署新的机器人容器吗？")) return;
    
    fetchWithAuth('/api/docker/add-bot', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(r => r.json())
    .then(res => {
        if (res.status === 'ok') {
            alert("机器人容器已在后台开始部署");
            if (window.loadDockerContainers) {
                window.loadDockerContainers(true);
            }
        } else {
            alert(res.message || "部署请求失败");
        }
    })
    .catch(err => alert("Error: " + err));
}

export function addWorkerContainer() {
    if (!confirm("确定要部署新的 Worker 节点容器吗？")) return;
    
    fetchWithAuth('/api/docker/add-worker', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(r => r.json())
    .then(res => {
        if (res.status === 'ok') {
            alert("Worker 节点容器已在后台开始部署");
            if (window.loadDockerContainers) {
                window.loadDockerContainers(true);
            }
        } else {
            alert(res.message || "部署请求失败");
        }
    })
    .catch(err => alert("Error: " + err));
}

export async function loadDockerContainers(silent = false) {
    const tbody = document.getElementById('docker-containers-tbody');
    if (!tbody) return;

    if (!silent) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center"><div class="spinner-border spinner-border-sm text-primary"></div></td></tr>';
    }

    try {
        const response = await fetchWithAuth('/api/docker/containers');
        const data = await response.json();
        window.dockerContainers = Array.isArray(data) ? data : (data.containers || []);
        renderDockerContainers(window.dockerContainers);
    } catch (err) {
        console.error('Failed to load docker containers:', err);
        if (!silent) {
            tbody.innerHTML = `<tr><td colspan="6" class="text-center text-danger">加载失败: ${err.message}</td></tr>`;
        }
    }
}
// End of module
