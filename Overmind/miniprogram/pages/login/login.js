// 登录页面逻辑
const MiniProgramAPI = require('../../utils/miniprogram_api.js');

Page({
  data: {
    username: '',
    password: '',
    isLoading: false,
    errorMessage: '',
    debugMode: true // 调试模式，显示测试账号信息
  },

  onLoad() {
    console.log('登录页面加载');
    // 检查是否已有token
    this.checkExistingToken();
  },

  // 检查是否已有有效的token
  async checkExistingToken() {
    try {
      const token = MiniProgramAPI.getStorageSync('token');
      if (token) {
        const result = await MiniProgramAPI.validateToken(token);
        if (result.valid) {
          // 已有有效token，直接跳转到首页
          this.switchToHomePage();
        }
      }
    } catch (error) {
      console.error('检查token失败:', error);
    }
  },

  // 用户名输入事件
  onUsernameInput(e) {
    this.setData({
      username: e.detail.value
    });
  },

  // 密码输入事件
  onPasswordInput(e) {
    this.setData({
      password: e.detail.value
    });
  },

  // 登录按钮点击事件
  async onLogin() {
    // 表单验证
    if (!this.validateForm()) {
      return;
    }

    try {
      this.setData({
        isLoading: true,
        errorMessage: ''
      });

      // 调用登录API
      const result = await MiniProgramAPI.login(this.data.username, this.data.password);

      if (result.success) {
        // 登录成功，保存用户信息到全局
        const app = getApp();
        const userData = result.data.user || {};
        app.globalData.userInfo = {
          username: this.data.username,
          role: userData.role || 'user'
        };

        // 跳转到首页
        this.switchToHomePage();
      } else {
        // 登录失败
        this.setData({
          errorMessage: result.error || '登录失败，请检查用户名和密码'
        });
      }
    } catch (error) {
      console.error('登录错误:', error);
      this.setData({
        errorMessage: '网络错误，请稍后重试'
      });
    } finally {
      this.setData({
        isLoading: false
      });
    }
  },

  // 表单验证
  validateForm() {
    if (!this.data.username.trim()) {
      this.setData({
        errorMessage: '请输入用户名'
      });
      return false;
    }

    if (!this.data.password.trim()) {
      this.setData({
        errorMessage: '请输入密码'
      });
      return false;
    }

    return true;
  },

  // 跳转到首页
  switchToHomePage() {
    wx.switchTab({
      url: '/pages/index/index',
      success: () => {
        console.log('登录成功，跳转到首页');
      },
      fail: (error) => {
        console.error('页面跳转失败:', error);
      }
    });
  },

  // 调试功能：自动填充测试账号
  onDebugFill(e) {
    if (this.data.debugMode) {
      this.setData({
        username: 'admin',
        password: 'admin123'
      });
    }
  }
});