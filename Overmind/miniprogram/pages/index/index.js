// 首页逻辑 - 与Overmind功能一致
const MiniProgramAPI = require('../../utils/miniprogram_api.js');

Page({
  data: {
    // 系统统计
    systemStats: {
      onlineBots: 0,
      totalGroups: 0,
      totalMessages: 0
    },
    
    // 系统状态
    systemStatus: {
      online: false,
      cpuPercent: 0,
      memoryPercent: 0,
      uptime: '0小时'
    },
    
    // 异常告警
    alerts: [],
    
    // 机器人列表
    botList: [],
    
    // 加载状态
    isLoaded: false,
    loadError: null,
    
    // 自动刷新定时器
    refreshTimer: null,
    
    // WebSocket连接状态
    wsConnected: false
  },

  onLoad() {
    console.log('首页加载');
    this.loadPageData();
    this.startAutoRefresh();
  },

  onShow() {
    // 页面显示时刷新数据
    if (this.data.isLoaded) {
      this.refreshStatus();
    }
  },

  onHide() {
    // 页面隐藏时停止自动刷新
    this.stopAutoRefresh();
  },

  onUnload() {
    // 页面卸载时清理资源
    this.stopAutoRefresh();
  },

  // 加载页面数据
  async loadPageData() {
    try {
      this.setData({ loadError: null });
      
      // 并行加载所有数据
      const [systemResult, botResult, alertResult] = await Promise.all([
        this.loadSystemStatus(),
        this.loadBotStatus(),
        this.loadAlerts()
      ]);

      if (systemResult.success && botResult.success) {
        this.updateSystemStats(botResult.data);
        this.setData({ 
          isLoaded: true,
          wsConnected: getApp().globalData.wsConnected
        });
      } else {
        throw new Error('数据加载失败');
      }
    } catch (error) {
      console.error('页面数据加载失败:', error);
      this.setData({
        loadError: error.message || '数据加载失败，请检查网络连接',
        isLoaded: true
      });
    }
  },

  // 加载系统状态
  async loadSystemStatus() {
    try {
      const result = await MiniProgramAPI.getSystemStatus();
      if (result.success) {
        const data = result.data;
        this.setData({
          systemStatus: {
            online: data.online || false,
            cpuPercent: data.cpu_percent || 0,
            memoryPercent: data.memory_percent || 0,
            uptime: this.formatUptime(data.uptime || 0)
          }
        });
      }
      return result;
    } catch (error) {
      console.error('系统状态加载失败:', error);
      return { success: false, error: error.message };
    }
  },

  // 加载机器人状态
  async loadBotStatus() {
    try {
      const result = await MiniProgramAPI.getBotStatus();
      if (result.success) {
        const bots = result.data.bots || [];
        this.setData({ botList: bots });
        getApp().globalData.botList = bots;
      }
      return result;
    } catch (error) {
      console.error('机器人状态加载失败:', error);
      return { success: false, error: error.message };
    }
  },

  // 加载异常告警
  async loadAlerts() {
    try {
      const result = await MiniProgramAPI.getAlerts();
      if (result.success) {
        const alerts = result.data.alerts || [];
        this.setData({ 
          alerts: alerts.map(alert => ({
            ...alert,
            time: this.formatTime(alert.timestamp)
          }))
        });
      }
      return result;
    } catch (error) {
      console.error('告警加载失败:', error);
      return { success: false, error: error.message };
    }
  },

  // 更新系统统计
  updateSystemStats(botData) {
    const bots = botData.bots || [];
    const onlineBots = bots.filter(bot => bot.enabled).length;
    const totalGroups = bots.reduce((sum, bot) => sum + (bot.groups || 0), 0);
    const totalMessages = bots.reduce((sum, bot) => sum + (bot.messages || 0), 0);

    this.setData({
      systemStats: {
        onlineBots,
        totalGroups,
        totalMessages
      }
    });
  },

  // 刷新状态
  async refreshStatus() {
    try {
      // 显示加载状态
      wx.showLoading({ title: '刷新中...' });
      
      // 并行刷新数据
      await Promise.all([
        this.loadSystemStatus(),
        this.loadBotStatus(),
        this.loadAlerts()
      ]);

      // 更新统计
      this.updateSystemStats({ bots: this.data.botList });
      
      wx.hideLoading();
      wx.showToast({
        title: '刷新成功',
        icon: 'success',
        duration: 1500
      });
    } catch (error) {
      wx.hideLoading();
      wx.showToast({
        title: '刷新失败',
        icon: 'error',
        duration: 2000
      });
    }
  },

  // 批量操作机器人
  async toggleAllBots(e) {
    const enable = e.currentTarget.dataset.enable;
    const actionText = enable ? '启动全部机器人' : '停止全部机器人';
    
    // 确认操作
    const confirm = await new Promise(resolve => {
      wx.showModal({
        title: '确认操作',
        content: `确定要${actionText}吗？`,
        success: (res) => resolve(res.confirm)
      });
    });

    if (!confirm) return;

    try {
      const result = await MiniProgramAPI.toggleAllBots(enable);
      if (result.success) {
        // 刷新机器人状态
        await this.loadBotStatus();
        this.updateSystemStats({ bots: this.data.botList });
        
        wx.showToast({
          title: enable ? '启动成功' : '停止成功',
          icon: 'success',
          duration: 2000
        });
      }
    } catch (error) {
      wx.showToast({
        title: '操作失败',
        icon: 'error',
        duration: 2000
      });
    }
  },

  // 导航到日志页面
  navigateToLogs() {
    wx.switchTab({
      url: '/pages/logs/logs'
    });
  },

  // 启动自动刷新
  startAutoRefresh() {
    // 每30秒自动刷新一次
    const timer = setInterval(() => {
      this.refreshStatus();
    }, 30000);
    
    this.setData({ refreshTimer: timer });
    
    // 订阅WebSocket事件
    this.subscribeWebSocketEvents();
  },

  // 停止自动刷新
  stopAutoRefresh() {
    if (this.data.refreshTimer) {
      clearInterval(this.data.refreshTimer);
      this.setData({ refreshTimer: null });
    }
    
    // 取消WebSocket事件订阅
    this.unsubscribeWebSocketEvents();
  },

  // WebSocket消息处理
  handleWebSocketMessage(data) {
    console.log('收到WebSocket消息:', data);
    
    switch (data.type) {
      case 'bot_status':
        // 机器人状态更新
        this.updateBotStatus(data.bots);
        break;
      case 'system_status':
        // 系统状态更新
        this.updateSystemStatus(data);
        break;
      case 'alert':
        // 异常告警
        this.addAlert(data.alert);
        break;
      case 'system_metrics':
        // 系统指标更新
        this.updateSystemMetrics(data.metrics);
        break;
      case 'bot_event':
        // 机器人事件
        this.handleBotEvent(data.event);
        break;
    }
  },

  // 更新机器人状态
  updateBotStatus(bots) {
    this.setData({ botList: bots });
    this.updateSystemStats({ bots });
    getApp().globalData.botList = bots;
  },

  // 更新系统状态
  updateSystemStatus(data) {
    this.setData({
      systemStatus: {
        online: data.online || this.data.systemStatus.online,
        cpuPercent: data.cpu_percent || this.data.systemStatus.cpuPercent,
        memoryPercent: data.memory_percent || this.data.systemStatus.memoryPercent,
        uptime: this.formatUptime(data.uptime || 0)
      }
    });
  },

  // 添加告警
  addAlert(alert) {
    const alerts = [{
      ...alert,
      time: this.formatTime(alert.timestamp || Date.now()),
      id: alert.id || Date.now()
    }, ...this.data.alerts].slice(0, 10); // 最多显示10条
    
    this.setData({ alerts });
    
    // 播放提示音（如果支持）
    if (alert.level === 'critical' || alert.level === 'error') {
      wx.vibrateShort();
    }
  },

  // 更新系统指标
  updateSystemMetrics(metrics) {
    this.setData({
      systemStatus: {
        ...this.data.systemStatus,
        cpuPercent: metrics.cpu_usage || this.data.systemStatus.cpuPercent,
        memoryPercent: metrics.memory_usage || this.data.systemStatus.memoryPercent
      }
    });
  },

  // 处理机器人事件
  handleBotEvent(event) {
    console.log('机器人事件:', event);
    
    switch (event.type) {
      case 'bot_started':
        this.showBotEventNotification('机器人已启动', event.bot_name);
        break;
      case 'bot_stopped':
        this.showBotEventNotification('机器人已停止', event.bot_name);
        break;
      case 'bot_error':
        this.showBotEventNotification('机器人错误', `${event.bot_name}: ${event.error}`);
        break;
      case 'bot_restart':
        this.showBotEventNotification('机器人重启', event.bot_name);
        break;
    }
    
    // 刷新机器人状态
    this.loadBotStatus();
  },

  // 显示机器人事件通知
  showBotEventNotification(title, content) {
    wx.showToast({
      title: title,
      icon: 'none',
      duration: 2000
    });
    
    // 可选：显示更详细的通知
    if (getApp().globalData.systemInfo.platform === 'android') {
      // Android平台可以显示更丰富的通知
      setTimeout(() => {
        wx.showModal({
          title: title,
          content: content,
          showCancel: false,
          confirmText: '知道了'
        });
      }, 2000);
    }
  },

  // 订阅WebSocket事件
  subscribeWebSocketEvents() {
    const app = getApp();
    
    // 订阅系统告警事件
    this.unsubscribeAlert = app.subscribeToEvent('systemAlert', (alert) => {
      this.addAlert(alert);
    });
    
    // 订阅状态更新事件
    this.unsubscribeStatusUpdate = app.subscribeToEvent('statusUpdate', (status) => {
      this.updateSystemStatus(status);
    });
    
    // 订阅机器人状态变化事件
    this.unsubscribeBotStatusChange = app.subscribeToEvent('botStatusChange', (change) => {
      this.handleBotEvent(change);
    });
    
    // 订阅系统指标事件
    this.unsubscribeSystemMetrics = app.subscribeToEvent('systemMetrics', (metrics) => {
      this.updateSystemMetrics(metrics);
    });
  },

  // 取消WebSocket事件订阅
  unsubscribeWebSocketEvents() {
    if (this.unsubscribeAlert) {
      this.unsubscribeAlert();
      this.unsubscribeAlert = null;
    }
    if (this.unsubscribeStatusUpdate) {
      this.unsubscribeStatusUpdate();
      this.unsubscribeStatusUpdate = null;
    }
    if (this.unsubscribeBotStatusChange) {
      this.unsubscribeBotStatusChange();
      this.unsubscribeBotStatusChange = null;
    }
    if (this.unsubscribeSystemMetrics) {
      this.unsubscribeSystemMetrics();
      this.unsubscribeSystemMetrics = null;
    }
  },

  // 工具函数：格式化运行时间
  formatUptime(seconds) {
    if (seconds < 60) return `${seconds}秒`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}小时`;
    return `${Math.floor(seconds / 86400)}天`;
  },

  // 工具函数：格式化时间
  formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
    
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  },

  // 下拉刷新
  async onPullDownRefresh() {
    try {
      await this.refreshStatus();
    } finally {
      wx.stopPullDownRefresh();
    }
  },

  // 分享功能
  onShareAppMessage() {
    return {
      title: 'BotMatrix - 机器人管理系统',
      path: '/pages/index/index',
      imageUrl: '/images/share.png'
    };
  }
});