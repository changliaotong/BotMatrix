import axios from 'axios';

const api = axios.create({
  baseURL: '/',
  timeout: 10000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('wxbot_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('wxbot_token');
      localStorage.removeItem('wxbot_role');
      // Trigger global event or redirect to login page
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
    return Promise.reject(error);
  }
);

export default api;
