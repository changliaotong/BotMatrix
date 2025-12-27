<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import LineChart from '@/components/charts/LineChart.vue';
import { 
  Activity, 
  Cpu, 
  Database, 
  MessageSquare, 
  Clock,
  LayoutDashboard,
  BarChart3,
  Terminal,
  ChevronRight,
  Bot,
  Users
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();

const recentLogs = ref<any[]>([]);
let refreshTimer: number | null = null;

onMounted(async () => {
  await botStore.fetchStats();
  await botStore.fetchBots();
  fetchLogs();
  
  timeTimer = window.setInterval(() => {
    currentTime.value = new Date().toLocaleTimeString();
  }, 1000);

  // Refresh stats every 5 seconds for more real-time monitoring
  refreshTimer = window.setInterval(async () => {
    await botStore.fetchStats();
    fetchLogs();
  }, 5000);
});

const fetchLogs = async () => {
  try {
    const data = await botStore.fetchSystemLogs();
    if (data.success && data.data && data.data.logs) {
      recentLogs.value = data.data.logs.slice(0, 10);
    }
  } catch (err) {
    console.error('Failed to fetch logs in dashboard:', err);
  }
};

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
  if (timeTimer) clearInterval(timeTimer);
});

const iconMap: Record<string, any> = {
  'Bot': Bot,
  'MessageSquare': MessageSquare,
  'Users': Users,
  'Activity': Activity,
  'Cpu': Cpu,
  'Database': Database,
  'Clock': Clock,
  'LayoutDashboard': LayoutDashboard,
  'BarChart3': BarChart3,
  'Terminal': Terminal
};

const t = (key: string) => systemStore.t(key);

const statsCards = computed(() => [
  { 
    id: 'bots',
    label: 'total_bots', 
    value: botStore.stats.bot_count_total || '0', 
    icon: 'Bot', 
    colorClass: 'bg-blue-500/10', 
    textColor: 'text-blue-500', 
    unit: 'unit_nodes',
    routeName: 'bots'
  },
  { 
    id: 'workers',
    label: 'active_workers', 
    value: botStore.stats.worker_count || '0', 
    icon: 'Users', 
    colorClass: 'bg-purple-500/10', 
    textColor: 'text-purple-500', 
    unit: 'unit_active',
    routeName: 'workers'
  },
  { 
    id: 'cpu',
    label: 'cpu_usage', 
    value: botStore.stats.cpu_usage ? botStore.stats.cpu_usage.toFixed(1) : '0', 
    icon: 'Cpu', 
    colorClass: 'bg-orange-500/10', 
    textColor: 'text-orange-500', 
    unit: '%',
    trend: botStore.stats.cpu_trend || [],
    routeName: 'monitor'
  },
  { 
    id: 'memory',
    label: 'memory_usage', 
    value: botStore.stats.memory_used_percent ? botStore.stats.memory_used_percent.toFixed(1) : '0', 
    icon: 'Database', 
    colorClass: 'bg-pink-500/10', 
    textColor: 'text-pink-500', 
    unit: '%',
    trend: (botStore.stats.mem_trend || []).map((v: number) => (v / (botStore.stats.memory_total || 1)) * 100),
    routeName: 'monitor'
  },
  { 
    id: 'throughput',
    label: 'throughput', 
    value: botStore.stats.message_count || '0', 
    icon: 'Activity', 
    colorClass: 'bg-[var(--matrix-color)]/10', 
    textColor: 'text-[var(--matrix-color)]', 
    unit: 'unit_msg',
    trend: botStore.stats.msg_trend || [],
    routeName: 'monitor'
  },
]);

const activeGroups = computed(() => botStore.stats.active_groups_today || 0);
const dragonKing = computed(() => {
  // Mocking dragon king data since backend doesn't provide it yet
  return { name: 'Matrix_User', count: botStore.stats.message_count > 0 ? Math.floor(botStore.stats.message_count * 0.2) : 0 };
});

const currentTime = ref(new Date().toLocaleTimeString());
let timeTimer: number | null = null;

</script>

