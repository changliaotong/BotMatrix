<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Cpu, 
  Activity, 
  Zap, 
  RefreshCw, 
  ShieldCheck, 
  AlertCircle,
  Clock,
  Network,
  Search
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const workers = ref<any[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const searchQuery = ref('');
const sortBy = ref<'id' | 'status' | 'handled_count' | 'avg_rtt'>('id');
const sortOrder = ref<'asc' | 'desc'>('asc');
let refreshTimer: number | null = null;

const filteredWorkers = computed(() => {
  let result = [...workers.value];
  
  // Filter
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    result = result.filter(w => 
      (w.id && w.id.toLowerCase().includes(q)) || 
      (w.name && w.name.toLowerCase().includes(q)) ||
      (w.remote_addr && w.remote_addr.toLowerCase().includes(q))
    );
  }
  
  // Sort
  result.sort((a, b) => {
    let valA: any = a[sortBy.value];
    let valB: any = b[sortBy.value];
    
    if (sortBy.value === 'avg_rtt') {
      // Parse "45ms" to 45
      valA = parseInt(valA) || 0;
      valB = parseInt(valB) || 0;
    }

    if (valA < valB) return sortOrder.value === 'asc' ? -1 : 1;
    if (valA > valB) return sortOrder.value === 'asc' ? 1 : -1;
    return 0;
  });
  
  return result;
});

const toggleSort = (field: 'id' | 'status' | 'handled_count' | 'avg_rtt') => {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortBy.value = field;
    sortOrder.value = 'asc';
  }
};

const fetchWorkers = async () => {
  error.value = null;
  loading.value = true;
  try {
    const data = await botStore.fetchWorkers();
    if (data.success) {
      workers.value = data.data?.workers || [];
    } else {
      error.value = data.message || 'Failed to fetch workers';
    }
  } catch (err: any) {
    console.error('Failed to fetch workers:', err);
    error.value = err.message || 'An error occurred while fetching workers';
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchWorkers();
  refreshTimer = window.setInterval(fetchWorkers, 5000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'online': return 'text-green-500 bg-green-500/10 border-green-500/20';
    case 'busy': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
    default: return 'text-red-500 bg-red-500/10 border-red-500/20';
  }
};

const getWorkerName = (worker: any) => {
  if (worker.name) return worker.name;
  return t(worker.id);
};

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('workers') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('workers_desc') }}</p>
      </div>
      
      <div class="flex flex-col sm:flex-row items-center gap-4">
        <div class="relative w-full sm:w-64 group">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
          <input 
            v-model="searchQuery"
            type="text" 
            :placeholder="t('search_workers')"
            class="w-full pl-11 pr-4 py-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl text-xs font-bold text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
          >
        </div>

        <div class="flex items-center gap-2 bg-[var(--bg-card)] border border-[var(--border-color)] p-1 rounded-2xl">
          <button 
            v-for="field in ['status', 'handled_count', 'avg_rtt']" 
            :key="field"
            @click="toggleSort(field as any)"
            :class="['px-3 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all', 
              sortBy === field ? 'bg-[var(--matrix-color)] text-black' : 'text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5']"
          >
            {{ t('sort_' + field) }}
            <span v-if="sortBy === field" class="ml-1">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
          </button>
        </div>

        <button 
          @click="fetchWorkers"
          class="p-3 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group"
        >
          <RefreshCw class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" :class="{ 'animate-spin': loading }" />
        </button>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center gap-3 text-red-500">
      <AlertCircle class="w-5 h-5" />
      <div class="flex-1">
        <p class="text-xs font-bold uppercase tracking-widest">{{ t('error') }}</p>
        <p class="text-sm font-black">{{ error }}</p>
      </div>
      <button @click="fetchWorkers" class="px-3 py-1 rounded-lg bg-red-500 text-white text-[10px] font-black uppercase tracking-widest">
        {{ t('retry') }}
      </button>
    </div>

    <!-- Workers Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="worker in filteredWorkers" 
        :key="worker.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500 relative overflow-hidden"
      >
        <!-- Top Info -->
        <div class="flex items-start justify-between mb-6">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <Cpu class="w-6 h-6 text-[var(--matrix-color)]" />
            </div>
            <div>
              <h3 class="font-black text-[var(--text-main)] break-all">{{ getWorkerName(worker) }}</h3>
              <div class="flex items-center gap-2 mt-1">
                <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t(worker.status.toLowerCase()) }}</span>
              </div>
            </div>
          </div>
          <div :class="getStatusColor(worker.status)" class="px-3 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest">
            {{ t(worker.status.toLowerCase()) }}
          </div>
        </div>

        <!-- Stats Grid -->
        <div class="grid grid-cols-2 gap-4 mb-6">
          <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div class="flex items-center gap-2 mb-1">
              <Activity class="w-3 h-3 text-[var(--text-muted)]" />
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('handled') }}</p>
            </div>
            <p class="text-xl font-black text-[var(--text-main)]">{{ worker.handled_count }}</p>
          </div>
          <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div class="flex items-center gap-2 mb-1">
              <Zap class="w-3 h-3 text-[var(--text-muted)]" />
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('avg_rtt') }}</p>
            </div>
            <p class="text-xl font-black text-[var(--text-main)]">{{ worker.avg_rtt }}</p>
          </div>
        </div>

        <!-- Details -->
        <div class="space-y-3">
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-2 text-[var(--text-muted)] font-bold">
              <Network class="w-3 h-3" /> {{ t('remote_addr') }}
            </div>
            <span class="font-mono text-[var(--text-main)]">{{ worker.remote_addr }}</span>
          </div>
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-2 text-[var(--text-muted)] font-bold">
              <Clock class="w-3 h-3" /> {{ t('connected_time') }}
            </div>
            <span class="font-mono text-[var(--text-main)]">{{ worker.connected }}</span>
          </div>
        </div>

        <!-- Progress Bar (Simulated load) -->
        <div class="mt-6 pt-6 border-t border-[var(--border-color)]">
          <div class="flex items-center justify-between mb-2">
            <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('current_load') }}</span>
            <span class="text-[10px] font-black text-[var(--matrix-color)]">{{ t('high_performance') }}</span>
          </div>
          <div class="w-full h-1.5 bg-black/10 dark:bg-white/10 rounded-full overflow-hidden">
            <div class="h-full bg-[var(--matrix-color)] w-1/3 shadow-[0_0_10px_var(--matrix-color)]"></div>
          </div>
        </div>
      </div>

      <!-- Add Worker Placeholder -->
      <router-link to="/docker" class="group p-6 rounded-3xl border-2 border-dashed border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all flex flex-col items-center justify-center gap-4 bg-black/5 dark:bg-white/5 min-h-[280px]">
        <div class="p-4 rounded-full bg-black/5 dark:bg-white/5 group-hover:scale-110 transition-transform">
          <Zap class="w-8 h-8 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
        </div>
        <div class="text-center">
          <h3 class="font-black text-[var(--text-main)] uppercase tracking-tighter">{{ t('deploy_new_worker') }}</h3>
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('via_docker_mgmt') }}</p>
        </div>
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && workers.length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
      <AlertCircle class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
      <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_workers_found') }}</h2>
      <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('check_infra_status') }}</p>
    </div>
  </div>
</template>