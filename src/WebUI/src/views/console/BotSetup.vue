<script setup lang="ts">
import { ref, onMounted } from 'vue';
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
  CheckCircle2,
  XCircle,
  Settings,
  ChevronDown
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const botStore = useBotStore();
const systemStore = useSystemStore();

const t = (key: string) => systemStore.t(key);

const botUin = ref<number>(0);
const allBots = ref<any[]>([]);
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
  web_ui_port: ''
});

const isLoading = ref(false);
const isSaving = ref(false);

const fetchBotInfo = async () => {
  const uinStr = route.query.bot_uin as string;
  const isAdmin = systemStore.user?.is_admin;
  
  if (!uinStr && !isAdmin) {
    router.push('/console/bots');
    return;
  }
  
  if (uinStr) {
    botUin.value = parseInt(uinStr);
  }
  
  isLoading.value = true;
  
  try {
    const res = await botStore.fetchMemberSetup();
    if (res.success && res.data && res.data.bots) {
      allBots.value = res.data.bots;
      
      if (botUin.value) {
        const found = allBots.value.find((b: any) => b.bot_uin === botUin.value);
        if (found) {
          botInfo.value = { ...found };
        } else if (!isAdmin) {
          alert(t('bot_not_found'));
          router.push('/console/bots');
        }
      } else if (isAdmin && allBots.value.length > 0) {
        // If admin and no bot specified, pick the first one as default
        botInfo.value = { ...allBots.value[0] };
        botUin.value = botInfo.value.bot_uin;
      }
    }
  } catch (err) {
    console.error('Failed to fetch bot info:', err);
  } finally {
    isLoading.value = false;
  }
};

const handleBotChange = (uin: number) => {
  const found = allBots.value.find((b: any) => b.bot_uin === uin);
  if (found) {
    botInfo.value = { ...found };
    botUin.value = uin;
    // Update query param without reloading
    router.replace({ query: { ...route.query, bot_uin: uin } });
  }
};

const handleSave = async () => {
  isSaving.value = true;
  try {
    const res = await botStore.updateMemberSetup(botInfo.value);
    if (res.success) {
      alert(t('save_success'));
    } else {
      alert(t('save_failed') + ': ' + (res.message || 'Unknown error'));
    }
  } catch (err) {
    alert(t('save_failed') + ': ' + err);
  } finally {
    isSaving.value = false;
  }
};

onMounted(() => {
  fetchBotInfo();
});
</script>

