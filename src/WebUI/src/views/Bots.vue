<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue';
import { useBotStore } from '@/stores/bot';
import { getPlatformIcon, getPlatformColor, isPlatformAvatar, getPlatformFromAvatar } from '@/utils/avatar';
import { 
  Bot, 
  MessageSquare, 
  Users, 
  Power, 
  Settings as SettingsIcon, 
  Trash2, 
  Plus, 
  Gamepad2, 
  Mail, 
  Slack, 
  Shield, 
  Globe,
  X,
  Loader2,
  RefreshCw,
  Terminal,
  Layers,
  Search,
  CheckCircle2,
  Circle,
  Send
} from 'lucide-vue-next';

import { useSystemStore } from '@/stores/system';

const systemStore = useSystemStore();
const botStore = useBotStore();

// Translation function
const t = (key: string) => systemStore.t(key);
const isLoading = ref(false);
const searchQuery = ref('');
const sortBy = ref<'id' | 'nickname' | 'msg_count' | 'connected'>('id');
const sortOrder = ref<'asc' | 'desc'>('asc');

const filteredBots = computed(() => {
  let result = [...botStore.bots];
  
  // Filter
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    result = result.filter(bot => 
      (bot.id && bot.id.toLowerCase().includes(q)) || 
      (bot.nickname && bot.nickname.toLowerCase().includes(q)) ||
      (bot.platform && bot.platform.toLowerCase().includes(q))
    );
  }
  
  // Sort
  result.sort((a, b) => {
    let valA: any = a[sortBy.value];
    let valB: any = b[sortBy.value];
    
    if (sortBy.value === 'connected') {
      valA = a.connected ? 1 : 0;
      valB = b.connected ? 1 : 0;
    }

    if (valA < valB) return sortOrder.value === 'asc' ? -1 : 1;
    if (valA > valB) return sortOrder.value === 'asc' ? 1 : -1;
    return 0;
  });
  
  return result;
});

const toggleSort = (field: 'id' | 'nickname' | 'msg_count' | 'connected') => {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortBy.value = field;
    sortOrder.value = 'asc';
  }
};
const showAddModal = ref(false);
const showLogModal = ref(false);
const showMatrixModal = ref(false);
const matrixLoading = ref(false);
const matrixTab = ref('groups'); // groups, friends, batch
const matrixSearchQuery = ref('');
const selectedBotForMatrix = ref<any>(null);
const matrixGroups = ref<any[]>([]);
const matrixFriends = ref<any[]>([]);
const selectedTargets = ref<Set<string>>(new Set());
const batchMessage = ref('');
const showMemberModal = ref(false);
const selectedGroupForMembers = ref<any>(null);
const groupMembers = ref<any[]>([]);
const memberLoading = ref(false);

const currentLogs = ref<string[]>([]);
const logTimer = ref<number | null>(null);
const selectedBot = ref<any>(null);

