<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 relative overflow-x-hidden" :class="[systemStore.style]">
    <!-- Unified Background -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.05))_0%,transparent_70%)] opacity-20"></div>
      <div class="absolute inset-0 bg-[grid-slate-800] [mask-image:radial-gradient(white,transparent)] opacity-20"></div>
    </div>

    <PortalHeader />

    <!-- Hero -->
    <header class="pt-48 pb-24 px-6 relative overflow-hidden">
      <div class="max-w-5xl mx-auto text-center">
        <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 text-[var(--matrix-color)] text-xs font-medium mb-10">
          <Newspaper class="w-4 h-4" /> {{ tt('portal.news.badge') }}
        </div>
        <h1 class="text-6xl md:text-8xl font-black mb-12 tracking-tighter leading-none uppercase" v-html="tt('portal.news.title')"></h1>
        <p class="text-[var(--text-muted)] max-w-3xl mx-auto text-xl leading-relaxed font-light">
          {{ tt('portal.news.subtitle') }}
        </p>
      </div>
    </header>

    <!-- News List -->
    <section class="py-24 pb-48 px-6">
      <div class="max-w-5xl mx-auto space-y-16">
        <article v-for="item in newsItems" :key="item.id" class="group relative bg-[var(--bg-body)]/50 rounded-[3rem] border border-[var(--border-color)] overflow-hidden hover:bg-[var(--bg-body)]/80 hover:border-[var(--matrix-color)]/30 transition-all duration-500">
          <div class="flex flex-col lg:flex-row items-stretch">
            <div class="lg:w-2/5 aspect-video lg:aspect-auto bg-[var(--bg-body)] relative overflow-hidden border-b lg:border-b-0 lg:border-r border-[var(--border-color)]">
              <div class="absolute inset-0 bg-gradient-to-br from-[var(--matrix-color)]/10 to-[var(--matrix-color)]/90 opacity-10 group-hover:scale-110 transition-transform duration-700"></div>
              <div class="absolute inset-0 flex items-center justify-center">
                <component :is="item.icon" class="w-24 h-24 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]/20 transition-colors duration-500" />
              </div>
            </div>
            <div class="flex-1 p-12 lg:p-16 flex flex-col justify-center">
              <div class="flex items-center gap-4 mb-8">
                <span class="px-4 py-1.5 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold uppercase tracking-widest border border-[var(--matrix-color)]/20">{{ item.category }}</span>
                <span class="text-[var(--text-muted)] text-[10px] font-bold uppercase tracking-widest">{{ item.date }}</span>
              </div>
              <h2 class="text-3xl font-bold mb-6 uppercase tracking-tight text-[var(--text-main)] group-hover:text-[var(--matrix-color)] transition-colors duration-500 cursor-pointer" @click="$router.push(`/news/${item.id}`)">{{ item.title }}</h2>
              <p class="text-[var(--text-muted)] mb-10 text-lg font-light leading-relaxed">{{ item.summary }}</p>
              <router-link :to="`/news/${item.id}`" class="flex items-center gap-3 text-[var(--text-muted)] font-bold uppercase tracking-[0.2em] text-[10px] hover:text-[var(--matrix-color)] hover:gap-6 transition-all duration-500">
                {{ tt('portal.news.view_detail') }} <ArrowRight class="w-5 h-5" />
              </router-link>
            </div>
          </div>
        </article>
      </div>
    </section>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { useSystemStore } from '@/stores/system';
import { 
  ArrowRight, 
  Newspaper, 
  Zap, 
  Rocket, 
  Cpu, 
  ShieldCheck 
} from 'lucide-vue-next';
import { useI18n } from '@/utils/i18n';
import { computed } from 'vue';
import { newsContent } from '@/locales/news_content';

const { t: tt, locale } = useI18n();
const systemStore = useSystemStore();

const newsItems = computed(() => {
  const lang = locale.value || 'zh-CN';
  const content = newsContent[lang] || newsContent['zh-CN'];
  
  // Map icons based on category or ID
  const icons = [Rocket, ShieldCheck, Zap, Rocket]; // Fallback icons
  
  return Object.entries(content)
    .map(([id, item]) => ({
      id: Number(id),
      ...item,
      icon: id === '4' ? Rocket : (id === '3' ? Zap : (id === '2' ? ShieldCheck : Cpu))
    }))
    .sort((a, b) => b.id - a.id); // Newest (highest ID) first
});
</script>
