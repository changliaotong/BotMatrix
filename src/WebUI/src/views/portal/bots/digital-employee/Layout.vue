<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from '@/utils/i18n';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { useSystemStore } from '@/stores/system';
import { 
  Home, Users, Zap, Target, ShieldCheck, Sparkles, ChevronRight
} from 'lucide-vue-next';

const route = useRoute();
const { t: tt } = useI18n();
const systemStore = useSystemStore();

const navLinks = computed(() => [
  { name: tt('portal.digital_employee.nav.home'), path: '#hero', icon: Home },
  { name: tt('portal.digital_employee.nav.vision'), path: '#vision', icon: Target },
  { name: tt('portal.digital_employee.nav.revolution'), path: '#revolution', icon: Zap },
  { name: tt('portal.digital_employee.nav.kpi'), path: '#kpi', icon: Users },
  { name: tt('portal.digital_employee.nav.security'), path: '#security', icon: ShieldCheck },
]);

const scrollTo = (id: string) => {
  const el = document.querySelector(id);
  if (el) {
    const headerOffset = 80;
    const elementPosition = el.getBoundingClientRect().top;
    const offsetPosition = elementPosition + window.pageYOffset - headerOffset;

    window.scrollTo({
      top: offsetPosition,
      behavior: 'smooth'
    });
  }
};
</script>

<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 overflow-x-hidden" :class="[systemStore.style]">
    <!-- Unified Background -->
    <div class="fixed inset-0 pointer-events-none z-0">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.05))_0%,transparent_70%)] opacity-20"></div>
      <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
    </div>

    <PortalHeader />
    
    <!-- Sub-navigation for Digital Employee -->
    <nav 
      class="fixed top-[73px] w-full z-[40] transition-all duration-300 border-b border-[var(--border-color)] bg-[var(--bg-body)]/50 backdrop-blur-md hidden md:block"
    >
      <div class="max-w-7xl mx-auto px-6 h-12 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-[10px] font-black uppercase tracking-widest text-[var(--matrix-color)]">{{ tt('common.digital_employee') }}</span>
          <ChevronRight class="w-3 h-3 text-[var(--text-muted)]" />
        </div>
        <div class="flex items-center gap-8">
          <button 
            v-for="link in navLinks" 
            :key="link.path"
            @click="scrollTo(link.path)"
            class="text-[11px] font-bold uppercase tracking-wider transition-colors hover:text-[var(--matrix-color)]"
            :class="route.hash === link.path ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'"
          >
            {{ link.name }}
          </button>
        </div>
      </div>
    </nav>

    <main>
      <router-view v-slot="{ Component }">
        <component :is="Component" />
      </router-view>
    </main>

    <PortalFooter />
  </div>
</template>

<style scoped>
.translate-y-[-10px] {
  transform: translateY(-10px);
}
</style>
