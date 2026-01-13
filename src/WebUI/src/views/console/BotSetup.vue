<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useBotStore } from '@/stores/bot';
import { useSystemStore } from '@/stores/system';
import { 
  ArrowLeft, 
  Save, 
  Bot, 
  Shield, 
  Zap, 
  Globe, 
  Lock,
  MessageSquare,
  CreditCard,
  Users,
  User,
  CheckCircle,
  XCircle,
  Settings,
  ChevronLeft,
  RefreshCw,
  Copy,
  Check
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const botStore = useBotStore();
const systemStore = useSystemStore();

const t = (key: string) => systemStore.t(key);

const botUin = ref<number>(0);
const allBots = ref<any[]>([]);
const loading = ref(true);
const saving = ref(false);
const activeTab = ref('basic');
const showCopyHint = ref(false);

const botInfo = ref<any>({
  bot_uin: 0,
  bot_name: '',
  bot_memo: '',
  wemcome_message: '',
  is_credit: false,
  is_group: false,
  is_private: false,
  valid: 1,
  is_freeze: false,
  is_block: false,
  is_vip: false,
  admin_id: 0,
  password: '',
  bot_type: 1,
  api_ip: '',
  api_port: '',
  api_key: '',
  is_signal_r: false,
  web_ui_token: '',
  web_ui_port: '',
  freeze_times: 0
});

const tabs = [
  { id: 'basic', name: t('basic_settings'), icon: Settings },
  { id: 'network', name: t('api_network_settings'), icon: Globe },
  { id: 'permissions', name: t('permission_settings'), icon: Lock },
  { id: 'security', name: t('security'), icon: Shield },
];

const passwordStrength = computed(() => {
  const pwd = botInfo.value.password || '';
  if (!pwd) return { score: 0, label: '', color: 'bg-gray-500/20' };
  
  let score = 0;
  if (pwd.length > 8) score++;
  if (/[A-Z]/.test(pwd)) score++;
  if (/[0-9]/.test(pwd)) score++;
  if (/[^A-Za-z0-9]/.test(pwd)) score++;
  
  if (score <= 1) return { score: 1, label: t('password_weak'), color: 'bg-red-500' };
  if (score <= 3) return { score: 2, label: t('password_medium'), color: 'bg-yellow-500' };
  return { score: 3, label: t('password_strong'), color: 'bg-green-500' };
});

const generateApiKey = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = 'bm_';
  for (let i = 0; i < 32; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  botInfo.value.api_key = result;
};

const copyApiKey = async () => {
  if (!botInfo.value.api_key) return;
  try {
    await navigator.clipboard.writeText(botInfo.value.api_key);
    showCopyHint.value = true;
    setTimeout(() => {
      showCopyHint.value = false;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy:', err);
  }
};

const fetchBotInfo = async () => {
  const uinStr = route.query.bot_uin as string;
  const adminIdStr = route.query.admin_id as string;
  const isAdmin = systemStore.user?.is_admin;
  
  if (!uinStr && !isAdmin) {
    router.push('/console/bots');
    return;
  }
  
  if (uinStr) {
    botUin.value = parseInt(uinStr);
  }
  
  loading.value = true;
  
  try {
    const res = await botStore.fetchMemberSetup(adminIdStr);
    if (res.success && res.data && res.data.bots) {
      allBots.value = res.data.bots;
      
      if (botUin.value) {
        const found = allBots.value.find((b: any) => b.bot_uin === botUin.value);
        if (found) {
          botInfo.value = { ...botInfo.value, ...found };
        } else if (!isAdmin) {
          router.push('/console/bots');
        }
      } else if (isAdmin && allBots.value.length > 0) {
        botInfo.value = { ...botInfo.value, ...allBots.value[0] };
        botUin.value = botInfo.value.bot_uin;
      }
    }
  } catch (err) {
    console.error('Failed to fetch bot info:', err);
  } finally {
    loading.value = false;
  }
};

const handleBotChange = (uin: number) => {
  const found = allBots.value.find((b: any) => b.bot_uin === uin);
  if (found) {
    botInfo.value = { ...botInfo.value, ...found };
    botUin.value = uin;
    router.replace({ query: { ...route.query, bot_uin: uin } });
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    const res = await botStore.updateMemberSetup(botInfo.value);
    if (res.success) {
      // Success toast would be nice here
    }
  } catch (err) {
    console.error('Save error:', err);
  } finally {
    saving.value = false;
  }
};

onMounted(fetchBotInfo);
</script>

<template>
  <div class="p-4 sm:p-6 max-w-5xl mx-auto space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <button @click="router.back()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
          <ChevronLeft class="w-6 h-6 text-[var(--text-muted)]" />
        </button>
        <div>
          <h1 class="text-xl font-black text-[var(--text-main)] tracking-tight flex items-center gap-3">
            <Bot class="w-8 h-8 text-[var(--matrix-color)]" /> {{ botInfo.bot_name || t('bot_setup') }}
          </h1>
          <p class="text-[10px] font-mono text-[var(--text-muted)] uppercase tracking-widest">UIN: {{ botInfo.bot_uin }}</p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <button 
          @click="handleSave" 
          :disabled="saving"
          class="w-full sm:w-auto px-6 py-2 bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest rounded-xl hover:opacity-90 transition-all flex items-center justify-center gap-2 shadow-lg shadow-[var(--matrix-color)]/20 disabled:opacity-50"
        >
          <Save v-if="!saving" class="w-4 h-4" />
          <div v-else class="w-4 h-4 border-2 border-black border-t-transparent rounded-full animate-spin"></div>
          {{ saving ? t('saving') : t('save_settings') }}
        </button>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- Sidebar Tabs -->
      <div class="flex md:flex-col overflow-x-auto pb-2 md:pb-0 gap-2 md:col-span-1 no-scrollbar">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="activeTab === tab.id ? 'bg-[var(--matrix-color)] text-black shadow-lg shadow-[var(--matrix-color)]/20' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap"
        >
          <component :is="tab.icon" class="w-5 h-5" /> {{ tab.name }}
        </button>

        <div class="hidden md:block mt-6 pt-6 border-t border-[var(--border-color)]">
          <h4 class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-4 px-4">{{ t('linked_bots') }}</h4>
          <div class="space-y-2">
            <button 
              v-for="bot in allBots.slice(0, 10)" 
              :key="bot.bot_uin"
              @click="handleBotChange(bot.bot_uin)"
              class="w-full flex items-center gap-3 p-3 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-all text-left"
              :class="{ 'bg-[var(--matrix-color)]/5 border border-[var(--matrix-color)]/20': bot.bot_uin === botInfo.bot_uin }"
            >
              <div class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center text-[10px] font-black">
                {{ bot.bot_name?.substring(0, 1) || 'B' }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-[10px] font-black text-[var(--text-main)] truncate">{{ bot.bot_name || bot.bot_uin }}</p>
                <p class="text-[8px] font-mono text-[var(--text-muted)]">{{ bot.bot_uin }}</p>
              </div>
            </button>
          </div>
        </div>
      </div>

      <!-- Content Area -->
      <div class="md:col-span-3 space-y-6">
        <div v-if="loading" class="space-y-6 animate-pulse">
          <div v-for="i in 3" :key="i" class="h-48 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
        </div>

        <div v-else class="space-y-6">
          <!-- Basic Tab -->
          <div v-if="activeTab === 'basic'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Settings class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('basic_settings') }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('bot_name') }}</label>
                  <input v-model="botInfo.bot_name" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('bot_type') }}</label>
                  <select v-model.number="botInfo.bot_type" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors appearance-none">
                    <option :value="1">{{ t('bot_type_normal') }}</option>
                    <option :value="2">{{ t('bot_type_manager') }}</option>
                    <option :value="3">{{ t('bot_type_system') }}</option>
                  </select>
                </div>
                <div class="space-y-2 sm:col-span-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('bot_memo') }}</label>
                  <input v-model="botInfo.bot_memo" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2 sm:col-span-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('welcome_message') }}</label>
                  <textarea v-model="botInfo.wemcome_message" rows="4" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors resize-none"></textarea>
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)]">
                <h4 class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-4">{{ t('status_controls') }}</h4>
                <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                  <div class="flex flex-col items-center gap-2 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <span class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('is_valid') }}</span>
                    <button @click="botInfo.valid = botInfo.valid === 1 ? 0 : 1" :class="botInfo.valid === 1 ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" class="transition-colors">
                      <CheckCircle v-if="botInfo.valid === 1" class="w-6 h-6" />
                      <div v-else class="w-6 h-6 rounded-full border-2 border-current"></div>
                    </button>
                  </div>
                  <div v-for="field in ['is_freeze', 'is_block', 'is_vip', 'is_signal_r']" :key="field" class="flex flex-col items-center gap-2 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <span class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
                    <button @click="botInfo[field] = !botInfo[field]" :class="botInfo[field] ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" class="transition-colors">
                      <CheckCircle v-if="botInfo[field]" class="w-6 h-6" />
                      <div v-else class="w-6 h-6 rounded-full border-2 border-current"></div>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Network Tab -->
          <div v-if="activeTab === 'network'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Globe class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('api_network_settings') }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('api_ip') }}</label>
                  <input v-model="botInfo.api_ip" type="text" placeholder="0.0.0.0" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('api_port') }}</label>
                  <input v-model="botInfo.api_port" type="text" placeholder="8080" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2 sm:col-span-2">
                  <div class="flex items-center justify-between">
                    <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('api_key') }}</label>
                    <div class="flex items-center gap-3">
                      <button @click="copyApiKey" class="text-[10px] font-black text-[var(--matrix-color)] hover:opacity-80 transition-all flex items-center gap-1">
                        <component :is="showCopyHint ? Check : Copy" class="w-3 h-3" /> {{ showCopyHint ? t('copy_success') : t('copy_key') }}
                      </button>
                      <button @click="generateApiKey" class="text-[10px] font-black text-[var(--matrix-color)] hover:opacity-80 transition-all flex items-center gap-1">
                        <RefreshCw class="w-3 h-3" /> {{ t('generate_key') }}
                      </button>
                    </div>
                  </div>
                  <input v-model="botInfo.api_key" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-mono font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('web_ui_token') }}</label>
                  <input v-model="botInfo.web_ui_token" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('web_ui_port') }}</label>
                  <input v-model="botInfo.web_ui_port" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>
            </div>
          </div>

          <!-- Permissions Tab -->
          <div v-if="activeTab === 'permissions'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Lock class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('permission_settings') }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div v-for="field in ['is_credit', 'is_group', 'is_private']" :key="field" class="flex flex-col items-center gap-4 p-6 rounded-3xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] transition-all hover:border-[var(--matrix-color)]/30">
                  <div class="p-3 rounded-2xl bg-white/5">
                    <CreditCard v-if="field === 'is_credit'" class="w-6 h-6 text-blue-500" />
                    <Users v-else-if="field === 'is_group'" class="w-6 h-6 text-purple-500" />
                    <User v-else class="w-6 h-6 text-orange-500" />
                  </div>
                  <div class="text-center">
                    <span class="text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest block mb-1">{{ t(field + '_bot') }}</span>
                    <p class="text-[8px] text-[var(--text-muted)] uppercase tracking-widest">{{ botInfo[field] ? t('status_enabled') : t('status_disabled') }}</p>
                  </div>
                  <button @click="botInfo[field] = !botInfo[field]" :class="botInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-12 h-6 rounded-full transition-colors flex items-center px-1">
                    <div :class="botInfo[field] ? 'translate-x-6 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-4 h-4 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Security Tab -->
          <div v-if="activeTab === 'security'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Shield class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('security') }}
              </h3>

              <div class="space-y-6">
                <div v-if="systemStore.user?.is_admin" class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('admin_id') }}</label>
                  <input v-model.number="botInfo.admin_id" type="number" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>

                <div class="space-y-4">
                  <div class="space-y-2">
                    <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('password') }}</label>
                    <input v-model="botInfo.password" type="password" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                  </div>
                  
                  <!-- Password Strength Meter -->
                  <div class="p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] space-y-3">
                    <div class="flex items-center justify-between">
                      <span class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('password_strength') }}</span>
                      <span class="text-[10px] font-black uppercase tracking-widest" :class="passwordStrength.color.replace('bg-', 'text-')">{{ passwordStrength.label }}</span>
                    </div>
                    <div class="flex gap-1">
                      <div v-for="i in 3" :key="i" class="h-1 flex-1 rounded-full transition-all duration-500" :class="i <= passwordStrength.score ? passwordStrength.color : 'bg-gray-500/10'"></div>
                    </div>
                  </div>
                </div>

                <div v-if="botInfo.freeze_times > 0" class="flex items-center justify-between p-4 rounded-2xl bg-red-500/5 border border-red-500/20">
                  <div class="flex items-center gap-3">
                    <XCircle class="w-5 h-5 text-red-500" />
                    <span class="text-xs font-bold text-[var(--text-main)]">{{ t('freeze_times') }}</span>
                  </div>
                  <span class="px-3 py-1 rounded-lg bg-red-500 text-white text-[10px] font-black">{{ botInfo.freeze_times }}</span>
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
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>