<template>
  <div class="p-4 sm:p-8 space-y-6 max-w-5xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <button 
          @click="router.back()" 
          class="p-2 hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors text-[var(--text-muted)]"
        >
          <ArrowLeft class="w-6 h-6" />
        </button>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('bot_setup') }}</h1>
          <div class="flex items-center gap-2">
            <p v-if="!systemStore.user?.is_admin || allBots.length <= 1" class="text-[var(--text-muted)] text-[10px] font-bold tracking-widest uppercase">ID: {{ botUin }}</p>
            <div v-else class="relative flex items-center gap-2">
              <span class="text-[var(--text-muted)] text-[10px] font-bold tracking-widest uppercase">ID:</span>
              <select 
                :value="botUin" 
                @change="(e: any) => handleBotChange(parseInt(e.target.value))"
                class="bg-transparent text-[10px] font-bold tracking-widest uppercase text-[var(--matrix-color)] outline-none cursor-pointer border-b border-[var(--matrix-color)]/30 hover:border-[var(--matrix-color)] transition-colors"
              >
                <option v-for="bot in allBots" :key="bot.bot_uin" :value="bot.bot_uin" class="bg-[var(--bg-card)] text-[var(--text-main)]">
                  {{ bot.bot_uin }} {{ bot.bot_name ? `(${bot.bot_name})` : '' }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div>
      
      <button 
        @click="handleSave"
        :disabled="isSaving"
        class="flex items-center gap-2 px-6 py-3 rounded-2xl bg-[var(--matrix-color)] text-black text-xs font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20 disabled:opacity-50"
      >
        <Save v-if="!isSaving" class="w-4 h-4" />
        <span v-else class="w-4 h-4 border-2 border-black/30 border-t-black rounded-full animate-spin"></span>
        {{ isSaving ? t('saving') : t('save_settings') }}
      </button>
    </div>

    <div v-if="isLoading" class="flex flex-col items-center justify-center py-20 gap-4">
      <div class="w-12 h-12 border-4 border-[var(--matrix-color)]/20 border-t-[var(--matrix-color)] rounded-full animate-spin"></div>
      <p class="text-[var(--text-muted)] text-xs font-bold uppercase tracking-widest">{{ t('loading_config') }}</p>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left Column: Basic Info -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Basic Settings -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-blue-500/10 rounded-xl">
              <Bot class="w-5 h-5 text-blue-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('basic_settings') }}</h2>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div v-if="systemStore.user?.is_admin" class="space-y-2 sm:col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('admin_id') }}</label>
              <input 
                v-model.number="botInfo.admin_id"
                type="number"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                :placeholder="t('admin_id_placeholder')"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('bot_name') }}</label>
              <input 
                v-model="botInfo.bot_name"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                :placeholder="t('bot_name_placeholder')"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('bot_type') }}</label>
              <select 
                v-model.number="botInfo.bot_type"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all appearance-none"
              >
                <option :value="1">{{ t('bot_type_normal') }}</option>
                <option :value="2">{{ t('bot_type_manager') }}</option>
                <option :value="3">{{ t('bot_type_system') }}</option>
              </select>
            </div>
            <div class="space-y-2 sm:col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('bot_memo') }}</label>
              <input 
                v-model="botInfo.bot_memo"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                :placeholder="t('bot_memo_placeholder')"
              />
            </div>
            <div class="space-y-2 sm:col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('welcome_message') }}</label>
              <textarea 
                v-model="botInfo.wemcome_message"
                rows="4"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all resize-none"
                :placeholder="t('welcome_message_placeholder')"
              ></textarea>
            </div>
          </div>
        </section>

        <!-- API & Network Settings -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-purple-500/10 rounded-xl">
              <Globe class="w-5 h-5 text-purple-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('api_network_settings') }}</h2>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('api_ip') }}</label>
              <input 
                v-model="botInfo.api_ip"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                placeholder="0.0.0.0"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('api_port') }}</label>
              <input 
                v-model="botInfo.api_port"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                placeholder="8080"
              />
            </div>
            <div class="space-y-2 sm:col-span-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('api_key') }}</label>
              <input 
                v-model="botInfo.api_key"
                type="password"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                placeholder="Enter API key"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('web_ui_token') }}</label>
              <input 
                v-model="botInfo.web_ui_token"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                placeholder="WebUI Access Token"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('web_ui_port') }}</label>
              <input 
                v-model="botInfo.web_ui_port"
                type="text"
                class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
                placeholder="WebUI Port"
              />
            </div>
          </div>
        </section>
      </div>

      <!-- Right Column: Status & Toggles -->
      <div class="space-y-6">
        <!-- Status Panel -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-green-500/10 rounded-xl">
              <Zap class="w-5 h-5 text-green-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('status_controls') }}</h2>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/20 transition-all">
              <div class="flex items-center gap-3">
                <CheckCircle2 class="w-4 h-4 text-green-500" />
                <span class="text-xs font-bold text-[var(--text-main)]">{{ t('is_valid') }}</span>
              </div>
              <select v-model.number="botInfo.valid" class="bg-transparent text-xs font-bold outline-none text-[var(--matrix-color)]">
                <option :value="1">{{ t('active') }}</option>
                <option :value="0">{{ t('inactive') }}</option>
              </select>
            </div>

            <div v-for="field in ['is_freeze', 'is_block', 'is_vip', 'is_signal_r']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/20 transition-all">
              <div class="flex items-center gap-3">
                <Shield v-if="field.includes('freeze') || field.includes('block')" class="w-4 h-4 text-red-500" />
                <Zap v-else class="w-4 h-4 text-yellow-500" />
                <span class="text-xs font-bold text-[var(--text-main)]">{{ t(field) }}</span>
              </div>
              <button 
                @click="botInfo[field] = !botInfo[field]"
                :class="['w-10 h-5 rounded-full relative transition-all duration-300', botInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/30']"
              >
                <div :class="['absolute top-1 w-3 h-3 rounded-full bg-white transition-all duration-300', botInfo[field] ? 'left-6' : 'left-1']"></div>
              </button>
            </div>

            <!-- Freeze Times Display -->
            <div v-if="botInfo.freeze_times > 0" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent">
              <div class="flex items-center gap-3">
                <Shield class="w-4 h-4 text-orange-500" />
                <span class="text-xs font-bold text-[var(--text-main)]">{{ t('freeze_times') }}</span>
              </div>
              <span class="text-xs font-mono font-bold text-[var(--text-muted)]">{{ botInfo.freeze_times }}</span>
            </div>
          </div>
        </section>

        <!-- Permissions Panel -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-orange-500/10 rounded-xl">
              <Lock class="w-5 h-5 text-orange-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('permission_settings') }}</h2>
          </div>

          <div class="space-y-4">
            <div v-for="field in ['is_credit', 'is_group', 'is_private']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-[var(--matrix-color)]/20 transition-all">
              <div class="flex items-center gap-3">
                <CreditCard v-if="field.includes('credit')" class="w-4 h-4 text-blue-500" />
                <Users v-else-if="field.includes('group')" class="w-4 h-4 text-cyan-500" />
                <User v-else class="w-4 h-4 text-indigo-500" />
                <span class="text-xs font-bold text-[var(--text-main)]">{{ t(field + '_bot') }}</span>
              </div>
              <button 
                @click="botInfo[field] = !botInfo[field]"
                :class="['w-10 h-5 rounded-full relative transition-all duration-300', botInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/30']"
              >
                <div :class="['absolute top-1 w-3 h-3 rounded-full bg-white transition-all duration-300', botInfo[field] ? 'left-6' : 'left-1']"></div>
              </button>
            </div>
          </div>
        </section>

        <!-- Security Panel -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-red-500/10 rounded-xl">
              <Lock class="w-5 h-5 text-red-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('security') }}</h2>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest ml-2">{{ t('password') }}</label>
            <input 
              v-model="botInfo.password"
              type="password"
              class="w-full px-5 py-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-all"
              :placeholder="t('password_placeholder')"
            />
          </div>
        </section>

        <!-- Stats & Dates Panel (Read-only) -->
        <section class="bg-[var(--bg-card)] rounded-[2.5rem] border border-[var(--border-color)] p-6 sm:p-8 space-y-6">
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-blue-500/10 rounded-xl">
              <Settings class="w-5 h-5 text-blue-500" />
            </div>
            <h2 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wider">{{ t('system_info') }}</h2>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div v-for="field in ['insert_date', 'valid_date', 'last_date', 'heartbeat_date', 'receive_date', 'block_date', 'freeze_date']" :key="field" class="flex flex-col gap-1 p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent">
              <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
              <span class="text-xs font-mono font-bold text-[var(--text-main)]">{{ botInfo[field] || 'N/A' }}</span>
            </div>
          </div>
        </section>
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
