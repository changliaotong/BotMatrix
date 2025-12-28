<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
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

const recentLogs = ref<any[]>([]);
const recentMessages = ref<any[]>([]);
const chatStats = ref<any>(null);
const currentTime = ref(new Date().toLocaleTimeString());
let timeTimer: number | null = null;
let refreshTimer: number | null = null;

onMounted(() => {
  // Initial fetch
  const initFetch = async () => {
    try {
      await Promise.allSettled([
        botStore.fetchStats(),
        botStore.fetchBots(),
        fetchLogs(),
        fetchMessages()
      ]);
      
      const cs = await botStore.fetchChatStats();
      if (cs) chatStats.value = cs;
    } catch (err) {
      console.error('Initial fetch failed:', err);
    }
  };
  
  initFetch();
  
  timeTimer = window.setInterval(() => {
    currentTime.value = new Date().toLocaleTimeString();
  }, 1000);
});

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

const refreshing = ref(false);

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

const fetchMessages = async () => {
  try {
    const messages = await botStore.fetchMessages(10);
    recentMessages.value = messages;
  } catch (err) {
    console.error('Failed to fetch messages in dashboard:', err);
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

const showActiveGroupsModal = ref(false);
const showDragonKingModal = ref(false);

const statsCards = computed(() => [
  { 
    id: 'bots',
    label: 'total_bots', 
    value: botStore.stats.bot_count || '0', 
    icon: 'Bot', 
    colorClass: 'bg-blue-500/10', 
    textColor: 'text-blue-500', 
    unit: 'unit_active',
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
    value: (botStore.stats.cpu_usage || 0).toFixed(1) + '%', 
    icon: 'Cpu', 
    colorClass: 'bg-orange-500/10', 
    textColor: 'text-orange-500', 
    unit: 'unit_live',
    trend: botStore.stats.cpu_trend || [],
    routeName: 'monitor'
  },
  { 
    id: 'memory',
    label: 'memory_usage', 
    value: (botStore.stats.memory_used_percent || 0).toFixed(1) + '%', 
    icon: 'Database', 
    colorClass: 'bg-blue-500/10', 
    textColor: 'text-blue-500', 
    unit: 'unit_live',
    trend: botStore.stats.mem_trend || [],
    routeName: 'monitor'
  },
  { 
    id: 'throughput',
    label: 'throughput', 
    value: botStore.stats.message_count || '0', 
    icon: 'Activity', 
    colorClass: 'bg-pink-500/10', 
    textColor: 'text-pink-500', 
    unit: 'unit_msg',
    trend: botStore.stats.msg_trend || [],
    routeName: 'monitor'
  },
  { 
    id: 'time',
    label: 'current_time', 
    value: currentTime.value, 
    icon: 'Clock', 
    colorClass: 'bg-green-500/10', 
    textColor: 'text-green-500', 
    unit: '',
    routeName: 'monitor'
  },
]);

const activeGroupsData = computed(() => {
  if (!chatStats.value || !chatStats.value.group_stats_today) return [];
  return Object.entries(chatStats.value.group_stats_today)
    .map(([id, count]) => ({
      id,
      name: chatStats.value.group_names[id] || id,
      avatar: getAvatarUrl(chatStats.value.group_avatars ? chatStats.value.group_avatars[id] : ''),
      count: count as number
    }))
    .sort((a, b) => b.count - a.count);
});

const activeGroups = computed(() => activeGroupsData.value.length);

const dragonKingsData = computed(() => {
  if (!chatStats.value || !chatStats.value.user_stats_today) return [];
  return Object.entries(chatStats.value.user_stats_today)
    .map(([id, count]) => ({
      id,
      name: chatStats.value.user_names[id] || id,
      avatar: getAvatarUrl(chatStats.value.user_avatars ? chatStats.value.user_avatars[id] : ''),
      count: count as number
    }))
    .sort((a, b) => b.count - a.count);
});

const dragonKing = computed(() => {
  return dragonKingsData.value[0] || { name: t('mock_user'), count: 0 };
});

</script>

<template>
  <div class="p-4 sm:p-6 space-y-4 sm:space-y-8">
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
          :disabled="refreshing"
        >
          <RefreshCw :class="{ 'animate-spin': refreshing }" class="w-4 h-4" />
          {{ t('refresh') }}
        </button>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 sm:gap-6">
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
            </div>
            <div>
              <p class="text-[10px] sm:text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider">{{ t(card.label) }}</p>
              <div class="flex items-baseline gap-1 sm:gap-2">
                <h2 class="text-2xl sm:text-3xl font-black text-[var(--text-main)] tracking-tight">{{ card.value }}</h2>
                <span class="text-[8px] sm:text-[10px] font-black text-[var(--text-muted)]">{{ card.unit.startsWith('unit_') ? t(card.unit) : card.unit }}</span>
              </div>
              <!-- Trend Curve -->
              <div v-if="card.trend" class="mt-2 h-12 w-full opacity-40 group-hover:opacity-100 transition-opacity">
                <LineChart 
                  :data="card.trend" 
                  :color="card.textColor.includes('orange') ? '#f97316' : card.textColor.includes('pink') ? '#ec4899' : card.textColor.includes('blue') ? '#3b82f6' : '#00ff41'" 
                  :fill="true" 
                />
              </div>
            </div>
          </div>
        </div>
      </router-link>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-6">
      <!-- Main Content Area -->
      <div class="lg:col-span-2 space-y-4 sm:space-y-6">
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
            <LineChart :data="botStore.stats.msg_trend" />
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
            <div v-if="recentMessages.length === 0" class="py-12 flex flex-col items-center justify-center text-[var(--text-muted)] space-y-2 opacity-50">
              <MessageSquare class="w-8 h-8" />
              <p class="text-xs font-bold mono uppercase tracking-widest">{{ t('no_messages') }}</p>
            </div>
            <div 
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

        <!-- Top Processes -->
        <div class="p-4 sm:p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] flex flex-col">
          <div class="flex items-center justify-between mb-4 sm:mb-6">
            <h3 class="text-base sm:text-lg font-bold flex items-center gap-3">
              <Activity class="w-5 h-5 text-orange-500" /> {{ t('top_processes') }}
            </h3>
            <div class="p-2 rounded-xl bg-orange-500/10 text-orange-500">
              <Cpu class="w-4 h-4" />
            </div>
          </div>

          <div class="space-y-3">
            <div v-if="!botStore.stats.top_processes || botStore.stats.top_processes.length === 0" class="py-8 text-center text-[var(--text-muted)]">
              <p class="text-xs font-bold uppercase tracking-widest">{{ t('no_data') }}</p>
            </div>
            <div 
              v-for="proc in botStore.stats.top_processes" 
              :key="proc.pid"
              class="p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--border-color)] transition-all group"
            >
              <div class="flex items-center justify-between">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 mb-1">
                    <span class="text-[10px] font-mono text-[var(--text-muted)]">#{{ proc.pid }}</span>
                    <p class="text-xs font-bold text-[var(--text-main)] truncate">{{ proc.name }}</p>
                  </div>
                  <div class="flex items-center gap-3">
                    <div class="flex items-center gap-1">
                      <div class="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                      <span class="text-[10px] font-bold text-[var(--text-muted)]">CPU: {{ proc.cpu.toFixed(1) }}%</span>
                    </div>
                    <div class="flex items-center gap-1">
                      <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                      <span class="text-[10px] font-bold text-[var(--text-muted)]">MEM: {{ (proc.memory / 1024 / 1024).toFixed(1) }} MB</span>
                    </div>
                  </div>
                </div>
                <div class="w-12 h-1.5 rounded-full bg-black/10 dark:bg-white/10 overflow-hidden">
                  <div 
                    class="h-full bg-orange-500 transition-all duration-500" 
                    :style="{ width: Math.min(proc.cpu, 100) + '%' }"
                  ></div>
                </div>
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
