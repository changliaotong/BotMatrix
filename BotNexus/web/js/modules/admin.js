/**
 * Admin module for user management
 */

import { fetchWithAuth } from './api.js';
import { showToast } from './ui.js?v=1.1.88';
import { authRole } from './auth.js';

/**
 * 显示创建用户模态框
 */
export function showCreateUserModal() {
    const modalElement = document.getElementById('createUserModal');
    if (modalElement) {
        const modal = bootstrap.Modal.getOrCreateInstance(modalElement);
        modal.show();
    }
}

/**
 * 创建新用户
 */
export async function createUser() {
    const usernameEl = document.getElementById('new-username');
    const passwordEl = document.getElementById('new-password');
    const isAdminEl = document.getElementById('new-is-admin');
    
    if (!usernameEl || !passwordEl || !isAdminEl) return;
    
    const username = usernameEl.value.trim();
    const password = passwordEl.value.trim();
    const isAdmin = isAdminEl.checked;

    if (!username || !password) {
        alert('请填写用户名和密码');
        return;
    }

    try {
        const response = await fetchWithAuth('/api/admin/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password, is_admin: isAdmin })
        });
        const result = await response.json();
        if (result.success) {
            showToast('用户创建成功', 'success');
            const modalElement = document.getElementById('createUserModal');
            if (modalElement) {
                const modal = bootstrap.Modal.getInstance(modalElement);
                if (modal) modal.hide();
            }
            fetchUsers(); // 刷新列表
            // 清空表单
            usernameEl.value = '';
            passwordEl.value = '';
            isAdminEl.checked = false;
        } else {
            alert('创建失败: ' + result.message);
        }
    } catch (err) {
        console.error('Create user error:', err);
        alert('创建用户时出错');
    }
}

/**
 * 获取用户列表
 */
export async function fetchUsers() {
    if (authRole !== 'admin' && authRole !== 'super') return;
    
    try {
        const response = await fetchWithAuth('/api/admin/users');
        const result = await response.json();
        if (result.success) {
            renderUsers(result.users || []);
        }
    } catch (err) {
        console.error('Fetch users error:', err);
    }
}

/**
 * 渲染用户列表
 * @param {Array} users 用户数组
 */
export function renderUsers(users) {
    const tbody = document.getElementById('users-table-body');
    if (!tbody) return;
    
    tbody.innerHTML = '';
    
    users.forEach(user => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${user.id}</td>
            <td>${user.username}</td>
            <td><span class="badge ${user.is_admin ? 'bg-danger' : 'bg-info'}">${user.is_admin ? 'Admin' : 'User'}</span></td>
            <td>${new Date(user.created_at).toLocaleString()}</td>
            <td class="text-end">
                <button class="btn btn-sm btn-outline-danger" onclick="deleteUser('${user.username}')" ${user.username === 'admin' ? 'disabled' : ''}>
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

/**
 * 删除用户
 * @param {string} username 用户名
 */
export async function deleteUser(username) {
    if (!confirm(`确定要删除用户 ${username} 吗？`)) return;
    
    try {
        const response = await fetchWithAuth(`/api/admin/users?username=${encodeURIComponent(username)}`, {
            method: 'DELETE'
        });
        const result = await response.json();
        if (result.success) {
            showToast('用户删除成功', 'success');
            fetchUsers();
        } else {
            alert('删除失败: ' + result.message);
        }
    } catch (err) {
        console.error('Delete user error:', err);
    }
}



/**
 * 重置用户密码
 * @param {string} username 用户名
 */
export async function resetUserPassword(username) {
    if (!username) return;
    const newPassword = prompt(`请输入用户 ${username} 的新密码:`);
    if (!newPassword) return;
    
    try {
        const response = await fetchWithAuth('/api/admin/users/password', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password: newPassword })
        });
        const result = await response.json();
        if (result.success) {
            showToast('密码重置成功', 'success');
        } else {
            alert('重置失败: ' + result.message);
        }
    } catch (err) {
        console.error('Reset password error:', err);
        alert('重置密码时出错');
    }
}

