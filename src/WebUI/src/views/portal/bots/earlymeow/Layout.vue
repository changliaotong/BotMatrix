<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { useMeowStore } from '@/stores/earlymeow';
import { useI18n } from '@/utils/i18n';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  Home, LayoutDashboard, 
  Layers, Compass, DollarSign, ChevronRight, Sparkles,
  BookOpen
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const systemStore = useSystemStore();
const authStore = useAuthStore();
const meowStore = useMeowStore();
const { t } = useI18n();

const isScrolled = ref(false);

const handleScroll = () => {
  isScrolled.value = window.scrollY > 20;
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
  meowStore.init();
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});

const navLinks = computed(() => [
  { name: t('earlymeow.nav.home'), path: '/', icon: Home },
  { name: t('earlymeow.nav.guide_angel'), path: '/guide-angel', icon: Sparkles },
  { name: t('common.digital_employee'), path: '/digital-employee', icon: LayoutDashboard },
  { name: t('earlymeow.nav.manual'), path: '/manual', icon: BookOpen },
  { name: t('earlymeow.nav.tech'), path: '/tech', icon: Layers },
  { name: t('earlymeow.nav.ecosystem'), path: '/ecosystem', icon: Compass },
  { name: t('earlymeow.nav.pricing'), path: '/pricing', icon: DollarSign },
]);
</script>

<template>
  <div 
    class="min-h-screen transition-colors duration-500 selection:bg-[var(--matrix-color)]/30 selection:text-[var(--text-main)] overflow-x-hidden bg-[var(--bg-body)] text-[var(--text-main)]"
    :class="[systemStore.style]"
  >
    <PortalHeader />
    
    <!-- Background grid -->
    <div class="fixed inset-0 z-0 opacity-20 pointer-events-none" 
      :style="{ backgroundImage: 'radial-gradient(circle at 2px 2px, var(--matrix-color,rgba(168,85,247,0.15)) 1px, transparent 0)', backgroundSize: '40px 40px' }">
    </div>
    
    <!-- Main Content -->
    <main class="relative z-10">
      <router-view v-slot="{ Component }">
        <transition 
          enter-active-class="transition duration-500 ease-out"
          enter-from-class="opacity-0 translate-y-8"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-300 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-8"
          mode="out-in"
        >
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <PortalFooter />
  </div>
</template>

<style scoped>
/* Removed old styles, using Tailwind */
</style>
