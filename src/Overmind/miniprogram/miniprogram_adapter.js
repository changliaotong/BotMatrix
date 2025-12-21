// 微信/QQ小程序统一适配器
class MiniProgramAdapter {
  // 平台检测
  static String get platform {
    if (typeof wx !== 'undefined') return 'wechat';
    if (typeof qq !== 'undefined') return 'qq';
    if (typeof swan !== 'undefined') return 'baidu';
    if (typeof my !== 'undefined') return 'alipay';
    return 'unknown';
  }

  // HTTP请求适配
  static request(options) {
    return new Promise((resolve, reject) => {
      const currentPlatform = this.getCurrentPlatform();
      
      const requestOptions = {
        url: options.url,
        method: options.method || 'GET',
        data: options.data || {},
        header: options.header || {
          'Content-Type': 'application/json'
        },
        timeout: options.timeout || 3000, // 小程序优化超时
        success: (res) => {
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve(res.data);
          } else {
            reject(new Error(`请求失败: ${res.statusCode}`));
          }
        },
        fail: (err) => {
          reject(new Error(`网络错误: ${err.errMsg}`));
        }
      };

      currentPlatform.request(requestOptions);
    });
  }

  // WebSocket连接适配
  static connectWebSocket(url) {
    return new Promise((resolve, reject) => {
      const currentPlatform = this.getCurrentPlatform();
      
      try {
        const socketTask = currentPlatform.connectSocket({
          url: url,
          success: () => {
            resolve(socketTask);
          },
          fail: (err) => {
            reject(new Error(`WebSocket连接失败: ${err.errMsg}`));
          }
        });
      } catch (error) {
        reject(error);
      }
    });
  }

  // 获取当前平台实例
  static getCurrentPlatform() {
    if (typeof wx !== 'undefined') return wx;
    if (typeof qq !== 'undefined') return qq;
    if (typeof swan !== 'undefined') return swan;
    if (typeof my !== 'undefined') return my;
    throw new Error('不支持的小程序平台');
  }

  // 本地存储适配
  static setStorage(key, data) {
    const currentPlatform = this.getCurrentPlatform();
    return new Promise((resolve, reject) => {
      currentPlatform.setStorage({
        key: key,
        data: typeof data === 'object' ? JSON.stringify(data) : data,
        success: resolve,
        fail: reject
      });
    });
  }

  static getStorage(key) {
    const currentPlatform = this.getCurrentPlatform();
    return new Promise((resolve, reject) => {
      currentPlatform.getStorage({
        key: key,
        success: (res) => {
          try {
            // 尝试解析JSON
            const data = JSON.parse(res.data);
            resolve(data);
          } catch (e) {
            // 如果不是JSON，直接返回
            resolve(res.data);
          }
        },
        fail: reject
      });
    });
  }

  // 显示提示信息
  static showToast(title, icon = 'none', duration = 2000) {
    const currentPlatform = this.getCurrentPlatform();
    currentPlatform.showToast({
      title: title,
      icon: icon,
      duration: duration
    });
  }

  // 显示加载中
  static showLoading(title = '加载中...') {
    const currentPlatform = this.getCurrentPlatform();
    currentPlatform.showLoading({
      title: title,
      mask: true
    });
  }

  static hideLoading() {
    const currentPlatform = this.getCurrentPlatform();
    currentPlatform.hideLoading();
  }
}

// 高频功能API封装（小程序版本）
class MiniProgramAPI {
  static const BASE_URL = 'http://bot-manager:5000';
  static const WS_URL = 'ws://bot-manager:3005';

  // 获取机器人状态
  static async getBotStatus() {
    try {
      MiniProgramAdapter.showLoading('获取状态中...');
      const data = await MiniProgramAdapter.request({
        url: `${BASE_URL}/api/bots`,
        method: 'GET',
        timeout: 3000
      });
      MiniProgramAdapter.hideLoading();
      return data;
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      console.error('获取机器人状态失败:', error);
      MiniProgramAdapter.showToast('获取状态失败，请检查网络');
      return { bots: [], error: error.message };
    }
  }

  // 获取系统状态
  static async getSystemStatus() {
    try {
      const data = await MiniProgramAdapter.request({
        url: `${BASE_URL}/api/stats`,
        method: 'GET',
        timeout: 3000
      });
      return data;
    } catch (error) {
      console.error('获取系统状态失败:', error);
      return { cpu_usage: 0, memory_usage: 0, error: error.message };
    }
  }

  // 切换机器人状态
  static async toggleBot(botId, enable) {
    try {
      MiniProgramAdapter.showLoading('操作中...');
      await MiniProgramAdapter.request({
        url: `${BASE_URL}/api/bot/toggle`,
        method: 'POST',
        data: { bot_id: botId, enable: enable },
        timeout: 5000
      });
      MiniProgramAdapter.hideLoading();
      MiniProgramAdapter.showToast(enable ? '机器人已启动' : '机器人已停止');
      return true;
    } catch (error) {
      MiniProgramAdapter.hideLoading();
      console.error('机器人操作失败:', error);
      MiniProgramAdapter.showToast('操作失败，请重试');
      return false;
    }
  }

  // 连接WebSocket
  static connectWebSocket(onMessage) {
    MiniProgramAdapter.connectWebSocket(WS_URL)
      .then(socketTask => {
        console.log('WebSocket连接成功');
        
        socketTask.onMessage((message) => {
          try {
            const data = JSON.parse(message.data);
            onMessage(data);
          } catch (error) {
            console.error('消息解析失败:', error);
          }
        });

        socketTask.onOpen(() => {
          console.log('WebSocket已打开');
        });

        socketTask.onClose(() => {
          console.log('WebSocket已关闭');
          // 3秒后重连
          setTimeout(() => {
            this.connectWebSocket(onMessage);
          }, 3000);
        });

        socketTask.onError((error) => {
          console.error('WebSocket错误:', error);
          // 3秒后重连
          setTimeout(() => {
            this.connectWebSocket(onMessage);
          }, 3000);
        });

        return socketTask;
      })
      .catch(error => {
        console.error('WebSocket连接失败:', error);
        // 3秒后重连
        setTimeout(() => {
          this.connectWebSocket(onMessage);
        }, 3000);
      });
  }
}