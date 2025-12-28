import { defineStore } from 'pinia';
import api from '@/api';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('wxbot_token') || null,
    role: localStorage.getItem('wxbot_role') || 'user',
    user: null as any,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.role === 'admin' || state.role === 'super',
  },
  actions: {
    setToken(token: string) {
      this.token = token;
      localStorage.setItem('wxbot_token', token);
    },
    setRole(role: string) {
      this.role = role;
      localStorage.setItem('wxbot_role', role);
    },
    async checkAuth() {
      if (!this.token) return false;
      try {
        const { data } = await api.get('/api/me');
        if (data && data.user) {
          this.user = data.user;
          const role = data.user.is_admin ? (this.role === 'super' ? 'super' : 'admin') : 'user';
          this.setRole(role);
          return true;
        }
        return false;
      } catch (error) {
        this.logout();
        return false;
      }
    },
    logout() {
      this.token = null;
      this.role = 'user';
      this.user = null;
      localStorage.removeItem('wxbot_token');
      localStorage.removeItem('wxbot_role');
    },
    async login(username: string, password: string) {
      try {
        const { data } = await api.post('/api/login', { username, password });
        if (data.success && data.token) {
          this.setToken(data.token);
          this.setRole(data.role || 'user');
          if (data.user) {
            this.user = data.user;
          }
          return true;
        }
        return false;
      } catch (error: any) {
        console.error('Login failed:', error);
        throw new Error(error.response?.data?.message || 'login_failed');
      }
    },
    async loginWithMagicToken(token: string) {
      try {
        const { data } = await api.post('/api/login/magic', { token });
        this.setToken(data.token);
        this.setRole(data.role);
        return true;
      } catch (error) {
        return false;
      }
    }
  },
});
