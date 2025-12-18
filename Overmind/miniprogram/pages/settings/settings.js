// pages/settings/settings.js
const api = require('../../utils/miniprogram_api.js');

Page({
  data: {
    settings: {
      system: {
        autoRefresh: true,
        refreshInterval: 30,
        theme: 'auto',
        language: 'zh-CN'
      },
      notifications: {
        enabled: true,
        sound: true,
        vibration: true,
        criticalOnly: false
      },
      performance: {
        cacheEnabled: true,
        compressionEnabled: true,
        timeout: 5000
      },
      network: {
        retryCount: 3,
        retryDelay: 1000,
        timeout: 30000
      }
    },
    originalSettings: {},
    hasChanges: false,
    loading: true,
    saving: false,
    activeTab: 'system',
    tabs: [
      { key: 'system', label: '系统设置', icon: '⚙️' },
      { key: 'account', label: '账户设置', icon: '👤' },
      { key: 'notifications', label: '通知设置', icon: '🔔' },
      { key: 'performance', label: '性能设置', icon: '⚡' },
      { key: 'network', label: '网络设置', icon: '🌐' },
      { key: 'about', label: '关于', icon: 'ℹ️' }
    ],
    userInfo: null,
    showPasswordModal: false,
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
    themes: [
      { value: 'light', label: '浅色主题' },
      { value: 'dark', label: '深色主题' },
      { value: 'auto', label: '跟随系统' }
    ],
    languages: [
      { value: 'zh-CN', label: '简体中文' },
      { value: 'en-US', label: 'English' }
    ],
    intervals: [
      { value: 10, label: '10秒' },
      { value: 30, label: '30秒' },
      { value: 60, label: '1分钟' },
      { value: 300, label: '5分钟' }
    ],
    timeouts: [
      { value: 3000, label: '3秒' },
      { value: 5000, label: '5秒' },
      { value: 10000, label: '10秒' },
      { value: 30000, label: '30秒' }
    ],
    retryCounts: [
      { value: 1, label: '1次' },
      { value: 3, label: '3次' },
      { value: 5, label: '5次' },
      { value: 10, label: '10次' }
    ],
    appInfo: {
      version: '1.0.0',
      build: '20240101',
      name: 'BotMatrix 小程序',
      description: 'BotMatrix 移动端管理应用',
      author: 'BotMatrix Team',
      website: 'https://botmatrix.com',
      github: 'https://github.com/botmatrix/botmatrix-miniprogram'
    },
    // 中间索引变量，用于替代WXML中的findIndex方法
    refreshIntervalIndex: 1, // 默认30秒
    themeIndex: 2, // 默认跟随系统
    languageIndex: 0, // 默认简体中文
    performanceTimeoutIndex: 1, // 默认5秒
    retryCountIndex: 1, // 默认3次
    networkTimeoutIndex: 3 // 默认30秒
  },

  onLoad() {
    this.loadSettings();
    this.loadUserInfo();
  },

  loadUserInfo() {
    const app = getApp();
    this.setData({
      userInfo: app.globalData.userInfo
    });
  },

  // 登出
  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          api.removeStorageSync('token');
          const app = getApp();
          app.globalData.userInfo = null;
          wx.reLaunch({
            url: '/pages/login/login'
          });
        }
      }
    });
  },

  // 显示修改密码弹窗
  showChangePassword() {
    this.setData({
      showPasswordModal: true,
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
  },

  // 隐藏修改密码弹窗
  hideChangePassword() {
    this.setData({
      showPasswordModal: false
    });
  },

  // 输入处理
  onPasswordInput(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({
      [field]: e.detail.value
    });
  },

  // 提交修改密码
  async submitChangePassword() {
    const { oldPassword, newPassword, confirmPassword } = this.data;

    if (!oldPassword || !newPassword || !confirmPassword) {
      wx.showToast({ title: '请填写完整信息', icon: 'none' });
      return;
    }

    if (newPassword !== confirmPassword) {
      wx.showToast({ title: '两次输入的密码不一致', icon: 'none' });
      return;
    }

    if (newPassword.length < 6) {
      wx.showToast({ title: '新密码长度至少为6位', icon: 'none' });
      return;
    }

    const result = await api.changePassword(oldPassword, newPassword);
    if (result.success) {
      this.hideChangePassword();
      wx.showModal({
        title: '成功',
        content: '密码修改成功，请重新登录',
        showCancel: false,
        success: () => {
          this.onLogout();
        }
      });
    }
  },
});

  onUnload() {
    if (this.data.hasChanges) {
      this.showUnsavedChangesWarning();
    }
  },

  // 加载设置
  async loadSettings() {
    this.setData({ loading: true });

    try {
      // 从本地存储加载设置
      const settings = wx.getStorageSync('botmatrix_settings');
      if (settings) {
        this.setData({
          settings: { ...this.data.settings, ...settings },
          originalSettings: JSON.parse(JSON.stringify(settings))
        });
      }

      // 从服务器获取最新配置
      const result = await api.getServerConfig();
      if (result.success) {
        const serverSettings = this.mapServerConfigToLocal(result.data);
        this.setData({
          settings: { ...this.data.settings, ...serverSettings },
          originalSettings: JSON.parse(JSON.stringify(this.data.settings))
        });
      }

      // 更新所有中间索引变量
      this.updateAllIndexVariables();

      this.setData({ loading: false });
    } catch (error) {
      console.error('加载设置失败:', error);
      this.setData({ loading: false });
      wx.showToast({
        title: '加载设置失败',
        icon: 'error'
      });
    }
  },

  // 将服务器配置映射到本地设置
  mapServerConfigToLocal(serverConfig) {
    return {
      system: {
        autoRefresh: serverConfig.auto_refresh !== false,
        refreshInterval: serverConfig.refresh_interval || 30,
        theme: serverConfig.theme || 'auto',
        language: serverConfig.language || 'zh-CN'
      },
      performance: {
        cacheEnabled: serverConfig.cache_enabled !== false,
        compressionEnabled: serverConfig.compression_enabled !== false,
        timeout: serverConfig.timeout || 5000
      },
      network: {
        retryCount: serverConfig.retry_count || 3,
        retryDelay: serverConfig.retry_delay || 1000,
        timeout: serverConfig.network_timeout || 30000
      }
    };
  },

  // 切换标签页
  switchTab(e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({ activeTab: tab });
  },

  // 设置值改变处理
  onSettingChange(e) {
    const { key, subkey } = e.currentTarget.dataset;
    const index = e.detail.value;
    let value;
    let indexUpdate = {};

    // 根据不同的设置项，从对应的数组中获取实际值，并更新索引变量
    if (key === 'system' && subkey === 'refreshInterval') {
      value = this.data.intervals[index].value;
      indexUpdate.refreshIntervalIndex = index;
    } else if (key === 'system' && subkey === 'theme') {
      value = this.data.themes[index].value;
      indexUpdate.themeIndex = index;
    } else if (key === 'system' && subkey === 'language') {
      value = this.data.languages[index].value;
      indexUpdate.languageIndex = index;
    } else if (key === 'performance' && subkey === 'timeout') {
      value = this.data.timeouts[index].value;
      indexUpdate.performanceTimeoutIndex = index;
    } else if (key === 'network' && subkey === 'retryCount') {
      value = this.data.retryCounts[index].value;
      indexUpdate.retryCountIndex = index;
    } else if (key === 'network' && subkey === 'timeout') {
      value = this.data.timeouts[index].value;
      indexUpdate.networkTimeoutIndex = index;
    }

    const newSettings = { ...this.data.settings };
    newSettings[key][subkey] = value;

    this.setData({
      settings: newSettings,
      ...indexUpdate,
      hasChanges: true
    });

    // 实时应用某些设置
    if (key === 'system' && subkey === 'theme') {
      this.applyTheme(value);
    }
  },

  // 开关切换处理
  onSwitchChange(e) {
    const { key, subkey } = e.currentTarget.dataset;
    const value = e.detail.value;

    const newSettings = { ...this.data.settings };
    newSettings[key][subkey] = value;

    this.setData({
      settings: newSettings,
      hasChanges: true
    });
  },

  // 应用主题
  applyTheme(theme) {
    const app = getApp();
    if (app && app.applyTheme) {
      app.applyTheme(theme);
    }
  },

  // 保存设置
  async saveSettings() {
    if (this.data.saving) return;

    this.setData({ saving: true });

    try {
      // 保存到本地存储
      wx.setStorageSync('botmatrix_settings', this.data.settings);

      // 同步到服务器
      const serverConfig = this.mapLocalConfigToServer(this.data.settings);
      const result = await api.updateServerConfig(serverConfig);

      if (result.success) {
        this.setData({
          originalSettings: JSON.parse(JSON.stringify(this.data.settings)),
          hasChanges: false,
          saving: false
        });

        wx.showToast({
          title: '设置已保存',
          icon: 'success'
        });

        // 通知其他页面设置已更新
        this.notifySettingsUpdated();
      } else {
        throw new Error(result.error || '保存失败');
      }
    } catch (error) {
      console.error('保存设置失败:', error);
      this.setData({ saving: false });
      wx.showToast({
        title: '保存失败',
        icon: 'error'
      });
    }
  },

  // 将本地配置映射到服务器配置
  mapLocalConfigToServer(localSettings) {
    return {
      auto_refresh: localSettings.system.autoRefresh,
      refresh_interval: localSettings.system.refreshInterval,
      theme: localSettings.system.theme,
      language: localSettings.system.language,
      notifications_enabled: localSettings.notifications.enabled,
      notification_sound: localSettings.notifications.sound,
      notification_vibration: localSettings.notifications.vibration,
      critical_notifications_only: localSettings.notifications.criticalOnly,
      cache_enabled: localSettings.performance.cacheEnabled,
      compression_enabled: localSettings.performance.compressionEnabled,
      timeout: localSettings.performance.timeout,
      retry_count: localSettings.network.retryCount,
      retry_delay: localSettings.network.retryDelay,
      network_timeout: localSettings.network.timeout
    };
  },

  // 重置设置
  resetSettings() {
    wx.showModal({
      title: '重置设置',
      content: '确定要重置所有设置为默认值吗？',
      success: (res) => {
        if (res.confirm) {
          const defaultSettings = {
            system: {
              autoRefresh: true,
              refreshInterval: 30,
              theme: 'auto',
              language: 'zh-CN'
            },
            notifications: {
              enabled: true,
              sound: true,
              vibration: true,
              criticalOnly: false
            },
            performance: {
              cacheEnabled: true,
              compressionEnabled: true,
              timeout: 5000
            },
            network: {
              retryCount: 3,
              retryDelay: 1000,
              timeout: 30000
            }
          };

          this.setData({
            settings: defaultSettings,
            hasChanges: true
          });

          wx.showToast({
            title: '已重置为默认值',
            icon: 'success'
          });
        }
      }
    });
  },

  // 恢复默认设置
  restoreDefaults() {
    this.resetSettings();
  },

  // 导出设置
  exportSettings() {
    const settingsJson = JSON.stringify(this.data.settings, null, 2);
    
    wx.setClipboardData({
      data: settingsJson,
      success: () => {
        wx.showToast({
          title: '设置已复制到剪贴板',
          icon: 'success'
        });
      },
      fail: () => {
        wx.showToast({
          title: '复制失败',
          icon: 'error'
        });
      }
    });
  },

  // 导入设置
  importSettings() {
    wx.showModal({
      title: '导入设置',
      content: '请粘贴设置JSON数据',
      editable: true,
      placeholderText: '粘贴设置JSON数据...',
      success: (res) => {
        if (res.confirm && res.content) {
          try {
            const importedSettings = JSON.parse(res.content);
            
            // 验证设置格式
            if (this.validateSettings(importedSettings)) {
              this.setData({
                settings: { ...this.data.settings, ...importedSettings },
                hasChanges: true
              });

              wx.showToast({
                title: '设置导入成功',
                icon: 'success'
              });
            } else {
              wx.showToast({
                title: '设置格式错误',
                icon: 'error'
              });
            }
          } catch (error) {
            wx.showToast({
              title: 'JSON格式错误',
              icon: 'error'
            });
          }
        }
      }
    });
  },

  // 验证设置格式
  validateSettings(settings) {
    const requiredKeys = ['system', 'notifications', 'performance', 'network'];
    const systemKeys = ['autoRefresh', 'refreshInterval', 'theme', 'language'];
    const notificationKeys = ['enabled', 'sound', 'vibration', 'criticalOnly'];
    const performanceKeys = ['cacheEnabled', 'compressionEnabled', 'timeout'];
    const networkKeys = ['retryCount', 'retryDelay', 'timeout'];

    try {
      return requiredKeys.every(key => settings[key]) &&
             systemKeys.every(key => typeof settings.system[key] !== 'undefined') &&
             notificationKeys.every(key => typeof settings.notifications[key] !== 'undefined') &&
             performanceKeys.every(key => typeof settings.performance[key] !== 'undefined') &&
             networkKeys.every(key => typeof settings.network[key] !== 'undefined');
    } catch (error) {
      return false;
    }
  },

  // 检查更新
  checkForUpdates() {
    wx.showModal({
      title: '检查更新',
      content: '当前版本: ' + this.data.appInfo.version + '\n\n点击确定检查更新',
      success: async (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '检查中...',
            mask: true
          });

          try {
            // 模拟检查更新
            await new Promise(resolve => setTimeout(resolve, 2000));
            
            wx.hideLoading();
            wx.showModal({
              title: '检查完成',
              content: '当前已是最新版本',
              showCancel: false
            });
          } catch (error) {
            wx.hideLoading();
            wx.showToast({
              title: '检查失败',
              icon: 'error'
            });
          }
        }
      }
    });
  },

  // 清除缓存
  clearCache() {
    wx.showModal({
      title: '清除缓存',
      content: '确定要清除所有缓存数据吗？这可能会导致需要重新登录。',
      success: (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '清除中...',
            mask: true
          });

          try {
            // 清除本地存储
            wx.clearStorageSync();
            
            // 清除临时文件
            wx.getFileSystemManager().readdir({
              dirPath: wx.env.USER_DATA_PATH,
              success: (res) => {
                res.files.forEach(file => {
                  if (file !== 'miniprogram.log') {
                    wx.getFileSystemManager().unlink({
                      filePath: wx.env.USER_DATA_PATH + '/' + file
                    });
                  }
                });
              }
            });

            wx.hideLoading();
            wx.showToast({
              title: '缓存已清除',
              icon: 'success'
            });

            // 重新加载设置
            setTimeout(() => {
              this.loadSettings();
            }, 1000);
          } catch (error) {
            wx.hideLoading();
            wx.showToast({
              title: '清除失败',
              icon: 'error'
            });
          }
        }
      }
    });
  },

  // 显示未保存更改警告
  showUnsavedChangesWarning() {
    wx.showModal({
      title: '未保存的更改',
      content: '您有未保存的设置更改，是否保存？',
      confirmText: '保存',
      cancelText: '放弃',
      success: (res) => {
        if (res.confirm) {
          this.saveSettings();
        }
      }
    });
  },

  // 更新所有中间索引变量
  updateAllIndexVariables() {
    const { settings } = this.data;
    const indexUpdates = {};

    // 刷新间隔索引
    indexUpdates.refreshIntervalIndex = this.data.intervals.findIndex(item => item.value === settings.system.refreshInterval) || 1;
    // 主题索引
    indexUpdates.themeIndex = this.data.themes.findIndex(item => item.value === settings.system.theme) || 2;
    // 语言索引
    indexUpdates.languageIndex = this.data.languages.findIndex(item => item.value === settings.system.language) || 0;
    // 性能超时索引
    indexUpdates.performanceTimeoutIndex = this.data.timeouts.findIndex(item => item.value === settings.performance.timeout) || 1;
    // 重试次数索引
    indexUpdates.retryCountIndex = this.data.retryCounts.findIndex(item => item.value === settings.network.retryCount) || 1;
    // 网络超时索引
    indexUpdates.networkTimeoutIndex = this.data.timeouts.findIndex(item => item.value === settings.network.timeout) || 3;

    this.setData(indexUpdates);
  },

  // 通知设置已更新
  notifySettingsUpdated() {
    const app = getApp();
    if (app && app.broadcastEvent) {
      app.broadcastEvent('settingsUpdated', this.data.settings);
    }
  },

  // 显示应用信息
  showAppInfo() {
    const info = this.data.appInfo;
    const content = `名称: ${info.name}\n版本: ${info.version}\n构建: ${info.build}\n描述: ${info.description}\n作者: ${info.author}\n官网: ${info.website}`;
    
    wx.showModal({
      title: '应用信息',
      content: content,
      confirmText: '访问官网',
      success: (res) => {
        if (res.confirm) {
          // 复制官网地址到剪贴板
          wx.setClipboardData({
            data: info.website,
            success: () => {
              wx.showToast({
                title: '官网地址已复制',
                icon: 'success'
              });
            }
          });
        }
      }
    });
  },

  // 打开GitHub页面
  openGitHub() {
    const githubUrl = this.data.appInfo.github;
    
    wx.setClipboardData({
      data: githubUrl,
      success: () => {
        wx.showModal({
          title: 'GitHub仓库',
          content: '项目地址已复制到剪贴板，请在浏览器中打开',
          showCancel: false
        });
      }
    });
  }
});