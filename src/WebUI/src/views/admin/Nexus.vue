<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Network, 
  Share2, 
  Activity, 
  RefreshCw, 
  Shield, 
  Zap,
  ArrowUpRight,
  ArrowDownRight,
  Server,
  Link,
  Bot
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const connections = ref<any[]>([]);
const nexusStats = ref<any>(null);
const strategies = ref<any[]>([]);
const shadowRules = ref<any[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
let refreshTimer: number | null = null;

const fetchNexus = async () => {
  error.value = null;
  loading.value = true;
  try {
    const [statusData, strategiesData, shadowData] = await Promise.all([
      botStore.fetchNexusStatus(),
      botStore.fetchStrategies(),
      botStore.fetchShadowRules()
    ]);
    
    if (statusData.success) {
      connections.value = statusData.connections || [];
      nexusStats.value = statusData.stats;
    }
    if (strategiesData.success) {
      strategies.value = strategiesData.data || [];
    }
    if (shadowData.success) {
      shadowRules.value = shadowData.data || [];
    }
  } catch (err: any) {
    console.error('Failed to fetch nexus data:', err);
    error.value = err.message || 'Failed to fetch nexus data';
  } finally {
    loading.value = false;
  }
};

const toggleStrategy = async (strategy: any) => {
  try {
    strategy.IsEnabled = !strategy.IsEnabled;
    await botStore.saveStrategy(strategy);
    await fetchNexus();
  } catch (err) {
    console.error('Failed to toggle strategy:', err);
  }
};

onMounted(() => {
  fetchNexus();
  refreshTimer = window.setInterval(fetchNexus, 5000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('nexus') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('nexus_desc') }}</p>
      </div>
      <button 
        @click="fetchNexus"
        class="p-3 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all group"
      >
        <RefreshCw class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" :class="{ 'animate-spin': loading }" />
      </button>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center gap-3 text-red-500">
      <Shield class="w-5 h-5" />
      <div class="flex-1">
        <p class="text-xs font-bold uppercase tracking-widest">{{ t('error') }}</p>
        <p class="text-sm font-black">{{ error }}</p>
      </div>
      <button @click="fetchNexus" class="px-3 py-1 rounded-lg bg-red-500 text-[var(--sidebar-text)] text-[10px] font-black uppercase tracking-widest">
        {{ t('retry') }}
      </button>
    </div>

    <!-- Nexus Visualization (Simplified) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Status Card -->
      <div class="lg:col-span-2 p-8 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] relative overflow-hidden group">
        <div class="absolute -right-10 -top-10 w-64 h-64 bg-[var(--matrix-color)]/5 rounded-full blur-3xl group-hover:bg-[var(--matrix-color)]/10 transition-all duration-700"></div>
        
        <div class="relative z-10 flex flex-col h-full justify-between gap-12">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-4">
              <div class="p-4 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)]">
                <Share2 class="w-8 h-8" />
              </div>
              <div>
                <h2 class="text-2xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('botnexus_core') }}</h2>
                <div class="flex items-center gap-2 mt-1">
                  <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                  <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('gateway_operational') }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
            <div class="space-y-2">
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('active_links') }}</p>
              <p class="text-4xl font-black text-[var(--text-main)]">{{ connections.length }}</p>
            </div>
            <div class="space-y-2">
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('bot_online') }}</p>
              <p class="text-4xl font-black text-blue-500">{{ nexusStats?.online_bots || 0 }}</p>
            </div>
            <div class="space-y-2">
              <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('worker_online') }}</p>
              <p class="text-4xl font-black text-purple-500">{{ nexusStats?.online_workers || 0 }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions/Info -->
      <div class="p-6 rounded-3xl bg-[var(--matrix-color)] text-black space-y-6">
        <h3 class="font-black uppercase tracking-widest text-sm">{{ t('security_layer') }}</h3>
        <div class="space-y-4">
          <div class="flex items-center justify-between p-3 rounded-2xl bg-black/10 border border-black/10">
            <div class="flex items-center gap-3">
              <Shield class="w-5 h-5" />
              <span class="text-xs font-bold uppercase">{{ t('ssl_wss') }}</span>
            </div>
            <span class="text-[10px] font-black bg-black text-[var(--sidebar-text)] px-2 py-0.5 rounded-md">{{ t('encrypted') }}</span>
          </div>
          <div class="flex items-center justify-between p-3 rounded-2xl bg-black/10 border border-black/10">
            <div class="flex items-center gap-3">
              <Zap class="w-5 h-5" />
              <span class="text-xs font-bold uppercase">{{ t('turbo_relay') }}</span>
            </div>
            <span class="text-[10px] font-black bg-black text-[var(--sidebar-text)] px-2 py-0.5 rounded-md">{{ t('enabled_caps') }}</span>
          </div>
        </div>
        <p class="text-[10px] font-bold opacity-60 uppercase tracking-widest leading-relaxed">
          {{ t('nexus_description') }}
        </p>
      </div>
    </div>

    <!-- Connections Table -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Strategies -->
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <h3 class="font-black text-[var(--text-main)] uppercase tracking-widest text-sm mb-6 flex items-center gap-2">
          <Shield class="w-4 h-4 text-[var(--matrix-color)]" /> {{ t('global_strategies') }}
        </h3>
        <div class="space-y-4">
          <div v-for="s in strategies" :key="s.ID" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div>
              <p class="font-black text-sm text-[var(--text-main)] uppercase">{{ t(s.Name) || s.Name }}</p>
              <p class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-wider">{{ t(s.Type) || s.Type }}</p>
            </div>
            <button 
              @click="toggleStrategy(s)"
              :class="s.IsEnabled ? 'bg-[var(--matrix-color)] text-black' : 'bg-red-500/10 text-red-500'"
              class="px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all"
            >
              {{ s.IsEnabled ? t('enabled') : t('disabled') }}
            </button>
          </div>
          <div v-if="strategies.length === 0" class="text-center py-8 text-[var(--text-muted)] text-[10px] font-bold uppercase">
            {{ t('no_strategies') }}
          </div>
        </div>
      </div>

      <!-- Shadow Rules -->
      <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
        <h3 class="font-black text-[var(--text-main)] uppercase tracking-widest text-sm mb-6 flex items-center gap-2">
          <Activity class="w-4 h-4 text-[var(--matrix-color)]" /> {{ t('shadow_rules') }}
        </h3>
        <div class="space-y-4">
          <div v-for="r in shadowRules" :key="r.ID" class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
            <div class="flex items-center justify-between mb-2">
              <p class="font-black text-sm text-[var(--text-main)] uppercase">{{ t(r.Name) || r.Name }}</p>
              <span class="text-[10px] font-black px-2 py-0.5 rounded bg-[var(--matrix-color)]/20 text-[var(--matrix-color)]">{{ r.Ratio * 100 }}% {{ t('traffic') }}</span>
            </div>
            <p class="text-[10px] text-[var(--text-muted)] font-bold uppercase truncate">{{ r.TargetWorkerID }}</p>
          </div>
          <div v-if="shadowRules.length === 0" class="text-center py-8 text-[var(--text-muted)] text-[10px] font-bold uppercase">
            {{ t('no_shadow_rules') }}
          </div>
        </div>
      </div>
    </div>

    <!-- Connections Table -->
    <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] overflow-hidden">
      <div class="flex items-center justify-between mb-6">
        <h3 class="font-black text-[var(--text-main)] uppercase tracking-widest text-sm flex items-center gap-2">
          <Link class="w-4 h-4 text-[var(--matrix-color)]" /> {{ t('active_connections') }}
        </h3>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)]">
              <th class="pb-4 px-4">{{ t('entity_id') }}</th>
              <th class="pb-4 px-4">{{ t('type') }}</th>
              <th class="pb-4 px-4">{{ t('remote_address') }}</th>
              <th class="pb-4 px-4">{{ t('uptime') }}</th>
              <th class="pb-4 px-4">{{ t('status') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--border-color)]">
            <tr v-for="conn in connections" :key="conn.id" class="group hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <td class="py-4 px-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center">
                    <Server v-if="conn.type === 'worker'" class="w-4 h-4 text-purple-500" />
                    <Bot v-else class="w-4 h-4 text-blue-500" />
                  </div>
                  <span class="font-black text-sm text-[var(--text-main)]">{{ conn.id }}</span>
                </div>
              </td>
              <td class="py-4 px-4">
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t(conn.type) }}</span>
              </td>
              <td class="py-4 px-4 font-mono text-xs text-[var(--text-muted)]">
                {{ conn.remote_addr }}
              </td>
              <td class="py-4 px-4 font-mono text-xs text-[var(--text-main)]">
                {{ conn.uptime }}
              </td>
              <td class="py-4 px-4">
                <span class="px-2 py-0.5 rounded-md text-[8px] font-black bg-green-500/10 text-green-500 border border-green-500/20 uppercase tracking-widest">{{ t('active') }}</span>
              </td>
            </tr>
            <tr v-if="connections.length === 0">
              <td colspan="5" class="py-12 text-center text-[var(--text-muted)] font-bold uppercase tracking-widest text-xs">
                {{ t('no_connections') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>