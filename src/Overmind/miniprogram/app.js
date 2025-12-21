// 小程序应用入口
const MiniProgramAPI = require('./utils/miniprogram_api.js');

App({
  globalData: {
    userInfo: null,
    systemInfo: null,
    botList: [],
    wsConnected: false,
    socketTask: null,
    apiBaseUrl: 'http://localhost:3001',
    wsUrl: 'ws://localhost:3001',
    eventSubscribers: {}, // 事件订阅者管理
    preloadedData: {} // 预加载数据缓存
  },

  onLaunch() {
    console.log('BotMatrix小程序启动');
    this.getSystemInfo();
    this.checkLoginStatus();
  },

  // 获取系统信息
  getSystemInfo() {
    const systemInfo = MiniProgramAPI.getSystemInfo();
    this.globalData.systemInfo = systemInfo;
    console.log('系统信息:', systemInfo);
  },

  // 检查登录状态
  async checkLoginStatus() {
    try {
      const token = MiniProgramAPI.getStorageSync('token');
      if (token) {
        // 验证token有效性
        const result = await MiniProgramAPI.validateToken(token);
        if (result.valid) {
          this.globalData.userInfo = result.userInfo;
          this.connectWebSocket();
          // 登录成功后立即开始预加载关键数据
          this.startPreload();
          return true;
        } else {
          MiniProgramAPI.removeStorageSync('token');
        }
      }
      // 未登录或token无效，跳转到登录页面
      this.navigateToLogin();
      return false;
    } catch (error) {
      console.error('检查登录状态失败:', error);
      // 错误情况下也跳转到登录页面
      this.navigateToLogin();
      return false;
    }
  },

  // 跳转到登录页面
  navigateToLogin() {
    // 获取当前页面栈
    const pages = getCurrentPages();
    // 如果当前不是登录页面，跳转到登录页面
    if (pages.length === 0 || pages[pages.length - 1].route !== 'pages/login/login') {
      wx.navigateTo({
        url: '/pages/login/login',
        success: () => {
          console.log('跳转到登录页面');
        },
        fail: (error) => {
          console.error('跳转到登录页面失败:', error);
        }
      });
    }
  },

  // 开始预加载关键数据
  async startPreload() {
    console.log('开始预加载数据...');
    try {
      // 并行预加载多种数据，提高效率
      await Promise.all([
        this.preloadData('botStatus', () => MiniProgramAPI.getBotStatus()),
        this.preloadData('systemStatus', () => MiniProgramAPI.getSystemStatus()),
        this.preloadData('botList', () => MiniProgramAPI.getBotList()),
        this.preloadData('alerts', () => MiniProgramAPI.getAlerts()),
        this.preloadData('systemMonitoring', () => MiniProgramAPI.getSystemMonitoring()),
        this.preloadData('networkStatus', () => MiniProgramAPI.getNetworkStatus())
      ]);
      console.log('预加载完成');
    } catch (error) {
      console.error('预加载失败:', error);
    }
  },

  // 预加载数据方法
  async preloadData(key, fetchFunction, expiry = 30000) {
    try {
      const data = await fetchFunction();
      if (data.success) {
        this.globalData.preloadedData[key] = {
          data: data.data,
          timestamp: Date.now(),
          expiry: expiry
        };
        console.log(`预加载数据 [${key}] 成功`);
        return data.data;
      }
    } catch (error) {
      console.error(`预加载数据 [${key}] 失败:`, error);
    }
    return null;
  },

  // 检查预加载数据是否有效
  getPreloadedData(key) {
    const preloadItem = this.globalData.preloadedData[key];
    if (!preloadItem) return null;
    
    // 检查数据是否过期
    const isExpired = Date.now() - preloadItem.timestamp > preloadItem.expiry;
    if (isExpired) {
      delete this.globalData.preloadedData[key];
      return null;
    }
    
    return preloadItem.data;
  },

  // 清理过期的预加载数据
  cleanExpiredPreloadData() {
    const now = Date.now();
    const preloadedData = this.globalData.preloadedData;
    
    for (const [key, item] of Object.entries(preloadedData)) {
      if (now - item.timestamp > item.expiry) {
        delete preloadedData[key];
        console.log(`清理过期预加载数据 [${key}]`);
      }
    }
  },

  // WebSocket连接
  connectWebSocket() {
    if (this.globalData.wsConnected) return;

    MiniProgramAPI.connectWebSocket(this.globalData.wsUrl, {
      onOpen: () => {
        console.log('WebSocket连接成功');
        this.globalData.wsConnected = true;
      },
      onMessage: (data) => {
        this.handleWebSocketMessage(data);
      },
      onClose: () => {
        console.log('WebSocket连接关闭');
        this.globalData.wsConnected = false;
        // 3秒后重连
        setTimeout(() => this.connectWebSocket(), 3000);
      },
      onError: (error) => {
        console.error('WebSocket错误:', error);
        this.globalData.wsConnected = false;
      }
    });
  },

  // 处理WebSocket消息
  handleWebSocketMessage(data) {
    console.log('收到WebSocket消息:', data);
    
    // 处理不同类型的消息
    if (data.type === 'alert') {
      this.handleSystemAlert(data.payload);
    } else if (data.type === 'status_update') {
      this.broadcastEvent('statusUpdate', data.payload);
    } else if (data.type === 'bot_status_change') {
      this.broadcastEvent('botStatusChange', data.payload);
    } else if (data.type === 'system_metrics') {
      this.broadcastEvent('systemMetrics', data.payload);
    }
    
    // 广播消息到各个页面
    const pages = getCurrentPages();
    if (pages.length > 0) {
      const currentPage = pages[pages.length - 1];
      if (currentPage.handleWebSocketMessage) {
        currentPage.handleWebSocketMessage(data);
      }
    }
  },

  // 全局方法：刷新机器人状态
  async refreshBotStatus() {
    try {
      const result = await MiniProgramAPI.getBotStatus();
      if (result.success) {
        this.globalData.botList = result.data.bots || [];
        return result.data;
      }
      return null;
    } catch (error) {
      console.error('刷新机器人状态失败:', error);
      return null;
    }
  },

  // 全局方法：切换机器人状态
  async toggleBot(botId, enable) {
    try {
      const result = await MiniProgramAPI.toggleBot(botId, enable);
      if (result.success) {
        // 更新本地缓存
        const bot = this.globalData.botList.find(b => b.id === botId);
        if (bot) {
          bot.enabled = enable;
        }
        return true;
      }
      return false;
    } catch (error) {
      console.error('切换机器人状态失败:', error);
      return false;
    }
  },

  // 全局方法：发送WebSocket消息
  sendWebSocketMessage(message) {
    if (this.globalData.wsConnected && this.globalData.socketTask) {
      try {
        const messageStr = typeof message === 'string' ? message : JSON.stringify(message);
        this.globalData.socketTask.send({
          data: messageStr,
          success: () => {
            console.log('WebSocket消息发送成功:', message);
          },
          fail: (error) => {
            console.error('WebSocket消息发送失败:', error);
          }
        });
      } catch (error) {
        console.error('WebSocket消息格式错误:', error);
      }
    } else {
      console.warn('WebSocket未连接，无法发送消息');
    }
  },

  // 全局方法：订阅特定事件
  subscribeToEvent(eventType, callback) {
    if (!this.globalData.eventSubscribers[eventType]) {
      this.globalData.eventSubscribers[eventType] = [];
    }
    this.globalData.eventSubscribers[eventType].push(callback);
    
    // 返回取消订阅函数
    return () => {
      const index = this.globalData.eventSubscribers[eventType].indexOf(callback);
      if (index > -1) {
        this.globalData.eventSubscribers[eventType].splice(index, 1);
      }
    };
  },

  // 全局方法：广播事件到订阅者
  broadcastEvent(eventType, data) {
    if (this.globalData.eventSubscribers[eventType]) {
      this.globalData.eventSubscribers[eventType].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`事件订阅者回调错误 [${eventType}]:`, error);
        }
      });
    }
  },

  // 全局方法：处理系统告警
  handleSystemAlert(alert) {
    console.log('收到系统告警:', alert);
    
    // 显示通知
    wx.showToast({
      title: alert.title || '系统告警',
      icon: 'none',
      duration: 3000
    });
    
    // 广播告警事件
    this.broadcastEvent('systemAlert', alert);
    
    // 如果告警级别较高，显示模态框
    if (alert.level === 'critical' || alert.level === 'error') {
      wx.showModal({
        title: alert.title || '系统告警',
        content: alert.message || '系统检测到异常',
        showCancel: false,
        confirmText: '知道了'
      });
    }
  },

  // 全局方法：获取系统状态概览
  async getSystemOverview() {
    try {
      const [systemResult, botResult] = await Promise.all([
        MiniProgramAPI.getSystemStatus(),
        MiniProgramAPI.getBotStatus()
      ]);
      
      const systemData = systemResult.success ? systemResult.data : {};
      const botData = botResult.success ? botResult.data : { bots: [] };
      
      return {
        success: true,
        data: {
          system: systemData,
          bots: botData.bots,
          timestamp: Date.now()
        }
      };
    } catch (error) {
      console.error('获取系统概览失败:', error);
      return {
        success: false,
        error: error.message,
        data: { system: {}, bots: [], timestamp: Date.now() }
      };
    }
  }
});