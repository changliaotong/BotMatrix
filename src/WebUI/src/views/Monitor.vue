<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import LineChart from '@/components/charts/LineChart.vue';
import { 
  Activity, 
  Cpu, 
  Database, 
  ChevronLeft,
  ArrowUpRight,
  ArrowDownRight,
  HardDrive,
  Network,
  ListRestart
} from 'lucide-vue-next';

const route = useRoute();
const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const activeType = ref((route.query.type as string) || 'cpu');
let refreshTimer: number | null = null;

onMounted(async () => {
  await botStore.fetchStats();
  refreshTimer = window.setInterval(async () => {
    await botStore.fetchStats();
  }, 3000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

const chartData = computed(() => {
  const stats = botStore.stats;
  switch (activeType.value) {
    case 'cpu':
      return {
        title: t('cpu_usage'),
        value: stats.cpu_usage?.toFixed(1) || '0',
        unit: '%',
        trend: stats.cpu_trend || [],
        color: '#3b82f6',
        icon: Cpu
      };
    case 'memory':
      return {
        title: t('memory_usage'),
        value: stats.memory_used_percent?.toFixed(1) || '0',
        unit: '%',
        trend: (stats.mem_trend || []).map((v: number) => (v / (stats.memory_total || 1)) * 100),
        color: '#a855f7',
        icon: Database
      };
    case 'throughput':
      return {
        title: t('throughput'),
        value: stats.message_count || '0',
        unit: t('unit_msg'),
        trend: stats.msg_trend || [],
        color: '#00ff41',
        icon: Activity
      };
    default:
      return null;
  }
});

const systemInfo = computed(() => [
  { label: 'OS', value: `${botStore.stats.os_platform} ${botStore.stats.os_version}` },
  { label: 'Arch', value: botStore.stats.os_arch },
  { label: 'CPU', value: botStore.stats.cpu_model },
  { label: 'Cores', value: `${botStore.stats.cpu_cores_physical} Physical / ${botStore.stats.cpu_cores_logical} Logical` },
  { label: 'Uptime', value: formatUptime(botStore.stats.start_time) },
]);

const formatUptime = (startTime: number) => {
  if (!startTime) return '-';
  const seconds = Math.floor(Date.now() / 1000 - startTime);
  const d = Math.floor(seconds / (3600 * 24));
  const h = Math.floor((seconds % (3600 * 24)) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${d}d ${h}h ${m}m`;
};

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <router-link to="/" class="p-2 rounded-full hover:bg-black/5 dark:hover:bg-white/5 transition-colors border border-[var(--border-color)]">
          <ChevronLeft class="w-6 h-6" />
        </router-link>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('monitor_detail') }}</h1>
      </div>
      <div class="flex gap-2">
        <button 
          v-for="type in ['cpu', 'memory', 'throughput']" 
          :key="type"
          @click="activeType = type"
          class="px-4 py-2 rounded-xl text-sm font-bold transition-all border"
          :class="activeType === type ? 'bg-[var(--matrix-color)] text-black border-transparent' : 'bg-[var(--bg-card)] text-[var(--text-muted)] border-[var(--border-color)] hover:border-[var(--matrix-color)]/30'"
        >
          {{ t(type === 'cpu' ? 'cpu_usage' : type === 'memory' ? 'memory_usage' : 'throughput') }}
        </button>
      </div>
    </div>

    <div v-if="chartData" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Chart Card -->
      <div class="lg:col-span-2 space-y-6">
        <div class="p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-8">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-4">
              <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                <component :is="chartData.icon" class="w-8 h-8" :style="{ color: chartData.color }" />
              </div>
              <div>
                <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ chartData.title }}</p>
                <div class="flex items-baseline gap-2">
                  <h2 class="text-5xl font-black text-[var(--text-main)] tracking-tight">{{ chartData.value }}</h2>
                  <span class="text-lg font-bold text-[var(--text-muted)]">{{ chartData.unit }}</span>
                </div>
              </div>
            </div>
            <div class="flex flex-col items-end gap-2">
              <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-green-500/10 text-green-500 text-xs font-bold">
                <ArrowUpRight class="w-4 h-4" /> 12%
              </div>
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">VS LAST HOUR</p>
            </div>
          </div>
          
          <div class="h-[400px] w-full">
            <LineChart :data="chartData.trend" :color="chartData.color" :labels="new Array(chartData.trend.length).fill('')" />
          </div>
        </div>

        <!-- Secondary Stats -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <!-- Network Traffic -->
          <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-4">
            <div class="flex items-center gap-3">
              <Network class="w-5 h-5 text-blue-500" />
              <h3 class="font-bold uppercase tracking-widest text-sm">{{ t('network_traffic') }}</h3>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mb-1">Upload</p>
                <p class="font-black text-[var(--text-main)]">{{ formatBytes(botStore.stats.net_sent_trend?.[botStore.stats.net_sent_trend.length-1] || 0) }}/s</p>
              </div>
              <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mb-1">Download</p>
                <p class="font-black text-[var(--text-main)]">{{ formatBytes(botStore.stats.net_recv_trend?.[botStore.stats.net_recv_trend.length-1] || 0) }}/s</p>
              </div>
            </div>
          </div>

          <!-- Disk Status -->
          <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-4">
            <div class="flex items-center gap-3">
              <HardDrive class="w-5 h-5 text-orange-500" />
              <h3 class="font-bold uppercase tracking-widest text-sm">{{ t('disk_status') }}</h3>
            </div>
            <div class="space-y-4">
              <div class="flex items-center justify-between">
                <p class="text-xs font-bold text-[var(--text-main)]">System Disk</p>
                <p class="text-xs font-bold text-[var(--text-muted)]">{{ botStore.stats.disk_usage }}</p>
              </div>
              <div class="w-full h-2 bg-black/10 dark:bg-white/10 rounded-full overflow-hidden">
                <div class="h-full bg-orange-500" :style="{ width: botStore.stats.disk_usage }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar Info -->
      <div class="space-y-6">
        <!-- System Specs -->
        <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-6">
          <h3 class="font-bold uppercase tracking-widest text-sm flex items-center gap-2">
            <ListRestart class="w-5 h-5 text-[var(--matrix-color)]" /> System Info
          </h3>
          <div class="space-y-4">
            <div v-for="info in systemInfo" :key="info.label" class="flex flex-col gap-1 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ info.label }}</p>
              <p class="text-sm font-bold text-[var(--text-main)] break-all">{{ info.value }}</p>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="p-6 rounded-3xl bg-[var(--matrix-color)] text-black space-y-4">
          <h3 class="font-black uppercase tracking-widest text-sm">System Control</h3>
          <p class="text-xs font-bold opacity-70">Perform critical system operations from this terminal.</p>
          <div class="grid grid-cols-2 gap-3">
            <button class="px-4 py-3 rounded-2xl bg-black text-white text-xs font-black hover:opacity-80 transition-opacity">RESTART</button>
            <button class="px-4 py-3 rounded-2xl bg-black/20 text-black text-xs font-black hover:bg-black/30 transition-colors">DUMP</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>