<template>
  <div class="p-4 sm:p-6 space-y-4 sm:space-y-6">
    <!-- Header with Time -->
    <div class="flex flex-col sm:flex-row items-center justify-between gap-4 bg-[var(--bg-card)] border border-[var(--border-color)] p-4 sm:px-8 rounded-3xl">
      <div class="flex items-center gap-4">
        <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)]">
          <LayoutDashboard class="w-6 h-6" />
        </div>
        <div>
          <h1 class="text-xl font-black text-[var(--text-main)] tracking-tight">{{ t('dashboard') }}</h1>
          <p class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full bg-[var(--matrix-color)] animate-pulse"></span>
            {{ t('neural_link_active') }}
          </p>
        </div>
      </div>
      <div class="flex items-center gap-6 px-6 py-2 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
        <div class="text-right">
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('current_time') }}</p>
          <p class="text-lg font-black text-[var(--matrix-color)] mono">{{ currentTime }}</p>
        </div>
        <Clock class="w-8 h-8 text-[var(--text-muted)] opacity-50" />
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4 sm:gap-6">
      <router-link 
        v-for="card in statsCards" 
        :key="card.label"
        :to="card.routeName === 'monitor' ? { name: 'monitor', query: { type: card.id } } : { name: card.routeName }"
        class="group p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500 relative overflow-hidden"
      >
        <!-- Background Decoration -->
        <div class="absolute -right-4 -bottom-4 w-24 h-24 opacity-5 group-hover:opacity-10 transition-opacity rotate-12">
          <component :is="iconMap[card.icon]" class="w-full h-full" />
        </div>

        <div class="flex items-start justify-between relative z-10">
          <div class="space-y-4 w-full">
            <div class="flex items-center justify-between">
              <div class="p-3 rounded-2xl inline-flex" :class="card.colorClass">
                <component :is="iconMap[card.icon]" class="w-6 h-6" :class="card.textColor" />
              </div>
              <div v-if="card.trend" class="w-20 h-10 opacity-50 group-hover:opacity-100 transition-opacity">
                <LineChart :data="card.trend" :color="card.textColor.includes('orange') ? '#f97316' : card.textColor.includes('pink') ? '#ec4899' : '#00ff41'" :fill="false" />
              </div>
            </div>
            <div>
              <p class="text-[10px] sm:text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider">{{ t(card.label) }}</p>
              <div class="flex items-baseline gap-1 sm:gap-2">
                <h2 class="text-2xl sm:text-3xl font-black text-[var(--text-main)] tracking-tight">{{ card.value }}</h2>
                <span class="text-[8px] sm:text-[10px] font-black text-[var(--text-muted)]">{{ card.unit.startsWith('unit_') ? t(card.unit) : card.unit }}</span>
              </div>
            </div>
          </div>
          <div class="absolute top-4 right-4 sm:top-6 sm:right-6 w-8 h-8 sm:w-10 sm:h-10 flex items-center justify-center rounded-full border border-[var(--border-color)] group-hover:border-[var(--matrix-color)]/30 transition-colors bg-[var(--bg-card)]">
            <ChevronRight class="w-4 h-4 sm:w-5 sm:h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
          </div>
        </div>
      </router-link>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-6">
      <!-- Main Content Area -->
      <div class="lg:col-span-2 space-y-4 sm:space-y-6">
        <!-- Active Groups & Dragon King -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
          <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl bg-orange-500/10 text-orange-500">
                  <Users class="w-5 h-5" />
                </div>
                <h3 class="font-bold">{{ t('active_groups') }}</h3>
              </div>
              <router-link to="/contacts" class="p-2 rounded-full hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                <ChevronRight class="w-4 h-4 text-[var(--text-muted)]" />
              </router-link>
            </div>
            <div class="flex items-baseline gap-2">
              <span class="text-4xl font-black text-[var(--text-main)]">{{ activeGroups }}</span>
              <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('unit_msg').replace('消息', '群组') }}</span>
            </div>
            <div class="mt-4 flex -space-x-2">
              <div v-for="i in 5" :key="i" class="w-8 h-8 rounded-full border-2 border-[var(--bg-card)] bg-[var(--border-color)] flex items-center justify-center text-[10px] font-bold">
                G{{ i }}
              </div>
            </div>
          </div>

          <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl bg-yellow-500/10 text-yellow-500">
                  <Activity class="w-5 h-5" />
                </div>
                <h3 class="font-bold">{{ t('dragon_king') }}</h3>
              </div>
              <div class="p-2 rounded-full hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer">
                <ChevronRight class="w-4 h-4 text-[var(--text-muted)]" />
              </div>
            </div>
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-2xl bg-yellow-500/20 flex items-center justify-center text-yellow-500 font-black">👑</div>
              <div>
                <p class="font-bold text-[var(--text-main)]">{{ dragonKing.name }}</p>
                <p class="text-xs text-[var(--text-muted)]">{{ dragonKing.count }} {{ t('unit_msg') }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Message Chart -->
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-4 sm:space-y-6">
          <div class="flex items-center justify-between">
            <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
              <BarChart3 class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('throughput') }}
            </h3>
            <div class="flex gap-2">
              <button class="px-3 py-1 text-[10px] font-bold rounded-lg bg-[var(--matrix-color)] text-black">1H</button>
              <button class="px-3 py-1 text-[10px] font-bold rounded-lg bg-black/10 dark:bg-white/5 text-[var(--text-muted)]">24H</button>
            </div>
          </div>
          <div class="h-48 sm:h-64 w-full">
            <LineChart :data="botStore.stats.msg_trend" />
          </div>
        </div>
      </div>

      <!-- Recent Logs -->
      <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex flex-col">
        <div class="flex items-center justify-between mb-4 sm:mb-6">
          <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
            <Terminal class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('recent_logs') }}
          </h3>
          <router-link to="/logs" class="text-xs font-bold text-[var(--matrix-color)] hover:underline">{{ t('view_all') }}</router-link>
        </div>
        
        <div class="flex-1 space-y-4 overflow-y-auto custom-scrollbar pr-2">
          <div v-if="recentLogs.length === 0" class="h-full flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
            <Database class="w-8 h-8" />
            <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('status_loading') }}...</p>
          </div>
          <div 
            v-for="(log, idx) in recentLogs" 
            :key="idx"
            class="p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--border-color)] transition-all group"
          >
            <div class="flex items-start gap-3">
              <div class="w-1 h-1 rounded-full mt-2" :class="log.level === 'ERROR' ? 'bg-red-500' : 'bg-[var(--matrix-color)]'"></div>
              <div class="space-y-1 min-w-0">
                <p class="text-[10px] font-bold text-[var(--text-muted)] mono uppercase">{{ log.time }}</p>
                <p class="text-xs text-[var(--text-main)] font-medium truncate group-hover:text-clip group-hover:whitespace-normal">{{ log.message }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-matrix {
  color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: rgba(0, 255, 65, 0.1);
}
</style>
