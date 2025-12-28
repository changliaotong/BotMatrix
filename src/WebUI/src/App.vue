<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import Header from './components/layout/Header.vue';
import Sidebar from './components/layout/Sidebar.vue';
import KawaiiSparkles from './components/common/KawaiiSparkles.vue';
import MatrixRain from './components/common/MatrixRain.vue';

const route = useRoute();
const systemStore = useSystemStore();
const authStore = useAuthStore();

const isBlankLayout = computed(() => route.meta.layout === 'blank');

onMounted(() => {
  systemStore.initTheme();
  authStore.checkAuth();
});
</script>

<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] font-[var(--font-main)]" :style="{ fontSize: 'var(--font-size-base)' }">
    <!-- Style-specific background effects -->
    <MatrixRain v-if="systemStore.style === 'matrix'" />
    <KawaiiSparkles v-if="systemStore.style === 'kawaii'" />
    
    <template v-if="isBlankLayout">
      <router-view />
    </template>
    <template v-else>
      <div class="flex h-screen overflow-hidden">
        <Sidebar />
        <div class="flex flex-col flex-1 overflow-hidden relative">
          <Header />
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
