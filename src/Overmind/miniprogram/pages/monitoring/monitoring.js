const MiniProgramAPI = require('../../utils/miniprogram_api.js');
const ChartUtil = require('../../utils/chart_util.js');
const MiniProgramAdapter = require('../../utils/miniprogram_adapter.js');
const app = getApp();

Page({
  data: {
    monitoringData: {
      cpu: { usage: 0, temperature: 0, cores: 1 },
      memory: { total: 0, used: 0, free: 0, usage: 0 },
      disk: { total: 0, used: 0, free: 0, usage: 0 },
      network: { upload: 0, download: 0, connections: 0 },
      processes: { total: 0, running: 0, sleeping: 0 },
      uptime: '0天0小时0分钟'
    },
    performanceData: {
      cpuHistory: [],
      memoryHistory: [],
      networkHistory: []
    },
    networkStatus: {
      status: 'unknown',
      latency: 0,
      packetLoss: 0
    },
    timeRange: '1h',
    refreshTimer: null,
    isRefreshing: false,
    lastUpdateTime: ''
  },

  onLoad() {
    this.loadMonitoringData();
    this.loadPerformanceData();
    this.loadNetworkStatus();
    this.startAutoRefresh();
  },

  onUnload() {
    this.stopAutoRefresh();
  },

  onPullDownRefresh() {
    this.refreshData();
    MiniProgramAdapter.stopPullDownRefresh();
  },

  // 加载系统监控数据
  async loadMonitoringData() {
    try {
      this.setData({ isLoading: true });
      
      // 先检查是否有预加载数据
      const preloadedMonitoringData = app.getPreloadedData('systemMonitoring');
      let monitoringData;
      
      if (preloadedMonitoringData) {
        console.log('使用预加载的系统监控数据');
        monitoringData = preloadedMonitoringData;
      } else {
        const result = await MiniProgramAPI.getSystemMonitoring();
        if (result.success) {
          monitoringData = result.data;
          // 缓存数据供其他页面使用
          app.preloadData('systemMonitoring', () => Promise.resolve({ success: true, data: monitoringData }), 10000);
        } else {
          throw new Error(result.error || '获取监控数据失败');
        }
      }
      
      this.setData({
        monitoringData: monitoringData,
        lastUpdateTime: new Date().toLocaleString(),
        isLoading: false
      });
    } catch (error) {
      console.error('加载监控数据失败:', error);
      this.setData({ isLoading: false });
      MiniProgramAdapter.showToast({
        title: '加载失败',
        icon: 'error'
      });
    }
  },

  // 加载性能数据
  async loadPerformanceData() {
    try {
      const result = await MiniProgramAPI.getPerformanceData(this.data.timeRange);
      if (result.success) {
        this.setData({
          performanceData: {
            cpuHistory: result.data.cpu || [],
            memoryHistory: result.data.memory || [],
            networkHistory: result.data.network || []
          }
        });
        // 绘制图表
        this.drawCharts();
      } else {
        console.error('获取性能数据失败:', result.error);
      }
    } catch (error) {
      console.error('加载性能数据失败:', error);
    }
  },

  // 加载网络状态
  async loadNetworkStatus() {
    try {
      // 先检查是否有预加载数据
      const preloadedNetworkStatus = app.getPreloadedData('networkStatus');
      let networkStatus;
      
      if (preloadedNetworkStatus) {
        console.log('使用预加载的网络状态数据');
        networkStatus = preloadedNetworkStatus;
      } else {
        const result = await MiniProgramAPI.getNetworkStatus();
        if (result.success) {
          networkStatus = result.data;
          // 缓存数据供其他页面使用
          app.preloadData('networkStatus', () => Promise.resolve({ success: true, data: networkStatus }), 10000);
        } else {
          throw new Error(result.error || '获取网络状态失败');
        }
      }
      
      this.setData({
        networkStatus: networkStatus
      });
    } catch (error) {
      console.error('加载网络状态失败:', error);
    }
  },

  // 刷新所有数据
  async refreshData() {
    if (this.data.isRefreshing) return;
    
    this.setData({ isRefreshing: true });
    
    try {
      await Promise.all([
        this.loadMonitoringData(),
        this.loadPerformanceData(),
        this.loadNetworkStatus()
      ]);
      
      MiniProgramAdapter.showToast({
        title: '刷新成功',
        icon: 'success'
      });
    } catch (error) {
      console.error('刷新数据失败:', error);
      MiniProgramAdapter.showToast({
        title: '刷新失败',
        icon: 'error'
      });
    } finally {
      this.setData({ isRefreshing: false });
    }
  },

  // 时间范围切换
  onTimeRangeChange(e) {
    const timeRange = e.currentTarget.dataset.range;
    this.setData({ timeRange });
    this.loadPerformanceData();
  },

  // 自动刷新
  startAutoRefresh() {
    this.stopAutoRefresh();
    const timer = setInterval(() => {
      if (!this.data.isRefreshing) {
        this.loadMonitoringData();
        this.loadNetworkStatus();
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

  // 格式化字节大小
  formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  // 格式化网络速度
  formatSpeed(bytesPerSecond) {
    return this.formatBytes(bytesPerSecond) + '/s';
  },

  // 格式化时间
  formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}天${hours}小时${minutes}分钟`;
  },

  // 绘制图表
  drawCharts() {
    const { performanceData } = this.data;
    
    // 转换CPU数据格式
    const cpuData = performanceData.cpuHistory.map(item => item.value);
    // 转换内存数据格式
    const memoryData = performanceData.memoryHistory.map(item => item.value);
    // 转换网络数据格式
    const networkData = performanceData.networkHistory.map(item => ({
      in: item.download,
      out: item.upload
    }));
    
    // 绘制CPU图表
    if (cpuData.length > 0) {
      ChartUtil.drawLineChart('cpuChart', cpuData, '#1989fa', 'CPU使用率');
    }
    
    // 绘制内存图表
    if (memoryData.length > 0) {
      ChartUtil.drawLineChart('memoryChart', memoryData, '#07c160', '内存使用率');
    }
    
    // 绘制网络图表
    if (networkData.length > 0) {
      ChartUtil.drawNetworkChart('networkChart', networkData);
    }
  },

  // 获取状态颜色
  getStatusColor(value, thresholds) {
    if (value < thresholds.warning) return '#52c41a'; // 绿色
    if (value < thresholds.critical) return '#faad14'; // 黄色
    return '#f5222d'; // 红色
  },

  // 获取CPU使用率颜色
  getCpuUsageColor(usage) {
    return this.getStatusColor(usage, { warning: 60, critical: 80 });
  },

  // 获取内存使用率颜色
  getMemoryUsageColor(usage) {
    return this.getStatusColor(usage, { warning: 70, critical: 85 });
  },

  // 获取磁盘使用率颜色
  getDiskUsageColor(usage) {
    return this.getStatusColor(usage, { warning: 75, critical: 90 });
  },

  // 获取网络状态颜色
  getNetworkStatusColor(status) {
    const colors = {
      'online': '#52c41a',
      'offline': '#f5222d',
      'slow': '#faad14',
      'unknown': '#d9d9d9'
    };
    return colors[status] || colors.unknown;
  },

  // 获取网络状态文本
  getNetworkStatusText(status) {
    const texts = {
      'online': '正常',
      'offline': '离线',
      'slow': '缓慢',
      'unknown': '未知'
    };
    return texts[status] || texts.unknown;
  }
});