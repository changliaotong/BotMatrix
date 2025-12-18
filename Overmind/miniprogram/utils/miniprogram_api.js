// 统一API封装 - 与Overmind功能一致
const MiniProgramAdapter = require('./miniprogram_adapter.js');

const API_BASE_URL = 'http://localhost:3001';
const WS_URL = 'ws://localhost:3001';

class MiniProgramAPI {
  // 系统状态获取 - 高频使用
  static async getSystemStatus() {
    try {
      MiniProgramAdapter.showLoading('获取系统状态中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/stats`,
        method: 'GET',
        timeout: 3000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 机器人状态获取 - 高频使用
  static async getBotStatus() {
    try {
      MiniProgramAdapter.showLoading('获取机器人状态中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots`,
        method: 'GET',
        timeout: 3000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { bots: [] } };
    }
  }

  // 切换机器人状态 - 高频操作
  static async toggleBot(botId, enable) {
    try {
      MiniProgramAdapter.showLoading('操作中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/toggle`,
        method: 'POST',
        data: {
          bot_id: botId,
          enable: enable
        },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      MiniProgramAdapter.showToast(enable ? '机器人已启动' : '机器人已停止');
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 获取机器人列表 - 机器人管理页面使用
  static async getBotList() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots/list`,
        method: 'GET',
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: [] };
    }
  }

  // 获取单个机器人详情
  static async getBotDetail(botId) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}`,
        method: 'GET',
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 重启机器人
  static async restartBot(botId) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}/restart`,
        method: 'POST',
        timeout: 10000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 删除机器人
  static async deleteBot(botId) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}`,
        method: 'DELETE',
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 获取机器人配置
  static async getBotConfig(botId) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}/config`,
        method: 'GET',
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 更新机器人配置
  static async updateBotConfig(botId, config) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}/config`,
        method: 'POST',
        data: config,
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 批量启动机器人
  static async batchStartBots(botIds) {
    try {
      MiniProgramAdapter.showLoading('批量启动中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots/batch-start`,
        method: 'POST',
        data: { bot_ids: botIds },
        timeout: 8000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 批量停止机器人
  static async batchStopBots(botIds) {
    try {
      MiniProgramAdapter.showLoading('批量停止中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots/batch-stop`,
        method: 'POST',
        data: { bot_ids: botIds },
        timeout: 8000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 批量重启机器人
  static async batchRestartBots(botIds) {
    try {
      MiniProgramAdapter.showLoading('批量重启中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots/batch-restart`,
        method: 'POST',
        data: { bot_ids: botIds },
        timeout: 10000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      return { success: false, error: error.message };
    }
  }

  // 批量切换机器人状态
  static async toggleAllBots(enable) {
    try {
      MiniProgramAdapter.showLoading('批量操作中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bots/toggle-all`,
        method: 'POST',
        data: { enable },
        timeout: 8000
      });
      MiniProgramAdapter.hideLoading();
      MiniProgramAdapter.showToast(enable ? '所有机器人已启动' : '所有机器人已停止');
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      MiniProgramAdapter.showToast('批量操作失败');
      return { success: false, error: error.message };
    }
  }

  // 获取系统日志
  static async getSystemLogs(limit = 50) {
    try {
      MiniProgramAdapter.showLoading('获取日志中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/logs`,
        method: 'GET',
        data: { limit },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { logs: [] } };
    }
  }

  // 获取机器人详细日志
  static async getBotLogs(botId, limit = 30) {
    try {
      MiniProgramAdapter.showLoading('获取机器人日志中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/bot/${botId}/logs`,
        method: 'GET',
        data: { limit },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { logs: [] } };
    }
  }

  // 用户登录
  static async login(username, password) {
    try {
      MiniProgramAdapter.showLoading('登录中...');
      const response = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/login`,
        method: 'POST',
        data: { username, password },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      
      if (response.success && response.token) {
        // 保存token
        MiniProgramAdapter.setStorageSync('token', response.token);
        MiniProgramAdapter.showToast('登录成功');
        return { success: true, data: response };
      } else {
        const errorMsg = response.message || '用户名或密码错误';
        MiniProgramAdapter.showToast(errorMsg);
        return { success: false, error: errorMsg };
      }
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 验证token - 通过请求用户信息接口验证
  static async validateToken(token) {
    try {
      if (!token) return { valid: false };
      
      const response = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/user/info`,
        method: 'GET',
        timeout: 5000
      });
      
      if (response.success) {
        return { 
          valid: true, 
          userInfo: response.data 
        };
      } else {
        return { valid: false, error: response.message };
      }
    } catch (error) {
      console.error('Token验证失败:', error);
      return { valid: false, error: error.message };
    }
  }

  // 修改密码
  static async changePassword(oldPassword, newPassword) {
    try {
      MiniProgramAdapter.showLoading('修改密码中...');
      const response = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/user/password`,
        method: 'POST',
        data: { 
          old_password: oldPassword, 
          new_password: newPassword 
        },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      
      if (response.success) {
        MiniProgramAdapter.showToast('密码修改成功');
        return { success: true };
      } else {
        MiniProgramAdapter.showToast(response.message || '修改失败');
        return { success: false, error: response.message };
      }
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      return { success: false, error: error.message };
    }
  }

  // WebSocket连接
  static connectWebSocket(url, callbacks) {
    return MiniProgramAdapter.connectWebSocket(url, callbacks);
  }

  // 获取系统信息
  static getSystemInfo() {
    return MiniProgramAdapter.getSystemInfo();
  }

  // 本地存储
  static setStorageSync(key, data) {
    return MiniProgramAdapter.setStorageSync(key, data);
  }

  static getStorageSync(key) {
    return MiniProgramAdapter.getStorageSync(key);
  }

  static removeStorageSync(key) {
    return MiniProgramAdapter.removeStorageSync(key);
  }

  // 获取异常告警
  static async getAlerts() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/alerts`,
        method: 'GET',
        timeout: 3000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { alerts: [] } };
    }
  }

  // 标记告警为已读
  static async markAlertRead(alertId) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/alerts/${alertId}/read`,
        method: 'POST',
        timeout: 3000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  // 获取服务器配置
  static async getServerConfig() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/config`,
        method: 'GET',
        timeout: 3000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: {} };
    }
  }

  // 更新服务器配置
  static async updateServerConfig(config) {
    try {
      MiniProgramAdapter.showLoading('保存配置中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/config`,
        method: 'POST',
        data: config,
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }

  // 获取系统监控数据
  static async getSystemMonitoring() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/monitoring`,
        method: 'GET',
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message, data: {} };
    }
  }

  // 获取性能监控数据
  static async getPerformanceData(timeRange = '1h') {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/performance`,
        method: 'GET',
        data: { time_range: timeRange },
        timeout: 8000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { cpu: [], memory: [], network: [] } };
    }
  }

  // 获取网络状态
  static async getNetworkStatus() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/network`,
        method: 'GET',
        timeout: 3000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { status: 'unknown', latency: 0 } };
    }
  }

  // 获取日志列表 - 支持分页和筛选
  static async getLogs(params = {}) {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/logs`,
        method: 'GET',
        data: params,
        timeout: 5000
      });
      return { success: true, data };
    } catch (error) {
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message, data: { logs: [], hasMore: false } };
    }
  }

  // 导出日志
  static async exportLogs({ logs, format, filters }) {
    try {
      MiniProgramAdapter.showLoading('导出中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/logs/export`,
        method: 'POST',
        data: { logs, format, filters },
        timeout: 10000
      });
      MiniProgramAdapter.hideLoading();
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      return { success: false, error: error.message };
    }
  }

  // 清空日志
  static async clearLogs() {
    try {
      MiniProgramAdapter.showLoading('清空中...');
      const data = await MiniProgramAdapter.request({
        url: `${API_BASE_URL}/api/logs/clear`,
        method: 'POST',
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      MiniProgramAdapter.showToast('日志已清空');
      return { success: true, data };
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      const errorMsg = MiniProgramAdapter.getErrorMsg(error);
      MiniProgramAdapter.showToast(errorMsg);
      return { success: false, error: error.message };
    }
  }
}

module.exports = MiniProgramAPI;