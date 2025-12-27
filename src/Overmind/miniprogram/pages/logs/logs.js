// pages/logs/logs.js
const api = require('../../utils/miniprogram_api.js');

Page({
  data: {
    logs: [],
    filteredLogs: [],
    loading: true,
    error: null,
    searchKeyword: '',
    selectedLevel: 'all',
    selectedLevelIndex: 0,
    selectedBot: 'all',
    selectedBotIndex: 0,
    levels: [
      { value: 'all', label: '全部级别' },
      { value: 'debug', label: '调试' },
      { value: 'info', label: '信息' },
      { value: 'warning', label: '警告' },
      { value: 'error', label: '错误' },
      { value: 'critical', label: '严重' }
    ],
    bots: [
      { value: 'all', label: '全部机器人' }
    ],
    page: 1,
    pageSize: 20,
    hasMore: true,
    autoRefresh: false,
    refreshInterval: null,
    sortBy: 'timestamp',
    sortOrder: 'desc',
    selectedLogIds: [],
    selectedLogs: [],
    showExportModal: false,
    exportFormat: 'json'
  },

  onLoad() {
    this.loadLogs();
    this.loadBots();
    this.startAutoRefresh();
  },

  onUnload() {
    this.stopAutoRefresh();
  },

  onPullDownRefresh() {
    this.refreshLogs();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMoreLogs();
    }
  },

  // 加载日志列表
  async loadLogs() {
    if (this.data.loading) return;

    this.setData({ loading: true, error: null });

    try {
      const result = await api.getLogs({
        page: this.data.page,
        pageSize: this.data.pageSize,
        level: this.data.selectedLevel === 'all' ? null : this.data.selectedLevel,
        botId: this.data.selectedBot === 'all' ? null : this.data.selectedBot,
        search: this.data.searchKeyword || null,
        sortBy: this.data.sortBy,
        sortOrder: this.data.sortOrder
      });

      if (result.success) {
        const logs = this.formatLogs(result.data.logs || []);
        const hasMore = result.data.hasMore || false;

        this.setData({
          logs,
          filteredLogs: logs,
          hasMore,
          loading: false
        });

        this.applyFilters();
      } else {
        this.setData({
          error: result.error || '加载日志失败',
          loading: false
        });
      }
    } catch (error) {
      console.error('加载日志失败:', error);
      this.setData({
        error: '网络错误，请检查网络连接',
        loading: false
      });
    }
  },

  // 加载更多日志
  async loadMoreLogs() {
    const nextPage = this.data.page + 1;
    
    try {
      const result = await api.getLogs({
        page: nextPage,
        pageSize: this.data.pageSize,
        level: this.data.selectedLevel === 'all' ? null : this.data.selectedLevel,
        botId: this.data.selectedBot === 'all' ? null : this.data.selectedBot,
        search: this.data.searchKeyword || null,
        sortBy: this.data.sortBy,
        sortOrder: this.data.sortOrder
      });

      if (result.success) {
        const newLogs = this.formatLogs(result.data.logs || []);
        const allLogs = [...this.data.logs, ...newLogs];
        const hasMore = result.data.hasMore || false;

        this.setData({
          logs: allLogs,
          filteredLogs: allLogs,
          page: nextPage,
          hasMore,
          loading: false
        });

        this.applyFilters();
      }
    } catch (error) {
      console.error('加载更多日志失败:', error);
    }
  },

  // 加载机器人列表
  async loadBots() {
    try {
      const result = await api.getBots();
      if (result.success) {
        const bots = result.data.bots || [];
        const botOptions = [
          { value: 'all', label: '全部机器人' },
          ...bots.map(bot => ({
            value: bot.id,
            label: bot.name || bot.id
          }))
        ];

        this.setData({ bots: botOptions });
      }
    } catch (error) {
      console.error('加载机器人列表失败:', error);
    }
  },

  // 格式化日志数据
  formatLogs(logs) {
    return logs.map(log => ({
      ...log,
      time: this.formatTime(log.timestamp),
      levelClass: this.getLevelClass(log.level),
      levelText: this.getLevelText(log.level),
      shortMessage: this.truncateMessage(log.message, 100),
      fullMessage: log.message
    }));
  },

  // 格式化时间
  formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;

    if (diff < 60000) {
      return '刚刚';
    } else if (diff < 3600000) {
      return `${Math.floor(diff / 60000)}分钟前`;
    } else if (diff < 86400000) {
      return `${Math.floor(diff / 3600000)}小时前`;
    } else {
      return date.toLocaleString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    }
  },

  // 获取日志级别样式类
  getLevelClass(level) {
    const levelClasses = {
      debug: 'level-debug',
      info: 'level-info',
      warning: 'level-warning',
      error: 'level-error',
      critical: 'level-critical'
    };
    return levelClasses[level] || 'level-info';
  },

  // 获取日志级别文本
  getLevelText(level) {
    const levelTexts = {
      debug: '调试',
      info: '信息',
      warning: '警告',
      error: '错误',
      critical: '严重'
    };
    return levelTexts[level] || level;
  },

  // 截断消息
  truncateMessage(message, maxLength) {
    if (!message) return '';
    if (message.length <= maxLength) return message;
    return message.substring(0, maxLength) + '...';
  },

  // 应用过滤器
  applyFilters() {
    let filteredLogs = [...this.data.logs];

    // 按级别筛选
    if (this.data.selectedLevel !== 'all') {
      filteredLogs = filteredLogs.filter(log => log.level === this.data.selectedLevel);
    }

    // 按机器人筛选
    if (this.data.selectedBot !== 'all') {
      filteredLogs = filteredLogs.filter(log => log.botId === this.data.selectedBot);
    }

    // 按关键词搜索
    if (this.data.searchKeyword) {
      const keyword = this.data.searchKeyword.toLowerCase();
      filteredLogs = filteredLogs.filter(log => 
        log.message.toLowerCase().includes(keyword) ||
        log.botId.toLowerCase().includes(keyword) ||
        log.level.toLowerCase().includes(keyword)
      );
    }

    this.setData({ filteredLogs });
  },

  // 搜索输入处理
  onSearchInput(e) {
    this.setData({ searchKeyword: e.detail.value });
    this.debounce(this.applyFilters, 300)();
  },

  // 级别选择处理
  onLevelChange(e) {
    const index = e.detail.value;
    const selectedLevel = this.data.levels[index].value;
    this.setData({ 
      selectedLevelIndex: index,
      selectedLevel: selectedLevel 
    });
    this.applyFilters();
  },

  // 机器人选择处理
  onBotChange(e) {
    const index = e.detail.value;
    const selectedBot = this.data.bots[index].value;
    this.setData({ 
      selectedBotIndex: index,
      selectedBot: selectedBot 
    });
    this.applyFilters();
  },

  // 日志项点击处理
  onLogItemTap(e) {
    const log = e.currentTarget.dataset.log;
    this.showLogDetail(log);
  },

  // 显示日志详情
  showLogDetail(log) {
    wx.showModal({
      title: '日志详情',
      content: `时间: ${log.time}\n级别: ${log.levelText}\n机器人: ${log.botId}\n消息: ${log.fullMessage}`,
      showCancel: false,
      confirmText: '关闭'
    });
  },

  // 日志项长按处理
  onLogItemLongPress(e) {
    const log = e.currentTarget.dataset.log;
    this.showLogContextMenu(log);
  },

  // 显示日志上下文菜单
  showLogContextMenu(log) {
    const itemList = ['复制消息', '查看详情', '标记为已读'];
    
    wx.showActionSheet({
      itemList,
      success: (res) => {
        switch (res.tapIndex) {
          case 0:
            this.copyLogMessage(log);
            break;
          case 1:
            this.showLogDetail(log);
            break;
          case 2:
            this.markLogAsRead(log);
            break;
        }
      }
    });
  },

  // 复制日志消息
  copyLogMessage(log) {
    wx.setClipboardData({
      data: log.fullMessage,
      success: () => {
        wx.showToast({
          title: '已复制到剪贴板',
          icon: 'success'
        });
      }
    });
  },

  // 标记日志为已读
  markLogAsRead(log) {
    // 这里可以添加标记已读的逻辑
    wx.showToast({
      title: '已标记为已读',
      icon: 'success'
    });
  },

  // 选择日志项
  onLogSelect(e) {
    const log = e.currentTarget.dataset.log;
    const selectedLogIds = this.data.selectedLogIds;
    const selectedLogs = this.data.selectedLogs;
    const index = selectedLogIds.indexOf(log.id);

    if (index > -1) {
      // 取消选择
      selectedLogIds.splice(index, 1);
      selectedLogs.splice(selectedLogs.findIndex(item => item.id === log.id), 1);
    } else {
      // 选择
      selectedLogIds.push(log.id);
      selectedLogs.push(log);
    }

    this.setData({ selectedLogIds, selectedLogs });
  },

  // 显示导出模态框
  showExportModal() {
    this.setData({ showExportModal: true });
  },

  // 隐藏导出模态框
  hideExportModal() {
    this.setData({ showExportModal: false });
  },

  // 导出格式选择
  onExportFormatChange(e) {
    this.setData({ exportFormat: e.detail.value });
  },

  // 导出日志
  async exportLogs() {
    const format = this.data.exportFormat;
    const logs = this.data.selectedLogs.length > 0 ? this.data.selectedLogs : this.data.filteredLogs;

    try {
      const result = await api.exportLogs({
        logs,
        format,
        filters: {
          level: this.data.selectedLevel,
          botId: this.data.selectedBot,
          keyword: this.data.searchKeyword
        }
      });

      if (result.success) {
        wx.showToast({
          title: '导出成功',
          icon: 'success'
        });
        this.hideExportModal();
      } else {
        wx.showToast({
          title: '导出失败',
          icon: 'error'
        });
      }
    } catch (error) {
      console.error('导出日志失败:', error);
      wx.showToast({
        title: '导出失败',
        icon: 'error'
      });
    }
  },

  // 清空日志
  clearLogs() {
    wx.showModal({
      title: '清空日志',
      content: '确定要清空所有日志吗？此操作不可恢复。',
      success: async (res) => {
        if (res.confirm) {
          try {
            const result = await api.clearLogs();
            if (result.success) {
              this.setData({
                logs: [],
                filteredLogs: [],
                selectedLogs: []
              });
              wx.showToast({
                title: '已清空',
                icon: 'success'
              });
            }
          } catch (error) {
            console.error('清空日志失败:', error);
            wx.showToast({
              title: '操作失败',
              icon: 'error'
            });
          }
        }
      }
    });
  },

  // 刷新日志
  refreshLogs() {
    this.setData({
      page: 1,
      hasMore: true,
      selectedLogs: []
    });
    this.loadLogs();
  },

  // 开始自动刷新
  startAutoRefresh() {
    if (this.data.autoRefresh) return;
    
    this.setData({ autoRefresh: true });
    
    const interval = setInterval(() => {
      this.loadLogs();
    }, 30000); // 30秒刷新一次

    this.setData({ refreshInterval: interval });
  },

  // 停止自动刷新
  stopAutoRefresh() {
    if (this.data.refreshInterval) {
      clearInterval(this.data.refreshInterval);
      this.setData({
        autoRefresh: false,
        refreshInterval: null
      });
    }
  },

  // 切换自动刷新
  toggleAutoRefresh() {
    if (this.data.autoRefresh) {
      this.stopAutoRefresh();
    } else {
      this.startAutoRefresh();
    }
  },

  // 防抖函数
  debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  },

  // WebSocket 消息处理
  handleWebSocketMessage(data) {
    if (data.type === 'log_update') {
      // 收到新的日志消息
      const newLog = this.formatLogs([data.payload])[0];
      const logs = [newLog, ...this.data.logs].slice(0, 200); // 保持最多200条
      
      this.setData({ logs });
      this.applyFilters();
      
      // 显示通知
      if (data.payload.level === 'error' || data.payload.level === 'critical') {
        this.showNewLogNotification(newLog);
      }
    }
  },

  // 显示新日志通知
  showNewLogNotification(log) {
    // 只在小程序前台时显示通知
    if (getApp().globalData.appInForeground) {
      wx.showToast({
        title: `新${log.levelText}日志`,
        icon: 'none',
        duration: 2000
      });
    }
  }
});