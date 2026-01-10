import { createRouter, createWebHistory, RouterView } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // --- Portal / Official Website ---
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/portal/bots/EarlyMeow.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/matrix',
      name: 'matrix-home',
      component: () => import('@/views/portal/Home.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/bots/early-meow',
      redirect: '/'
    },
    {
      path: '/bots/nexus-guard',
      name: 'bot-nexus-guard',
      component: () => import('@/views/portal/bots/NexusGuard.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/bots/digital-employee',
      name: 'bot-digital-employee',
      component: () => import('@/views/portal/bots/DigitalEmployee.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('@/views/portal/About.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('@/views/portal/Pricing.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('@/views/portal/Docs.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/news',
      name: 'news',
      component: () => import('@/views/portal/News.vue'),
      meta: { layout: 'blank' }
    },

    // --- Auth ---
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/Register.vue'),
      meta: { layout: 'blank' }
    },
    {
      path: '/auth/token-login',
      name: 'token-login',
      component: () => import('@/views/auth/TokenLogin.vue'),
      meta: { layout: 'blank' }
    },

    // --- User Console ---
    {
      path: '/console',
      component: RouterView,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'console-dashboard',
          component: () => import('@/views/console/Dashboard.vue'),
        },
        {
          path: 'bots',
          name: 'console-bots',
          component: () => import('@/views/console/Bots.vue'),
        },
        {
          path: 'contacts',
          name: 'console-contacts',
          component: () => import('@/views/console/Contacts.vue'),
        },
        {
          path: 'messages',
          name: 'console-messages',
          component: () => import('@/views/console/Messages.vue'),
        },
        {
          path: 'tasks',
          name: 'console-tasks',
          component: () => import('@/views/console/Tasks.vue'),
        },
        {
          path: 'fission',
          name: 'console-fission',
          component: () => import('@/views/console/Fission.vue'),
        },
        {
          path: 'manual',
          name: 'console-manual',
          component: () => import('@/views/console/Manual.vue'),
        },
        {
          path: 'settings',
          name: 'console-settings',
          component: () => import('@/views/console/Settings.vue'),
        }
      ]
    },

    // --- Admin Dashboard ---
    {
      path: '/admin',
      component: RouterView,
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: 'workers',
          name: 'admin-workers',
          component: () => import('@/views/admin/Workers.vue'),
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/views/admin/Users.vue'),
        },
        {
          path: 'logs',
          name: 'admin-logs',
          component: () => import('@/views/admin/Logs.vue'),
        },
        {
          path: 'monitor',
          name: 'admin-monitor',
          component: () => import('@/views/admin/Monitor.vue'),
        },
        {
          path: 'nexus',
          name: 'admin-nexus',
          component: () => import('@/views/admin/Nexus.vue'),
        },
        {
          path: 'ai',
          name: 'admin-ai',
          component: () => import('@/views/admin/NexusAI.vue'),
        },
        {
          path: 'routing',
          name: 'admin-routing',
          component: () => import('@/views/admin/Routing.vue'),
        },
        {
          path: 'docker',
          name: 'admin-docker',
          component: () => import('@/views/admin/Docker.vue'),
        },
        {
          path: 'plugins',
          name: 'admin-plugins',
          component: () => import('@/views/admin/Plugins.vue'),
        }
      ]
    },

    // --- Fallback ---
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ],
});

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  const token = localStorage.getItem('wxbot_token');
  
  // Check if the route or any of its parents require authentication
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  const isPublic = !requiresAuth;

  console.log(`[Router] Navigating to: ${String(to.name || to.path)}, hasToken: ${!!token}, isPublic: ${isPublic}`);

  if (requiresAuth) {
    if (!token) {
      console.log('[Router] No token found, redirecting to login');
      authStore.logout(); // Ensure state is cleared
      next({ name: 'login' });
    } else {
      // Check auth if user info is missing
      if (!authStore.user) {
        try {
          const isValid = await authStore.checkAuth();
          if (!isValid) {
            console.log('[Router] Auth check failed, redirecting to login');
            authStore.logout();
            next({ name: 'login' });
            return;
          }
        } catch (err) {
          console.error('[Router] Auth check error:', err);
          authStore.logout();
          next({ name: 'login' });
          return;
        }
      }

      // Check Admin permissions for admin routes
      const requiresAdmin = to.matched.some(record => record.meta.requiresAdmin);
      if (requiresAdmin && !authStore.isAdmin) {
        console.warn('[Router] Access denied: Admin role required');
        next({ name: 'console-dashboard' });
        return;
      }

      next();
    }
  } else {
    // Public paths
    // If already logged in and visiting login/register, redirect to console
    if (token && (to.name === 'login' || to.name === 'register')) {
      next({ name: 'console-dashboard' });
    } else {
      next();
    }
  }
});

export default router;
