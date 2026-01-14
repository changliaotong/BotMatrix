<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useBotStore } from '@/stores/bot';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  Bot, 
  Settings, 
  Users, 
  MessageSquare, 
  Plus, 
  Search,
  ChevronRight,
  Shield,
  Zap,
  Globe,
  RefreshCw,
  Power,
  Trash2,
  ExternalLink,
  Activity,
  CheckCircle2,
  XCircle,
  AlertCircle
} from 'lucide-vue-next';

const router = useRouter();
const botStore = useBotStore();
const systemStore = useSystemStore();
const authStore = useAuthStore();

const t = (key: string) => systemStore.t(key);

const loading = ref(true);
const searchQuery = ref('');
const adminQQ = ref(authStore.user?.qq || '');
const isSavingQQ = ref(false);

const fetchBots = async () => {
  loading.value = true;
  try {
    const adminId = authStore.user?.id ? String(authStore.user.id) : undefined;
    const qq = adminQQ.value || undefined;
    await botStore.fetchMemberSetup(adminId, qq);
  } catch (err) {
    console.error('Failed to fetch bots:', err);
  } finally {
    loading.value = false;
  }
};

const savePersonalQQ = async () => {
  if (!adminQQ.value) return;
  isSavingQQ.value = true;
  try {
    const res = await botStore.updateUserProfile({ qq: adminQQ.value });
    if (res.success) {
      // Update local auth store user info if needed
      if (authStore.user) {
        authStore.user.qq = adminQQ.value;
      }
      await fetchBots();
    }
  } catch (err) {
    console.error('Failed to save QQ:', err);
  } finally {
    isSavingQQ.value = false;
  }
};

const filteredBots = computed(() => {
  if (!searchQuery.value) return botStore.bots;
  const q = searchQuery.value.toLowerCase();
  return botStore.bots.filter(bot => 
    (bot.bot_name && bot.bot_name.toLowerCase().includes(q)) || 
    (bot.bot_uin && bot.bot_uin.toString().includes(q)) ||
    (bot.bot_memo && bot.bot_memo.toLowerCase().includes(q))
  );
});

const getStatusColor = (bot: any) => {
  if (!bot.valid) return 'text-red-500 bg-red-500/10';
  if (bot.is_freeze) return 'text-orange-500 bg-orange-500/10';
  return 'text-green-500 bg-green-500/10';
};

const getStatusLabel = (bot: any) => {
  if (!bot.valid) return t('status_invalid');
  if (bot.is_freeze) return t('status_frozen');
  return t('status_online');
};

const navigateToSetup = (bot: any) => {
  router.push({
    name: 'portal-bot-setup',
    query: { bot_uin: bot.bot_uin, admin_id: bot.admin_id }
  });
};

const navigateToGroupSetup = (bot: any) => {
  router.push({
    name: 'portal-group-setup',
    query: { robot_owner: bot.bot_uin }
  });
};

const showLogModal = ref(false);
const selectedBot = ref<any>(null);
const logs = ref<string[]>([]);
const logLoading = ref(false);

const deleteBot = async (bot: any) => {
  if (!confirm(t('confirm_delete_bot'))) return;
  try {
    const res = await botStore.deleteMemberSetup(bot.bot_uin);
    if (res.success) {
      await fetchBots();
    }
  } catch (err) {
    console.error('Delete failed:', err);
  }
};

const viewLogs = async (bot: any) => {
  selectedBot.value = bot;
  showLogModal.value = true;
  logLoading.value = true;
  logs.value = [];
  
  try {
    // We'll use the bot_uin as the identifier for logs if possible, 
    // or the bot might have an 'id' for the container.
    // Let's assume bot.id exists as it's common in this codebase.
    const res = await botStore.getLogs(bot.id || bot.bot_uin.toString());
    if (res && res.data) {
      logs.value = Array.isArray(res.data) ? res.data : res.data.split('\n');
    }
  } catch (err) {
    console.error('Fetch logs failed:', err);
    logs.value = [t('failed_to_fetch_logs')];
  } finally {
    logLoading.value = false;
  }
};

