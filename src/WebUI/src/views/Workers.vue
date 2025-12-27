<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
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
  Network
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const workers = ref<any[]>([]);
const loading = ref(true);
let refreshTimer: number | null = null;

const fetchWorkers = async () => {
  loading.value = true;
  try {
    const data = await botStore.fetchWorkers();
    if (data.success) {
      workers.value = data.workers;
    }
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

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('workers') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('workers_desc') }}</p>
      </div>
      <button 
        @click="fetchWorkers"
        class="p-3 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group"
      >
        <RefreshCw class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" :class="{ 'animate-spin': loading }" />
      </button>
    </div>

    <!-- Workers Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="worker in workers" 
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
              <h3 class="font-black text-[var(--text-main)] break-all">{{ worker.id }}</h3>
              <div class="flex items-center gap-2 mt-1">
                <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ worker.status }}</span>
              </div>
            </div>
          </div>
          <div :class="getStatusColor(worker.status)" class="px-3 py-1 rounded-full border text-[10px] font-black uppercase tracking-widest">
            {{ worker.status }}
          </div>
        </div>

        <!-- Stats Grid -->
        <div class="grid grid-cols-2 gap-4 mb-6">
          <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div class="flex items-center gap-2 mb-1">
              <Activity class="w-3 h-3 text-[var(--text-muted)]" />
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">Handled</p>
            </div>
            <p class="text-xl font-black text-[var(--text-main)]">{{ worker.handled_count }}</p>
          </div>
          <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div class="flex items-center gap-2 mb-1">
              <Zap class="w-3 h-3 text-[var(--text-muted)]" />
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">Avg RTT</p>
            </div>
            <p class="text-xl font-black text-[var(--text-main)]">{{ worker.avg_rtt }}</p>
          </div>
        </div>

        <!-- Details -->
        <div class="space-y-3">
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-2 text-[var(--text-muted)] font-bold">
              <Network class="w-3 h-3" /> ADDR
            </div>
            <span class="font-mono text-[var(--text-main)]">{{ worker.remote_addr }}</span>
          </div>
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-2 text-[var(--text-muted)] font-bold">
              <Clock class="w-3 h-3" /> CONNECTED
            </div>
            <span class="font-mono text-[var(--text-main)]">{{ worker.connected }}</span>
          </div>
        </div>

        <!-- Progress Bar (Simulated load) -->
        <div class="mt-6 pt-6 border-t border-[var(--border-color)]">
          <div class="flex items-center justify-between mb-2">
            <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">Current Load</span>
            <span class="text-[10px] font-black text-[var(--matrix-color)]">High Performance</span>
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
          <h3 class="font-black text-[var(--text-main)] uppercase tracking-tighter">Deploy New Worker</h3>
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">Via Docker Management</p>
        </div>
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && workers.length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
      <AlertCircle class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
      <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_workers_found') || 'No Workers Connected' }}</h2>
      <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">Check infrastructure status or deploy new nodes</p>
    </div>
  </div>
</template>