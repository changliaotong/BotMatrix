<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted, watch } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { useAuthStore } from '@/stores/auth';
import LineChart from '@/components/charts/LineChart.vue';
import { getPlatformIcon, getPlatformColor, isPlatformAvatar, getPlatformFromAvatar, getAvatarUrl } from '@/utils/avatar';
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
  Users,
  X,
  RefreshCw
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const authStore = useAuthStore();

const recentLogs = ref<any[]>([]);
const recentMessages = ref<any[]>([]);
const chatStats = ref<any>(null);
const loading = ref(true);
const currentTime = ref(new Date().toLocaleTimeString());
let timeTimer: number | null = null;
let refreshTimer: number | null = null;

const isInitialFetching = ref(false);
const refreshing = ref(false);
const fetchError = ref<string | null>(null);

const fetchLogs = async () => {
  if (!authStore.isAdmin) {
    console.log('Skipping logs fetch for non-admin user');
    return;
  }
  try {
    const data = await botStore.fetchSystemLogs();
    if (data && data.success && data.data && data.data.logs) {
      recentLogs.value = data.data.logs.slice(0, 10);
    }
  } catch (err) {
    console.error('Failed to fetch logs in dashboard:', err);
  }
};

const fetchMessages = async () => {
  if (!authStore.isAdmin) {
    console.log('Skipping messages fetch for non-admin user');
    return;
  }
  try {
    const messages = await botStore.fetchMessages(10);
    if (Array.isArray(messages)) {
      recentMessages.value = messages;
    }
  } catch (err) {
    console.error('Failed to fetch messages in dashboard:', err);
  }
};

const initFetch = async () => {
  if (!authStore.isAuthenticated) {
    loading.value = false;
    return;
  }
  if (isInitialFetching.value) return;
  
  isInitialFetching.value = true;
  loading.value = true;
  fetchError.value = null;
  
  try {
      const promises: Promise<any>[] = [
        botStore.fetchStats(),
        botStore.fetchBots(),
        botStore.fetchChatStats()
      ];

      // 只有管理员才获取系统日志和消息
      if (authStore.isAdmin) {
        promises.push(fetchLogs());
        promises.push(fetchMessages());
      }

      const results = await Promise.allSettled(promises);
    
    // Check if critical fetches failed
    const statsFailed = results[0].status === 'rejected';
    const botsFailed = results[1].status === 'rejected';

    if (statsFailed || botsFailed) {
      console.error('Critical data fetch failed:', { statsFailed, botsFailed });
      fetchError.value = 'failed_to_load_critical_data';
    }

    // Index 2 is always fetchChatStats
    if (results[2].status === 'fulfilled' && results[2].value) {
      chatStats.value = results[2].value;
    }
  } catch (err) {
    console.error('Initial fetch failed:', err);
    fetchError.value = 'failed_to_load_dashboard_data';
  } finally {
    loading.value = false;
    isInitialFetching.value = false;
  }
};

const refreshData = async () => {
  if (refreshing.value) return;
  refreshing.value = true;
  try {
    await Promise.allSettled([
      botStore.fetchStats(),
      fetchLogs(),
      fetchMessages()
    ]);
    
    const cs = await botStore.fetchChatStats();
    if (cs) chatStats.value = cs;
  } catch (err) {
    console.error('Dashboard refresh failed:', err);
  } finally {
    refreshing.value = false;
  }
};

onMounted(() => {
  timeTimer = window.setInterval(() => {
    currentTime.value = new Date().toLocaleTimeString();
  }, 1000);
  
  // If already authenticated on mount, fetch data
  if (authStore.isAuthenticated) {
    initFetch();
  }
});

// Watch for auth changes to fetch data
watch(() => authStore.isAuthenticated, (newVal) => {
  if (newVal) {
    initFetch();
  }
}, { immediate: true });

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

const showActiveGroupsModal = ref(false);
const showDragonKingModal = ref(false);

