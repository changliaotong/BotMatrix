import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from '@/views/Dashboard.vue';
import { useAuthStore } from '@/stores/auth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard,
    },
    {
      path: '/bots',
      name: 'bots',
      component: () => import('@/views/Bots.vue'),
    },
    {
      path: '/workers',
      name: 'workers',
      component: () => import('@/views/Workers.vue'),
    },
    {
      path: '/plugins',
      name: 'plugins',
      component: () => import('@/views/Plugins.vue'),
    },
    {
      path: '/contacts',
      name: 'contacts',
      component: () => import('@/views/Contacts.vue'),
    },
    {
      path: '/messages',
      name: 'messages',
      component: () => import('@/views/Messages.vue'),
    },
    {
      path: '/nexus',
      name: 'nexus',
      component: () => import('@/views/Nexus.vue'),
    },
    {
      path: '/ai',
      name: 'ai',
      component: () => import('@/views/NexusAI.vue'),
    },
    {
      path: '/visualization',
      name: 'visualization',
      component: () => import('@/views/Nexus.vue'),
    },
    {
      path: '/tasks',
      name: 'tasks',
      component: () => import('@/views/Tasks.vue'),
    },
    {
      path: '/fission',
      name: 'fission',
      component: () => import('@/views/Fission.vue'),
    },
    {
      path: '/docker',
      name: 'docker',
      component: () => import('@/views/Docker.vue'),
    },
    {
      path: '/routing',
      name: 'routing',
      component: () => import('@/views/Routing.vue'),
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('@/views/Users.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/Settings.vue'),
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/Logs.vue'),
    },
    {
      path: '/manual',
      name: 'manual',
      component: () => import('@/views/Manual.vue'),
    },
    {
      path: '/monitor',
      name: 'monitor',
      component: () => import('@/views/Monitor.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { layout: 'blank' }
    }
  ],
});

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  const token = localStorage.getItem('wxbot_token');
  
  console.log(`[Router] Navigating to: ${String(to.name)}, hasToken: ${!!token}, hasUser: ${!!authStore.user}`);

  if (to.name !== 'login') {
    if (!token) {
      console.log('[Router] No token found, redirecting to login');
      next({ name: 'login' });
    } else {
      // 如果没有用户信息，尝试验证 token
      if (!authStore.user) {
        try {
          console.log('[Router] Token exists but no user info, checking auth...');
          const isValid = await authStore.checkAuth();
          console.log(`[Router] Auth check result: ${isValid}, hasToken after check: ${!!authStore.token}`);
          
          // 只有在明确验证失败（token被清除）的情况下才跳转登录
          if (!isValid && !authStore.token) {
            console.log('[Router] Auth check failed and token cleared, redirecting to login');
            next({ name: 'login' });
            return;
          }
        } catch (err) {
          console.error('[Router] Router auth check error:', err);
          // 网络错误等情况不强制跳转，让页面尝试加载
        }
      }
      next();
    }
  } else {
    // 如果已登录且访问登录页，重定向到首页
    if (token && authStore.user) {
      console.log('[Router] Already logged in, redirecting to dashboard');
      next({ name: 'dashboard' });
    } else {
      next();
    }
  }
});

export default router;
