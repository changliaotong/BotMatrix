<script setup lang="ts">
import { 
  Heart, 
  Sparkles, 
  History, 
  ArrowLeft,
  ChevronRight,
  Bot,
  Zap,
  Shield,
  MessageSquare
} from 'lucide-vue-next';
import { useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';

const router = useRouter();
const systemStore = useSystemStore();

const t = (key: string, def?: string) => systemStore.t(key) || def || key;

const features = [
  {
    title: 'guide_angel.original_intention_title',
    desc: 'guide_angel.original_intention_desc',
    icon: Heart,
    color: 'text-pink-500'
  },
  {
    title: 'guide_angel.brand_upgrade_title',
    desc: 'guide_angel.brand_upgrade_desc',
    icon: Sparkles,
    color: 'text-amber-500'
  },
  {
    title: 'guide_angel.nostalgia_moment_title',
    desc: 'guide_angel.nostalgia_moment_desc',
    icon: History,
    color: 'text-blue-500'
  }
];
</script>

<template>
  <div class="relative w-full min-h-full pt-32">
    <!-- Background Effects (Contained) -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-pink-500/5 blur-[120px] rounded-full"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 blur-[120px] rounded-full"></div>
    </div>

    <div class="relative z-10 p-6 md:p-12 max-w-7xl mx-auto">
      <!-- Header / Breadcrumb -->
      <div class="flex items-center gap-2 mb-12 text-[var(--text-muted)]">
        <router-link to="/meow" class="flex items-center gap-1 hover:text-[var(--text-main)] transition-colors">
          <ArrowLeft class="w-4 h-4" />
          <span class="text-xs font-black uppercase tracking-widest">{{ t('common.back') }}</span>
        </router-link>
        <span class="opacity-20">/</span>
        <div class="flex items-center gap-2">
          <Heart class="w-3 h-3 text-pink-500" />
          <span class="text-xs font-black uppercase tracking-widest text-[var(--text-main)]">{{ t('guide_angel.nostalgia') }}</span>
        </div>
      </div>

      <!-- Hero Section -->
      <div class="text-center mb-16">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-pink-500/10 text-pink-500 text-[10px] font-black uppercase tracking-[0.2em] mb-6 border border-pink-500/20">
          <Sparkles class="w-3 h-3" />
          {{ t('guide_angel.legacy_memory') }}
        </div>
        
        <h1 class="text-4xl md:text-6xl font-black text-[var(--text-main)] tracking-tighter uppercase italic mb-6">
          {{ t('guide_angel.title') }} <span class="text-[var(--text-muted)] opacity-30">/</span> <span class="text-pink-500">{{ t('guide_angel.subtitle') }}</span>
        </h1>
        
        <p class="text-base md:text-lg text-[var(--text-muted)] font-medium max-w-3xl mx-auto leading-relaxed mb-10 px-4">
          {{ t('guide_angel.description') }}
        </p>

        <div class="flex flex-wrap justify-center gap-4">
          <router-link to="/meow" class="px-8 py-4 bg-pink-500 text-white rounded-2xl font-black text-sm uppercase tracking-widest hover:bg-pink-600 transition-all shadow-xl shadow-pink-500/20 group">
            <span class="flex items-center gap-3">
              {{ t('guide_angel.enter_earlymeow') }}
              <ChevronRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </span>
          </router-link>
          <a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="px-8 py-4 bg-[var(--bg-card)] border border-[var(--border-color)] text-[var(--text-main)] rounded-2xl font-black text-sm uppercase tracking-widest hover:border-pink-500/50 transition-all">
            {{ t('guide_angel.learn_origin') }}
          </a>
        </div>
      </div>

      <!-- Feature Grid -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8 mb-24">
        <div v-for="f in features" :key="f.title" class="p-8 rounded-[2.5rem] bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-pink-500/30 transition-all group relative overflow-hidden">
          <div class="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.08] transition-opacity">
            <component :is="f.icon" class="w-32 h-32" />
          </div>
          
          <div class="w-12 h-12 rounded-2xl bg-[var(--bg-body)] border border-[var(--border-color)] flex items-center justify-center mb-6 group-hover:scale-110 transition-transform" :class="f.color">
            <component :is="f.icon" class="w-6 h-6" />
          </div>
          
          <h3 class="text-xl font-black text-[var(--text-main)] uppercase italic mb-4">{{ t(f.title) }}</h3>
          <p class="text-[var(--text-muted)] font-bold leading-relaxed">{{ t(f.desc) }}</p>
        </div>
      </div>

      <!-- Timeline/Story -->
      <div class="p-12 rounded-[3rem] bg-gradient-to-br from-pink-500/5 to-blue-500/5 border border-[var(--border-color)] relative overflow-hidden">
        <div class="flex flex-col md:flex-row gap-12 items-center">
          <div class="flex-1">
            <h2 class="text-3xl font-black text-[var(--text-main)] uppercase italic mb-6">{{ t('guide_angel.story_title') }}</h2>
            <div class="space-y-6 text-[var(--text-muted)] font-bold">
              <p>{{ t('guide_angel.story_p1') }}</p>
              <p>{{ t('guide_angel.story_p2') }}</p>
              <p>{{ t('guide_angel.story_p3') }}</p>
            </div>
          </div>
          <div class="w-full md:w-72 aspect-square rounded-[2rem] bg-pink-500/10 border border-pink-500/20 flex flex-col items-center justify-center p-8 text-center group">
            <div class="w-24 h-24 rounded-full bg-white flex items-center justify-center mb-6 shadow-2xl group-hover:scale-110 transition-transform duration-500">
              <Bot class="w-12 h-12 text-pink-500" />
            </div>
            <div class="text-xs font-black uppercase tracking-[0.3em] text-pink-500 mb-2">{{ t('guide_angel.legacy_bot') }}</div>
            <div class="text-xl font-black text-[var(--text-main)] italic">{{ t('guide_angel.title') }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Message -->
    <footer class="py-12 text-center border-t border-[var(--border-color)]">
      <div class="flex items-center justify-center gap-2 text-[var(--text-muted)] text-[10px] font-black uppercase tracking-[0.4em]">
        <span>{{ t('guide_angel.footer_copyright') }}</span>
        <div class="w-1 h-1 rounded-full bg-pink-500"></div>
        <span>{{ t('guide_angel.now_earlymeow') }}</span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.italic {
  font-style: italic;
}
</style> 
