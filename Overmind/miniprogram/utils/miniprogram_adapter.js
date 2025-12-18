// 微信小程序适配器 - 封装微信小程序API
class MiniProgramAdapter {
  // 显示加载提示
  static showLoading(title = '加载中...') {
    wx.showLoading({
      title: title,
      mask: true
    });
  }

  // 隐藏加载提示
  static hideLoading() {
    wx.hideLoading();
  }

  // 显示消息提示
  static showToast(title, icon = 'none', duration = 2000) {
    wx.showToast({
      title: title,
      icon: icon,
      duration: duration
    });
  }

  // 识别错误类型并返回友好提示
  static getErrorMsg(error) {
    const errorMap = {
      'request:fail timeout': '网络请求超时，请检查网络连接',
      'request:fail 网络错误': '网络连接失败，请检查网络设置',
      'request:fail 404': '请求的资源不存在',
      'request:fail 401': '登录已过期，请重新登录',
      'request:fail 500': '服务器内部错误，请稍后重试'
    };
    
    // 匹配错误信息
    for (const [key, value] of Object.entries(errorMap)) {
      if (error.message.includes(key)) {
        return value;
      }
    }
    
    // 默认错误提示
    return error.message || '操作失败，请稍后重试';
  }

  // 发起网络请求
  static request(options) {
    return new Promise((resolve, reject) => {
      // 自动注入token
      const token = this.getStorageSync('token');
      const header = {
        'content-type': 'application/json',
        ...options.header
      };
      
      if (token) {
        header['Authorization'] = `Bearer ${token}`;
      }

      wx.request({
        url: options.url,
        method: options.method || 'GET',
        data: options.data || {},
        header: header,
        timeout: options.timeout || 10000,
        success: (res) => {
          if (res.statusCode === 200) {
            resolve(res.data);
          } else if (res.statusCode === 401) {
            // 登录失效，清除token并跳转
            this.removeStorageSync('token');
            const pages = getCurrentPages();
            if (pages.length > 0 && pages[pages.length - 1].route !== 'pages/login/login') {
              wx.reLaunch({
                url: '/pages/login/login'
              });
            }
            reject(new Error('request:fail 401'));
          } else {
            reject(new Error(`request:fail ${res.statusCode}`));
          }
        },
        fail: (error) => {
          reject(error);
        }
      });
    });
  }

  // 设置本地存储
  static setStorageSync(key, data) {
    try {
      wx.setStorageSync(key, data);
      return true;
    } catch (error) {
      console.error('设置存储失败:', error);
      return false;
    }
  }

  // 获取本地存储
  static getStorageSync(key) {
    try {
      return wx.getStorageSync(key);
    } catch (error) {
      console.error('获取存储失败:', error);
      return null;
    }
  }

  // 删除本地存储
  static removeStorageSync(key) {
    try {
      wx.removeStorageSync(key);
      return true;
    } catch (error) {
      console.error('删除存储失败:', error);
      return false;
    }
  }

  // 建立WebSocket连接
  static connectWebSocket(url, callbacks) {
    // 自动注入token到URL参数中
    const token = this.getStorageSync('token');
    let finalUrl = url;
    if (token) {
      const separator = url.indexOf('?') !== -1 ? '&' : '?';
      finalUrl = `${url}${separator}token=${token}`;
    }

    const socketTask = wx.connectSocket({
      url: finalUrl,
      header: {
        'content-type': 'application/json'
      },
      protocols: ['protocol1']
    });

    socketTask.onOpen(() => {
      if (callbacks.onOpen) callbacks.onOpen();
    });

    socketTask.onMessage((message) => {
      if (callbacks.onMessage) callbacks.onMessage(message);
    });

    socketTask.onClose(() => {
      if (callbacks.onClose) callbacks.onClose();
    });

    socketTask.onError((error) => {
      if (callbacks.onError) callbacks.onError(error);
    });

    return socketTask;
  }

  // 获取系统信息
  static getSystemInfo() {
    try {
      return wx.getSystemInfoSync();
    } catch (error) {
      console.error('获取系统信息失败:', error);
      return null;
    }
  }

  // 显示模态对话框
  static showModal(options) {
    return new Promise((resolve) => {
      wx.showModal({
        title: options.title || '',
        content: options.content || '',
        showCancel: options.showCancel !== undefined ? options.showCancel : true,
        cancelText: options.cancelText || '取消',
        confirmText: options.confirmText || '确定',
        success: (res) => {
          resolve(res);
        }
      });
    });
  }

  // 显示操作菜单
  static showActionSheet(options) {
    return new Promise((resolve, reject) => {
      wx.showActionSheet({
        itemList: options.itemList || [],
        itemColor: options.itemColor || '#000000',
        success: (res) => {
          resolve(res);
        },
        fail: (error) => {
          reject(error);
        }
      });
    });
  }

  // 停止下拉刷新
  static stopPullDownRefresh() {
    wx.stopPullDownRefresh();
  }
}

module.exports = MiniProgramAdapter;