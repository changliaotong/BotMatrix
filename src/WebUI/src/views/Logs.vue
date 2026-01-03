<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useBotStore } from '@/stores/bot';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { 
  Search, 
  Filter, 
  Trash2, 
  RefreshCw,
  Download,
  Bot,
  AlertCircle,
  Info,
  AlertTriangle,
  Terminal,
  Activity
} from 'lucide-vue-next';

const botStore = useBotStore();
const systemStore = useSystemStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);

const logs = ref<any[]>([]);
const loading = ref(false);
const autoRefresh = ref(true);
let refreshTimer: number | null = null;

// Filters
const filterSource = ref('all');
const filterLevel = ref('all');
const searchQuery = ref('');

const fetchLogs = async () => {
  if (loading.value || !authStore.isAdmin) return;
  loading.value = true;
  try {
    const params: any = {};
    if (filterSource.value !== 'all') params.bot_id = filterSource.value;
    if (filterLevel.value !== 'all') params.level = filterLevel.value;
    if (searchQuery.value) params.search = searchQuery.value;

    const data = await botStore.fetchSystemLogs(params);
    if (data.success && data.data) {
      logs.value = data.data.logs || [];
    }
  } catch (err) {
    console.error('Failed to fetch logs:', err);
  } finally {
    loading.value = false;
  }
};

const clearLogs = async () => {
  if (!confirm(t('confirm_clear_logs'))) return;
  try {
    const data = await botStore.clearSystemLogs();
    if (data.success) {
      logs.value = [];
    }
  } catch (err) {
    console.error('Failed to clear logs:', err);
  }
};

