<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useMeowStore } from '@/stores/earlymeow';
import { useSystemStore } from '@/stores/system';
import { t } from '@/utils/i18n';
import { 
  Activity, Settings, MessageSquare, Power, 
  BarChart3, Plus, Bell, RefreshCw, 
  Play, Pause, Trash2, Edit3, Search
} from 'lucide-vue-next';

const meowStore = useMeowStore();
const systemStore = useSystemStore();

const tt = (key: string, defaultText?: string) => {
  const res = t(key);
  return res === key ? (defaultText || key) : res;
};

const bots = ref([
  { id: 1, name: tt('portal.喵喵助手01'), status: 'online', messages: 1240, mode: 'gentle' },
  { id: 2, name: tt('meow.slacker'), status: 'offline', messages: 856, mode: 'focus' },
  { id: 3, name: tt('meow.late_night_radio'), status: 'online', messages: 210, mode: 'sleep' }
]);

const activeTab = ref('dashboard');

onMounted(() => {
  meowStore.init();
});
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto min-h-screen relative z-10">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 mb-12">
      <div class="space-y-1">
        <h1 class="text-3xl font-black tracking-tight text-[var(--text-main)]">{{ tt('earlymeow.console.header.title') }}</h1>
        <p class="text-sm text-[var(--text-muted)] font-medium">{{ tt('earlymeow.console.header.desc') }}</p>
      </div>
      
      <div class="flex items-center gap-4">
        <button class="p-3 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all relative group">
          <Bell class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
          <div class="absolute top-2.5 right-2.5 w-2 h-2 bg-[var(--matrix-color)] rounded-full shadow-[0_0_10px_var(--matrix-color)]"></div>
        </button>
        <button class="px-6 py-3 rounded-xl bg-[var(--matrix-color)] text-white font-black flex items-center gap-2 hover:opacity-90 hover:shadow-[0_0_20px_var(--matrix-color)]/40 transition-all">
          <Plus class="w-5 h-5" />
          {{ tt('earlymeow.console.header.add_bot') }}
        </button>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
      <div class="p-8 rounded-[32px] bg-[var(--bg-body)]/40 border border-[var(--border-color)] backdrop-blur-xl space-y-4 hover:border-[var(--matrix-color)]/20 transition-all group">
        <div class="flex justify-between items-start">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 group-hover:scale-110 transition-transform">
            <MessageSquare class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--matrix-color)] bg-[var(--matrix-color)]/10 px-2 py-1 rounded-full">+12.5%</div>
        </div>
        <div>
          <div class="text-4xl font-black text-[var(--text-main)]">2,405</div>
          <div class="text-xs font-bold text-[var(--text-muted)]/60 uppercase tracking-widest">{{ tt('earlymeow.console.stats.messages') }}</div>
        </div>
      </div>

      <div class="p-8 rounded-[32px] bg-[var(--bg-body)]/40 border border-[var(--border-color)] backdrop-blur-xl space-y-4 hover:border-[var(--matrix-color)]/20 transition-all group">
        <div class="flex justify-between items-start">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 group-hover:scale-110 transition-transform">
            <Activity class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--matrix-color)] bg-[var(--matrix-color)]/10 px-2 py-1 rounded-full">99.9%</div>
        </div>
        <div>
          <div class="text-4xl font-black text-[var(--text-main)]">12.4h</div>
          <div class="text-xs font-bold text-[var(--text-muted)]/60 uppercase tracking-widest">{{ tt('earlymeow.console.stats.online') }}</div>
        </div>
      </div>

      <div class="p-8 rounded-[32px] bg-[var(--bg-body)]/40 border border-[var(--border-color)] backdrop-blur-xl space-y-4 hover:border-[var(--matrix-color)]/20 transition-all group">
        <div class="flex justify-between items-start">
          <div class="p-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20 group-hover:scale-110 transition-transform">
            <RefreshCw class="w-6 h-6" />
          </div>
          <div class="text-[10px] font-black text-[var(--matrix-color)] bg-[var(--matrix-color)]/10 px-2 py-1 rounded-full">Active</div>
        </div>
        <div>
          <div class="text-4xl font-black text-[var(--text-main)]">2.4.0</div>
          <div class="text-xs font-bold text-[var(--text-muted)]/60 uppercase tracking-widest">{{ tt('earlymeow.console.stats.version') }}</div>
        </div>
      </div>
    </div>

    <!-- Main Content Tabs -->
    <div class="bg-[var(--bg-body)]/40 border border-[var(--border-color)] rounded-[40px] overflow-hidden backdrop-blur-xl">
      <div class="flex border-b border-[var(--border-color)] px-8">
        <button 
          v-for="tab in ['dashboard', 'bots', 'logs', 'settings']" 
          :key="tab"
          @click="activeTab = tab"
          class="px-8 py-6 text-sm font-black uppercase tracking-widest transition-all relative"
          :class="activeTab === tab ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
        >
          {{ tt(`earlymeow.console.tabs.${tab}`, tab) }}
          <div v-if="activeTab === tab" class="absolute bottom-0 left-8 right-8 h-1 bg-[var(--matrix-color)] rounded-full shadow-[0_0_10px_var(--matrix-color)]/50"></div>
        </button>
      </div>

      <div class="p-8">
        <!-- Bots Table -->
        <div v-if="activeTab === 'bots' || activeTab === 'dashboard'" class="space-y-6">
          <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-4">
            <h3 class="text-xl font-black text-[var(--text-main)]">{{ tt('earlymeow.console.bots.title') }}</h3>
            <div class="relative w-full md:w-64">
              <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
              <input type="text" :placeholder="tt('earlymeow.console.bots.search')" class="w-full pl-12 pr-6 py-2.5 rounded-xl bg-[var(--bg-body)]/50 border border-[var(--border-color)] text-sm text-[var(--text-main)] focus:border-[var(--matrix-color)] outline-none transition-all" />
            </div>
          </div>

          <div class="space-y-4">
            <div 
              v-for="bot in bots" 
              :key="bot.id"
              class="flex flex-col md:flex-row items-center justify-between p-6 rounded-3xl bg-[var(--bg-body)]/30 border border-[var(--border-color)] hover:border-[var(--matrix-color)]/20 transition-all group"
            >
              <div class="flex items-center gap-6 mb-4 md:mb-0 w-full md:w-auto">
                <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-[var(--bg-body)] to-[var(--bg-body)]/50 border border-[var(--border-color)] flex items-center justify-center text-3xl shadow-inner group-hover:scale-105 transition-transform">
                  {{ bot.id === 1 ? '🐱' : bot.id === 2 ? '🎣' : '📻' }}
                </div>
                <div>
                  <div class="flex items-center gap-3">
                    <h4 class="font-black text-lg text-[var(--text-main)]">{{ bot.name }}</h4>
                    <span 
                      class="px-2 py-0.5 rounded-md text-[10px] font-black uppercase"
                      :class="bot.status === 'online' ? 'bg-[var(--matrix-color)]/20 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20' : 'bg-[var(--bg-body)] text-[var(--text-muted)]'"
                    >
                      {{ tt(`earlymeow.console.bots.status.${bot.status}`, bot.status) }}
                    </span>
                  </div>
                  <div class="text-xs font-bold text-[var(--text-muted)]/60">{{ tt('earlymeow.console.bots.mode') }}: {{ bot.mode }} • {{ tt('earlymeow.console.bots.processed') }}: {{ bot.messages }}</div>
                </div>
              </div>

              <div class="flex items-center gap-3 w-full md:w-auto justify-end">
                <button class="p-3 rounded-xl bg-[var(--bg-body)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all text-[var(--text-muted)] hover:text-[var(--text-main)]">
                  <Edit3 class="w-5 h-5" />
                </button>
                <button class="p-3 rounded-xl bg-[var(--bg-body)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all text-[var(--text-muted)] hover:text-[var(--text-main)]">
                  <BarChart3 class="w-5 h-5" />
                </button>
                <div class="w-px h-8 bg-[var(--border-color)] mx-2"></div>
                <button 
                  class="p-3 rounded-xl transition-all"
                  :class="bot.status === 'online' ? 'bg-[var(--matrix-color)]/20 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/30 border border-[var(--matrix-color)]/20' : 'bg-[var(--matrix-color)]/20 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)]/30 border border-[var(--matrix-color)]/20'"
                >
                  <component :is="bot.status === 'online' ? Pause : Play" class="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="py-20 text-center space-y-4">
          <div class="text-4xl">🚧</div>
          <h3 class="text-xl font-black text-[var(--text-main)]">{{ tt('earlymeow.console.dev.title') }}</h3>
          <p class="text-sm text-[var(--text-muted)] font-medium">{{ tt('earlymeow.console.dev.desc') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Simplified styles */
</style>
