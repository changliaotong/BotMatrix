import { createRouter, createWebHistory, RouterView } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useSystemStore } from '@/stores/system';
import { t } from '@/utils/i18n';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // --- EarlyMeow (Standalone) ---
    {
      path: '/meow',
      component: () => import('@/views/portal/bots/earlymeow/Layout.vue'),
      meta: { layout: 'blank', title: 'title.meow' },
      children: [
        {
          path: '',
          name: 'early-meow-home',
          component: () => import('@/views/portal/bots/earlymeow/pages/Home.vue'),
        },
        {
          path: 'tech',
          name: 'early-meow-tech',
          component: () => import('@/views/portal/bots/earlymeow/pages/Tech.vue'),
        },
        {
          path: 'pricing',
          name: 'early-meow-pricing',
          component: () => import('@/views/portal/bots/earlymeow/pages/Pricing.vue'),
        },
        {
          path: 'ecosystem',
          name: 'early-meow-ecosystem',
          component: () => import('@/views/portal/bots/earlymeow/pages/Ecosystem.vue'),
        },
        {
          path: 'console',
          name: 'early-meow-console',
          component: () => import('@/views/portal/bots/earlymeow/Console.vue'),
          meta: { requiresAuth: true }
        }
      ]
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/portal/Home.vue'),
      meta: { layout: 'blank', title: 'title.home' }
    },
    {
      path: '/matrix',
      redirect: '/'
    },
    {
      path: '/bots/early-meow',
      redirect: '/meow'
    },
    {
      path: '/bots/nexus-guard',
      name: 'bot-nexus-guard',
      component: () => import('@/views/portal/bots/NexusGuard.vue'),
      meta: { layout: 'blank', title: 'title.home' }
    },
    {
      path: '/bots/digital-employee',
      component: () => import('@/views/portal/bots/digital-employee/Layout.vue'),
      meta: { layout: 'blank', title: 'title.digital_employee' },
      children: [
        {
          path: '',
          name: 'bot-digital-employee',
          component: () => import('@/views/portal/bots/digital-employee/pages/Home.vue'),
        }
      ]
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('@/views/portal/About.vue'),
      meta: { layout: 'blank', title: 'title.about' }
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('@/views/portal/Pricing.vue'),
      meta: { layout: 'blank', title: 'title.pricing' }
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('@/views/portal/Docs.vue'),
      meta: { layout: 'blank', title: 'title.docs' }
    },
    {
      path: '/docs/:id',
      name: 'docs-detail',
      component: () => import('@/views/portal/DocsDetail.vue'),
      meta: { layout: 'blank', title: 'title.docs' }
    },
    {
      path: '/news',
      name: 'news',
      component: () => import('@/views/portal/News.vue'),
      meta: { layout: 'blank', title: 'title.news' }
    },
    {
      path: '/news/:id',
      name: 'news-detail',
      component: () => import('@/views/portal/NewsDetail.vue'),
      meta: { layout: 'blank', title: 'title.news' }
    },
    {
      path: '/industrial-test',
      name: 'industrial-test',
      component: () => import('@/views/portal/IndustrialTest.vue'),
      meta: { layout: 'blank', title: 'title.test' }
    },

    // --- Portal Setup ---
    {
      path: '/setup',
      component: RouterView,
      meta: { requiresAuth: true, layout: 'blank', title: 'title.setup' },
      children: [
        {
          path: 'bot',
          name: 'portal-bot-setup',
          component: () => import('@/views/portal/setup/BotSetup.vue'),
        },
        {
          path: 'group',
          name: 'portal-group-setup',
          component: () => import('@/views/portal/setup/GroupSetup.vue'),
        }
      ]
    },

    // --- Auth ---
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { layout: 'blank', title: 'title.login' }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/Register.vue'),
      meta: { layout: 'blank', title: 'title.register' }
    },
    {
      path: '/auth/token-login',
      name: 'token-login',
      component: () => import('@/views/auth/TokenLogin.vue'),
      meta: { layout: 'blank', title: 'title.login' }
    },

    // --- User Console ---
    {
      path: '/console',
      component: RouterView,
      meta: { requiresAuth: true, title: 'title.console' },
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
      meta: { requiresAuth: true, requiresAdmin: true, title: 'title.admin' },
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
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else if (to.hash) {
      return {
        el: to.hash,
        behavior: 'smooth',
        top: 80, // Offset for fixed header
      };
    } else {
      return { top: 0 };
    }
  },
});

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  const systemStore = useSystemStore();
  const token = localStorage.getItem('wxbot_token');
  
  // Check if the route or any of its parents require authentication
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  const isPublic = !requiresAuth;

  console.log(`[Router] Navigating to: ${String(to.name || to.path)}, hasToken: ${!!token}, isPublic: ${isPublic}`);

  if (requiresAuth) {
    if (!token) {
      console.log('[Router] No token found, redirecting to login');
      authStore.logout(); // Ensure state is cleared
      next({ 
        name: 'login',
        query: { redirect: to.fullPath }
      });
    } else {
      // Check auth if user info is missing
      if (!authStore.user) {
        try {
          const isValid = await authStore.checkAuth();
          if (!isValid) {
            console.log('[Router] Auth check failed, redirecting to login');
            authStore.logout();
            next({ 
              name: 'login',
              query: { redirect: to.fullPath }
            });
            return;
          }
        } catch (err) {
          console.error('[Router] Auth check error:', err);
          authStore.logout();
          next({ 
            name: 'login',
            query: { redirect: to.fullPath }
          });
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
      const redirect = to.query.redirect as string;
      if (redirect) {
        next(redirect);
      } else {
        // Default landing pages
        if (authStore.isAdmin) {
          next({ name: 'console-dashboard' });
        } else {
          next({ name: 'console-bot-setup' });
        }
      }
    } else {
      next();
    }
  }
});

export default router;
