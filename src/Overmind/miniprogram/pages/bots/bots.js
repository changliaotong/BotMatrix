const MiniProgramAPI = require('../../utils/miniprogram_api.js');

Page({
  data: {
    bots: [],
    filteredBots: [],
    searchKeyword: '',
    statusFilter: 'all', // all, online, offline, error
    sortBy: 'name', // name, status, platform, created
    isLoading: false,
    isLoaded: false,
    loadError: null,
    selectedBots: [],
    isSelecting: false,
    refreshTimer: null
  },

  onLoad() {
    this.loadBots();
    this.startAutoRefresh();
  },

  onUnload() {
    this.stopAutoRefresh();
  },

  onShow() {
    // 页面显示时刷新数据
    if (this.data.isLoaded) {
      this.loadBots();
    }
  },

  // 加载机器人列表
  async loadBots() {
    try {
      this.setData({ isLoading: true, loadError: null });
      
      const result = await MiniProgramAPI.getBotList();
      if (result.success) {
        const bots = result.data.map(bot => ({
          ...bot,
          statusText: this.getStatusText(bot.status),
          statusColor: this.getStatusColor(bot.status),
          platformText: this.getPlatformText(bot.platform),
          lastSeenText: this.formatLastSeen(bot.lastSeen)
        }));

        this.setData({
          bots,
          filteredBots: this.filterBots(bots),
          isLoaded: true,
          isLoading: false
        });
      } else {
        throw new Error(result.error || '加载机器人列表失败');
      }
    } catch (error) {
      console.error('加载机器人列表失败:', error);
      this.setData({
        loadError: error.message || '加载失败，请检查网络连接',
        isLoading: false,
        isLoaded: true
      });
    }
  },

  // 搜索功能
  onSearchInput(e) {
    const keyword = e.detail.value.toLowerCase();
    this.setData({ searchKeyword: keyword });
    this.updateFilteredBots();
  },

  // 状态筛选
  onStatusFilterChange(e) {
    const status = e.currentTarget.dataset.status;
    this.setData({ statusFilter: status });
    this.updateFilteredBots();
  },

  // 排序
  onSortChange(e) {
    const sortBy = e.currentTarget.dataset.sort;
    this.setData({ sortBy });
    this.updateFilteredBots();
  },

  // 更新筛选后的机器人列表
  updateFilteredBots() {
    const filtered = this.filterBots(this.data.bots);
    const sorted = this.sortBots(filtered);
    this.setData({ filteredBots: sorted });
  },

  // 筛选机器人
  filterBots(bots) {
    return bots.filter(bot => {
      // 搜索关键词筛选
      if (this.data.searchKeyword) {
        const keyword = this.data.searchKeyword;
        const matchName = bot.name.toLowerCase().includes(keyword);
        const matchId = bot.id.toString().includes(keyword);
        const matchPlatform = bot.platform.toLowerCase().includes(keyword);
        if (!matchName && !matchId && !matchPlatform) {
          return false;
        }
      }

      // 状态筛选
      if (this.data.statusFilter !== 'all') {
        if (this.data.statusFilter === 'online' && bot.status !== 'online') return false;
        if (this.data.statusFilter === 'offline' && bot.status !== 'offline') return false;
        if (this.data.statusFilter === 'error' && bot.status !== 'error') return false;
      }

      return true;
    });
  },

  // 排序机器人
  sortBots(bots) {
    const sorted = [...bots];
    switch (this.data.sortBy) {
      case 'name':
        return sorted.sort((a, b) => a.name.localeCompare(b.name));
      case 'status':
        const statusOrder = { online: 1, offline: 2, error: 3 };
        return sorted.sort((a, b) => statusOrder[a.status] - statusOrder[b.status]);
      case 'platform':
        return sorted.sort((a, b) => a.platform.localeCompare(b.platform));
      case 'created':
        return sorted.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
      default:
        return sorted;
    }
  },

  // 机器人详情
  onBotTap(e) {
    const botId = e.currentTarget.dataset.id;
    if (this.data.isSelecting) {
      this.toggleBotSelection(botId);
    } else {
      wx.navigateTo({
        url: `/pages/bot-detail/bot-detail?id=${botId}`
      });
    }
  },

  // 切换机器人状态
  async toggleBotStatus(e) {
    const botId = e.currentTarget.dataset.id;
    const bot = this.data.bots.find(b => b.id === botId);
    if (!bot) return;

    try {
      const newStatus = bot.status === 'online' ? 'offline' : 'online';
      const result = await MiniProgramAPI.toggleBot(botId, newStatus);
      
      if (result.success) {
        // 更新本地状态
        const bots = this.data.bots.map(b => 
          b.id === botId ? { 
            ...b, 
            status: newStatus,
            statusText: this.getStatusText(newStatus),
            statusColor: this.getStatusColor(newStatus)
          } : b
        );
        
        this.setData({ bots });
        this.updateFilteredBots();
        
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

  // 批量选择模式
  startSelection() {
    this.setData({ 
      isSelecting: true, 
      selectedBots: [] 
    });
  },

  cancelSelection() {
    this.setData({ 
      isSelecting: false, 
      selectedBots: [] 
    });
  },

  toggleBotSelection(botId) {
    const selected = this.data.selectedBots;
    const index = selected.indexOf(botId);
    
    if (index > -1) {
      selected.splice(index, 1);
    } else {
      selected.push(botId);
    }
    
    this.setData({ selectedBots: selected });
  },

  // 批量操作
  async batchOperation(e) {
    const operation = e.currentTarget.dataset.operation;
    const selectedBots = this.data.selectedBots;
    
    if (selectedBots.length === 0) {
      wx.showToast({
        title: '请选择机器人',
        icon: 'none'
      });
      return;
    }

    try {
      let result;
      switch (operation) {
        case 'start':
          result = await MiniProgramAPI.batchStartBots(selectedBots);
          break;
        case 'stop':
          result = await MiniProgramAPI.batchStopBots(selectedBots);
          break;
        case 'restart':
          result = await MiniProgramAPI.batchRestartBots(selectedBots);
          break;
        default:
          return;
      }

      if (result.success) {
        wx.showToast({
          title: '批量操作成功',
          icon: 'success'
        });
        
        // 刷新数据
        this.loadBots();
        this.cancelSelection();
      } else {
        throw new Error(result.error || '批量操作失败');
      }
    } catch (error) {
      console.error('批量操作失败:', error);
      wx.showToast({
        title: '批量操作失败',
        icon: 'error'
      });
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    await this.loadBots();
    wx.stopPullDownRefresh();
  },

  // 自动刷新
  startAutoRefresh() {
    this.stopAutoRefresh();
    const timer = setInterval(() => {
      if (!this.data.isSelecting) {
        this.loadBots();
      }
    }, 30000); // 30秒自动刷新
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
  }
});