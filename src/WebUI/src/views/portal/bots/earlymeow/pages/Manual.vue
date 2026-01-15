<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { 
  BookOpen, 
  Search, 
  ChevronRight, 
  ArrowLeft,
  HelpCircle,
  MessageSquare,
  Sparkles
} from 'lucide-vue-next';
import { useBotStore } from '@/stores/bot';
import { useSystemStore } from '@/stores/system';
import { useI18n } from '@/utils/i18n';

const botStore = useBotStore();
const systemStore = useSystemStore();
const { t } = useI18n();

const manualData = ref<any>(null);
const loading = ref(true);
const search = ref('');
const selectedSectionIndex = ref(0);

const fetchManual = async () => {
  loading.value = true;
  try {
    const data = await botStore.fetchManual();
    console.log('Manual data fetched:', data);
    if (data.success && data.data) {
      manualData.value = data.data;
    } else {
      console.warn('Manual data fetch success but no data:', data);
    }
  } catch (err) {
    console.error('Failed to fetch manual:', err);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchManual);

const filteredSections = computed(() => {
  if (!manualData.value?.sections) return [];
  if (!search.value) return manualData.value.sections;
  
  const query = search.value.toLowerCase();
  return manualData.value.sections.filter((s: any) => 
    s.title.toLowerCase().includes(query) || 
    s.content.toLowerCase().includes(query)
  );
});

const currentSection = computed(() => {
  return filteredSections.value[selectedSectionIndex.value] || null;
});

const selectSection = (index: number) => {
  console.log('Selecting section index:', index);
  selectedSectionIndex.value = index;
};

// Simple Markdown-like to HTML converter
const renderContent = (content: string) => {
  if (!content) return '';
  
  let html = content
    .replace(/^### (.*$)/gim, '<h3 class="text-xl font-black text-[var(--matrix-color)] mt-8 mb-4 uppercase tracking-tight italic">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 class="text-2xl font-black text-[var(--text-main)] mt-10 mb-6 uppercase tracking-tight italic">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 class="text-3xl font-black text-[var(--text-main)] mt-12 mb-8 uppercase tracking-tight italic">$1</h1>')
    .replace(/^\d+\. (.*$)/gim, '<li class="ml-4 mb-2 text-[var(--text-muted)] font-medium list-decimal">$1</li>')
    .replace(/^\* (.*$)/gim, '<li class="ml-4 mb-2 text-[var(--text-muted)] font-medium list-disc">$1</li>')
    .replace(/^- (.*$)/gim, '<li class="ml-4 mb-2 text-[var(--text-muted)] font-medium list-disc">$1</li>')
    .replace(/\*\*(.*?)\*\*/g, '<strong class="text-[var(--text-main)] font-black">$1</strong>')
    .replace(/`(.*?)`/g, '<code class="bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] px-1.5 py-0.5 rounded font-mono text-xs">$1</code>')
    .replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2" target="_blank" class="text-[var(--matrix-color)] hover:underline font-bold">$1</a>')
    .replace(/\$(.*?)\$/g, '<span class="font-mono bg-[var(--bg-body)] px-1 rounded text-[var(--text-main)] border border-[var(--border-color)]">$1</span>')
    .replace(/\n\n/g, '<div class="mb-4"></div>')
    .replace(/\n/g, '<br>');
    
  return html;
};
</script>

<template>
  <div class="relative w-full min-h-full pt-12">
    <!-- Background Effects -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-[var(--matrix-color)]/5 blur-[120px] rounded-full"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 blur-[120px] rounded-full"></div>
    </div>

    <div class="relative z-10 p-6 md:p-12 max-w-7xl mx-auto">
      <!-- Hero Section -->
      <div class="mb-16">
        <h1 class="text-4xl md:text-6xl font-black text-[var(--text-main)] tracking-tighter uppercase italic mb-6">
          {{ t('earlymeow.nav.manual') }}
        </h1>
        
        <div class="relative max-w-xl">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)]" />
          <input 
            v-model="search"
            type="text"
            :placeholder="t('search_docs')"
            class="w-full pl-12 pr-6 py-4 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-sm font-bold text-[var(--text-main)] transition-all shadow-xl"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
        <!-- Sidebar Navigation -->
        <div class="lg:col-span-1 space-y-2">
          <div v-if="loading" class="space-y-2">
            <div v-for="i in 6" :key="i" class="h-12 bg-[var(--bg-card)] rounded-xl animate-pulse"></div>
          </div>
          <template v-else>
            <button 
              v-for="(section, index) in filteredSections" 
              :key="index"
              @click="selectSection(index)"
              :class="selectedSectionIndex === index ? 'text-[var(--matrix-color)] bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/30' : 'text-[var(--text-muted)] hover:text-[var(--text-main)] hover:bg-[var(--bg-card)] border-transparent'"
              class="w-full text-left px-5 py-3 rounded-xl text-xs font-black uppercase tracking-widest transition-all flex items-center justify-between group border"
            >
              {{ section.title }}
              <ChevronRight class="w-4 h-4 transition-all" :class="selectedSectionIndex === index ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0'" />
            </button>
          </template>

          <div class="p-8 rounded-[2rem] bg-[var(--matrix-color)] text-black mt-12 relative overflow-hidden">
            <div class="relative z-10">
              <HelpCircle class="w-10 h-10 mb-4" />
              <h3 class="font-black text-lg uppercase tracking-tight leading-none mb-2">NEED HELP?</h3>
              <p class="text-[10px] font-bold uppercase tracking-widest opacity-80 mb-6">JOIN OUR COMMUNITY FOR SUPPORT</p>
              <button class="w-full py-4 rounded-2xl bg-black text-white text-[10px] font-black uppercase tracking-widest hover:scale-105 transition-all flex items-center justify-center gap-2">
                <MessageSquare class="w-4 h-4" />
                JOIN DISCORD
              </button>
            </div>
            <div class="absolute -bottom-4 -right-4 opacity-10">
              <HelpCircle class="w-32 h-32" />
            </div>
          </div>
        </div>

        <!-- Main Content -->
        <div class="lg:col-span-3">
          <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[3rem] p-8 md:p-16 min-h-[600px] relative overflow-hidden shadow-2xl">
            <div v-if="loading" class="space-y-8 animate-pulse">
              <div class="h-12 w-1/2 bg-[var(--bg-body)] rounded-2xl"></div>
              <div class="space-y-4">
                <div class="h-4 w-full bg-[var(--bg-body)] rounded-lg"></div>
                <div class="h-4 w-full bg-[var(--bg-body)] rounded-lg"></div>
                <div class="h-4 w-3/4 bg-[var(--bg-body)] rounded-lg"></div>
              </div>
              <div class="h-64 w-full bg-[var(--bg-body)] rounded-[2rem]"></div>
            </div>

            <div v-else-if="currentSection" class="relative z-10">
              <h2 class="text-3xl md:text-5xl font-black text-[var(--text-main)] uppercase tracking-tighter italic mb-12 border-b border-[var(--border-color)] pb-8">
                {{ currentSection.title }}
              </h2>
              
              <div 
                class="text-[var(--text-muted)] font-medium leading-relaxed prose prose-invert max-w-none"
                v-html="renderContent(currentSection.content)"
              ></div>
            </div>

            <div v-else class="flex flex-col items-center justify-center py-32 text-center">
              <div class="w-24 h-24 rounded-full bg-[var(--bg-body)] flex items-center justify-center mb-8 border border-[var(--border-color)]">
                <Search class="w-10 h-10 text-[var(--text-muted)] opacity-20" />
              </div>
              <h2 class="text-2xl font-black text-[var(--text-main)] uppercase tracking-tight">NOT FOUND</h2>
              <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-4 max-w-sm">
                TRY ANOTHER KEYWORD OR SELECT A SECTION FROM THE SIDEBAR.
              </p>
            </div>

            <!-- Decorative element -->
            <div class="absolute top-0 right-0 p-12 opacity-[0.02] pointer-events-none">
              <BookOpen class="w-64 h-64" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Any additional custom styles can go here */
</style>