const statsCards = computed(() => {
  const colorMap: Record<string, string> = {
    'text-blue-500': '#3b82f6',
    'text-purple-500': '#a855f7',
    'text-orange-500': '#f97316',
    'text-pink-500': '#ec4899',
    'text-green-500': '#22c55e',
    'text-cyan-500': '#06b6d4',
    'text-emerald-500': '#10b981'
  };

  return [
    { 
      id: 'bots',
      label: 'total_bots', 
      value: botStore.stats.bot_count || '0', 
      icon: 'Bot', 
      colorClass: 'bg-blue-500/10', 
      textColor: 'text-blue-500', 
      chartColor: colorMap['text-blue-500'],
      unit: 'unit_active',
      routeName: 'console-bots'
    },
    { 
      id: 'workers',
      label: 'active_workers', 
      value: botStore.stats.worker_count || '0', 
      icon: 'Users', 
      colorClass: 'bg-purple-500/10', 
      textColor: 'text-purple-500', 
      chartColor: colorMap['text-purple-500'],
      unit: 'unit_active',
      routeName: 'admin-workers'
    },
    { 
      id: 'cpu',
      label: 'cpu_usage', 
      value: (botStore.stats.cpu_usage || 0).toFixed(1) + '%', 
      icon: 'Cpu', 
      colorClass: 'bg-orange-500/10', 
      textColor: 'text-orange-500', 
      chartColor: colorMap['text-orange-500'],
      unit: 'unit_live',
      trend: botStore.stats.cpu_trend || [],
      routeName: 'admin-monitor'
    },
    { 
      id: 'memory',
      label: 'memory_usage', 
      value: (botStore.stats.memory_used_percent || 0).toFixed(1) + '%', 
      icon: 'Database', 
      colorClass: 'bg-blue-500/10', 
      textColor: 'text-blue-500', 
      chartColor: colorMap['text-blue-500'],
      unit: 'unit_live',
      trend: botStore.stats.mem_trend || [],
      routeName: 'admin-monitor'
    },
    { 
      id: 'throughput',
      label: 'throughput', 
      value: botStore.stats.message_count || '0', 
      icon: 'Activity', 
      colorClass: 'bg-pink-500/10', 
      textColor: 'text-pink-500', 
      chartColor: colorMap['text-pink-500'],
      unit: 'unit_msg',
      trend: botStore.stats.msg_trend || [],
      routeName: 'admin-monitor'
    },
    { 
      id: 'time',
      label: 'current_time', 
      value: currentTime.value, 
      icon: 'Clock', 
      colorClass: 'bg-green-500/10', 
      textColor: 'text-green-500', 
      chartColor: colorMap['text-green-500'],
      unit: '',
      routeName: 'admin-monitor'
    },
  ];
});

const activeGroupsData = computed(() => {
  if (!chatStats.value || !chatStats.value.group_stats_today) return [];
  const group_names = chatStats.value.group_names || {};
  const group_avatars = chatStats.value.group_avatars || {};
  
  return Object.entries(chatStats.value.group_stats_today)
    .map(([id, count]) => ({
      id,
      name: group_names[id] || id,
      avatar: getAvatarUrl(group_avatars[id] || ''),
      count: count as number
    }))
    .sort((a, b) => b.count - a.count);
});

const activeGroups = computed(() => activeGroupsData.value.length);

const dragonKingsData = computed(() => {
  if (!chatStats.value || !chatStats.value.user_stats_today) return [];
  const user_names = chatStats.value.user_names || {};
  const user_avatars = chatStats.value.user_avatars || {};

  return Object.entries(chatStats.value.user_stats_today)
    .map(([id, count]) => ({
      id,
      name: user_names[id] || id,
      avatar: getAvatarUrl(user_avatars[id] || ''),
      count: count as number
    }))
    .sort((a, b) => b.count - a.count);
});

