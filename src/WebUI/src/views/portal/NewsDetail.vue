<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 relative overflow-x-hidden" :class="[systemStore.style]">
    <!-- Unified Background -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,var(--matrix-color,rgba(168,85,247,0.05))_0%,transparent_70%)] opacity-20"></div>
    </div>

    <PortalHeader />

    <main class="pt-48 pb-24 px-6 relative z-10">
      <div class="max-w-3xl mx-auto">
        <!-- Back Button -->
        <router-link 
          to="/news" 
          class="inline-flex items-center gap-2 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors mb-12 group"
        >
          <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
          <span class="text-xs font-bold uppercase tracking-widest">{{ tt('common.back') }}</span>
        </router-link>

        <template v-if="article">
          <div class="flex items-center gap-4 mb-8">
            <span class="px-4 py-1.5 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold uppercase tracking-widest border border-[var(--matrix-color)]/20">
              {{ article.category }}
            </span>
            <span class="text-[var(--text-muted)] text-[10px] font-bold uppercase tracking-widest">
              {{ article.date }}
            </span>
          </div>

          <h1 class="text-4xl md:text-6xl font-black mb-12 tracking-tighter leading-tight uppercase">
            {{ article.title }}
          </h1>

          <div class="prose prose-invert max-w-none news-content" v-html="article.content"></div>
        </template>

        <div v-else class="py-24 text-center">
          <h2 class="text-2xl font-bold text-[var(--text-muted)] mb-8">{{ tt('common.not_found') }}</h2>
          <router-link 
            to="/news" 
            class="px-8 py-4 bg-[var(--matrix-color)] text-white rounded-xl font-bold transition-all"
          >
            {{ tt('common.back_to_list') }}
          </router-link>
        </div>
      </div>
    </main>

    <PortalFooter />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { ArrowLeft } from 'lucide-vue-next';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { useSystemStore } from '@/stores/system';
import { useI18n } from '@/utils/i18n';
import { newsContent } from '@/locales/news_content';

const route = useRoute();
const { t: tt, locale } = useI18n();
const systemStore = useSystemStore();

const articleId = computed(() => Number(route.params.id));

const article = computed(() => {
  const lang = locale.value || 'zh-CN';
  const content = newsContent[lang] || newsContent['zh-CN'];
  return content[articleId.value];
});
</script>

<style scoped>
@reference "tailwindcss";

.news-content :deep(h3) {
  @apply text-2xl font-bold mt-12 mb-6 text-[var(--text-main)];
}

.news-content :deep(p) {
  @apply text-lg leading-relaxed text-[var(--text-muted)] mb-6 font-light;
}

.news-content :deep(ul) {
  @apply list-disc list-inside mb-8 space-y-4 text-[var(--text-muted)];
}

.news-content :deep(li strong) {
  @apply text-[var(--matrix-color)];
}
</style>