const platforms = [
  { 
    id: 'Kook', 
    name: 'platform_kook', 
    icon: Gamepad2, 
    color: 'text-purple-500', 
    image: 'botmatrix-kookbot',
    fields: [
      { key: 'BOT_TOKEN', label: 'bot_token', type: 'password', placeholder: 'kook_bot_token' }
    ]
  },
  { 
    id: 'DingTalk', 
    name: 'platform_dingtalk', 
    icon: MessageSquare, 
    color: 'text-blue-500', 
    image: 'botmatrix-dingtalkbot',
    fields: [
      { key: 'DINGTALK_APP_KEY', label: 'app_key', type: 'text', placeholder: 'dingtalk_app_key' },
      { key: 'DINGTALK_APP_SECRET', label: 'app_secret', type: 'password', placeholder: 'dingtalk_app_secret' }
    ]
  },
  { 
    id: 'Email', 
    name: 'platform_email', 
    icon: Mail, 
    color: 'text-orange-500', 
    image: 'botmatrix-emailbot',
    fields: [
      { key: 'EMAIL_HOST', label: 'smtp_host', type: 'text', placeholder: 'smtp_host_placeholder' },
      { key: 'EMAIL_PORT', label: 'smtp_port', type: 'text', placeholder: 'smtp_port_placeholder' },
      { key: 'EMAIL_USER', label: 'email_account', type: 'text', placeholder: 'email_account_placeholder' },
      { key: 'EMAIL_PASS', label: 'email_password', type: 'password', placeholder: 'smtp_auth_code' }
    ]
  },
  { 
    id: 'Slack', 
    name: 'platform_slack', 
    icon: Slack, 
    color: 'text-red-500', 
    image: 'botmatrix-slackbot',
    fields: [
      { key: 'SLACK_BOT_TOKEN', label: 'bot_token', type: 'password', placeholder: 'slack_bot_token_placeholder' },
      { key: 'SLACK_APP_TOKEN', label: 'app_token', type: 'password', placeholder: 'slack_app_token_placeholder' }
    ]
  },
  { 
    id: 'TencentCloud', 
    name: 'platform_tencent', 
    icon: Shield, 
    color: 'text-indigo-500', 
    image: 'botmatrix-tencentbot',
    fields: [
      { key: 'TENCENT_APP_ID', label: 'app_id', type: 'text', placeholder: 'tencent_appid_placeholder' },
      { key: 'TENCENT_TOKEN', label: 'token', type: 'password', placeholder: 'tencent_token_placeholder' },
      { key: 'TENCENT_SECRET', label: 'secret', type: 'password', placeholder: 'tencent_secret_placeholder' }
    ]
  },
  { 
    id: 'WeCom', 
    name: 'platform_wecom', 
    icon: Globe, 
    color: 'text-blue-600', 
    image: 'botmatrix-wecombot',
    fields: [
      { key: 'WECOM_CORP_ID', label: 'corp_id', type: 'text', placeholder: 'wecom_corpid_placeholder' },
      { key: 'WECOM_AGENT_ID', label: 'agent_id', type: 'text', placeholder: 'wecom_agentid_placeholder' },
      { key: 'WECOM_SECRET', label: 'secret', type: 'password', placeholder: 'wecom_secret_placeholder' },
      { key: 'WECOM_TOKEN', label: 'token', type: 'password', placeholder: 'wecom_token_placeholder' },
      { key: 'WECOM_AES_KEY', label: 'encoding_aes_key', type: 'password', placeholder: 'wecom_aeskey_placeholder' }
    ]
  },
  { 
    id: 'Web', 
    name: 'platform_web', 
    icon: Globe, 
    color: 'text-emerald-500', 
    image: 'botmatrix-webbot',
    fields: [
      { key: 'WEB_SITE_ID', label: 'site_id', type: 'text', placeholder: 'site_id_placeholder' },
      { key: 'WEB_TITLE', label: 'window_title', type: 'text', placeholder: 'online_service' },
      { key: 'WEB_WELCOME_MSG', label: 'welcome_msg', type: 'text', placeholder: 'welcome_msg_placeholder' }
    ]
  },
];

const newBot = reactive({
  platform: 'Kook',
  image: 'botmatrix-kookbot',
  env: {} as Record<string, string>,
});

const currentFields = ref<any[]>(platforms[0].fields);

const updatePlatform = (p: any) => {
  newBot.platform = p.id;
  newBot.image = p.image;
  currentFields.value = p.fields;
  // Reset env
  newBot.env = {};
  p.fields.forEach((f: any) => {
    newBot.env[f.key] = '';
  });
};

onMounted(async () => {
  await botStore.fetchBots();
  updatePlatform(platforms[0]);
});

const getStatusColor = (connected: boolean) => {
  return connected ? 'bg-matrix' : 'bg-gray-400';
};

const getPlatformIcon = (platform: string) => {
  const p = platforms.find(p => p.id === platform);
  return p ? p.icon : Bot;
};

const getPlatformColor = (platform: string) => {
  const p = platforms.find(p => p.id === platform);
  return p ? p.color : 'text-matrix';
};

const getPlatformName = (platform: string) => {
  const p = platforms.find(p => p.id === platform);
  return p ? t(p.name) : platform;
};

const handleAddBot = async () => {
  try {
    isLoading.value = true;
    const config = {
      platform: newBot.platform,
      image: newBot.image,
      env: {
        ...newBot.env
      }
    };
    await botStore.addBot(config);
    showAddModal.value = false;
    await botStore.fetchBots();
  } catch (err) {
    alert(t('add_failed') + ': ' + err);
  } finally {
    isLoading.value = false;
  }
};

const handleDeleteBot = async (botId: string) => {
  if (confirm(t('confirm_delete_bot'))) {
    try {
      await botStore.removeBot(botId);
    } catch (err) {
      alert(t('delete_failed') + ': ' + err);
    }
  }
};

const viewLogs = async (bot: any) => {
  selectedBot.value = bot;
  showLogModal.value = true;
  await fetchLogs();
  
  if (logTimer.value) clearInterval(logTimer.value);
  logTimer.value = window.setInterval(fetchLogs, 3000);
};

const fetchLogs = async () => {
  if (!selectedBot.value) return;
  try {
    const data = await botStore.getLogs(selectedBot.value.id);
    if (data.status === 'ok') {
      // Handle Docker log header (8 bytes) and split by line
      const rawLogs = data.logs || '';
      // This is a simple handling, actual might be more complex as io.ReadAll reads raw byte stream
      // In JS, since logs are strings, we can only try to clean up non-printable characters
      currentLogs.value = rawLogs
        .split('\n')
        .map((line: string) => line.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '').trim())
        .filter((line: string) => line.length > 0);
    }
  } catch (err) {
    console.error(t('failed_fetch_logs'), err);
  }
};

