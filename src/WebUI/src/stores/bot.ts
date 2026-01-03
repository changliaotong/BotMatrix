import { defineStore } from 'pinia';
import api from '@/api';

export const useBotStore = defineStore('bot', {
  state: () => ({
    currentBotId: localStorage.getItem('wxbot_current_bot_id') || null,
    bots: [] as any[],
    messages: [] as any[],
    stats: {} as any,
    pendingRequests: new Map<string, { resolve: Function; reject: Function; timeout: number }>(),
  }),
  actions: {
    // --- Advanced Task & Strategy API ---
    async fetchStrategies() {
      try {
        const { data } = await api.get('/api/admin/strategies');
        return data;
      } catch (err) {
        console.error('Failed to fetch strategies:', err);
        return { success: false, data: [] };
      }
    },
    async saveStrategy(strategy: any) {
      try {
        const { data } = await api.post('/api/admin/strategies', strategy);
        return data;
      } catch (err) {
        console.error('Failed to save strategy:', err);
        throw err;
      }
    },
    async deleteStrategy(id: number) {
      try {
        const { data } = await api.delete(`/api/admin/strategies?id=${id}`);
        return data;
      } catch (err) {
        console.error('Failed to delete strategy:', err);
        throw err;
      }
    },

    async fetchIdentities() {
      try {
        const { data } = await api.get('/api/admin/identities');
        return data;
      } catch (err) {
        console.error('Failed to fetch identities:', err);
        return { success: false, data: [] };
      }
    },
    async saveIdentity(identity: any) {
      try {
        const { data } = await api.post('/api/admin/identities', identity);
        return data;
      } catch (err) {
        console.error('Failed to save identity:', err);
        throw err;
      }
    },
    async deleteIdentity(id: number) {
      try {
        const { data } = await api.delete(`/api/admin/identities?id=${id}`);
        return data;
      } catch (err) {
        console.error('Failed to delete identity:', err);
        throw err;
      }
    },

    async fetchShadowRules() {
      try {
        const { data } = await api.get('/api/admin/shadow-rules');
        return data;
      } catch (err) {
        console.error('Failed to fetch shadow rules:', err);
        return { success: false, data: [] };
      }
    },
    async saveShadowRule(rule: any) {
      try {
        const { data } = await api.post('/api/admin/shadow-rules', rule);
        return data;
      } catch (err) {
        console.error('Failed to save shadow rule:', err);
        throw err;
      }
    },
    async deleteShadowRule(id: number) {
      try {
        const { data } = await api.delete(`/api/admin/shadow-rules?id=${id}`);
        return data;
      } catch (err) {
        console.error('Failed to delete shadow rule:', err);
        throw err;
      }
    },

    async fetchTaskCapabilities() {
      try {
        const { data } = await api.get('/api/admin/tasks/capabilities');
        return data;
      } catch (err) {
        console.error('Failed to fetch task capabilities:', err);
        return { success: false, data: { actions: [], interceptors: [] } };
      }
    },

    reset() {
      this.bots = [];
      this.messages = [];
      this.stats = {};
      this.currentBotId = null;
      localStorage.removeItem('wxbot_current_bot_id');
    },

    // --- Plugin Management API ---
    async fetchPlugins() {
      try {
        const { data } = await api.get('/api/admin/plugins/list');
        return data;
      } catch (err) {
        console.error('Failed to fetch plugins:', err);
        return { success: false, message: 'Failed to fetch plugins' };
      }
    },
    async pluginAction(id: string, action: string, type: string, source: string) {
      try {
        const { data } = await api.post('/api/admin/plugins/action', {
          id,
          action,
          type,
          source
        });
        return data;
      } catch (err) {
        console.error('Plugin action failed:', err);
        return { success: false, message: 'Action failed' };
      }
    },
    async installPlugin(file: File, target: string = 'nexus') {
      try {
        const formData = new FormData();
        formData.append('plugin', file);
        formData.append('target', target);
        
        const { data } = await api.post('/api/admin/plugins/install', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        });
        return data;
      } catch (err) {
        console.error('Plugin installation failed:', err);
        return { success: false, message: 'Installation failed' };
      }
    },
    async deletePlugin(id: string, version: string, source: string) {
      try {
        const { data } = await api.post('/api/admin/plugins/delete', {
          id,
          version,
          source
        });
        return data;
      } catch (err) {
        console.error('Plugin deletion failed:', err);
        return { success: false, message: 'Deletion failed' };
      }
    },

    // --- Bot & Worker Management ---
    async fetchBots() {
      try {
        const { data } = await api.get('/api/bots');
        if (data.success && data.data) {
          this.bots = data.data.bots || [];
        }
      } catch (err) {
        console.error('Failed to fetch bots:', err);
        throw err;
      }
    },
    async addBot(config: any) {
      try {
        const { data } = await api.post('/api/admin/docker/add-bot', config);
        return data;
      } catch (err) {
        console.error('Failed to add bot:', err);
        throw err;
      }
    },
    async removeBot(botId: string) {
      // Note: This might need a different API endpoint if it's a Docker container
      try {
        const { data } = await api.post('/api/admin/docker/action', {
          container_id: botId, // This assumes botId is the container ID
          action: 'delete'
        });
        await this.fetchBots();
        return data;
      } catch (err) {
        console.error('Failed to remove bot:', err);
        throw err;
      }
    },
    async getLogs(botId: string) {
      try {
        const { data } = await api.get(`/api/admin/docker/logs?id=${botId}`);
        return data;
      } catch (err) {
        console.error('Failed to get logs:', err);
        throw err;
      }
    },
    async fetchSystemLogs(params: { bot_id?: string; search?: string; level?: string } = {}) {
      try {
        const query = new URLSearchParams();
        if (params.bot_id) query.append('botId', params.bot_id);
        if (params.search) query.append('search', params.search);
        if (params.level) query.append('level', params.level);
        
        const { data } = await api.get(`/api/logs?${query.toString()}`);
        return data;
      } catch (err) {
        console.error('Failed to fetch system logs:', err);
        throw err;
      }
    },
    async clearSystemLogs() {
      try {
        const { data } = await api.post('/api/admin/logs/clear');
        return data;
      } catch (err) {
        console.error('Failed to clear system logs:', err);
        throw err;
      }
    },
    async fetchStats() {
      try {
        const { data } = await api.get('/api/stats');
        if (data.success && data.data) {
          // Backend returns { stats: { ... } }, so we need to extract data.data.stats
          const rawStats = data.data.stats || data.data;
          
          // Map backend 'top_processes' to frontend 'top_processes' if needed
          // or ensure it's consistently named.
          this.stats = {
            ...rawStats,
            top_processes: rawStats.top_processes || rawStats.processes || []
          };
          console.log('Fetched stats with processes:', this.stats.top_processes?.length);
        } else {
          // Keep existing stats or set to empty object if null
          this.stats = this.stats || {};
        }
      } catch (err) {
        console.error('Failed to fetch stats:', err);
        this.stats = this.stats || {};
        throw err; // Propagate error for Dashboard to handle
      }
    },

    async fetchMessages(limit = 50) {
      try {
        const { data } = await api.get(`/api/admin/messages?limit=${limit}`);
        if (data.success && data.data) {
          const msgs = data.data.messages || [];
          // 合并并去重
          const existingIds = new Set(this.messages.map(m => m.id));
          const newMsgs = msgs.filter((m: any) => !existingIds.has(m.id));
          this.messages = [...this.messages, ...newMsgs].sort((a, b) => (a.time || 0) - (b.time || 0));
          return msgs;
        }
        return [];
      } catch (err) {
        console.error('Failed to fetch messages:', err);
        return [];
      }
    },
    async fetchWorkers() {
      try {
        const { data } = await api.get('/api/workers');
        return data;
      } catch (err) {
        console.error('Failed to fetch workers:', err);
        return { success: false, workers: [] };
      }
    },
    async fetchRoutingRules() {
      try {
        const { data } = await api.get('/api/admin/routing');
        return data;
      } catch (err) {
        console.error('Failed to fetch routing rules:', err);
        return { success: false, rules: [] };
      }
    },
    async setRoutingRule(rule: { key: string; worker_id: string }) {
      try {
        const { data } = await api.post('/api/admin/routing', rule);
        return data;
      } catch (err) {
        console.error('Failed to set routing rule:', err);
        throw err;
      }
    },
    async deleteRoutingRule(key: string) {
      try {
        const { data } = await api.delete(`/api/admin/routing?key=${key}`);
        return data;
      } catch (err) {
        console.error('Failed to delete routing rule:', err);
        throw err;
      }
    },
    async fetchDockerContainers() {
      try {
        const { data } = await api.get('/api/admin/docker/list');
        return data;
      } catch (err) {
        console.error('Failed to fetch docker containers:', err);
        return { success: false, containers: [] };
      }
    },
    async dockerAction(containerId: string, action: string) {
      try {
        const { data } = await api.post('/api/admin/docker/action', {
          container_id: containerId,
          action: action
        });
        return data;
      } catch (err) {
        console.error('Failed to perform docker action:', err);
        throw err;
      }
    },
    async fetchContacts(botId?: string) {
      try {
        const url = botId ? `/api/contacts?bot_id=${botId}` : '/api/contacts';
        const { data } = await api.get(url);
        return data;
      } catch (err) {
        console.error('Failed to fetch contacts:', err);
        return { success: false, contacts: [] };
      }
    },
    async fetchGroupMembers(botId: string, groupId: string, refresh = false) {
      try {
        const { data } = await api.get(`/api/admin/group/members?bot_id=${botId}&group_id=${groupId}&refresh=${refresh}`);
        return data;
      } catch (err) {
        console.error('Failed to fetch group members:', err);
        return { success: false, data: [] };
      }
    },
    async syncContacts(botId: string) {
      try {
        const { data } = await api.post('/api/admin/contacts/sync', { bot_id: botId });
        return data;
      } catch (err) {
        console.error('Failed to sync contacts:', err);
        throw err;
      }
    },
    async fetchNexusStatus() {
      try {
        const { data } = await api.get('/api/admin/nexus/status');
        return data;
      } catch (err) {
        console.error('Failed to fetch nexus status:', err);
        return { success: false, connections: [] };
      }
    },
    async fetchTasks() {
      try {
        const { data } = await api.get('/api/admin/tasks');
        return data;
      } catch (err) {
        console.error('Failed to fetch tasks:', err);
        return { success: false, tasks: [] };
      }
    },
    async createTask(task: any) {
      try {
        const { data } = await api.post('/api/admin/tasks', task);
        return data;
      } catch (err) {
        console.error('Failed to create task:', err);
        throw err;
      }
    },
    async toggleTask(taskId: string, status: string) {
      try {
        const { data } = await api.post(`/api/admin/tasks/toggle`, { id: taskId, status });
        return data;
      } catch (err) {
        console.error('Failed to toggle task:', err);
        throw err;
      }
    },
    async deleteTask(taskId: string) {
      try {
        const { data } = await api.delete(`/api/admin/tasks?id=${taskId}`);
        return data;
      } catch (err) {
        console.error('Failed to delete task:', err);
        throw err;
      }
    },
    async updateTask(taskId: string, task: any) {
      try {
        const { data } = await api.put(`/api/admin/tasks?id=${taskId}`, task);
        return data;
      } catch (err) {
        console.error('Failed to update task:', err);
        throw err;
      }
    },
    async fetchConfig() {
      try {
        const { data } = await api.get('/api/admin/config');
        return data;
      } catch (err) {
        console.error('Failed to fetch config:', err);
        return { success: false };
      }
    },
    async updateConfig(config: any) {
      try {
        const { data } = await api.post('/api/admin/config', config);
        return data;
      } catch (err) {
        console.error('Failed to update config:', err);
        throw err;
      }
    },
    async fetchSystemCapabilities() {
      try {
        const { data } = await api.get('/api/system/capabilities');
        return data;
      } catch (err) {
        console.error('Failed to fetch system capabilities:', err);
        return { success: false };
      }
    },
    async manageTags(action: 'add' | 'remove', targetType: string, targetId: string, tagName: string) {
      try {
        const { data } = await api.post('/api/tags', {
          action,
          target_type: targetType,
          target_id: targetId,
          tag_name: tagName
        });
        return data;
      } catch (err) {
        console.error('Failed to manage tags:', err);
        throw err;
      }
    },
    async fetchDetailedSystemStats() {
      try {
        const { data } = await api.get('/api/system/stats');
        if (data.success) {
          return data.stats;
        }
        return data;
      } catch (err) {
        console.error('Failed to fetch detailed system stats:', err);
        return null;
      }
    },
    async fetchChatStats() {
      try {
        const { data } = await api.get('/api/stats/chat');
        if (data.success) {
          return data.data;
        }
        return null;
      } catch (err) {
        console.error('Failed to fetch chat stats:', err);
        throw err;
      }
    },
    async fetchUsers() {
      try {
        const { data } = await api.get('/api/admin/users');
        return data;
      } catch (err) {
        console.error('Failed to fetch users:', err);
        return { success: false, users: [] };
      }
    },
    async manageUser(userData: any) {
      try {
        const { data } = await api.post('/api/admin/users', userData);
        return data;
      } catch (err) {
        console.error('Failed to manage user:', err);
        throw err;
      }
    },
    async deleteUser(userId: string) {
      try {
        const { data } = await api.delete(`/api/admin/users?id=${userId}`);
        return data;
      } catch (err) {
        console.error('Failed to delete user:', err);
        throw err;
      }
    },
    async fetchManual() {
      try {
        const { data } = await api.get('/api/admin/manual');
        return data;
      } catch (err) {
        console.error('Failed to fetch manual:', err);
        return { success: false, content: '' };
      }
    },
    async fetchFissionConfig() {
      try {
        const { data } = await api.get('/api/admin/fission/config');
        return data;
      } catch (err) {
        console.error('Failed to fetch fission config:', err);
        return { success: false };
      }
    },
    async updateFissionConfig(config: any) {
      try {
        const { data } = await api.post('/api/admin/fission/config', config);
        return data;
      } catch (err) {
        console.error('Failed to update fission config:', err);
        throw err;
      }
    },
    async fetchFissionTasks() {
      try {
        const { data } = await api.get('/api/admin/fission/tasks');
        return data;
      } catch (err) {
        console.error('Failed to fetch fission tasks:', err);
        return { success: false, tasks: [] };
      }
    },
    async saveFissionTask(task: any) {
      try {
        const { data } = await api.post('/api/admin/fission/tasks', task);
        return data;
      } catch (err) {
        console.error('Failed to save fission task:', err);
        throw err;
      }
    },
    async deleteFissionTask(taskId: string) {
      try {
        const { data } = await api.delete(`/api/admin/fission/tasks?id=${taskId}`);
        return data;
      } catch (err) {
        console.error('Failed to delete fission task:', err);
        throw err;
      }
    },
    async fetchFissionStats() {
      try {
        const { data } = await api.get('/api/admin/fission/stats');
        return data;
      } catch (err) {
        console.error('Failed to fetch fission stats:', err);
        return { success: false, stats: null };
      }
    },
    async fetchFissionLeaderboard() {
      try {
        const { data } = await api.get('/api/admin/fission/leaderboard');
        return data;
      } catch (err) {
        console.error('Failed to fetch fission leaderboard:', err);
        return { success: false, leaderboard: [] };
      }
    },
    async fetchFissionInvitations() {
      try {
        const { data } = await api.get('/api/admin/fission/invitations');
        return data;
      } catch (err) {
        console.error('Failed to fetch fission invitations:', err);
        return { success: false, invitations: [] };
      }
    },
    async fetchTaskExecutions() {
      try {
        const { data } = await api.get('/api/admin/tasks/executions');
        return data;
      } catch (err) {
        console.error('Failed to fetch task executions:', err);
        return { success: false, executions: [] };
      }
    },
    async fetchRedisConfig() {
      try {
        const { data } = await api.get('/api/admin/redis/config');
        return data;
      } catch (err) {
        console.error('Failed to fetch redis config:', err);
        return { success: false };
      }
    },
    async updateRedisConfig(config: any) {
      try {
        const { data } = await api.post('/api/admin/redis/config', config);
        return data;
      } catch (err) {
        console.error('Failed to update redis config:', err);
        throw err;
      }
    },
    async aiParse(input: string, actionType?: string) {
      try {
        const { data } = await api.post('/api/ai/parse', { 
          input, 
          action_type: actionType 
        });
        return data;
      } catch (err) {
        console.error('Failed to AI parse:', err);
        throw err;
      }
    },
    async aiConfirm(draftId: string) {
      try {
        const { data } = await api.post('/api/ai/confirm', { draft_id: draftId });
        return data;
      } catch (err) {
        console.error('Failed to AI confirm:', err);
        throw err;
      }
    },
    async smartAction(text: string) {
      try {
        const { data } = await api.post('/api/smart_action', { text });
        return data;
      } catch (err) {
        console.error('Failed to perform smart action:', err);
        throw err;
      }
    },

    setCurrentBotId(id: string) {
      this.currentBotId = id;
      localStorage.setItem('wxbot_current_bot_id', id);
    },
    async callBotApi(action: string, params: any = {}, botId: string | null = null) {
      const targetBotId = botId || this.currentBotId;
      if (!targetBotId) throw new Error('No bot selected');

      const { data } = await api.post('/api/action', {
        bot_id: targetBotId,
        action,
        params,
      });

      if (data.error) throw new Error(data.error);

      if (data.echo) {
        return new Promise((resolve, reject) => {
          const timeout = window.setTimeout(() => {
            this.pendingRequests.delete(data.echo);
            reject(new Error("Request timed out (30s)"));
          }, 30000);

          this.pendingRequests.set(data.echo, { resolve, reject, timeout });
        });
      }

      return data;
    },
    handleWebSocketMessage(message: any) {
      if (message.echo && this.pendingRequests.has(message.echo)) {
        const req = this.pendingRequests.get(message.echo)!;
        clearTimeout(req.timeout);
        this.pendingRequests.delete(message.echo);
        if (message.status === 'failed' || message.error) {
          req.reject(new Error(message.message || message.error));
        } else {
          req.resolve(message);
        }
      }

      // 处理广播事件
      if (message.post_type === 'message') {
        // 如果是消息事件，添加到消息列表
        const newMessage = {
          id: message.message_id,
          bot_id: message.self_id,
          self_id: message.self_id,
          user_id: message.user_id,
          target_id: message.target_id,
          group_id: message.group_id,
          message_type: message.message_type,
          content: message.message || message.raw_message,
          raw_message: message.raw_message,
          sender: message.sender,
          time: message.time,
          created_at: new Date(message.time * 1000).toISOString()
        };
        
        // 检查是否已经存在（避免重复）
        const exists = this.messages.some(m => m.id === newMessage.id);
        if (!exists) {
          this.messages.push(newMessage);
          // 保持消息数量在合理范围内
          if (this.messages.length > 500) {
            this.messages = this.messages.slice(-500);
          }
        }
      } else if (message.type === 'skill_result') {
        // 处理技能执行结果广播
        console.log('Received skill result:', message.data);
        const skillResult = message.data;
        if (skillResult && skillResult.result) {
          // 如果技能有结果，将其作为一条系统消息或机器人消息显示
          const resultMessage = {
            id: `skill_${skillResult.execution_id || Date.now()}`,
            bot_id: skillResult.bot_id,
            self_id: skillResult.bot_id,
            user_id: skillResult.bot_id, // 机器人发送的结果
            target_id: skillResult.user_id,
            group_id: skillResult.group_id,
            message_type: skillResult.group_id ? 'group' : 'private',
            content: `[Skill Result] ${skillResult.result}`,
            raw_message: skillResult.result,
            time: Math.floor(Date.now() / 1000),
            created_at: new Date().toISOString(),
            is_skill_result: true
          };
          this.messages.push(resultMessage);
        }
      } else if (message.type === 'worker_update') {
        // 处理 Worker 更新广播
        // 可以在这里更新 workers 列表状态
      }
    },
  },
});
