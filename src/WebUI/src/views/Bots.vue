<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { useBotStore } from '@/stores/bot';
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
  Terminal
} from 'lucide-vue-next';

import { useSystemStore } from '@/stores/system';

const systemStore = useSystemStore();
const botStore = useBotStore();

// Translation function
const t = (key: string) => systemStore.t(key);
const isLoading = ref(false);
const showAddModal = ref(false);
const showLogModal = ref(false);
const currentLogs = ref<string[]>([]);
const logTimer = ref<number | null>(null);
const selectedBot = ref<any>(null);

const platforms = [
  { 
    id: 'Kook', 
    name: 'KOOK', 
    icon: Gamepad2, 
    color: 'text-purple-500', 
    image: 'botmatrix-kookbot',
    fields: [
      { key: 'BOT_TOKEN', label: 'Bot Token', type: 'password', placeholder: 'Kook Bot Token' }
    ]
  },
  { 
    id: 'DingTalk', 
    name: '钉钉', 
    icon: MessageSquare, 
    color: 'text-blue-500', 
    image: 'botmatrix-dingtalkbot',
    fields: [
      { key: 'DINGTALK_APP_KEY', label: 'App Key', type: 'text', placeholder: 'DingTalk App Key' },
      { key: 'DINGTALK_APP_SECRET', label: 'App Secret', type: 'password', placeholder: 'DingTalk App Secret' }
    ]
  },
  { 
    id: 'Email', 
    name: '邮件', 
    icon: Mail, 
    color: 'text-orange-500', 
    image: 'botmatrix-emailbot',
    fields: [
      { key: 'EMAIL_HOST', label: 'SMTP 主机', type: 'text', placeholder: 'smtp.gmail.com' },
      { key: 'EMAIL_PORT', label: 'SMTP 端口', type: 'text', placeholder: '587' },
      { key: 'EMAIL_USER', label: '邮箱账号', type: 'text', placeholder: 'your-email@example.com' },
      { key: 'EMAIL_PASS', label: '邮箱密码', type: 'password', placeholder: 'SMTP 授权码' }
    ]
  },
  { 
    id: 'Slack', 
    name: 'Slack', 
    icon: Slack, 
    color: 'text-red-500', 
    image: 'botmatrix-slackbot',
    fields: [
      { key: 'SLACK_BOT_TOKEN', label: 'Bot Token', type: 'password', placeholder: 'xoxb-...' },
      { key: 'SLACK_APP_TOKEN', label: 'App Token', type: 'password', placeholder: 'xapp-...' }
    ]
  },
  { 
    id: 'Tencent', 
    name: '腾讯云', 
    icon: Shield, 
    color: 'text-blue-600', 
    image: 'botmatrix-tencentbot',
    fields: [
      { key: 'TENCENT_APP_ID', label: 'App ID', type: 'text', placeholder: '腾讯云 Bot AppID' },
      { key: 'TENCENT_TOKEN', label: 'Token', type: 'password', placeholder: '腾讯云 Bot Token' },
      { key: 'TENCENT_SECRET', label: 'Secret', type: 'password', placeholder: '腾讯云 Bot Secret' }
    ]
  },
  { 
    id: 'WeCom', 
    name: '企业微信', 
    icon: Globe, 
    color: 'text-green-500', 
    image: 'botmatrix-wecombot',
    fields: [
      { key: 'WECOM_CORP_ID', label: 'Corp ID', type: 'text', placeholder: '企业 ID' },
      { key: 'WECOM_AGENT_ID', label: 'Agent ID', type: 'text', placeholder: '应用 ID' },
      { key: 'WECOM_SECRET', label: 'Secret', type: 'password', placeholder: '应用 Secret' },
      { key: 'WECOM_TOKEN', label: 'Token', type: 'password', placeholder: '接收消息 Token' },
      { key: 'WECOM_AES_KEY', label: 'EncodingAESKey', type: 'password', placeholder: '消息加密 Key' }
    ]
  },
  { 
    id: 'Web', 
    name: 'Web 机器人', 
    icon: Globe, 
    color: 'text-cyan-500', 
    image: 'botmatrix-webbot',
    fields: [
      { key: 'WEB_SITE_ID', label: '站点 ID', type: 'text', placeholder: 'Site-001' },
      { key: 'WEB_TITLE', label: '窗口标题', type: 'text', placeholder: '在线客服' },
      { key: 'WEB_WELCOME_MSG', label: '欢迎语', type: 'text', placeholder: '您好，请问有什么可以帮您？' }
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
    alert('添加失败: ' + err);
  } finally {
    isLoading.value = false;
  }
};

const handleDeleteBot = async (botId: string) => {
  if (confirm('确定要删除这个机器人吗？此操作不可逆。')) {
    try {
      await botStore.removeBot(botId);
    } catch (err) {
      alert('删除失败: ' + err);
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
      // 处理 Docker 日志头 (8 字节) 并按行分割
      const rawLogs = data.logs || '';
      // 这是一个简单的处理，实际可能更复杂，因为 io.ReadAll 读到的是原始字节流
      // 在 JS 中，由于 logs 是字符串，我们只能尝试清理不可见字符
      currentLogs.value = rawLogs
        .split('\n')
        .map((line: string) => line.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '').trim())
        .filter((line: string) => line.length > 0);
    }
  } catch (err) {
    console.error('Failed to fetch logs:', err);
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
      
      <button 
        @click="showAddModal = true"
        class="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-[var(--matrix-color)] text-black text-xs font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20"
      >
        <Plus class="w-4 h-4" />
        {{ t('deploy_bot') }}
      </button>
    </div>

    <!-- Empty State -->
    <div v-if="botStore.bots.length === 0" class="flex flex-col items-center justify-center py-10 sm:py-20 bg-[var(--bg-card)]/50 backdrop-blur-md rounded-3xl border border-dashed border-[var(--border-color)]">
      <div class="w-12 h-12 sm:w-16 sm:h-16 bg-[var(--matrix-color)]/10 rounded-2xl flex items-center justify-center mb-4">
        <Bot class="w-6 h-6 sm:w-8 sm:h-8 text-[var(--matrix-color)]" />
      </div>
      <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)] mb-2">{{ t('no_bots') }}</h3>
      <p class="text-[var(--text-muted)] text-[10px] sm:text-sm mb-6 text-center max-w-xs px-4">{{ t('no_bots_desc') }}</p>
      <button 
        @click="showAddModal = true"
        class="px-6 py-2 border border-[var(--matrix-color)] text-[var(--matrix-color)] font-bold rounded-xl hover:bg-[var(--matrix-color)] hover:text-black transition-all"
      >
        {{ t('start_now') }}
      </button>
    </div>

    <!-- Bot Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
      <div v-for="bot in botStore.bots" :key="bot.id" class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] backdrop-blur-xl border border-[var(--border-color)] shadow-sm hover:shadow-xl transition-all duration-300 group relative overflow-hidden">
        <!-- Platform Background Icon -->
        <component :is="getPlatformIcon(bot.platform)" class="absolute -right-4 -bottom-4 w-24 sm:w-32 h-24 sm:h-32 opacity-[0.03] dark:opacity-[0.05] pointer-events-none" />

        <div class="flex items-start justify-between mb-4 sm:mb-6 relative z-10">
          <div class="flex items-center gap-3 sm:gap-4">
            <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center relative shadow-inner">
              <component :is="getPlatformIcon(bot.platform)" :class="['w-5 h-5 sm:w-6 sm:h-6', getPlatformColor(bot.platform)]" />
              <div :class="['absolute -bottom-1 -right-1 w-3.5 h-3.5 sm:w-4 sm:h-4 rounded-full border-4 border-[var(--bg-card)]', getStatusColor(bot.connected)]"></div>
            </div>
            <div class="min-w-0">
              <h3 class="font-bold text-sm sm:text-base text-[var(--text-main)] truncate">{{ bot.nickname || bot.id }}</h3>
              <p class="text-[8px] sm:text-[10px] text-[var(--text-muted)] uppercase tracking-widest font-bold">{{ bot.platform }}</p>
            </div>
          </div>
          <div class="flex gap-1 shrink-0">
            <button 
              @click="viewLogs(bot)"
              class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors"
              :title="t('view_logs')"
            >
              <Terminal class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
            <button class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors">
              <SettingsIcon class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </button>
            <button 
              @click="handleDeleteBot(bot.id)"
              class="p-1.5 sm:p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-[var(--text-muted)] hover:text-red-500 transition-colors"
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
          <button class="px-3 sm:px-4 py-2.5 sm:py-3 rounded-2xl border border-[var(--border-color)] text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 transition-all">
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
                  <span :class="['text-[8px] sm:text-[10px] font-bold uppercase tracking-tighter', newBot.platform === p.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]']">{{ p.name }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('docker_image') }}</label>
              <input 
                v-model="newBot.image"
                type="text" 
                class="w-full px-4 sm:px-5 py-2.5 sm:py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-[var(--text-main)] mono text-xs sm:text-sm"
                placeholder="镜像名称 (e.g. botmatrix-kookbot)"
              />
            </div>

            <div v-for="field in currentFields" :key="field.key" class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ field.label }}</label>
              <input 
                v-model="newBot.env[field.key]"
                :type="field.type" 
                class="w-full px-4 sm:px-5 py-2.5 sm:py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] focus:ring-0 transition-all text-[var(--text-main)] mono text-xs sm:text-sm"
                :placeholder="field.placeholder"
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
              {{ isLoading ? (t('deploying') || '正在部署...') : (t('deploy_now') || '立即部署') }}
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
                <h2 class="text-lg sm:text-xl font-bold text-[var(--text-main)] truncate max-w-[150px] sm:max-w-none">{{ t('realtime_logs') || '实时运行日志' }}</h2>
                <p class="text-[10px] text-[var(--text-muted)] font-medium uppercase tracking-widest truncate max-w-[150px] sm:max-w-none">{{ selectedBot?.nickname || selectedBot?.id }} ({{ selectedBot?.platform }})</p>
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
