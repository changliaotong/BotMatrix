<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 overflow-x-hidden" :class="[systemStore.style]">
    <PortalHeader />

    <!-- 1. HERO SECTION -->
    <header class="relative pt-48 pb-32 px-6 overflow-hidden">
      <!-- Ambient Background -->
      <div class="absolute inset-0 pointer-events-none">
        <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.15))_0%,transparent_70%)] opacity-20"></div>
        <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
      </div>

      <div class="container mx-auto max-w-7xl relative z-10">
        <div class="flex flex-col items-center text-center">
          <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold mb-10 animate-pulse shadow-[var(--matrix-glow)]">
            <Sparkles class="w-4 h-4" />
            {{ tt('botmatrix_hero_subtitle') }}
          </div>

          <h1 class="text-6xl md:text-9xl font-black mb-12 tracking-tighter leading-none" :style="{ fontFamily: systemStore.style === 'matrix' ? 'var(--font-main)' : 'inherit' }">
            {{ tt('botmatrix_hero_title') }}<br/>
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-[var(--matrix-color)] to-blue-500">
              {{ tt('botmatrix_hero_title_accent') }}
            </span>
          </h1>

          <p class="text-xl md:text-2xl text-[var(--text-muted)] mb-16 max-w-4xl mx-auto font-light leading-relaxed">
            {{ tt('botmatrix_hero_desc') }}
          </p>

          <div class="flex flex-col sm:flex-row gap-6 justify-center items-center">
            <button 
              @click="router.push(authStore.isAuthenticated ? '/console' : '/login')"
              class="px-10 py-5 bg-[var(--matrix-color)] hover:opacity-90 text-white rounded-2xl text-xl font-bold transition-all shadow-[var(--matrix-glow)] flex items-center gap-3 group"
            >
              {{ tt('earlymeow_hero_init_link') }} <ArrowRight class="w-6 h-6 group-hover:translate-x-2 transition-transform" />
            </button>
            <button 
              @click="router.push('/docs')"
              class="px-10 py-5 bg-[var(--bg-card)] hover:bg-[var(--bg-body)] text-[var(--text-main)] rounded-2xl text-xl font-bold transition-all border border-[var(--border-color)] backdrop-blur-sm"
            >
              {{ tt('earlymeow_hero_read_docs') }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- 2. CORE FEATURES (GRID) -->
    <section class="py-32 relative">
      <div class="container mx-auto max-w-7xl px-6">
        <div class="grid lg:grid-cols-3 gap-8">
          <div 
            v-for="feat in architectureFeatures" 
            :key="feat.title"
            class="group p-10 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[3rem] hover:border-[var(--matrix-color)]/30 transition-all hover:-translate-y-2 shadow-sm"
          >
            <div class="w-16 h-16 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] flex items-center justify-center mb-8 group-hover:bg-[var(--matrix-color)] group-hover:text-white transition-all">
              <component :is="feat.icon" class="w-8 h-8" />
            </div>
            <h3 class="text-2xl font-bold mb-4 uppercase tracking-tight">{{ feat.title }}</h3>
            <p class="text-[var(--text-muted)] leading-relaxed">{{ feat.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 3. FEATURED: EARLYMEOW -->
    <section class="py-32 bg-[var(--bg-card)]/30">
      <div class="container mx-auto max-w-7xl px-6">
        <div class="bg-gradient-to-br from-[var(--bg-card)] to-[var(--matrix-color)]/5 rounded-[4rem] p-12 md:p-24 border border-[var(--border-color)] relative overflow-hidden group">
          <div class="absolute top-0 right-0 w-full h-full bg-[radial-gradient(circle_at_top_right,var(--matrix-color),transparent_50%)] opacity-5"></div>
          
          <div class="flex flex-col lg:flex-row gap-20 items-center relative z-10">
            <div class="flex-1 space-y-10">
              <div class="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-bold uppercase tracking-widest">
                <Cat class="w-4 h-4" /> {{ tt('botmatrix_home_flagship') }}
              </div>
              <h2 class="text-5xl md:text-7xl font-black italic uppercase tracking-tighter leading-tight text-[var(--text-main)]">
                {{ tt('botmatrix_showcase_earlymeow_title') }}
              </h2>
              <p class="text-xl text-[var(--text-muted)] leading-relaxed">
                {{ tt('botmatrix_showcase_earlymeow_desc') }}
              </p>
              <router-link 
                to="/meow" 
                class="inline-flex items-center gap-4 px-8 py-4 bg-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/80 text-white font-bold rounded-2xl transition-all shadow-[0_0_30px_rgba(var(--matrix-color-rgb),0.2)]"
              >
                {{ tt('botmatrix_home_explore_meow') }} <ArrowRight class="w-5 h-5" />
              </router-link>
            </div>
            
            <div class="flex-1 relative">
              <div class="w-80 h-80 bg-[var(--bg-body)] border-2 border-[var(--matrix-color)]/30 rounded-[3rem] flex items-center justify-center relative group-hover:scale-105 transition-transform duration-700">
                <div class="absolute inset-0 bg-[var(--matrix-color)]/5 blur-[60px] rounded-full animate-pulse"></div>
                <Cat class="w-40 h-40 text-[var(--matrix-color)] drop-shadow-[0_0_30px_rgba(var(--matrix-color-rgb),0.4)]" />
                
                <!-- Orbiting elements -->
                <div class="absolute -top-6 -right-6 p-4 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl shadow-2xl animate-bounce-slow">
                  <Zap class="w-6 h-6 text-[var(--matrix-color)]" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 4. WORKFORCE & PERFORMANCE -->
    <section class="py-40 relative">
      <div class="container mx-auto max-w-7xl px-6">
        <div class="text-center mb-24">
          <h2 class="text-5xl md:text-7xl font-black mb-8 italic uppercase tracking-tighter">{{ tt('botmatrix_workforce_title') }}</h2>
          <p class="text-xl text-[var(--text-muted)] max-w-3xl mx-auto font-light leading-relaxed">
            {{ tt('botmatrix_workforce_subtitle') }}
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div 
            v-for="feat in workforceFeatures" 
            :key="feat.title"
            class="p-10 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2.5rem] hover:border-[var(--matrix-color)]/30 transition-all"
          >
            <div class="flex items-center gap-4 mb-8">
              <div class="w-12 h-12 rounded-xl bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)]">
                <component :is="feat.icon" class="w-6 h-6" />
              </div>
              <h4 class="text-xl font-bold uppercase tracking-wide">{{ feat.title }}</h4>
            </div>
            <p class="text-[var(--text-muted)] leading-relaxed">{{ feat.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 5. REAL-TIME STATS -->
    <section class="py-40 bg-[var(--bg-card)]/30 overflow-hidden">
      <div class="container mx-auto max-w-7xl px-6">
        <div class="grid lg:grid-cols-3 gap-12">
          <div 
            v-for="stat in monitorStats" 
            :key="stat.title"
            class="flex flex-col items-center text-center p-12 bg-[var(--bg-body)]/50 border border-[var(--border-color)] rounded-[3rem] relative group"
          >
            <div class="w-20 h-20 rounded-full bg-[var(--bg-card)] flex items-center justify-center mb-8 border border-[var(--border-color)] group-hover:border-[var(--matrix-color)]/50 transition-all">
              <component :is="stat.icon" class="w-8 h-8" :style="{ color: stat.color }" />
            </div>
            <div class="text-5xl font-black mb-4 tracking-tighter" :style="{ color: stat.color }">{{ stat.value }}</div>
            <h3 class="text-xl font-bold mb-4 uppercase tracking-tight text-[var(--text-main)]">{{ stat.title }}</h3>
            <p class="text-[var(--text-muted)] text-sm leading-relaxed">{{ stat.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 6. CALL TO ACTION -->
    <section class="py-48 text-center px-6 relative">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,var(--matrix-color),transparent_70%)] opacity-5"></div>
      <h2 class="text-6xl md:text-9xl font-black mb-12 tracking-tighter leading-none italic uppercase">
        {{ tt('botmatrix_home_ready_title') }}
      </h2>
      <button 
        @click="router.push('/login')"
        class="px-12 py-6 bg-gradient-to-r from-[var(--matrix-color)] to-blue-600 hover:opacity-90 text-white rounded-3xl text-2xl font-black transition-all shadow-2xl shadow-[var(--matrix-color)]/40"
      >
        {{ tt('earlymeow_hero_init_link') }}
      </button>
    </section>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useSystemStore } from '../../stores/system';
import { useAuthStore } from '../../stores/auth';
import { useI18n } from '../../utils/i18n';
import PortalHeader from '../../components/layout/PortalHeader.vue';
import PortalFooter from '../../components/layout/PortalFooter.vue';
import { 
  ArrowRight, 
  Cpu, 
  Zap, 
  Network, 
  Sparkles,
  Cat,
  Box,
  BrainCircuit,
  Route,
  Share2,
  Cloud,
  Workflow,
  BarChart3,
  Globe2,
  Gauge
} from 'lucide-vue-next';

const router = useRouter();
const systemStore = useSystemStore();
const authStore = useAuthStore();
const { t: tt } = useI18n();

const architectureFeatures = computed(() => [
  {
    title: tt('botmatrix_arch_onebot_title'),
    desc: tt('botmatrix_arch_onebot_desc'),
    icon: Share2,
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_arch_proxy_title'),
    desc: tt('botmatrix_arch_proxy_desc'),
    icon: Cloud,
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_arch_router_title'),
    desc: tt('botmatrix_arch_router_desc'),
    icon: Route,
    color: 'var(--matrix-color)'
  }
]);

const workforceFeatures = computed(() => [
  {
    title: tt('botmatrix_workforce_ai_title'),
    desc: tt('botmatrix_workforce_ai_desc'),
    icon: BrainCircuit,
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_workforce_cluster_title'),
    desc: tt('botmatrix_workforce_cluster_desc'),
    icon: Box,
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_workforce_automation_title'),
    desc: tt('botmatrix_workforce_automation_desc'),
    icon: Workflow,
    color: 'var(--matrix-color)'
  }
]);

const monitorStats = computed(() => [
  {
    title: tt('botmatrix_monitor_rtt_title'),
    desc: tt('botmatrix_monitor_rtt_desc'),
    icon: Gauge,
    value: '12ms',
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_monitor_throughput_title'),
    desc: tt('botmatrix_monitor_throughput_desc'),
    icon: BarChart3,
    value: '1.2 PB/s',
    color: 'var(--matrix-color)'
  },
  {
    title: tt('botmatrix_monitor_cluster_title'),
    desc: tt('botmatrix_monitor_cluster_desc'),
    icon: Globe2,
    value: '99.99%',
    color: 'var(--matrix-color)'
  }
]);
</script>

<style scoped>
@keyframes bounce-slow {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
.animate-bounce-slow {
  animation: bounce-slow 4s ease-in-out infinite;
}

@keyframes spin-slow {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
.animate-spin-slow {
  animation: spin-slow 12s linear infinite;
}

.perspective-2000 {
  perspective: 2000px;
}
</style>
