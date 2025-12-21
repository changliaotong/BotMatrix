const MiniProgramAPI = require('../../utils/miniprogram_api.js');
const ChartUtil = require('../../utils/chart_util.js');
const app = getApp();

Page({
  data: {
    systemInfo: null,
    performanceData: {
      cpu: [],
      memory: [],
      network: [],
      timestamps: []
    },
    realTimeStats: {
      cpuUsage: 0,
      memoryUsage: 0,
      networkIn: 0,
      networkOut: 0
    },
    isLoading: true,
    isLoaded: false,
    loadError: null,
    refreshTimer: null,
    chartTimer: null,
    activeTab: 'overview', // overview, performance, network, processes
    processList: [],
    networkConnections: [],
    diskUsage: []
  },

  onLoad() {
    this.loadSystemData();
    this.startAutoRefresh();
    this.startChartUpdate();
  },

  onUnload() {
    this.stopAutoRefresh();
    this.stopChartUpdate();
  },

  onShow() {
    if (this.data.isLoaded) {
      this.loadSystemData();
    }
  },

  // 加载系统数据
  async loadSystemData() {
    try {
      this.setData({ isLoading: true, loadError: null });
      
      // 先检查是否有预加载数据
      const preloadedSystemData = app.getPreloadedData('systemStatus');
      let systemData;
      
      if (preloadedSystemData) {
        console.log('使用预加载的系统状态数据');
        systemData = preloadedSystemData;
      } else {
        const result = await MiniProgramAPI.getSystemStatus();
        if (result.success) {
          systemData = result.data;
          // 缓存数据供其他页面使用
          app.preloadData('systemStatus', () => Promise.resolve({ success: true, data: systemData }), 10000);
        } else {
          throw new Error(result.error || '加载系统数据失败');
        }
      }
      
      const systemInfo = {
        ...systemData,
        uptimeText: this.formatUptime(systemData.uptime),
        bootTimeText: this.formatBootTime(systemData.bootTime)
      };

      this.setData({
        systemInfo,
        isLoaded: true,
        isLoading: false
      });

      // 加载其他数据
      this.loadPerformanceData();
      this.loadProcessList();
      this.loadNetworkConnections();
      this.loadDiskUsage();
    } catch (error) {
      console.error('加载系统数据失败:', error);
      this.setData({
        loadError: error.message || '加载失败，请检查网络连接',
        isLoading: false,
        isLoaded: true
      });
    }
  },

  // 加载性能数据
  async loadPerformanceData() {
    try {
      // 模拟性能数据获取
      const performanceData = await this.generatePerformanceData();
      this.updatePerformanceCharts(performanceData);
    } catch (error) {
      console.error('加载性能数据失败:', error);
    }
  },

  // 生成性能数据（模拟）
  async generatePerformanceData() {
    const now = Date.now();
    const performanceData = this.data.performanceData;
    
    // 生成模拟数据
    const cpuUsage = Math.random() * 80 + 10; // 10-90%
    const memoryUsage = Math.random() * 70 + 20; // 20-90%
    const networkIn = Math.random() * 1000; // 0-1000 KB/s
    const networkOut = Math.random() * 800; // 0-800 KB/s

    // 保持最近20个数据点
    const maxPoints = 20;
    performanceData.cpu.push(cpuUsage);
    performanceData.memory.push(memoryUsage);
    performanceData.network.push({ in: networkIn, out: networkOut });
    performanceData.timestamps.push(now);

    if (performanceData.cpu.length > maxPoints) {
      performanceData.cpu.shift();
      performanceData.memory.shift();
      performanceData.network.shift();
      performanceData.timestamps.shift();
    }

    return {
      cpu: performanceData.cpu,
      memory: performanceData.memory,
      network: performanceData.network,
      timestamps: performanceData.timestamps,
      realTimeStats: {
        cpuUsage: Math.round(cpuUsage),
        memoryUsage: Math.round(memoryUsage),
        networkIn: Math.round(networkIn),
        networkOut: Math.round(networkOut)
      }
    };
  },

  // 更新性能图表
  updatePerformanceCharts(data) {
    this.setData({
      performanceData: {
        cpu: data.cpu,
        memory: data.memory,
        network: data.network,
        timestamps: data.timestamps
      },
      realTimeStats: data.realTimeStats
    });

    // 绘制图表
    this.drawCharts();
  },

  // 绘制图表
  drawCharts() {
    // CPU使用率图表
    ChartUtil.drawLineChart('cpuChart', this.data.performanceData.cpu, '#1989fa', 'CPU使用率');
    
    // 内存使用率图表
    ChartUtil.drawLineChart('memoryChart', this.data.performanceData.memory, '#07c160', '内存使用率');
    
    // 网络流量图表
    ChartUtil.drawNetworkChart('networkChart', this.data.performanceData.network);
  },

  // 加载进程列表
  async loadProcessList() {
    try {
      // 模拟进程数据
      const processes = [
        { pid: 1, name: 'BotMatrix Core', cpu: 15, memory: 256, status: 'running' },
        { pid: 2, name: 'QQ Bot #1', cpu: 8, memory: 128, status: 'running' },
        { pid: 3, name: 'WeChat Bot #1', cpu: 12, memory: 192, status: 'running' },
        { pid: 4, name: 'Telegram Bot #1', cpu: 5, memory: 96, status: 'running' },
        { pid: 5, name: 'System Monitor', cpu: 3, memory: 64, status: 'running' }
      ];
      
      this.setData({ processList: processes });
    } catch (error) {
      console.error('加载进程列表失败:', error);
    }
  },

  // 加载网络连接
  async loadNetworkConnections() {
    try {
      // 模拟网络连接数据
      const connections = [
        { protocol: 'TCP', localAddr: '127.0.0.1:3000', remoteAddr: '127.0.0.1:5000', status: 'ESTABLISHED' },
        { protocol: 'TCP', localAddr: '127.0.0.1:3001', remoteAddr: '127.0.0.1:6700', status: 'ESTABLISHED' },
        { protocol: 'TCP', localAddr: '127.0.0.1:3002', remoteAddr: '127.0.0.1:8080', status: 'ESTABLISHED' },
        { protocol: 'UDP', localAddr: '0.0.0.0:53', remoteAddr: '0.0.0.0:0', status: 'LISTEN' }
      ];
      
      this.setData({ networkConnections: connections });
    } catch (error) {
      console.error('加载网络连接失败:', error);
    }
  },

  // 加载磁盘使用情况
  async loadDiskUsage() {
    try {
      // 模拟磁盘使用数据
      const diskUsage = [
        { mount: '/', total: 100, used: 65, free: 35, usage: 65 },
        { mount: '/data', total: 500, used: 320, free: 180, usage: 64 },
        { mount: '/logs', total: 50, used: 12, free: 38, usage: 24 }
      ];
      
      this.setData({ diskUsage });
    } catch (error) {
      console.error('加载磁盘使用失败:', error);
    }
  },

  // 切换标签页
  onTabChange(e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({ activeTab: tab });
    
    // 根据标签页加载相应数据
    switch (tab) {
      case 'performance':
        this.loadPerformanceData();
        break;
      case 'processes':
        this.loadProcessList();
        break;
      case 'network':
        this.loadNetworkConnections();
        break;
    }
  },

  // 下拉刷新
  async onPullDownRefresh() {
    await this.loadSystemData();
    wx.stopPullDownRefresh();
  },

  // 自动刷新
  startAutoRefresh() {
    this.stopAutoRefresh();
    const timer = setInterval(() => {
      this.loadSystemData();
    }, 30000); // 30秒自动刷新
    this.setData({ refreshTimer: timer });
  },

  stopAutoRefresh() {
    if (this.data.refreshTimer) {
      clearInterval(this.data.refreshTimer);
      this.setData({ refreshTimer: null });
    }
  },

  // 图表更新
  startChartUpdate() {
    this.stopChartUpdate();
    const timer = setInterval(() => {
      if (this.data.activeTab === 'performance') {
        this.loadPerformanceData();
      }
    }, 5000); // 5秒更新图表
    this.setData({ chartTimer: timer });
  },

  stopChartUpdate() {
    if (this.data.chartTimer) {
      clearInterval(this.data.chartTimer);
      this.setData({ chartTimer: null });
    }
  },

  // 工具函数
  formatUptime(uptime) {
    if (!uptime) return '未知';
    
    const seconds = Math.floor(uptime / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days}天${hours % 24}小时`;
    if (hours > 0) return `${hours}小时${minutes % 60}分钟`;
    if (minutes > 0) return `${minutes}分钟`;
    return `${seconds}秒`;
  },

  formatBootTime(timestamp) {
    if (!timestamp) return '未知';
    return new Date(timestamp).toLocaleString('zh-CN');
  },

  formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  formatPercent(usage, total) {
    if (!total || total === 0) return '0%';
    return Math.round((usage / total) * 100) + '%';
  },

  getProcessStatusColor(status) {
    const colorMap = {
      running: '#07c160',
      sleeping: '#1989fa',
      stopped: '#ee0a24',
      zombie: '#ff976a'
    };
    return colorMap[status] || '#969799';
  },

  getConnectionStatusColor(status) {
    const colorMap = {
      ESTABLISHED: '#07c160',
      LISTEN: '#1989fa',
      TIME_WAIT: '#ff976a',
      CLOSE_WAIT: '#ee0a24'
    };
    return colorMap[status] || '#969799';
  }
});