const dragonKing = computed(() => {
  return dragonKingsData.value[0] || { name: t('mock_user'), count: 0 };
});

</script>

<template>
  <div class="p-4 sm:p-6 space-y-4 sm:space-y-8 relative min-h-[400px]">
    <!-- Loading Overlay -->
    <div v-if="loading" class="absolute inset-0 z-50 flex items-center justify-center bg-[var(--bg-body)]/50 backdrop-blur-sm rounded-3xl transition-opacity duration-300">
      <div class="flex flex-col items-center gap-4">
        <div class="relative w-16 h-16">
          <div class="absolute inset-0 border-4 border-[var(--matrix-color)]/20 rounded-full"></div>
          <div class="absolute inset-0 border-4 border-t-[var(--matrix-color)] rounded-full animate-spin"></div>
        </div>
        <p class="text-[var(--matrix-color)] font-black text-xs uppercase tracking-[0.2em] animate-pulse">{{ t('loading') }}</p>
      </div>
    </div>

    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
          <LayoutDashboard class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('dashboard') }}</h1>
          <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold tracking-widest uppercase">{{ t('dashboard_description') }}</p>
        </div>
      </div>
      
      <div class="flex items-center gap-3">
        <button 
          @click="refreshData" 
          class="flex items-center gap-2 px-6 py-2 rounded-xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20 disabled:opacity-50"
          :disabled="refreshing || loading"
        >
          <RefreshCw :class="{ 'animate-spin': refreshing }" class="w-4 h-4" />
          {{ t('refresh') }}
        </button>
      </div>
    </div>

    <!-- Error State (as a banner instead of replacing the whole page) -->
    <div v-if="fetchError && !loading" class="flex items-center justify-between p-4 rounded-2xl bg-red-500/10 border border-red-500/20">
      <div class="flex items-center gap-3">
        <X class="w-5 h-5 text-red-500" />
        <div>
          <h3 class="font-bold text-sm text-[var(--text-main)]">{{ t('load_error') }}</h3>
          <p class="text-xs text-[var(--text-muted)]">{{ t(fetchError) }}</p>
        </div>
      </div>
      <button 
        @click="initFetch" 
        class="px-4 py-2 rounded-xl bg-red-500 text-[var(--sidebar-text)] font-bold text-xs uppercase tracking-widest hover:opacity-90 transition-all"
      >
        {{ t('retry') }}
      </button>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 sm:gap-6">
      <router-link 
        v-for="card in statsCards" 
        :key="card.label"
        :to="card.routeName === 'admin-monitor' ? { name: 'admin-monitor', query: { type: card.id } } : { name: card.routeName }"
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
                <component :is="iconMap[card.icon]" class="w-4 h-4 sm:w-5 sm:h-5" :class="card.textColor" />
              </div>
              <div v-if="card.trend && card.trend.length > 0" class="flex flex-col items-end">
                <div class="w-12 h-6 sm:w-16 sm:h-8">
                  <LineChart :data="card.trend" :color="card.chartColor" />
                </div>
              </div>
            </div>
            <div>
              <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold uppercase tracking-widest mb-1">{{ t(card.label) }}</p>
              <div class="flex items-baseline gap-1">
                <span class="text-lg sm:text-xl font-black text-[var(--text-main)] mono tracking-tighter">{{ card.value }}</span>
                <span class="text-[8px] sm:text-[10px] text-[var(--text-muted)] font-bold uppercase">{{ card.unit.startsWith('unit_') ? t(card.unit) : card.unit }}</span>
              </div>
            </div>
          </div>
        </div>
      </router-link>
    </div>

    <!-- Charts & Lists -->
    <div v-if="authStore.isAdmin" class="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-6">
      <!-- System Activity Section -->
      <div v-if="authStore.isAdmin" class="lg:col-span-2 space-y-4 sm:space-y-6">
        <!-- Active Groups Today -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
          <div 
            @click="showActiveGroupsModal = true"
            class="bg-[var(--bg-card)] border border-[var(--border-color)] p-4 sm:p-6 rounded-3xl hover:shadow-xl transition-all cursor-pointer group"
          >
            <div class="flex items-center justify-between mb-4 sm:mb-6">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 sm:w-10 sm:h-10 rounded-xl bg-orange-500/10 flex items-center justify-center">
                  <Users class="w-4 h-4 sm:w-5 sm:h-5 text-orange-500" />
                </div>
                <div>
                  <h3 class="font-bold text-sm sm:text-base text-[var(--text-main)]">{{ t('active_groups_today') }}</h3>
                  <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest font-bold">{{ activeGroups }} {{ t('unit_active') }}</p>
                </div>
              </div>
              <button 
                @click.stop="showActiveGroupsModal = true"
                class="text-[var(--matrix-color)] text-[10px] sm:text-xs font-bold uppercase tracking-widest hover:underline flex items-center gap-1"
              >
                {{ t('view_all') }}
                <ChevronRight class="w-3 h-3" />
              </button>
            </div>

            <div class="space-y-2 sm:space-y-3">
              <div v-for="(group, idx) in activeGroupsData.slice(0, 5)" :key="group.id" class="flex items-center justify-between p-2 sm:p-3 rounded-2xl bg-black/5 dark:bg-white/5 group/item hover:bg-[var(--matrix-color)]/5 transition-colors">
                <div class="flex items-center gap-2 sm:gap-3">
                  <div class="relative">
                    <div class="w-8 h-8 sm:w-10 sm:h-10 rounded-xl overflow-hidden bg-[var(--matrix-color)]/10 flex items-center justify-center border-2 border-transparent group-hover/item:border-[var(--matrix-color)]/30 transition-all">
                      <template v-if="group.avatar && !isPlatformAvatar(group.avatar)">
                        <img :src="group.avatar" class="w-full h-full object-cover" />
                      </template>
                      <template v-else>
                        <component 
                          :is="isPlatformAvatar(group.avatar) ? getPlatformIcon(getPlatformFromAvatar(group.avatar)) : Users" 
                          :class="['w-4 h-4 sm:w-5 sm:h-5', isPlatformAvatar(group.avatar) ? getPlatformColor(getPlatformFromAvatar(group.avatar)) : 'text-[var(--matrix-color)]']" 
                        />
                      </template>
                    </div>
                    <div class="absolute -top-1 -left-1 w-4 h-4 sm:w-5 sm:h-5 rounded-lg bg-[var(--matrix-color)] flex items-center justify-center font-bold text-black text-[8px] sm:text-[10px] shadow-lg">
                      {{ idx + 1 }}
                    </div>
                  </div>
                  <span class="text-[var(--text-main)] text-xs sm:text-sm font-medium truncate max-w-[100px] sm:max-w-[150px]">{{ group.name }}</span>
                </div>
                <div class="flex items-center gap-1 sm:gap-2">
                  <span class="text-[var(--text-main)] font-bold mono text-xs sm:text-sm">{{ group.count }}</span>
                  <span class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase font-bold">{{ t('unit_msg') }}</span>
                </div>
              </div>
              
              <div v-if="activeGroupsData.length === 0" class="py-6 sm:py-8 text-center">
                <p class="text-[var(--text-muted)] text-[10px] sm:text-xs">{{ t('no_data') }}</p>
              </div>
            </div>
          </div>

          <!-- Today's Dragon King -->
          <div 
            @click="showDragonKingModal = true"
            class="bg-[var(--bg-card)] border border-[var(--border-color)] p-4 sm:p-6 rounded-3xl hover:shadow-xl transition-all cursor-pointer group"
          >
            <div class="flex items-center justify-between mb-4 sm:mb-6">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 sm:w-10 sm:h-10 rounded-xl bg-yellow-500/10 flex items-center justify-center">
                  <Bot class="w-4 h-4 sm:w-5 sm:h-5 text-yellow-500" />
                </div>
                <div>
                  <h3 class="font-bold text-sm sm:text-base text-[var(--text-main)]">{{ t('today_dragon_king') }}</h3>
                  <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest font-bold">{{ t('most_active_user') }}</p>
                </div>
              </div>
              <button 
                @click.stop="showDragonKingModal = true"
                class="text-[var(--matrix-color)] text-[10px] sm:text-xs font-bold uppercase tracking-widest hover:underline flex items-center gap-1"
              >
                {{ t('view_all') }}
                <ChevronRight class="w-3 h-3" />
              </button>
            </div>

            <div class="space-y-2 sm:space-y-3">
              <div v-for="(user, idx) in dragonKingsData.slice(0, 5)" :key="user.id" class="flex items-center justify-between p-2 sm:p-3 rounded-2xl bg-black/5 dark:bg-white/5 group/item hover:bg-yellow-500/5 transition-colors">
                <div class="flex items-center gap-2 sm:gap-3">
                  <div class="relative">
                    <div class="w-8 h-8 sm:w-10 sm:h-10 rounded-xl overflow-hidden bg-yellow-500/10 flex items-center justify-center border-2 border-transparent group-hover/item:border-yellow-500/30 transition-all">
                      <template v-if="user.avatar && !isPlatformAvatar(user.avatar)">
                        <img :src="user.avatar" class="w-full h-full object-cover" />
                      </template>
                      <template v-else>
                        <component 
                          :is="isPlatformAvatar(user.avatar) ? getPlatformIcon(getPlatformFromAvatar(user.avatar)) : Bot" 
                          :class="['w-4 h-4 sm:w-5 sm:h-5', isPlatformAvatar(user.avatar) ? getPlatformColor(getPlatformFromAvatar(user.avatar)) : 'text-yellow-500']" 
                        />
                      </template>
                    </div>
                    <div class="absolute -top-1 -left-1 w-4 h-4 sm:w-5 sm:h-5 rounded-lg bg-yellow-500 flex items-center justify-center font-bold text-black text-[8px] sm:text-[10px] shadow-lg">
                      {{ idx === 0 ? '👑' : idx + 1 }}
                    </div>
                  </div>
                  <span class="text-[var(--text-main)] text-xs sm:text-sm font-medium truncate max-w-[100px] sm:max-w-[150px]">{{ user.name }}</span>
                </div>
                <div class="flex items-center gap-1 sm:gap-2">
                  <span class="text-[var(--text-main)] font-bold mono text-xs sm:text-sm">{{ user.count }}</span>
                  <span class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase font-bold">{{ t('unit_msg') }}</span>
                </div>
              </div>
              
              <div v-if="dragonKingsData.length === 0" class="py-6 sm:py-8 text-center">
                <p class="text-[var(--text-muted)] text-[10px] sm:text-xs">{{ t('no_data') }}</p>
              </div>
            </div>
          </div>
        </div>

      <!-- Detail Modals -->
      <Teleport to="body">
        <template v-if="authStore.isAdmin">
          <!-- Active Groups Modal -->
          <div v-if="showActiveGroupsModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6 bg-black/60 backdrop-blur-sm" @click="showActiveGroupsModal = false">
            <div class="w-full max-w-lg bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200" @click.stop>
              <div class="p-6 sm:p-8 border-b border-[var(--border-color)] flex items-center justify-between">
                <div>
                  <h3 class="text-xl font-black text-[var(--text-main)] flex items-center gap-3 uppercase tracking-tight">
                    <Users class="w-6 h-6 text-[var(--matrix-color)]" />
                    {{ t('active_groups_detail') }}
                  </h3>
                  <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('realtime_group_activity') }}</p>
                </div>
                <button @click="showActiveGroupsModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                  <X class="w-6 h-6 text-[var(--text-muted)]" />
                </button>
              </div>
              <div class="p-6 sm:p-8 max-h-[60vh] overflow-y-auto custom-scrollbar">
                <div v-if="activeGroupsData.length === 0" class="text-center py-12 text-[var(--text-muted)]">
                  <Database class="w-12 h-12 mx-auto mb-4 opacity-20" />
                  <p class="text-xs font-bold uppercase tracking-widest">{{ t('no_leaderboard_data') }}</p>
                </div>
                <div v-else class="space-y-4">
                  <div v-for="(group, idx) in activeGroupsData" :key="group.id" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--border-color)] transition-all">
                    <div class="flex items-center gap-4">
                      <div class="relative">
                        <div class="w-12 h-12 rounded-xl overflow-hidden bg-[var(--matrix-color)]/10 flex items-center justify-center border-2 border-transparent hover:border-[var(--matrix-color)]/30 transition-all">
                          <template v-if="group.avatar && !isPlatformAvatar(group.avatar)">
                            <img :src="group.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            <component 
                              :is="isPlatformAvatar(group.avatar) ? getPlatformIcon(getPlatformFromAvatar(group.avatar)) : Users" 
                              :class="['w-6 h-6', isPlatformAvatar(group.avatar) ? getPlatformColor(getPlatformFromAvatar(group.avatar)) : 'text-[var(--matrix-color)]']" 
                            />
                          </template>
                        </div>
                        <div class="absolute -top-2 -left-2 w-6 h-6 rounded-lg bg-[var(--matrix-color)] flex items-center justify-center font-black text-black text-xs shadow-xl">
                          #{{ idx + 1 }}
                        </div>
                      </div>
                      <div>
                        <p class="font-bold text-[var(--text-main)] text-sm">{{ group.name }}</p>
                        <p class="text-[10px] text-[var(--text-muted)] font-mono opacity-60">{{ group.id }}</p>
                      </div>
                    </div>
                    <div class="text-right">
                      <p class="font-black text-lg text-[var(--text-main)] leading-none">{{ group.count }}</p>
                      <p class="text-[10px] text-[var(--text-muted)] uppercase font-bold tracking-widest">{{ t('unit_msg') }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Dragon King Modal -->
          <div v-if="showDragonKingModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6 bg-black/60 backdrop-blur-sm" @click="showDragonKingModal = false">
            <div class="w-full max-w-lg bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200" @click.stop>
              <div class="p-6 sm:p-8 border-b border-[var(--border-color)] flex items-center justify-between">
                <div>
                  <h3 class="text-xl font-black text-[var(--text-main)] flex items-center gap-3 uppercase tracking-tight">
                    <Activity class="w-6 h-6 text-[var(--matrix-color)]" />
                    {{ t('dragon_king_detail') }}
                  </h3>
                  <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest mt-1">{{ t('top_active_users_today') }}</p>
                </div>
                <button @click="showDragonKingModal = false" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                  <X class="w-6 h-6 text-[var(--text-muted)]" />
                </button>
              </div>
              <div class="p-6 sm:p-8 max-h-[60vh] overflow-y-auto custom-scrollbar">
                <div v-if="dragonKingsData.length === 0" class="text-center py-12 text-[var(--text-muted)]">
                  <Database class="w-12 h-12 mx-auto mb-4 opacity-20" />
                  <p class="text-xs font-bold uppercase tracking-widest">{{ t('no_leaderboard_data') }}</p>
                </div>
                <div v-else class="space-y-4">
                  <div v-for="(user, idx) in dragonKingsData" :key="user.id" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--border-color)] transition-all">
                    <div class="flex items-center gap-4">
                      <div class="relative">
                        <div class="w-12 h-12 rounded-xl overflow-hidden bg-yellow-500/10 flex items-center justify-center border-2 border-transparent hover:border-yellow-500/30 transition-all">
                          <template v-if="user.avatar && !isPlatformAvatar(user.avatar)">
                            <img :src="user.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            <component 
                              :is="isPlatformAvatar(user.avatar) ? getPlatformIcon(getPlatformFromAvatar(user.avatar)) : Bot" 
                              :class="['w-6 h-6', isPlatformAvatar(user.avatar) ? getPlatformColor(getPlatformFromAvatar(user.avatar)) : 'text-yellow-500']" 
                            />
                          </template>
                        </div>
                        <div class="absolute -top-2 -left-2 w-6 h-6 rounded-lg bg-yellow-500 flex items-center justify-center font-black text-black text-xs shadow-xl">
                          {{ idx === 0 ? '👑' : `#${idx + 1}` }}
                        </div>
                      </div>
                      <div>
                        <p class="font-bold text-[var(--text-main)] text-sm">{{ user.name }}</p>
                        <p class="text-[10px] text-[var(--text-muted)] font-mono opacity-60">{{ user.id }}</p>
                      </div>
                    </div>
                    <div class="text-right">
                      <p class="font-black text-lg text-[var(--text-main)] leading-none">{{ user.count }}</p>
                      <p class="text-[10px] text-[var(--text-muted)] uppercase font-bold tracking-widest">{{ t('unit_msg') }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </Teleport>

      <!-- Message Chart -->
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-4 sm:space-y-6">
          <div class="flex items-center justify-between">
            <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
              <BarChart3 class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('throughput') }}
            </h3>
            <div class="flex gap-2">
              <button class="px-3 py-1 text-[10px] font-bold rounded-lg bg-[var(--matrix-color)] text-black">{{ t('last_hour') }}</button>
              <button class="px-3 py-1 text-[10px] font-bold rounded-lg bg-black/10 dark:bg-white/5 text-[var(--text-muted)]">{{ t('last_24h') }}</button>
            </div>
          </div>
          <div class="h-48 sm:h-64 w-full">
            <LineChart :data="botStore.stats.msg_trend" color="#10b981" />
          </div>
        </div>

        <!-- Latest Messages -->
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] space-y-4 sm:space-y-6">
          <div class="flex items-center justify-between">
            <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
              <MessageSquare class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('recent_messages') }}
            </h3>
            <router-link to="/messages" class="text-xs font-bold text-[var(--matrix-color)] hover:underline">{{ t('view_all') }}</router-link>
          </div>
          
          <div class="space-y-3">
            <div v-if="!authStore.isAdmin" class="py-12 flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
              <Activity class="w-8 h-8" />
              <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('admin_required') }}</p>
            </div>
            <div v-else-if="recentMessages.length === 0" class="py-12 flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
              <MessageSquare class="w-8 h-8" />
              <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('no_messages') }}</p>
            </div>
            <div 
              v-else
              v-for="msg in recentMessages" 
              :key="msg.id"
              class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--border-color)] transition-all group"
            >
              <div class="flex items-start gap-4">
                <div class="w-10 h-10 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center flex-shrink-0 overflow-hidden">
                  <template v-if="msg.user_avatar && !isPlatformAvatar(msg.user_avatar)">
                    <img :src="msg.user_avatar" class="w-full h-full object-cover" />
                  </template>
                  <template v-else>
                    <component 
                      :is="isPlatformAvatar(msg.user_avatar) ? getPlatformIcon(getPlatformFromAvatar(msg.user_avatar)) : Users" 
                      :class="['w-5 h-5', isPlatformAvatar(msg.user_avatar) ? getPlatformColor(getPlatformFromAvatar(msg.user_avatar)) : 'text-[var(--matrix-color)]']" 
                    />
                  </template>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between mb-1">
                    <p class="font-bold text-[var(--text-main)] truncate">{{ msg.user_name || msg.user_id }}</p>
                    <p class="text-[10px] text-[var(--text-muted)] mono">{{ new Date(msg.created_at).toLocaleTimeString() }}</p>
                  </div>
                  <p class="text-xs text-[var(--text-muted)] truncate">{{ msg.content }}</p>
                  <div class="mt-2 flex items-center gap-2">
                    <span class="text-[10px] px-2 py-0.5 rounded-md bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] font-bold uppercase tracking-tight">
                      {{ msg.bot_id }}
                    </span>
                    <span v-if="msg.group_id && msg.group_id !== '0'" class="text-[10px] px-2 py-0.5 rounded-md bg-blue-500/10 text-blue-500 font-bold uppercase tracking-tight">
                      {{ msg.group_name || msg.group_id }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Logs -->
      <div class="space-y-4 sm:space-y-6 flex flex-col">
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex flex-col h-[400px]">
          <div class="flex items-center justify-between mb-4 sm:mb-6">
            <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
              <Terminal class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('recent_logs') }}
            </h3>
            <router-link to="/logs" class="text-xs font-bold text-[var(--matrix-color)] hover:underline">{{ t('view_all') }}</router-link>
          </div>
          
          <div class="flex-1 space-y-4 overflow-y-auto custom-scrollbar pr-2">
            <div v-if="!authStore.isAdmin" class="h-full flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
              <Activity class="w-8 h-8" />
              <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('admin_required') }}</p>
            </div>
            <div v-else-if="recentLogs.length === 0" class="h-full flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
              <Database class="w-8 h-8" />
              <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('status_loading') }}...</p>
            </div>
            <div 
              v-else
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

        <!-- Top Processes -->
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex flex-col shadow-sm">
          <div class="flex items-center justify-between mb-4 sm:mb-6 border-b border-[var(--border-color)] pb-4">
            <h3 class="text-base sm:text-lg font-black flex items-center gap-3 text-[var(--text-main)] uppercase tracking-tight">
              <Activity class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('top_processes') }}
            </h3>
            <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-black uppercase tracking-widest">
              <Cpu class="w-3 h-3" />
              SYSTEM_TOP
            </div>
          </div>

          <div class="space-y-4">
            <div v-if="!botStore.stats.top_processes || botStore.stats.top_processes.length === 0" class="py-12 text-center">
              <Terminal class="w-12 h-12 mx-auto mb-4 text-[var(--text-muted)] opacity-20" />
              <p class="text-xs font-bold uppercase tracking-widest text-[var(--text-muted)]">{{ t('no_data') }}</p>
            </div>
            <div 
              v-for="(proc, idx) in botStore.stats.top_processes" 
              :key="proc.pid"
              class="group relative"
            >
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-3">
                  <span class="text-[10px] font-black mono text-[var(--matrix-color)] opacity-50">{{ (idx + 1).toString().padStart(2, '0') }}</span>
                  <p class="text-xs font-black text-[var(--text-main)] truncate max-w-[120px]">{{ proc.name }}</p>
                  <span class="text-[9px] px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/5 text-[var(--text-muted)] mono font-bold">PID:{{ proc.pid }}</span>
                </div>
                <div class="text-right">
                  <span class="text-xs font-black text-[var(--matrix-color)] mono">{{ proc.cpu.toFixed(1) }}%</span>
                </div>
              </div>
              <div class="w-full h-1.5 rounded-full bg-black/5 dark:bg-white/5 overflow-hidden border border-[var(--border-color)]">
                <div 
                  class="h-full bg-[var(--matrix-color)] transition-all duration-1000 ease-out" 
                  :style="{ width: Math.min(proc.cpu, 100) + '%' }"
                ></div>
              </div>
              <div class="flex justify-between mt-1 px-1">
                 <span class="text-[8px] font-bold text-[var(--text-muted)] uppercase tracking-tighter">Usage</span>
                 <span class="text-[8px] font-bold text-[var(--text-muted)] mono">{{ (proc.memory / 1024 / 1024).toFixed(1) }} MB</span>
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
