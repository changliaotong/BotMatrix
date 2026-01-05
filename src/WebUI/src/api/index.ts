import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/',
  timeout: 10000,
});

api.interceptors.request.use((config) => {
  // Skip token for login and public endpoints
  const isPublic = config.url?.includes('/api/login');
  
  if (!isPublic) {
    const token = localStorage.getItem('wxbot_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  
  // 添加当前语言到请求头
  const lang = localStorage.getItem('wxbot_lang') || 'zh-CN';
  config.headers['Accept-Language'] = lang;
  
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Don't clear localStorage here, just trigger the event
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
    return Promise.reject(error);
  }
);

export default api;
