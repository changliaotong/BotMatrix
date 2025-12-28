<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  BookOpen, 
  Search, 
  ChevronRight, 
  ExternalLink,
  MessageSquare,
  Zap,
  Shield,
  Cpu,
  Globe,
  HelpCircle
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const manualContent = ref('');
const loading = ref(true);
const search = ref('');
const selectedSection = ref(t('system_overview'));

const fetchManual = async () => {
  loading.value = true;
  try {
    const data = await botStore.fetchManual();
    if (data.success && data.content) {
      manualContent.value = data.content;
    }
  } finally {
    loading.value = false;
  }
};

onMounted(fetchManual);

const selectSection = (item: string) => {
  selectedSection.value = item;
};

const categories = computed(() => [
  {
    title: t('getting_started'),
    icon: Zap,
    items: [t('system_overview'), t('installation_guide'), t('first_bot_setup')]
  },
  {
    title: t('core_features'),
    icon: Cpu,
    items: [t('worker_mgmt'), t('routing_rules'), t('docker_integration')]
  },
  {
    title: t('security_access'),
    icon: Shield,
    items: [t('user_roles'), t('api_auth'), t('permission_matrix')]
  },
  {
    title: t('advanced'),
    icon: Globe,
    items: [t('nexus_gateway'), t('fission_campaigns'), t('custom_tasks')]
  }
]);
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('manual') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('manual_desc') }}</p>
      </div>
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <input 
          v-model="search"
          type="text"
          :placeholder="t('search_docs')"
          class="pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] w-64 transition-all"
        />
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
      <!-- Sidebar Navigation -->
      <div class="lg:col-span-1 space-y-6">
        <div v-for="cat in categories" :key="cat.title" class="space-y-3">
          <div class="flex items-center gap-2 px-2">
            <component :is="cat.icon" class="w-4 h-4 text-[var(--matrix-color)]" />
            <span class="text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest">{{ cat.title }}</span>
          </div>
          <div class="space-y-1">
            <button 
              v-for="item in cat.items" 
              :key="item"
              @click="selectSection(item)"
              :class="selectedSection === item ? 'text-[var(--matrix-color)] bg-[var(--matrix-color)]/5' : 'text-[var(--text-muted)] hover:text-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/5'"
              class="w-full text-left px-4 py-2.5 rounded-xl text-xs font-bold transition-all flex items-center justify-between group"
            >
              {{ item }}
              <ChevronRight class="w-3 h-3 transition-all" :class="selectedSection === item ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0'" />
            </button>
          </div>
        </div>

        <div class="p-6 rounded-3xl bg-[var(--matrix-color)] text-black mt-8">
          <HelpCircle class="w-8 h-8 mb-4" />
          <h3 class="font-black text-sm uppercase tracking-tight">{{ t('need_help') }}</h3>
          <p class="text-[10px] font-bold uppercase tracking-widest mt-1 opacity-80">{{ t('join_community_desc') }}</p>
          <button class="w-full mt-4 py-3 rounded-xl bg-black text-white text-[10px] font-black uppercase tracking-widest hover:bg-black/80 transition-all flex items-center justify-center gap-2">
            <MessageSquare class="w-3 h-3" />
            {{ t('join_discord') }}
          </button>
        </div>
      </div>

      <!-- Main Content -->
      <div class="lg:col-span-3">
        <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] p-8 md:p-12 min-h-[600px] relative overflow-hidden">
          <div v-if="loading" class="space-y-6 animate-pulse">
            <div class="h-8 w-1/3 bg-black/5 dark:bg-white/5 rounded-xl"></div>
            <div class="space-y-3">
              <div class="h-4 w-full bg-black/5 dark:bg-white/5 rounded-lg"></div>
              <div class="h-4 w-full bg-black/5 dark:bg-white/5 rounded-lg"></div>
              <div class="h-4 w-2/3 bg-black/5 dark:bg-white/5 rounded-lg"></div>
            </div>
            <div class="h-64 w-full bg-black/5 dark:bg-white/5 rounded-[2rem]"></div>
          </div>

          <div v-else class="prose prose-invert max-w-none">
            <div v-if="manualContent" v-html="manualContent"></div>
            <div v-else-if="selectedSection === t('system_overview')" class="space-y-8">
              <div class="flex items-center gap-4 text-[var(--matrix-color)]">
                <BookOpen class="w-10 h-10" />
                <h2 class="text-3xl font-black text-[var(--text-main)] uppercase tracking-tight m-0">{{ t('system_overview') }}</h2>
              </div>
              
              <div class="p-8 rounded-[2rem] bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                <h3 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight mt-0">{{ t('architecture') }}</h3>
                <p class="text-[var(--text-muted)] font-medium leading-relaxed mt-4">
                  {{ t('architecture_desc') }}
                </p>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
                  <div class="p-6 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <div class="text-[10px] font-black text-[var(--matrix-color)] uppercase tracking-widest mb-2">{{ t('centralized') }}</div>
                    <div class="text-sm font-black text-[var(--text-main)] uppercase">{{ t('nexus_gateway') }}</div>
                    <p class="text-xs text-[var(--text-muted)] font-medium mt-2">{{ t('nexus_gateway_desc') }}</p>
                  </div>
                  <div class="p-6 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <div class="text-[10px] font-black text-[var(--matrix-color)] uppercase tracking-widest mb-2">{{ t('distributed') }}</div>
                    <div class="text-sm font-black text-[var(--text-main)] uppercase">{{ t('workers') }}</div>
                    <p class="text-xs text-[var(--text-muted)] font-medium mt-2">{{ t('workers_desc_manual') }}</p>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="selectedSection === t('installation_guide')" class="space-y-8">
              <div class="flex items-center gap-4 text-[var(--matrix-color)]">
                <Zap class="w-10 h-10" />
                <h2 class="text-3xl font-black text-[var(--text-main)] uppercase tracking-tight m-0">{{ t('installation_guide') }}</h2>
              </div>
              <div class="space-y-6">
                <div class="p-6 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <h3 class="text-lg font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('prerequisites') }}</h3>
                  <ul class="list-disc list-inside mt-4 space-y-2 text-sm text-[var(--text-muted)] font-medium">
                    <li>{{ t('prerequisite_docker') }}</li>
                    <li>{{ t('prerequisite_go') }}</li>
                    <li>{{ t('prerequisite_node') }}</li>
                  </ul>
                </div>
                <div class="p-6 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] font-mono text-xs">
                  <p class="text-[var(--matrix-color)] mb-2"># {{ t('clone_and_start') }}</p>
                  <p class="text-[var(--text-main)]">git clone https://github.com/{{ t('botmatrix') }}/{{ t('botmatrix') }}.git</p>
                  <p class="text-[var(--text-main)]">cd {{ t('botmatrix') }}</p>
                  <p class="text-[var(--text-main)]">docker-compose up -d</p>
                </div>
              </div>
            </div>

            <div v-else class="flex flex-col items-center justify-center py-20 text-center">
              <HelpCircle class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
              <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ selectedSection }}</h2>
              <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2 max-w-md">
                {{ t('coming_soon_desc') }}
              </p>
              <button class="mt-8 px-8 py-3 rounded-2xl bg-[var(--matrix-color)] text-black text-[10px] font-black uppercase tracking-widest hover:scale-105 transition-all">
                {{ t('visit_wiki') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.prose) {
  color: var(--text-muted);
}
:deep(.prose h1), :deep(.prose h2), :deep(.prose h3) {
  color: var(--text-main);
  text-transform: uppercase;
  letter-spacing: -0.025em;
  font-weight: 900;
}
:deep(.prose strong) {
  color: var(--text-main);
}
:deep(.prose a) {
  color: var(--matrix-color);
  text-decoration: none;
}
</style>
