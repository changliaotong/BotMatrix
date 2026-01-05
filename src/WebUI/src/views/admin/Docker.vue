<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Box, 
  Play, 
  Square, 
  RotateCcw, 
  Trash2, 
  Terminal, 
  Search,
  RefreshCw,
  Cpu,
  HardDrive,
  Activity,
  CheckCircle2,
  XCircle,
  AlertCircle
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const containers = ref<any[]>([]);
const loading = ref(true);
const refreshing = ref(false);
const search = ref('');
const autoRefresh = ref(true);
const showLogsModal = ref(false);
const currentLogs = ref('');
const selectedContainerId = ref('');
const loadingLogs = ref(false);

const fetchContainers = async (isRefresh = false) => {
  if (isRefresh) refreshing.value = true;
  else loading.value = true;
  
  try {
    const data = await botStore.fetchDockerContainers();
    if (data.success && data.data) {
      containers.value = data.data.containers || [];
    }
  } finally {
    loading.value = false;
    refreshing.value = false;
  }
};

const handleAction = async (containerId: string, action: string) => {
  if (action === 'delete') {
    if (!confirm(t('confirm_delete_container'))) return;
  }
  
  try {
    const data = await botStore.dockerAction(containerId, action);
    if (data.status === 'ok' || data.success) {
      await fetchContainers(true);
    }
  } catch (err) {
    console.error(`Failed to perform ${action} on container ${containerId}:`, err);
  }
};

const showLogs = async (containerId: string) => {
  selectedContainerId.value = containerId;
  showLogsModal.value = true;
  loadingLogs.value = true;
  currentLogs.value = '';
  
  try {
    const data = await botStore.getLogs(containerId);
    if (data.status === 'ok' || data.success) {
      currentLogs.value = data.logs;
    }
  } catch (err) {
    console.error('Failed to fetch logs:', err);
    currentLogs.value = t('failed_fetch_logs');
  } finally {
    loadingLogs.value = false;
  }
};

const getStatusColor = (status: string) => {
  if (status.includes('Up')) return 'text-green-500 bg-green-500/10 border-green-500/20';
  if (status.includes('Exited')) return 'text-red-500 bg-red-500/10 border-red-500/20';
  return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
};

const filteredContainers = () => {
  return containers.value.filter(c => {
    const name = (c.Names && c.Names[0]) || c.name || '';
    const id = c.Id || c.id || '';
    const image = c.Image || c.image || '';
    
    return name.toLowerCase().includes(search.value.toLowerCase()) ||
           id.toLowerCase().includes(search.value.toLowerCase()) ||
           image.toLowerCase().includes(search.value.toLowerCase());
  });
};

let refreshInterval: number;

const startRefresh = () => {
  stopRefresh();
  refreshInterval = window.setInterval(() => {
    if (autoRefresh.value) fetchContainers(true);
  }, 5000);
};

const stopRefresh = () => {
  if (refreshInterval) clearInterval(refreshInterval);
};

onMounted(() => {
  fetchContainers();
  startRefresh();
});