const closeLogModal = () => {
  showLogModal.value = false;
  if (logTimer.value) {
    clearInterval(logTimer.value);
    logTimer.value = null;
  }
};

const updateImage = () => {
  const p = platforms.find(p => p.id === newBot.platform);
  if (p) {
    newBot.image = p.image;
  }
};

const openMatrixModal = async (bot: any) => {
  selectedBotForMatrix.value = bot;
  showMatrixModal.value = true;
  matrixLoading.value = true;
  matrixTab.value = 'groups';
  selectedTargets.value.clear();
  
  try {
    const res = await botStore.fetchContacts(bot.id);
    if (res.success) {
      matrixGroups.value = res.contacts.filter((c: any) => c.type === 'group');
      matrixFriends.value = res.contacts.filter((c: any) => c.type === 'private');
    }
  } catch (err) {
    console.error('Failed to fetch contacts:', err);
  } finally {
    matrixLoading.value = false;
  }
};

const syncContacts = async () => {
  if (!selectedBotForMatrix.value) return;
  matrixLoading.value = true;
  try {
    await botStore.syncContacts(selectedBotForMatrix.value.id);
    // Refresh lists after a short delay to allow sync to complete
    setTimeout(async () => {
      const res = await botStore.fetchContacts(selectedBotForMatrix.value.id);
      if (res.success) {
        matrixGroups.value = res.contacts.filter((c: any) => c.type === 'group');
        matrixFriends.value = res.contacts.filter((c: any) => c.type === 'private');
      }
      matrixLoading.value = false;
    }, 2000);
  } catch (err) {
    alert(t('sync_failed'));
    matrixLoading.value = false;
  }
};

const toggleTarget = (target: any) => {
  const key = `${target.type}:${target.id}`;
  if (selectedTargets.value.has(key)) {
    selectedTargets.value.delete(key);
  } else {
    selectedTargets.value.add(key);
  }
};

const isTargetSelected = (target: any) => {
  return selectedTargets.value.has(`${target.type}:${target.id}`);
};

const sendBatchMessage = async () => {
  if (!batchMessage.value.trim() || selectedTargets.value.size === 0) return;
  
  const targets = Array.from(selectedTargets.value).map(key => {
    const [type, id] = key.split(':');
    return { type, id, bot_id: selectedBotForMatrix.value.id };
  });

  try {
    isLoading.value = true;
    await botStore.callBotApi('batch_send_msg', {
      message: batchMessage.value,
      targets: targets
    }, selectedBotForMatrix.value.id);
    
    alert(t('batch_send_success'));
    batchMessage.value = '';
    selectedTargets.value.clear();
    matrixTab.value = 'groups';
  } catch (err) {
    alert(t('batch_send_failed') + ': ' + err);
  } finally {
    isLoading.value = false;
  }
};

const viewMembers = async (group: any, refresh = false) => {
  selectedGroupForMembers.value = group;
  showMemberModal.value = true;
  memberLoading.value = true;
  
  if (refresh) {
    groupMembers.value = [];
  }
  
  try {
    const res = await botStore.fetchGroupMembers(selectedBotForMatrix.value.id, group.id, refresh);
    
    if (res && res.success) {
      groupMembers.value = res.data;
    }
  } catch (err) {
    console.error('Failed to fetch members:', err);
  } finally {
    memberLoading.value = false;
  }
};

const filteredGroups = computed(() => {
  if (!matrixSearchQuery.value) return matrixGroups.value;
  const q = matrixSearchQuery.value.toLowerCase();
  return matrixGroups.value.filter(g => 
    (g.name && g.name.toLowerCase().includes(q)) || 
    (g.id && g.id.toLowerCase().includes(q))
  );
});

const filteredFriends = computed(() => {
  if (!matrixSearchQuery.value) return matrixFriends.value;
  const q = matrixSearchQuery.value.toLowerCase();
  return matrixFriends.value.filter(f => 
    (f.name && f.name.toLowerCase().includes(q)) || 
    (f.id && f.id.toLowerCase().includes(q))
  );
});
</script>