const downloadLogs = () => {
  const content = logs.value.map(l => `[${l.time}] [${l.level}] ${l.source ? `[@${l.source}] ` : ''}${l.message}`).join('\n');
  const blob = new Blob([content], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `botmatrix-logs-${new Date().toISOString().split('T')[0]}.log`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
};

onMounted(async () => {
  await botStore.fetchBots();
  fetchLogs();
  
  refreshTimer = window.setInterval(() => {
    if (autoRefresh.value && !loading.value) {
      fetchLogs();
    }
  }, 5000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

const getLevelColor = (level: string) => {
  switch (level) {
    case 'ERROR': return 'text-red-500 bg-red-500/10';
    case 'WARN': return 'text-yellow-500 bg-yellow-500/10';
    case 'INFO': return 'text-[var(--matrix-color)] bg-[var(--matrix-color)]/10';
    default: return 'text-[var(--text-muted)] bg-[var(--text-muted)]/10';
  }
};

const getLevelIcon = (level: string) => {
  switch (level) {
    case 'ERROR': return AlertCircle;
    case 'WARN': return AlertTriangle;
    case 'INFO': return Info;
    default: return Terminal;
  }
};
</script>

<template>
  <div class="p-4 sm:p-8 space-y-4 sm:space-y-8">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
          <Terminal class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('system_logs') }}</h1>
          <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold tracking-widest uppercase">{{ t('centralized_log_mgmt') }}</p>
        </div>
      </div>
      
      <div class="flex items-center gap-2 overflow-x-auto pb-2 md:pb-0 no-scrollbar">
        <button 
          @click="autoRefresh = !autoRefresh"
          :class="[
            'flex items-center gap-2 px-3 sm:px-4 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all border shrink-0',
            autoRefresh 
              ? 'bg-[var(--matrix-color)]/10 border-[var(--matrix-color)]/20 text-[var(--matrix-color)]' 
              : 'bg-black/5 dark:bg-white/5 border-transparent text-[var(--text-muted)]'
          ]"
        >
          <RefreshCw :class="['w-3 h-3', { 'animate-spin': autoRefresh && loading }]" />
          <span class="hidden sm:inline">{{ autoRefresh ? t('auto_refresh_on') : t('auto_refresh_off') }}</span>
          <span class="sm:hidden">{{ autoRefresh ? 'ON' : 'OFF' }}</span>
        </button>
        <button 
          @click="downloadLogs"
          class="flex items-center gap-2 px-3 sm:px-4 py-2 rounded-xl bg-black/5 dark:bg-white/5 text-[var(--text-muted)] text-[10px] font-black uppercase tracking-widest hover:bg-black/10 dark:hover:bg-white/10 transition-all shrink-0"
        >
          <Download class="w-3 h-3" />
          {{ t('export') }}
        </button>
        <button 
          @click="clearLogs"
          class="flex items-center gap-2 px-3 sm:px-4 py-2 rounded-xl bg-red-500/10 text-red-500 text-[10px] font-black uppercase tracking-widest hover:bg-red-500/20 transition-all shrink-0"
        >
          <Trash2 class="w-3 h-3" />
          {{ t('clear') }}
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
      <div class="relative group">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
        <input 
          v-model="searchQuery"
          @input="fetchLogs"
          type="text" 
          :placeholder="t('search_logs')"
          class="w-full bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl py-3 pl-12 pr-4 text-xs font-bold tracking-wider text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
        />
      </div>

      <div class="relative">
        <Filter class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <select 
          v-model="filterSource"
          @change="fetchLogs"
          class="w-full bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl py-3 pl-12 pr-4 text-xs font-bold tracking-wider text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 appearance-none transition-all cursor-pointer"
        >
          <option value="all">{{ t('all_sources') }}</option>
          <option value="system">{{ t('system') }}</option>
          <option v-for="bot in botStore.bots" :key="bot.id" :value="bot.id">
            {{ t('bot_prefix') }}: {{ bot.nickname || bot.id }}
          </option>
        </select>
      </div>

      <div class="relative">
        <AlertCircle class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <select 
          v-model="filterLevel"
          @change="fetchLogs"
          class="w-full bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl py-3 pl-12 pr-4 text-xs font-bold tracking-wider text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 appearance-none transition-all cursor-pointer"
        >
          <option value="all">{{ t('all_levels') }}</option>
          <option value="INFO">{{ t('info') }}</option>
          <option value="WARN">{{ t('warning') }}</option>
          <option value="ERROR">{{ t('error') }}</option>
        </select>
      </div>

      <div class="flex items-center justify-end px-4">
        <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">
          {{ logs.length }} {{ t('entries_found') }}
        </span>
      </div>
    </div>

    <!-- Logs Table -->
    <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] overflow-hidden shadow-sm flex flex-col h-[calc(100vh-380px)] sm:h-[calc(100vh-320px)] transition-colors duration-300">
      <div class="flex-1 overflow-y-auto custom-scrollbar p-4 sm:p-6">
        <div v-if="!authStore.isAdmin" class="h-full flex flex-col items-center justify-center space-y-4 opacity-30">
          <Activity class="w-12 h-12 text-[var(--text-main)]" />
          <p class="text-[10px] font-black uppercase tracking-[0.2em] text-[var(--text-main)]">{{ t('admin_required') }}</p>
        </div>
        <div v-else-if="logs.length > 0" class="space-y-2">
          <div v-for="(log, index) in logs" :key="index" class="group flex items-start gap-4 p-3 rounded-2xl hover:bg-[var(--matrix-color)]/5 transition-all">
            <div :class="['mt-1 p-1.5 rounded-lg shrink-0', getLevelColor(log.level)]">
              <component :is="getLevelIcon(log.level)" class="w-3 h-3" />
            </div>
            
            <div class="flex-1 min-w-0 space-y-1">
              <div class="flex items-center gap-3">
                <span class="text-[10px] font-black text-[var(--text-muted)] mono tracking-tighter opacity-50 group-hover:opacity-100 transition-opacity">
                  {{ log.time }}
                </span>
                <span v-if="log.source" class="flex items-center gap-1 px-2 py-0.5 rounded-full bg-[var(--matrix-color)]/5 text-[8px] font-bold text-[var(--matrix-color)] uppercase tracking-widest border border-[var(--matrix-color)]/10">
                  <Bot v-if="log.source !== 'system'" class="w-2 h-2" />
                  {{ t(log.source) }}
                </span>
              </div>
              <p :class="[
                'text-[11px] leading-relaxed break-all font-medium transition-colors',
                log.level === 'ERROR' ? 'text-red-400' : 'text-[var(--text-main)]/80'
              ]">
                {{ log.message }}
              </p>
            </div>
          </div>
        </div>
        
        <div v-else-if="loading" class="h-full flex flex-col items-center justify-center space-y-4 opacity-50">
          <RefreshCw class="w-8 h-8 text-[var(--matrix-color)] animate-spin" />
          <p class="text-[10px] font-black uppercase tracking-[0.2em] text-[var(--text-main)]">{{ t('fetching_logs') }}</p>
        </div>
        
        <div v-else class="h-full flex flex-col items-center justify-center space-y-4 opacity-30">
          <Terminal class="w-12 h-12 text-[var(--text-main)]" />
          <p class="text-[10px] font-black uppercase tracking-[0.2em] text-[var(--text-main)]">{{ t('no_logs') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 10px;
}
.mono {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
</style>