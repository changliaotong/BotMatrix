import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from '@/views/Dashboard.vue';

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

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('wxbot_token');
  if (to.name !== 'login' && !token) {
    next({ name: 'login' });
  } else if (to.name === 'login' && token) {
    next({ name: 'dashboard' });
  } else {
    next();
  }
});

export default router;
