const MiniProgramAPI = require('../../utils/miniprogram_api.js');

Page({
  data: {
    botId: null,
    bot: null,
    logs: [],
    isLoading: true,
    isLoaded: false,
    loadError: null,
    refreshTimer: null,
    activeTab: 'info', // info, logs, config
    config: {
      autoReply: false,
      groupManage: false,
      welcomeMessage: '',
      commands: []
    },
    editConfig: null // 编辑中的配置
  },

  onLoad(options) {
    const botId = options.id;
    if (!botId) {
      wx.showToast({
        title: '参数错误',
        icon: 'error'
      });
      wx.navigateBack();
      return;
    }
    
    this.setData({ botId });
    this.loadBotDetail();
    this.startAutoRefresh();
  },

  onUnload() {
    this.stopAutoRefresh();
  },

  // 加载机器人详情
  async loadBotDetail() {
    try {
      this.setData({ isLoading: true, loadError: null });
      
      // 并行加载机器人信息和日志
      const [botResult, logsResult] = await Promise.all([
        this.loadBotInfo(),
        this.loadBotLogs()
      ]);

      if (botResult.success) {
        const bot = {
          ...botResult.data,
          statusText: this.getStatusText(botResult.data.status),
          statusColor: this.getStatusColor(botResult.data.status),
          platformText: this.getPlatformText(botResult.data.platform),
          lastSeenText: this.formatLastSeen(botResult.data.lastSeen),
          uptimeText: this.formatUptime(botResult.data.uptime)
        };

        this.setData({
          bot,
          logs: logsResult.success ? logsResult.data.logs || [] : [],
          config: botResult.data.config || this.data.config,
          isLoaded: true,
          isLoading: false
        });
      } else {
        throw new Error(botResult.error || '加载机器人信息失败');
      }
    } catch (error) {
      console.error('加载机器人详情失败:', error);
      this.setData({
        loadError: error.message || '加载失败，请检查网络连接',
        isLoading: false,
        isLoaded: true
      });
    }
  },

  // 加载机器人信息
  async loadBotInfo() {
    try {
      return await MiniProgramAPI.getBotDetail(this.data.botId);
    } catch (error) {
      return { success: false, error: error.message };
    }
  },

  // 加载机器人日志
  async loadBotLogs() {
    try {
      const result = await MiniProgramAPI.getBotLogs(this.data.botId, 100);
      return result;
    } catch (error) {
      return { success: false, error: error.message, data: { logs: [] } };
    }
  },

  // 切换标签页
  onTabChange(e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({ activeTab: tab });
    
    if (tab === 'logs' && this.data.logs.length === 0) {
      this.loadBotLogs();
    }
  },

  // 切换机器人状态
  async toggleBotStatus() {
    const bot = this.data.bot;
    if (!bot) return;

    try {
      const newStatus = bot.status === 'online' ? 'offline' : 'online';
      const result = await MiniProgramAPI.toggleBot(bot.id, newStatus === 'online');
      
      if (result.success) {
        // 更新本地状态
        const updatedBot = {
          ...bot,
          status: newStatus,
          statusText: this.getStatusText(newStatus),
          statusColor: this.getStatusColor(newStatus)
        };
        
        this.setData({ bot: updatedBot });
        
        wx.showToast({
          title: newStatus === 'online' ? '机器人已启动' : '机器人已停止',
          icon: 'success'
        });
      } else {
        throw new Error(result.error || '操作失败');
      }
    } catch (error) {
      console.error('切换机器人状态失败:', error);
      wx.showToast({
        title: '操作失败',
        icon: 'error'
      });
    }
  },

  // 重启机器人
  async restartBot() {
    try {
      wx.showModal({
        title: '确认重启',
        content: '确定要重启这个机器人吗？',
        success: async (res) => {
          if (res.confirm) {
            MiniProgramAPI.showLoading('重启中...');
            
            try {
              const result = await MiniProgramAPI.restartBot(this.data.botId);
              
              MiniProgramAPI.hideLoading();
              if (result.success) {
                wx.showToast({
                  title: '重启成功',
                  icon: 'success'
                });
                // 延迟刷新状态
                setTimeout(() => this.loadBotDetail(), 2000);
              } else {
                throw new Error(result.error);
              }
            } catch (error) {
              MiniProgramAPI.hideLoading();
              wx.showToast({
                title: '重启失败',
                icon: 'error'
              });
            }
          }
        }
      });
    } catch (error) {
      console.error('重启机器人失败:', error);
    }
  },

  // 删除机器人
  async deleteBot() {
    try {
      wx.showModal({
        title: '确认删除',
        content: '确定要删除这个机器人吗？此操作不可恢复。',
        confirmText: '删除',
        confirmColor: '#ee0a24',
        success: async (res) => {
          if (res.confirm) {
            MiniProgramAPI.showLoading('删除中...');
            
            try {
              const result = await MiniProgramAPI.deleteBot(this.data.botId);
              
              MiniProgramAPI.hideLoading();
              if (result.success) {
                wx.showToast({
                  title: '删除成功',
                  icon: 'success'
                });
                // 返回上一页
                setTimeout(() => wx.navigateBack(), 1500);
              } else {
                throw new Error(result.error);
              }
            } catch (error) {
              MiniProgramAPI.hideLoading();
              wx.showToast({
                title: '删除失败',
                icon: 'error'
              });
            }
          }
        }
      });
    } catch (error) {
      console.error('删除机器人失败:', error);
    }
  },

  // 配置编辑
  startConfigEdit() {
    this.setData({ 
      editConfig: JSON.parse(JSON.stringify(this.data.config))
    });
  },

  cancelConfigEdit() {
    this.setData({ editConfig: null });
  },

  onConfigInput(e) {
    const field = e.currentTarget.dataset.field;
    const value = e.detail.value;
    
    if (this.data.editConfig) {
      const editConfig = { ...this.data.editConfig };
      editConfig[field] = value;
      this.setData({ editConfig });
    }
  },

  onConfigSwitch(e) {
    const field = e.currentTarget.dataset.field;
    const value = e.detail.value;
    
    if (this.data.editConfig) {
      const editConfig = { ...this.data.editConfig };
      editConfig[field] = value;
      this.setData({ editConfig });
    }
  },

  // 保存配置
  async saveConfig() {
    if (!this.data.editConfig) return;

    try {
      MiniProgramAPI.showLoading('保存配置中...');
      
      const result = await MiniProgramAPI.updateBotConfig(this.data.botId, this.data.editConfig);
      
      MiniProgramAPI.hideLoading();
      
      if (result.success) {
        this.setData({
          config: this.data.editConfig,
          editConfig: null
        });
        
        wx.showToast({
          title: '配置保存成功',
          icon: 'success'
        });
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      MiniProgramAPI.hideLoading();
      wx.showToast({
        title: '配置保存失败',
        icon: 'error'
      });
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    await this.loadBotDetail();
    wx.stopPullDownRefresh();
  },

  // 自动刷新
  startAutoRefresh() {
    this.stopAutoRefresh();
    const timer = setInterval(() => {
      this.loadBotInfo().then(result => {
        if (result.success && this.data.activeTab === 'info') {
          const bot = {
            ...result.data,
            statusText: this.getStatusText(result.data.status),
            statusColor: this.getStatusColor(result.data.status),
            platformText: this.getPlatformText(result.data.platform),
            lastSeenText: this.formatLastSeen(result.data.lastSeen),
            uptimeText: this.formatUptime(result.data.uptime)
          };
          this.setData({ bot });
        }
      });
    }, 10000); // 10秒自动刷新
    this.setData({ refreshTimer: timer });
  },

  stopAutoRefresh() {
    if (this.data.refreshTimer) {
      clearInterval(this.data.refreshTimer);
      this.setData({ refreshTimer: null });
    }
  },

  // 工具函数
  getStatusText(status) {
    const statusMap = {
      online: '在线',
      offline: '离线',
      error: '错误',
      starting: '启动中',
      stopping: '停止中'
    };
    return statusMap[status] || '未知';
  },

  getStatusColor(status) {
    const colorMap = {
      online: '#07c160',
      offline: '#969799',
      error: '#ee0a24',
      starting: '#1989fa',
      stopping: '#ff976a'
    };
    return colorMap[status] || '#969799';
  },

  getPlatformText(platform) {
    const platformMap = {
      qq: 'QQ',
      wechat: '微信',
      telegram: 'Telegram',
      discord: 'Discord'
    };
    return platformMap[platform] || platform;
  },

  formatLastSeen(timestamp) {
    if (!timestamp) return '从未上线';
    
    const now = new Date();
    const lastSeen = new Date(timestamp);
    const diff = now - lastSeen;
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
    return `${Math.floor(diff / 86400000)}天前`;
  },

  formatUptime(uptime) {
    if (!uptime) return '未运行';
    
    const seconds = Math.floor(uptime / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days}天${hours % 24}小时`;
    if (hours > 0) return `${hours}小时${minutes % 60}分钟`;
    if (minutes > 0) return `${minutes}分钟`;
    return `${seconds}秒`;
  },

  formatLogTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  },

  getLogLevelColor(level) {
    const colorMap = {
      info: '#1989fa',
      warn: '#ff976a',
      error: '#ee0a24',
      debug: '#969799'
    };
    return colorMap[level] || '#333';
  }
});