<template>
  <div class="p-4 sm:p-8 space-y-4 sm:space-y-8">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
          <Bot class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('bot_instances') }}</h1>
          <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold tracking-widest uppercase">{{ t('manage_active_bots') }}</p>
        </div>
      </div>
      
      <div class="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
        <div class="relative w-full sm:w-64 group">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
          <input 
            v-model="searchQuery"
            type="text" 
            :placeholder="t('search_bots')"
            class="w-full pl-11 pr-4 py-3 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl text-xs font-bold text-[var(--text-main)] focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
          >
        </div>

        <div class="flex items-center gap-2 bg-[var(--bg-card)] border border-[var(--border-color)] p-1 rounded-2xl">
          <button 
            v-for="field in ['nickname', 'msg_count', 'connected']" 
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
          @click="showAddModal = true"
          class="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-[var(--matrix-color)] text-black text-xs font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20"
        >
          <Plus class="w-4 h-4" />
          {{ t('deploy_bot') }}
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredBots.length === 0" class="flex flex-col items-center justify-center py-10 sm:py-20 bg-[var(--bg-card)]/50 backdrop-blur-md rounded-3xl border border-dashed border-[var(--border-color)]">
      <div class="w-12 h-12 sm:w-16 sm:h-16 bg-[var(--matrix-color)]/10 rounded-2xl flex items-center justify-center mb-4">
        <Bot class="w-6 h-6 sm:w-8 sm:h-8 text-[var(--matrix-color)]" />
      </div>
      <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)] mb-2">{{ searchQuery ? t('no_search_results') : t('no_bots') }}</h3>
      <p class="text-[var(--text-muted)] text-[10px] sm:text-sm mb-6 text-center max-w-xs px-4">{{ searchQuery ? t('try_other_keywords') : t('no_bots_desc') }}</p>
      <button 
        v-if="!searchQuery"
        @click="showAddModal = true"
        class="px-6 py-2 border border-[var(--matrix-color)] text-[var(--matrix-color)] font-bold rounded-xl hover:bg-[var(--matrix-color)] hover:text-black transition-all"
      >
        {{ t('start_now') }}
      </button>
      <button 
        v-else
        @click="searchQuery = ''"
        class="px-6 py-2 border border-[var(--border-color)] text-[var(--text-muted)] font-bold rounded-xl hover:bg-black/5 transition-all"
      >
        {{ t('clear_search') }}
      </button>
    </div>

    <!-- Bot Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
      <div v-for="bot in filteredBots" :key="bot.id" class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] backdrop-blur-xl border border-[var(--border-color)] shadow-sm hover:shadow-xl transition-all duration-300 group relative overflow-hidden">
        <!-- Platform Background Icon -->
        <component :is="getPlatformIcon(bot.platform)" class="absolute -right-4 -bottom-4 w-24 sm:w-32 h-24 sm:h-32 opacity-[0.03] dark:opacity-[0.05] pointer-events-none" />

        <div class="flex items-start justify-between mb-4 sm:mb-6 relative z-10">
          <div class="flex items-center gap-3 sm:gap-4">
            <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center relative shadow-inner overflow-hidden">
                  <template v-if="bot.avatar && !isPlatformAvatar(bot.avatar)">
                    <img :src="bot.avatar" class="w-full h-full object-cover" />
                  </template>
                  <template v-else>
                    <component 
                      :is="isPlatformAvatar(bot.avatar) ? getPlatformIcon(getPlatformFromAvatar(bot.avatar)) : getPlatformIcon(bot.platform)" 
                      :class="['w-5 h-5 sm:w-6 sm:h-6', isPlatformAvatar(bot.avatar) ? getPlatformColor(getPlatformFromAvatar(bot.avatar)) : getPlatformColor(bot.platform)]" 
                    />
                  </template>
                  <div :class="['absolute -bottom-1 -right-1 w-3.5 h-3.5 sm:w-4 sm:h-4 rounded-full border-4 border-[var(--bg-card)]', getStatusColor(bot.connected)]"></div>
                </div>
            <div class="min-w-0">
              <h3 class="font-bold text-sm sm:text-base text-[var(--text-main)] truncate">{{ bot.nickname || bot.id }}</h3>
              <div class="flex items-center gap-2">
                <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest font-bold">{{ getPlatformName(bot.platform) }}</p>
                <span class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[8px] text-[var(--text-muted)] font-bold mono lowercase">ID: {{ bot.id }}</span>
              </div>
            </div>
          </div>
          <div class="flex gap-1 shrink-0">
            <button 
              @click="openMatrixModal(bot)"
              class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors"
              :title="t('bot_matrix_details')"
            >
              <Layers class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
            <button 
              @click="viewLogs(bot)"
              class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors"
              :title="t('view_logs')"
            >
              <Terminal class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
            <button class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors" :title="t('settings')">
              <SettingsIcon class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
            <button 
              @click="handleDeleteBot(bot.id)"
              class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-red-500 transition-colors"
              :title="t('delete')"
            >
              <Trash2 class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-2 sm:gap-4 mb-4 sm:mb-6 relative z-10">
          <div class="bg-black/5 dark:bg-white/5 p-2 sm:p-3 rounded-2xl text-center">
            <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest mb-1 font-bold">{{ t('messages') }}</p>
            <p class="font-bold text-[var(--text-main)] mono text-xs sm:text-sm">{{ bot.msg_count || 0 }}</p>
          </div>
          <div class="bg-black/5 dark:bg-white/5 p-2 sm:p-3 rounded-2xl text-center">
            <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest mb-1 font-bold">{{ t('group_chats') }}</p>
            <p class="font-bold text-[var(--text-main)] mono text-xs sm:text-sm">{{ bot.group_count || 0 }}</p>
          </div>
          <div class="bg-black/5 dark:bg-white/5 p-2 sm:p-3 rounded-2xl text-center">
            <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest mb-1 font-bold">{{ t('friends') }}</p>
            <p class="font-bold text-[var(--text-main)] mono text-xs sm:text-sm">{{ bot.friend_count || 0 }}</p>
          </div>
        </div>

        <div class="flex gap-2 relative z-10">
          <button 
            @click="botStore.setCurrentBotId(bot.id)"
            class="flex-1 py-2.5 sm:py-3 rounded-2xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] sm:text-xs font-bold uppercase tracking-widest hover:bg-[var(--matrix-color)] hover:text-black transition-all duration-300"
          >
            {{ botStore.currentBotId === bot.id ? t('currently_selected') : t('set_as_current') }}
          </button>
          <button class="px-3 sm:px-4 py-2.5 sm:py-3 rounded-2xl border border-[var(--border-color)] text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 transition-all" :title="t('refresh')">
            <RefreshCw class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Add Bot Modal -->
    <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-lg bg-[var(--bg-card)] rounded-[2rem] sm:rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-4 sm:space-y-6 max-h-[90vh] overflow-y-auto no-scrollbar">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 sm:gap-4">
              <div class="p-2.5 sm:p-3 bg-[var(--matrix-color)]/10 rounded-2xl">
                <Plus class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h2 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">{{ t('deploy_new_bot') }}</h2>
                <p class="text-[10px] text-[var(--text-muted)] font-medium uppercase tracking-widest">{{ t('select_and_configure') }}</p>
              </div>
            </div>
            <button @click="showAddModal = false" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
              <X class="w-5 h-5 text-[var(--text-muted)]" />
            </button>
          </div>

          <div class="space-y-4">
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('select_platform') }}</label>
              <div class="grid grid-cols-3 gap-2">
                <button 
                  v-for="p in platforms" 
                  :key="p.id"
                  @click="updatePlatform(p)"
                  :class="['p-2.5 sm:p-3 rounded-2xl border flex flex-col items-center gap-2 transition-all', 
                    newBot.platform === p.id ? 'border-[var(--matrix-color)] bg-[var(--matrix-color)]/5 ring-1 ring-[var(--matrix-color)]' : 'border-[var(--border-color)] hover:bg-black/5 dark:hover:bg-white/5']"
                >
                  <component :is="p.icon" :class="['w-4 h-4 sm:w-5 sm:h-5', newBot.platform === p.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]']" />
                  <span :class="['text-[8px] sm:text-[10px] font-bold uppercase tracking-tighter', newBot.platform === p.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]']">{{ t(p.name) || p.name }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('docker_image') }}</label>
              <input 
                v-model="newBot.image"
                type="text" 
                class="w-full px-4 sm:px-5 py-2.5 sm:py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-[var(--text-main)] mono text-xs sm:text-sm"
                :placeholder="t('docker_image_placeholder')"
              />
            </div>

            <div v-for="field in currentFields" :key="field.key" class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t(field.label) || field.label }}</label>
              <input 
                v-model="newBot.env[field.key]"
                :type="field.type" 
                class="w-full px-4 sm:px-5 py-2.5 sm:py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-[var(--text-main)] mono text-xs sm:text-sm"
                :placeholder="t(field.placeholder) || field.placeholder"
              />
            </div>
          </div>

          <div class="flex gap-4 pt-2">
            <button 
              @click="showAddModal = false"
              class="flex-1 py-3.5 sm:py-4 rounded-2xl border border-[var(--border-color)] text-[10px] sm:text-xs font-bold uppercase tracking-widest hover:bg-black/5 dark:hover:bg-white/5 transition-all text-[var(--text-muted)]"
            >
              {{ t('cancel') }}
            </button>
            <button 
              @click="handleAddBot"
              :disabled="isLoading"
              class="flex-1 py-3.5 sm:py-4 rounded-2xl bg-[var(--matrix-color)] text-black text-[10px] sm:text-xs font-bold uppercase tracking-widest hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2 shadow-lg shadow-[var(--matrix-color)]/20"
            >
              <Loader2 v-if="isLoading" class="w-4 h-4 animate-spin" />
              {{ isLoading ? t('deploying') : t('deploy_now') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Log Modal -->
    <div v-if="showLogModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-3xl bg-[var(--bg-card)] rounded-[2rem] sm:rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-4 sm:space-y-6 h-[85vh] sm:h-[80vh] flex flex-col">
          <div class="flex items-center justify-between flex-shrink-0">
            <div class="flex items-center gap-3 sm:gap-4">
              <div class="p-2.5 sm:p-3 bg-[var(--matrix-color)]/10 rounded-2xl">
                <Terminal class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h2 class="text-lg sm:text-xl font-bold text-[var(--text-main)] truncate max-w-[150px] sm:max-w-none">{{ t('realtime_logs') }}</h2>
                <div class="flex items-center gap-2 mt-0.5">
                  <p class="text-[10px] text-[var(--text-muted)] font-medium uppercase tracking-widest truncate max-w-[150px] sm:max-w-none">{{ selectedBot?.nickname || selectedBot?.id }} ({{ getPlatformName(selectedBot?.platform) }})</p>
                  <span class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[8px] text-[var(--text-muted)] font-bold mono lowercase">ID: {{ selectedBot?.id }}</span>
                </div>
              </div>
            </div>
            <button @click="closeLogModal" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
              <X class="w-5 h-5 text-[var(--text-muted)]" />
            </button>
          </div>

          <div class="flex-1 min-h-0 bg-black rounded-[1.5rem] p-4 sm:p-6 overflow-y-auto font-mono text-[10px] sm:text-xs space-y-1 custom-scrollbar">
            <div v-for="(log, idx) in currentLogs" :key="idx" class="text-[var(--matrix-color)]/80 break-all">
              <span class="text-gray-600 mr-2">{{ idx + 1 }}</span> {{ log }}
            </div>
          </div>

          <div class="flex items-center justify-between pt-2 flex-shrink-0">
            <div class="flex items-center gap-2 text-[8px] sm:text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">
              <div class="w-1.5 h-1.5 sm:w-2 h-2 bg-[var(--matrix-color)] rounded-full animate-pulse"></div>
              {{ t('syncing_live') }}
            </div>
            <button 
              @click="fetchLogs"
              class="px-3 sm:px-4 py-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[8px] sm:text-[10px] font-bold uppercase tracking-widest hover:bg-black/10 transition-all text-[var(--text-main)]"
            >
              {{ t('refresh') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bot Matrix Modal -->
    <div v-if="showMatrixModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-4xl bg-[var(--bg-card)] rounded-[2rem] sm:rounded-[2.5rem] border border-[var(--border-color)] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-4 sm:space-y-6 h-[85vh] flex flex-col">
          <!-- Modal Header -->
          <div class="flex items-center justify-between flex-shrink-0">
            <div class="flex items-center gap-3 sm:gap-4">
              <div class="p-2.5 sm:p-3 bg-[var(--matrix-color)]/10 rounded-2xl">
                <Layers class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
              </div>
              <div>
                <h2 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">{{ t('bot_matrix_details') }}</h2>
                <div class="flex items-center gap-2 mt-0.5">
                  <p class="text-[10px] text-[var(--text-muted)] font-medium uppercase tracking-widest">{{ selectedBotForMatrix?.nickname || selectedBotForMatrix?.id }}</p>
                  <span class="px-1.5 py-0.5 rounded-md bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[9px] text-[var(--text-muted)] font-bold mono uppercase">{{ getPlatformName(selectedBotForMatrix?.platform) }}</span>
                  <span class="px-1.5 py-0.5 rounded-md bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[9px] text-[var(--text-main)] font-bold mono lowercase">ID: {{ selectedBotForMatrix?.id }}</span>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button 
                @click="syncContacts" 
                :disabled="matrixLoading"
                class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors text-[var(--text-muted)] hover:text-[var(--matrix-color)]"
                :title="t('sync_contacts')"
              >
                <RefreshCw :class="['w-5 h-5', matrixLoading ? 'animate-spin' : '']" />
              </button>
              <button @click="showMatrixModal = false" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
                <X class="w-5 h-5 text-[var(--text-muted)]" />
              </button>
            </div>
          </div>

          <!-- Tabs and Search -->
          <div class="flex flex-col sm:flex-row gap-4 flex-shrink-0">
            <div class="flex bg-black/5 dark:bg-white/5 p-1 rounded-xl w-fit">
              <button 
                v-for="tab in ['groups', 'friends', 'batch']" 
                :key="tab"
                @click="matrixTab = tab"
                :class="['px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-widest transition-all', 
                  matrixTab === tab ? 'bg-[var(--bg-card)] text-[var(--matrix-color)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]']"
              >
                {{ t(tab === 'groups' ? 'group_list' : tab === 'friends' ? 'friend_list' : 'batch_messaging') }}
                <span v-if="tab === 'batch' && selectedTargets.size > 0" class="ml-1 px-1.5 py-0.5 rounded-full bg-[var(--matrix-color)] text-black text-[8px]">
                  {{ selectedTargets.size }}
                </span>
              </button>
            </div>
            <div v-if="matrixTab !== 'batch'" class="relative flex-1">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
              <input 
                v-model="matrixSearchQuery"
                type="text" 
                :placeholder="t('search')"
                class="w-full pl-10 pr-4 py-2 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-xs text-[var(--text-main)]"
              />
            </div>
          </div>

          <!-- Content Area -->
          <div class="flex-1 min-h-0 overflow-y-auto custom-scrollbar pr-2">
            <!-- Loading State -->
            <div v-if="matrixLoading" class="h-full flex flex-col items-center justify-center gap-4 text-[var(--text-muted)]">
              <Loader2 class="w-8 h-8 animate-spin text-[var(--matrix-color)]" />
              <p class="text-[10px] font-bold uppercase tracking-widest">{{ t('loading') }}</p>
            </div>

            <!-- Groups Tab -->
            <div v-else-if="matrixTab === 'groups'" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div v-if="filteredGroups.length === 0" class="col-span-full py-20 text-center text-[var(--text-muted)]">
                <p class="text-xs uppercase tracking-widest font-bold">{{ t('no_groups') }}</p>
              </div>
              <div 
                v-for="group in filteredGroups" 
                :key="group.id"
                class="group p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/30 transition-all flex items-center justify-between"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <div class="w-10 h-10 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center flex-shrink-0 overflow-hidden">
                    <template v-if="group.avatar && !isPlatformAvatar(group.avatar)">
                      <img :src="group.avatar" class="w-full h-full object-cover" />
                    </template>
                    <template v-else>
                      <component 
                        :is="isPlatformAvatar(group.avatar) ? getPlatformIcon(getPlatformFromAvatar(group.avatar)) : Users" 
                        :class="['w-5 h-5', isPlatformAvatar(group.avatar) ? getPlatformColor(getPlatformFromAvatar(group.avatar)) : 'text-[var(--matrix-color)]']" 
                      />
                    </template>
                  </div>
                  <div class="min-w-0">
                    <h4 class="text-sm font-bold text-[var(--text-main)] truncate">{{ group.name || group.id }}</h4>
                    <p class="text-[10px] text-[var(--text-muted)] mono truncate">{{ group.id }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <button 
                    @click="viewMembers(group)"
                    class="p-2 rounded-lg hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors"
                    :title="t('member_management')"
                  >
                    <Users class="w-4 h-4" />
                  </button>
                  <button 
                    @click="toggleTarget(group)"
                    :class="['p-2 rounded-lg transition-all', isTargetSelected(group) ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-black/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)]']"
                  >
                    <CheckCircle2 v-if="isTargetSelected(group)" class="w-4 h-4" />
                    <Circle v-else class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Friends Tab -->
            <div v-else-if="matrixTab === 'friends'" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div v-if="filteredFriends.length === 0" class="col-span-full py-20 text-center text-[var(--text-muted)]">
                <p class="text-xs uppercase tracking-widest font-bold">{{ t('no_friends') }}</p>
              </div>
              <div 
                v-for="friend in filteredFriends" 
                :key="friend.id"
                class="group p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/30 transition-all flex items-center justify-between"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <div class="w-10 h-10 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center flex-shrink-0 overflow-hidden">
                    <template v-if="friend.avatar && !isPlatformAvatar(friend.avatar)">
                      <img :src="friend.avatar" class="w-full h-full object-cover" />
                    </template>
                    <template v-else>
                      <component 
                        :is="isPlatformAvatar(friend.avatar) ? getPlatformIcon(getPlatformFromAvatar(friend.avatar)) : Bot" 
                        :class="['w-5 h-5', isPlatformAvatar(friend.avatar) ? getPlatformColor(getPlatformFromAvatar(friend.avatar)) : 'text-blue-500']" 
                      />
                    </template>
                  </div>
                  <div class="min-w-0">
                    <h4 class="text-sm font-bold text-[var(--text-main)] truncate">{{ friend.name || friend.id }}</h4>
                    <p class="text-[10px] text-[var(--text-muted)] mono truncate">{{ friend.id }}</p>
                  </div>
                </div>
                <button 
                  @click="toggleTarget(friend)"
                  :class="['p-2 rounded-lg transition-all', isTargetSelected(friend) ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-black/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)]']"
                >
                  <CheckCircle2 v-if="isTargetSelected(friend)" class="w-4 h-4" />
                  <Circle v-else class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Batch Tab -->
            <div v-else-if="matrixTab === 'batch'" class="space-y-6">
              <div class="space-y-2">
                <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('select_targets') }} ({{ selectedTargets.size }})</label>
                <div v-if="selectedTargets.size === 0" class="p-8 rounded-2xl border border-dashed border-[var(--border-color)] text-center text-[var(--text-muted)]">
                  <p class="text-xs font-bold uppercase tracking-widest">{{ t('no_targets_selected') }}</p>
                </div>
                <div v-else class="flex flex-wrap gap-2 p-4 rounded-2xl bg-black/5 dark:bg-white/5">
                  <div 
                    v-for="key in selectedTargets" 
                    :key="key"
                    class="px-3 py-1.5 rounded-lg bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] text-[10px] font-bold border border-[var(--matrix-color)]/30 flex items-center gap-2"
                  >
                    <span>{{ key.split(':')[1] }}</span>
                    <button @click="selectedTargets.delete(key)" class="hover:text-red-500">
                      <X class="w-3 h-3" />
                    </button>
                  </div>
                </div>
              </div>

              <div class="space-y-2">
                <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('message_content') }}</label>
                <textarea 
                  v-model="batchMessage"
                  rows="6"
                  class="w-full px-5 py-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-sm text-[var(--text-main)] resize-none"
                  :placeholder="t('message_content_placeholder')"
                ></textarea>
              </div>

              <button 
                @click="sendBatchMessage"
                :disabled="isLoading || !batchMessage.trim() || selectedTargets.size === 0"
                class="w-full py-4 rounded-2xl bg-[var(--matrix-color)] text-black text-xs font-black uppercase tracking-widest hover:opacity-90 disabled:opacity-50 transition-all flex items-center justify-center gap-2 shadow-lg shadow-[var(--matrix-color)]/20"
              >
                <Send class="w-4 h-4" />
                {{ isLoading ? t('sending') : t('send_now') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Group Member Modal -->
    <div v-if="showMemberModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-2xl bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="p-6 sm:p-8 space-y-6 h-[70vh] flex flex-col">
          <div class="flex items-center justify-between flex-shrink-0">
            <div class="flex items-center gap-4">
              <div class="p-3 bg-purple-500/10 rounded-2xl">
                <Users class="w-6 h-6 text-purple-500" />
              </div>
              <div>
                <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('group_members') }}</h2>
                <p class="text-[10px] text-[var(--text-muted)] font-bold uppercase tracking-widest truncate max-w-[300px]">{{ selectedGroupForMembers?.nickname || selectedGroupForMembers?.id }}</p>
              </div>
            </div>
            <button @click="showMemberModal = false" class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
              <X class="w-5 h-5 text-[var(--text-muted)]" />
            </button>
          </div>

          <div class="flex-1 min-h-0 overflow-y-auto custom-scrollbar pr-2">
            <div v-if="memberLoading" class="h-full flex flex-col items-center justify-center gap-4 text-[var(--text-muted)]">
              <Loader2 class="w-8 h-8 animate-spin text-[var(--matrix-color)]" />
              <p class="text-[10px] font-bold uppercase tracking-widest">{{ t('loading') }}</p>
            </div>
            <div v-else class="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <div v-if="groupMembers.length === 0" class="col-span-full py-20 text-center text-[var(--text-muted)]">
                <p class="text-xs font-bold uppercase tracking-widest opacity-40">{{ t('no_members') }}</p>
              </div>
              <div 
                v-for="member in groupMembers" 
                :key="member.user_id"
                class="p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/20 transition-all flex items-center gap-3"
              >
                <div class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center flex-shrink-0 overflow-hidden">
                  <template v-if="member.avatar && !isPlatformAvatar(member.avatar)">
                    <img :src="member.avatar" class="w-full h-full object-cover" />
                  </template>
                  <template v-else>
                    <component 
                      :is="isPlatformAvatar(member.avatar) ? getPlatformIcon(getPlatformFromAvatar(member.avatar)) : Bot" 
                      :class="['w-4 h-4', isPlatformAvatar(member.avatar) ? getPlatformColor(getPlatformFromAvatar(member.avatar)) : 'text-[var(--text-muted)]']" 
                    />
                  </template>
                </div>
                <div class="min-w-0">
                  <h5 class="text-[10px] font-black text-[var(--text-main)] truncate">{{ member.nickname || member.card || member.user_id }}</h5>
                  <p class="text-[8px] text-[var(--text-muted)] mono truncate uppercase tracking-widest">{{ member.user_id }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="flex-shrink-0 pt-4 border-t border-[var(--border-color)] flex justify-between items-center">
            <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('total_members') }}: {{ groupMembers.length }}</p>
            <button 
              @click="viewMembers(selectedGroupForMembers, true)"
              class="px-6 py-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest hover:bg-black/10 transition-all text-[var(--text-main)]"
            >
              {{ t('refresh_members') }}
            </button>
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
.bg-matrix {
  background-color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: rgba(0, 255, 65, 0.1);
}
.border-matrix {
  border-color: var(--matrix-color);
}
.shadow-matrix\/20 {
  box-shadow: 0 10px 15px -3px rgba(0, 255, 65, 0.2);
}
.ring-matrix {
  --tw-ring-color: var(--matrix-color);
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}
/* Scrollbar */
.scrollbar-thin::-webkit-scrollbar {
  width: 6px;
}
.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}
</style>
