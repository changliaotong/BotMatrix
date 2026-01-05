<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import Header from './components/layout/Header.vue';
import GlobalChatWindow from './components/layout/GlobalChatWindow.vue';
import Sidebar from './components/layout/Sidebar.vue';
import KawaiiSparkles from './components/common/KawaiiSparkles.vue';
import MatrixRain from './components/common/MatrixRain.vue';

const route = useRoute();
const router = useRouter();
const systemStore = useSystemStore();
const authStore = useAuthStore();
const isInitializing = ref(true);

const isBlankLayout = computed(() => route.meta.layout === 'blank');

const handleUnauthorized = () => {
  authStore.logout();
  if (route.name !== 'login') {
    router.push('/login');
  }
};

onMounted(async () => {
  systemStore.initTheme();
  try {
    if (authStore.token) {
      // Use checkAuth but don't automatically logout if it fails due to network/500
      const isAuthed = await authStore.checkAuth();
      if (!isAuthed) {
        // Only redirect if we're sure the session is invalid (token was cleared by checkAuth)
        if (!authStore.token && route.name !== 'login') {
          handleUnauthorized();
        }
      } else if (route.name === 'login') {
        router.push('/');
      }
    } else if (route.name !== 'login') {
      router.push('/login');
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
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] font-[var(--font-main)]" :style="{ fontSize: 'var(--font-size-base)' }">
    <!-- Style-specific background effects -->
    <MatrixRain v-if="systemStore.style === 'matrix'" />
    <KawaiiSparkles v-if="systemStore.style === 'kawaii'" />
    
    <div v-if="isInitializing" class="fixed inset-0 z-[100] flex items-center justify-center bg-[var(--bg-body)]">
      <div class="flex flex-col items-center gap-6">
        <div class="relative w-24 h-24">
          <div class="absolute inset-0 border-8 border-[var(--matrix-color)]/10 rounded-full"></div>
          <div class="absolute inset-0 border-8 border-t-[var(--matrix-color)] rounded-full animate-spin"></div>
          <div class="absolute inset-4 border-4 border-b-[var(--matrix-color)]/40 rounded-full animate-spin-slow"></div>
        </div>
        <div class="text-center space-y-2">
          <h2 class="text-2xl font-black text-[var(--text-main)] tracking-tighter uppercase italic">BOT<span class="text-[var(--matrix-color)]">MATRIX</span></h2>
          <p class="text-[var(--matrix-color)] font-bold text-[10px] uppercase tracking-[0.4em] animate-pulse">Initializing Neural Link...</p>
        </div>
      </div>
    </div>

    <template v-else-if="isBlankLayout">
      <router-view />
    </template>
    <template v-else>
      <div class="flex h-screen overflow-hidden">
        <Sidebar />
        <div class="flex flex-col flex-1 overflow-hidden relative">
          <Header />
      <GlobalChatWindow />
      <main class="flex-1 overflow-y-auto p-2 lg:p-4 bg-[var(--bg-body)] transition-colors duration-300">
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
