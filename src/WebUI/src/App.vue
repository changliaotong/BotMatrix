<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import Header from './components/layout/Header.vue';
import GlobalChatWindow from './components/layout/GlobalChatWindow.vue';
import Sidebar from './components/layout/Sidebar.vue';
import MatrixRain from './components/common/MatrixRain.vue';

const route = useRoute();
const router = useRouter();
const systemStore = useSystemStore();
const authStore = useAuthStore();
const isInitializing = ref(true);

const isBlankLayout = computed(() => {
  return route.meta.layout === 'blank' || route.matched.some(record => record.meta.layout === 'blank');
});

const isDarkMode = computed(() => systemStore.mode === 'dark');

const updateHtmlClasses = () => {
  const html = document.documentElement;
  // Always handle dark mode class
  if (isDarkMode.value) {
    html.classList.add('dark');
  } else {
    html.classList.remove('dark');
  }

  // Handle style classes
  const styles = ['classic', 'matrix'];
  html.classList.remove(...styles);
  html.classList.add(systemStore.style);
};

const toggleTheme = () => {
  systemStore.toggleMode();
};

import { watch } from 'vue';
import { t } from '@/utils/i18n';

// Watch for layout or system style changes
watch([isBlankLayout, () => systemStore.style, isDarkMode], () => {
  updateHtmlClasses();
}, { immediate: true });

// Update page title
watch([() => route.meta.title, () => systemStore.lang], ([titleKey]) => {
  const key = (titleKey as string) || 'title.home';
  document.title = t(key);
}, { immediate: true });

const handleUnauthorized = () => {
  authStore.logout();
  
  // Only redirect if the current route requires authentication
  if (route.meta.requiresAuth) {
    router.push('/login');
  }
};

onMounted(async () => {
  systemStore.initTheme();
  
  // Wait for router to be ready to ensure route.meta is populated
  await router.isReady();

  try {
    // If we have a token but no user info, try to fetch it
    if (authStore.token && !authStore.user) {
      await authStore.checkAuth();
    }
  } catch (err) {
    console.error('Initial auth check failed:', err);
  } finally {
    isInitializing.value = false;
  }
  window.addEventListener('auth:unauthorized', handleUnauthorized);
});

onUnmounted(() => {
  window.removeEventListener('auth:unauthorized', handleUnauthorized);
});
</script>

<template>
  <div 
    class="min-h-screen transition-colors duration-300 bg-[var(--bg-body)] text-[var(--text-main)] font-[var(--font-main)]"
    :style="{ fontSize: 'var(--font-size-base)' }"
  >
    <!-- Style-specific background effects -->
    <MatrixRain v-if="systemStore.style === 'matrix'" />
    
    <div v-if="isInitializing" class="fixed inset-0 z-[100] flex items-center justify-center bg-[var(--bg-body)]">
      <div class="flex flex-col items-center gap-6">
        <div class="relative w-24 h-24">
          <div class="absolute inset-0 border-8 border-[var(--matrix-color)]/10 rounded-full" :class="{ 'border-white/5': isBlankLayout }"></div>
          <div class="absolute inset-0 border-8 border-t-[var(--matrix-color)] rounded-full animate-spin" :class="{ 'border-t-cyber-neon': isBlankLayout }"></div>
          <div class="absolute inset-4 border-4 border-b-[var(--matrix-color)]/40 rounded-full animate-spin-slow" :class="{ 'border-b-cyber-pink/40': isBlankLayout }"></div>
        </div>
        <div class="text-center space-y-2">
          <h2 class="text-2xl font-black text-[var(--text-main)] tracking-tighter uppercase italic" :class="{ 'text-white': isBlankLayout && isDarkMode, 'text-meow-dark': isBlankLayout && !isDarkMode }">
            <template v-if="isBlankLayout">EARLY<span class="text-cyber-neon">MEOW</span></template>
            <template v-else>BOT<span class="text-[var(--matrix-color)]">NEXUS</span></template>
          </h2>
          <p class="font-bold text-[10px] uppercase tracking-[0.4em] animate-pulse" :class="isBlankLayout ? 'text-cyber-neon' : 'text-[var(--matrix-color)]'">
            {{ isBlankLayout ? 'Initializing Meow Kernel...' : 'Loading System...' }}
          </p>
        </div>
      </div>
    </div>

    <template v-else-if="isBlankLayout">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </template>
    <template v-else>
      <div class="flex h-screen overflow-hidden">
        <Sidebar />
        <div class="flex flex-col flex-1 overflow-hidden relative">
          <Header />
          <GlobalChatWindow />
          <main class="flex-1 overflow-y-auto p-2 md:p-4 bg-[var(--bg-body)] transition-colors duration-300">
            <router-view v-slot="{ Component }">
              <transition name="fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </main>
        </div>
      </div>
    </template>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>
