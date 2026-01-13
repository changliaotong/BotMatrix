import axios from 'axios';

const getBaseURL = () => {
  const envBaseURL = import.meta.env.VITE_API_BASE_URL;
  if (envBaseURL) return envBaseURL;
  
  // 如果没有配置环境变量，则根据当前访问的域名动态生成
  // 默认假设后端在 5000 端口
  if (typeof window !== 'undefined') {
    const { protocol, hostname } = window.location;
    // 如果是开发环境且是通过 IP 访问，或者是非 localhost 访问
    if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
      return `${protocol}//${hostname}:5000`;
    }
  }
  return '/';
};

const api = axios.create({
  baseURL: getBaseURL(),
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