onMounted(fetchBots);
</script>

<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 overflow-x-hidden" :class="[systemStore.style]">
    <PortalHeader />
    
    <div class="pt-40 pb-20 p-4 sm:p-6 max-w-7xl mx-auto space-y-8">
      <!-- Personal Setup Section -->
      <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2.5rem] p-6 sm:p-8 shadow-sm overflow-hidden relative group">
        <div class="absolute inset-0 bg-gradient-to-br from-[var(--matrix-color)]/5 to-transparent opacity-50"></div>
        <div class="relative flex flex-col md:flex-row items-center gap-6">
          <div class="w-16 h-16 bg-[var(--matrix-color)]/10 rounded-2xl flex items-center justify-center text-[var(--matrix-color)] shrink-0 shadow-inner">
            <User class="w-8 h-8" />
          </div>
          <div class="flex-1 text-center md:text-left">
            <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight italic">{{ t('personal_qq_setup', 'Personal Identity') }}</h2>
            <p class="text-xs text-[var(--text-muted)] font-medium mt-1 uppercase tracking-widest">
              {{ t('qq_setup_desc', 'Set your personal QQ to automatically filter and manage your bots and groups.') }}
            </p>
          </div>
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 w-full md:w-auto">
            <div class="relative">
              <input 
                v-model="adminQQ"
                type="text" 
                :placeholder="t('enter_personal_qq', 'Enter Personal QQ')"
                class="w-full sm:w-64 pl-4 pr-4 py-3 bg-black/5 dark:bg-white/5 border border-[var(--border-color)] rounded-xl text-sm font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
                @keyup.enter="savePersonalQQ"
              >
            </div>
            <button 
              @click="savePersonalQQ"
              :disabled="isSavingQQ || !adminQQ"
              class="px-6 py-3 bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest rounded-xl hover:opacity-90 transition-all flex items-center justify-center gap-2 disabled:opacity-50 shadow-lg shadow-[var(--matrix-color)]/10"
            >
              <div v-if="isSavingQQ" class="w-3 h-3 border-2 border-black border-t-transparent rounded-full animate-spin"></div>
              <Save v-else class="w-4 h-4" />
              {{ isSavingQQ ? t('saving') : t('save_identity', 'Apply Identity') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Header Section -->
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <div class="space-y-2">
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 bg-[var(--matrix-color)]/10 rounded-2xl flex items-center justify-center text-[var(--matrix-color)] shadow-lg shadow-[var(--matrix-color)]/5">
              <Bot class="w-7 h-7" />
            </div>
            <h1 class="text-3xl font-black tracking-tight text-[var(--text-main)] uppercase italic">
              {{ t('my_bots') }}
            </h1>
          </div>
          <p class="text-sm text-[var(--text-muted)] font-medium max-w-md">
            {{ t('my_bots_description', 'Manage and configure your digital assistants in one place.') }}
          </p>
        </div>

        <div class="flex items-center gap-3">
          <div class="relative group">
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
            <input 
              v-model="searchQuery"
              type="text" 
              :placeholder="t('search_bots')"
              class="pl-11 pr-6 py-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl text-sm font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 focus:ring-4 focus:ring-[var(--matrix-color)]/10 transition-all w-full md:w-64"
            >
          </div>
          <button 
            @click="fetchBots"
            class="p-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl hover:border-[var(--matrix-color)]/50 transition-all group"
            :title="t('refresh')"
          >
            <RefreshCw class="w-5 h-5 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)] transition-all" :class="{ 'animate-spin': loading }" />
          </button>
          <button 
            @click="router.push('/console/bots')"
            class="p-3 bg-[var(--matrix-color)] text-black rounded-2xl hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20"
            :title="t('add_bot')"
          >
            <Plus class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Bots Grid -->
      <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="i in 6" :key="i" class="h-64 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl animate-pulse"></div>
      </div>

      <div v-else-if="filteredBots.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <div 
          v-for="bot in filteredBots" 
          :key="bot.bot_uin"
          class="group relative bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 rounded-3xl transition-all duration-500 hover:shadow-2xl hover:shadow-[var(--matrix-color)]/5 overflow-hidden flex flex-col"
        >
          <!-- Card Header -->
          <div class="p-6 pb-4 flex items-start justify-between gap-4">
            <div class="flex items-center gap-4">
              <div class="relative">
                <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-[var(--matrix-color)]/20 to-[var(--matrix-color)]/5 flex items-center justify-center text-[var(--matrix-color)] font-black text-xl overflow-hidden">
                  <span v-if="bot.bot_name">{{ bot.bot_name.charAt(0).toUpperCase() }}</span>
                  <Bot v-else class="w-7 h-7" />
                </div>
                <div class="absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-4 border-[var(--bg-card)] flex items-center justify-center" :class="getStatusColor(bot)">
                  <div class="w-2 h-2 rounded-full bg-current shadow-[0_0_8px_currentColor]"></div>
                </div>
              </div>
              <div class="flex flex-col min-w-0">
                <h3 class="font-black text-[var(--text-main)] truncate group-hover:text-[var(--matrix-color)] transition-colors">
                  {{ bot.bot_name || 'Unnamed Bot' }}
                </h3>
                <span class="text-[10px] font-mono text-[var(--text-muted)] uppercase tracking-widest truncate">UIN: {{ bot.bot_uin }}</span>
                <span v-if="bot.admin_qq" class="text-[9px] font-mono text-[var(--matrix-color)]/70 uppercase tracking-widest truncate mt-0.5">Admin: {{ bot.admin_qq }}</span>
              </div>
            </div>
            
            <div class="flex items-center gap-1">
              <button @click="viewLogs(bot)" class="p-2 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors" :title="t('logs')">
                <Terminal class="w-4 h-4" />
              </button>
              <button @click="deleteBot(bot)" class="p-2 rounded-lg hover:bg-red-500/10 text-[var(--text-muted)] hover:text-red-500 transition-colors" :title="t('delete')">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Card Body -->
          <div class="px-6 py-2 flex-grow">
            <p class="text-xs text-[var(--text-muted)] line-clamp-2 min-h-[2.5rem]">
              {{ bot.bot_memo || t('no_memo', 'No description provided for this bot.') }}
            </p>

            <div class="mt-4 grid grid-cols-2 gap-3">
              <div class="bg-black/5 dark:bg-white/5 rounded-2xl p-3 flex flex-col gap-1">
                <span class="text-[9px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('type') }}</span>
                <span class="text-xs font-bold text-[var(--text-main)]">{{ bot.bot_type === 1 ? 'Standard' : 'Enterprise' }}</span>
              </div>
              <div class="bg-black/5 dark:bg-white/5 rounded-2xl p-3 flex flex-col gap-1">
                <span class="text-[9px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('groups') }}</span>
                <span class="text-xs font-bold text-[var(--text-main)] flex items-center gap-1">
                  <Users class="w-3 h-3 text-[var(--matrix-color)]" />
                  {{ bot.is_group ? 'Enabled' : 'Disabled' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Card Actions -->
          <div class="p-6 pt-4 grid grid-cols-2 gap-3 border-t border-[var(--border-color)]/50 mt-4">
            <button 
              @click="navigateToSetup(bot)"
              class="flex items-center justify-center gap-2 px-4 py-2.5 bg-black/5 dark:bg-white/5 hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] hover:text-[var(--matrix-color)] rounded-xl text-xs font-black uppercase tracking-widest transition-all group/btn"
            >
              <Settings class="w-3.5 h-3.5 group-hover/btn:rotate-90 transition-transform" />
              {{ t('setup') }}
            </button>
            <button 
              @click="navigateToGroupSetup(bot)"
              class="flex items-center justify-center gap-2 px-4 py-2.5 bg-black/5 dark:bg-white/5 hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] hover:text-[var(--matrix-color)] rounded-xl text-xs font-black uppercase tracking-widest transition-all group/btn"
            >
              <Users class="w-3.5 h-3.5 group-hover/btn:scale-110 transition-transform" />
              {{ t('groups') }}
            </button>
          </div>
          
          <!-- Hover Glow -->
          <div class="absolute inset-0 bg-gradient-to-br from-[var(--matrix-color)]/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"></div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="flex flex-col items-center justify-center py-20 px-6 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[3rem] border-dashed">
        <div class="w-20 h-20 bg-[var(--matrix-color)]/5 rounded-full flex items-center justify-center text-[var(--matrix-color)] mb-6">
          <Bot class="w-10 h-10 opacity-20" />
        </div>
        <h2 class="text-xl font-black text-[var(--text-main)] uppercase italic tracking-tight mb-2">{{ t('no_bots_found') }}</h2>
        <p class="text-sm text-[var(--text-muted)] text-center max-w-xs mb-8">
          {{ t('no_bots_description', "You don't have any bots yet, or no bots match your search criteria.") }}
        </p>
        <button 
          @click="router.push('/console/bots')"
          class="flex items-center gap-3 px-8 py-3 bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest rounded-2xl hover:opacity-90 transition-all shadow-xl shadow-[var(--matrix-color)]/20"
        >
          <Plus class="w-4 h-4" />
          {{ t('add_new_bot') }}
        </button>
      </div>
    </div>

    <!-- Log Modal -->
    <transition name="fade">
      <div v-if="showLogModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-md">
        <div class="w-full max-w-4xl bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
          <div class="p-6 sm:p-8 flex items-center justify-between border-b border-[var(--border-color)]">
            <div class="flex items-center gap-4">
              <div class="p-3 bg-[var(--matrix-color)]/10 rounded-2xl">
                <Terminal class="w-6 h-6 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h2 class="text-xl font-black text-[var(--text-main)] uppercase italic tracking-tight">{{ t('bot_logs') }}</h2>
                <p class="text-[10px] font-mono text-[var(--text-muted)] uppercase tracking-widest">UIN: {{ selectedBot?.bot_uin }}</p>
              </div>
            </div>
            <button @click="showLogModal = false" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
              <XCircle class="w-6 h-6 text-[var(--text-muted)]" />
            </button>
          </div>
          
          <div class="flex-grow overflow-y-auto p-6 font-mono text-xs bg-black/95 text-[var(--matrix-color)]/80 custom-scrollbar">
            <div v-if="logLoading" class="flex flex-col items-center justify-center py-20 space-y-4">
              <RefreshCw class="w-8 h-8 animate-spin opacity-50" />
              <p class="uppercase tracking-widest text-[10px] font-black">{{ t('loading_logs') }}</p>
            </div>
            <template v-else-if="logs.length > 0">
              <div v-for="(log, idx) in logs" :key="idx" class="py-0.5 border-l-2 border-transparent hover:border-[var(--matrix-color)]/30 hover:bg-white/5 px-2 transition-colors">
                <span class="text-white/20 mr-4 select-none">{{ idx + 1 }}</span>
                <span>{{ log }}</span>
              </div>
            </template>
            <div v-else class="text-center py-20 opacity-30 uppercase tracking-widest font-black">
              {{ t('no_logs_available') }}
            </div>
          </div>
          
          <div class="p-4 bg-black/50 border-t border-[var(--border-color)] flex items-center justify-between">
            <div class="flex items-center gap-2 text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">
              <div class="w-2 h-2 bg-[var(--matrix-color)] rounded-full animate-pulse shadow-[0_0_8px_var(--matrix-color)]"></div>
              Live Sync Active
            </div>
            <button @click="viewLogs(selectedBot)" class="flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-black uppercase tracking-widest hover:bg-[var(--matrix-color)]/20 transition-all">
              <RefreshCw class="w-3 h-3" />
              {{ t('refresh_logs') }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <PortalFooter />
  </div>
</template>

<style scoped>
.selection\:bg-\[var\(--matrix-color\)\]\/30 ::selection {
  background-color: rgba(var(--matrix-color-rgb), 0.3);
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