onUnmounted(() => {
  stopRefresh();
});
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('docker') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('docker_desc') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input 
            v-model="search"
            type="text"
            :placeholder="t('search_containers')"
            class="pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] w-64 transition-all"
          />
        </div>
        <button 
          @click="autoRefresh = !autoRefresh"
          :class="autoRefresh ? 'text-[var(--matrix-color)] border-[var(--matrix-color)]/30 bg-[var(--matrix-color)]/5' : 'text-[var(--text-muted)] border-[var(--border-color)] bg-[var(--bg-card)]'"
          class="flex items-center gap-2 px-3 py-2 rounded-2xl border font-black text-[10px] uppercase tracking-widest transition-all hover:scale-105 active:scale-95"
        >
          <div :class="autoRefresh ? 'bg-[var(--matrix-color)] animate-pulse' : 'bg-[var(--text-muted)]/30'" class="w-1.5 h-1.5 rounded-full"></div>
          {{ t('auto_refresh') }}
        </button>
        <button 
          @click="fetchContainers(true)"
          :disabled="refreshing"
          class="p-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all disabled:opacity-50"
        >
          <RefreshCw :class="{ 'animate-spin': refreshing }" class="w-5 h-5" />
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-blue-500/10 text-blue-500">
            <Box class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('total') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ containers.length }}</div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-green-500/10 text-green-500">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('running') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">
              {{ containers.filter(c => c.status.includes('Up')).length }}
            </div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-red-500/10 text-red-500">
            <XCircle class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('stopped') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">
              {{ containers.filter(c => !c.status.includes('Up')).length }}
            </div>
          </div>
        </div>
      </div>
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-yellow-500/10 text-yellow-500">
            <Activity class="w-6 h-6" />
          </div>
          <div>
            <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('health') }}</div>
            <div class="text-2xl font-black text-[var(--text-main)]">{{ t('good') }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Container List -->
    <div v-if="loading" class="space-y-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-32 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="container in filteredContainers()" 
        :key="container.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500"
      >
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div class="flex items-center gap-6">
            <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <Box class="w-8 h-8 text-[var(--matrix-color)]" />
            </div>
            <div>
              <div class="flex items-center gap-3">
                <h3 class="font-black text-lg text-[var(--text-main)] tracking-tight">{{ container.name }}</h3>
                <div :class="getStatusColor(container.status)" class="px-3 py-0.5 rounded-full border text-[10px] font-black uppercase tracking-widest">
                  {{ container.status }}
                </div>
              </div>
              <div class="flex flex-wrap items-center gap-4 mt-2">
                <div class="flex items-center gap-1.5">
                  <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('id') }}:</span>
                  <span class="text-[10px] font-bold text-[var(--text-main)] opacity-60">{{ container.id.substring(0, 12) }}</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('image') }}:</span>
                  <span class="text-[10px] font-bold text-[var(--text-main)] opacity-60">{{ container.image }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <div class="flex items-center gap-6 mr-6 border-r border-[var(--border-color)] pr-6">
              <div class="text-center">
                <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('cpu') }}</div>
                <div class="text-xs font-black text-[var(--text-main)]">{{ container.cpu || '0.0%' }}</div>
              </div>
              <div class="text-center">
                <div class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-1">{{ t('memory') }}</div>
                <div class="text-xs font-black text-[var(--text-main)]">{{ container.memory || '0B' }}</div>
              </div>
            </div>

            <div class="flex items-center gap-2">
              <button 
                v-if="!container.status.includes('Up')"
                @click="handleAction(container.id, 'start')"
                class="p-2.5 rounded-xl bg-green-500/10 border border-green-500/20 text-green-500 hover:bg-green-500 hover:text-[var(--sidebar-text)] transition-all"
                :title="t('start')"
              >
                <Play class="w-4 h-4" />
              </button>
              <button 
                v-else
                @click="handleAction(container.id, 'stop')"
                class="p-2.5 rounded-xl bg-yellow-500/10 border border-yellow-500/20 text-yellow-500 hover:bg-yellow-500 hover:text-[var(--sidebar-text)] transition-all"
                :title="t('stop')"
              >
                <Square class="w-4 h-4" />
              </button>
              
              <button 
                @click="handleAction(container.id, 'restart')"
                class="p-2.5 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-500 hover:bg-blue-500 hover:text-[var(--sidebar-text)] transition-all"
                :title="t('restart')"
              >
                <RotateCcw class="w-4 h-4" />
              </button>
              
              <button 
                @click="showLogs(container.id)"
                class="p-2.5 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-muted)] hover:border-[var(--matrix-color)]/30 hover:text-[var(--matrix-color)] transition-all"
                :title="t('logs')"
              >
                <Terminal class="w-4 h-4" />
              </button>

              <button 
                @click="handleAction(container.id, 'delete')"
                class="p-2.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 hover:bg-red-500 hover:text-[var(--sidebar-text)] transition-all"
                :title="t('delete')"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredContainers().length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
        <Box class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_containers_found') }}</h2>
        <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('no_containers_desc') }}</p>
      </div>
    </div>
  </div>

  <!-- Logs Modal -->
  <div v-if="showLogsModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
    <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] w-full max-w-4xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
      <div class="p-6 sm:p-8 border-b border-[var(--border-color)] flex items-center justify-between bg-black/5 dark:bg-white/5">
        <div class="flex items-center gap-3">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)]">
            <Terminal class="w-6 h-6" />
          </div>
          <div>
            <h2 class="text-xl font-black text-[var(--text-main)] tracking-tight uppercase">{{ t('container_logs') }}</h2>
            <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1 opacity-60">{{ selectedContainerId }}</p>
          </div>
        </div>
        <button 
          @click="showLogsModal = false"
          class="p-2 rounded-xl hover:bg-black/10 dark:hover:bg-white/10 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all"
        >
          <XCircle class="w-8 h-8" />
        </button>
      </div>
      
      <div class="flex-1 p-6 sm:p-8 overflow-hidden">
        <div class="h-full bg-black/90 rounded-[1.5rem] p-6 overflow-y-auto font-mono text-xs text-green-500/90 leading-relaxed scrollbar-thin scrollbar-thumb-[var(--matrix-color)]/20 shadow-inner">
          <div v-if="loadingLogs" class="flex flex-col items-center justify-center h-full gap-4 text-[var(--matrix-color)]/50">
            <RefreshCw class="w-8 h-8 animate-spin" />
            <span class="text-xs font-black uppercase tracking-widest">{{ t('fetching_logs') }}</span>
          </div>
          <pre v-else class="whitespace-pre-wrap break-all">{{ currentLogs || t('no_logs_available') }}</pre>
        </div>
      </div>

      <div class="p-6 sm:p-8 border-t border-[var(--border-color)] flex justify-end gap-4 bg-black/5 dark:bg-white/5">
        <button 
          @click="showLogs(selectedContainerId)"
          class="px-8 py-3 rounded-2xl bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] text-xs font-black uppercase tracking-widest hover:scale-105 active:scale-95 transition-all shadow-lg shadow-[var(--matrix-color)]/20"
        >
          {{ t('refresh_logs') }}
        </button>
        <button 
          @click="showLogsModal = false"
          class="px-8 py-3 rounded-2xl bg-black/5 dark:hover:bg-black/10 dark:bg-white/5 border border-[var(--border-color)] text-[var(--text-main)] text-xs font-black uppercase tracking-widest transition-all"
        >
          {{ t('close') }}
        </button>
      </div>
    </div>
  </div>
</template>
