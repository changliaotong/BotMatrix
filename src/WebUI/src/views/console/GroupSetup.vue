<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  Settings, 
  Save, 
  ChevronLeft, 
  MessageSquare, 
  Shield, 
  Zap, 
  Bot, 
  Bell,
  Search,
  Users,
  AlertTriangle,
  Lock,
  UserX,
  UserMinus,
  CheckCircle,
  Plus
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const loading = ref(true);
const saving = ref(false);
const groupInfo = ref<any>({
  id: 0,
  group_name: '',
  is_open: true,
  welcome_message: '',
  system_prompt: '',
  recall_keyword: '',
  warn_keyword: '',
  mute_keyword: '',
  kick_keyword: '',
  black_keyword: '',
  white_keyword: '',
  credit_keyword: '',
  recall_time: 120,
  is_recall: true,
  use_right: true,
  teach_right: true,
  admin_right: true,
  is_power_on: true,
  is_welcome_hint: true,
  is_exit_hint: false,
  is_kick_hint: false,
  is_change_hint: false,
  is_right_hint: false,
  is_ai: true,
  is_use_knowledgebase: false,
  is_mult_ai: false,
  context_count: 5,
  is_prop: false,
  is_pet: false,
  is_credit_system: true,
  is_auto_signin: true,
  is_owner_pay: false,
  is_send_help_info: true,
  card_name_prefix_boy: '',
  card_name_prefix_girl: '',
  card_name_prefix_manager: '',
  // 新增缺失字段初始化
  is_accept_new_member: true,
  reject_message: '',
  regex_request_join: '',
  mute_enter_count: 0,
  mute_keyword_count: 0,
  kick_count: 0,
  black_count: 0,
  is_mute_refresh: false,
  mute_refresh_count: 0,
  is_black_refresh: false,
  is_confirm_new: false,
  is_credit: false,
  is_hint_close: false,
  is_invite: false,
  invite_credit: 0,
  is_reply_image: false,
  is_reply_recall: false,
  is_voice_reply: false,
  voice_id: '',
  // 补充更多字段
  is_cloud_black: false,
  is_cloud_answer: 2, // 1:开启, 2:关闭
  is_require_prefix: false,
  is_sz84: false,
  is_warn: false,
  is_block: false,
  is_white: false,
  is_close_manager: false,
  is_black_exit: false,
  is_black_kick: false,
  is_black_share: false,
  is_change_enter: false,
  is_mute_enter: false,
  is_change_message: false,
  parent_group: 0,
  block_min: 0,
  city_name: '',
  fans_name: '',
});

const relatedGroups = ref<any[]>([]);
const activeTab = ref('basic');

// 关键词矩阵配置
const keywordTypes = [
  { key: 'spam', name: 'spam' },
  { key: 'image', name: 'image' },
  { key: 'url', name: 'url' },
  { key: 'dirty_word', name: 'dirty_word' },
  { key: 'ad', name: 'ad' },
  { key: 'recommend_group', name: 'recommend_group' },
  { key: 'recommend_friend', name: 'recommend_friend' },
  { key: 'merged_forward', name: 'merged_forward' },
];

const actionFields = [
  { key: 'recall_keyword', name: 'recall' },
  { key: 'credit_keyword', name: 'credit_deduct' },
  { key: 'warn_keyword', name: 'warn' },
  { key: 'mute_keyword', name: 'mute' },
  { key: 'kick_keyword', name: 'kick' },
  { key: 'black_keyword', name: 'blacklist' },
];

const getKeywordStatus = (field: string, typeName: string) => {
  const keyword = groupInfo.value[field] || '';
  const keys = keyword.split('|').map((k: string) => k.trim());
  const translatedType = t(typeName);
  return keys.includes(translatedType);
};

const toggleKeyword = (field: string, typeName: string) => {
  const keyword = groupInfo.value[field] || '';
  let keys = keyword.split('|').map((k: string) => k.trim()).filter((k: string) => k !== '');
  const translatedType = t(typeName);
  
  if (keys.includes(translatedType)) {
    keys = keys.filter((k: string) => k !== translatedType);
  } else {
    keys.push(translatedType);
  }
  
  groupInfo.value[field] = keys.join('|');
};

const fetchGroupData = async () => {
  const groupId = route.query.id;
  if (!groupId) {
    router.push({ name: 'console-contacts' });
    return;
  }

  loading.value = true;
  try {
    const data = await botStore.fetchGroupSetup();
    if (data.success && data.data && data.data.groups) {
      const groups = data.data.groups;
      relatedGroups.value = groups;
      const current = groups.find((g: any) => g.id.toString() === groupId.toString());
      if (current) {
        groupInfo.value = { ...current };
      } else {
        // If not found in list, try to fetch specific one if API supports it
        // For now, redirect if not found in the list we have access to
        router.push({ name: 'console-contacts' });
      }
    }
  } finally {
    loading.value = false;
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    const res = await botStore.updateGroupSetup(groupInfo.value);
    if (res.success) {
      // Success notification could go here
    }
  } finally {
    saving.value = false;
  }
};

onMounted(fetchGroupData);

const tabs = [
  { id: 'basic', name: t('group_basic_settings'), icon: Settings },
  { id: 'message', name: t('group_message_settings'), icon: MessageSquare },
  { id: 'keywords', name: t('group_keyword_settings'), icon: Shield },
  { id: 'ai', name: t('group_ai_settings'), icon: Bot },
  { id: 'advanced', name: t('group_advanced_settings'), icon: Zap },
];

const keywordFields = [
  { key: 'recall_keyword', name: t('recall_keyword'), color: 'text-orange-500' },
  { key: 'warn_keyword', name: t('warn_keyword'), color: 'text-yellow-500' },
  { key: 'mute_keyword', name: t('mute_keyword'), color: 'text-purple-500' },
  { key: 'kick_keyword', name: t('kick_keyword'), color: 'text-red-500' },
  { key: 'black_keyword', name: t('black_keyword'), color: 'text-gray-800' },
  { key: 'white_keyword', name: t('white_keyword'), color: 'text-green-500' },
  { key: 'credit_keyword', name: t('credit_keyword'), color: 'text-blue-500' },
];

</script>

<template>
  <div class="p-4 sm:p-6 max-w-5xl mx-auto space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <button 
          @click="router.back()"
          class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        >
          <ChevronLeft class="w-6 h-6 text-[var(--text-muted)]" />
        </button>
        <div>
          <h1 class="text-xl font-black text-[var(--text-main)] tracking-tight flex items-center gap-3">
            <Users class="w-8 h-8 text-[var(--matrix-color)]" /> {{ groupInfo.group_name || t('group_setup') }}
          </h1>
          <p class="text-[10px] font-mono text-[var(--text-muted)] uppercase tracking-widest">ID: {{ groupInfo.id }}</p>
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
          <h4 class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest mb-4 px-4">{ t('groupsetup.linked_groups') }</h4>
          <div class="space-y-2">
            <button 
              v-for="related in relatedGroups.slice(0, 5)"
              :key="related.id"
              @click="router.push({ name: 'console-group-setup', query: { id: related.id } })"
              class="w-full flex items-center gap-3 p-3 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-all text-left"
              :class="{ 'bg-[var(--matrix-color)]/5 border border-[var(--matrix-color)]/20': related.id === groupInfo.id }"
            >
              <div class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center text-[10px] font-black">
                {{ related.group_name?.substring(0, 1) }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-[10px] font-black text-[var(--text-main)] truncate">{{ related.group_name }}</p>
                <p class="text-[8px] font-mono text-[var(--text-muted)]">{{ related.id }}</p>
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
          <!-- Basic Settings Tab -->
          <div v-if="activeTab === 'basic'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <div class="flex items-center justify-between">
                <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                  <Settings class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('core_config') }}
                </h3>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)]">{{ t('enable_bot_service') }}</span>
                  <button 
                    @click="groupInfo.is_open = !groupInfo.is_open"
                    :class="groupInfo.is_open ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                    class="relative w-10 h-5 rounded-full transition-colors flex items-center px-1"
                  >
                    <div :class="groupInfo.is_open ? 'translate-x-5 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('group_name') }}</label>
                  <input 
                    v-model="groupInfo.group_name"
                    type="text" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('group_memo') }}</label>
                  <input 
                    v-model="groupInfo.group_memo"
                    type="text" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('robot_owner_name') }}</label>
                  <input 
                    v-model="groupInfo.robot_owner_name"
                    type="text" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('group_owner_name') }}</label>
                  <input 
                    v-model="groupInfo.group_owner_name"
                    type="text" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 pt-4 border-t border-[var(--border-color)]">
                <div v-for="field in ['use_right', 'teach_right', 'admin_right']" :key="field" class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <select 
                    v-model.number="groupInfo[field]"
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors appearance-none"
                  >
                    <option :value="1">{{ t('right_everyone') }}</option>
                    <option :value="2">{{ t('right_admin') }}</option>
                    <option :value="3">{{ t('right_white') }}</option>
                    <option :value="4">{{ t('right_owner') }}</option>
                  </select>
                </div>
              </div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-4 border-t border-[var(--border-color)]">
                <div v-for="field in ['is_power_on', 'is_require_prefix', 'is_sz84', 'is_cloud_black']" :key="field" class="flex flex-col items-center gap-2 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button 
                    @click="groupInfo[field] = !groupInfo[field]"
                    :class="groupInfo[field] ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'"
                    class="transition-colors"
                  >
                    <CheckCircle v-if="groupInfo[field]" class="w-6 h-6" />
                    <div v-else class="w-6 h-6 rounded-full border-2 border-current"></div>
                  </button>
                </div>
              </div>

              <div class="space-y-2 pt-4 border-t border-[var(--border-color)]">
                <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('is_cloud_answer') }}</label>
                <select 
                  v-model.number="groupInfo.is_cloud_answer"
                  class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors appearance-none"
                >
                  <option :value="1">{{ 开启全自 }}</option>
                  <option :value="2">{ t('common.close') }</option>
                  <option :value="3">{ t('groupsetup.keyword_only') }</option>
                </select>
              </div>

              <!-- Group System Info -->
              <div class="pt-6 border-t border-[var(--border-color)] space-y-4">
                <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                  <Settings class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('system_info') }}
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div v-for="field in ['insert_date', 'last_date']" :key="field" class="flex flex-col gap-1 p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent">
                    <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
                    <span class="text-xs font-mono font-bold text-[var(--text-main)]">{{ groupInfo[field] || 'N/A' }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Message Tab -->
          <div v-if="activeTab === 'message'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <MessageSquare class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('welcome_exit_message') }}
              </h3>
              
              <div class="space-y-4">
                <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <div class="space-y-1">
                    <span class="text-xs font-bold text-[var(--text-main)]">{{ t('welcome_message_title') }}</span>
                    <p class="text-[10px] text-[var(--text-muted)]">{{ groupInfo.is_welcome_hint ? t('status_enabled') : t('status_disabled') }}</p>
                  </div>
                  <button 
                    @click="groupInfo.is_welcome_hint = !groupInfo.is_welcome_hint"
                    :class="groupInfo.is_welcome_hint ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                    class="relative w-10 h-5 rounded-full transition-colors flex items-center px-1"
                  >
                    <div :class="groupInfo.is_welcome_hint ? 'translate-x-5 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
                
                <textarea 
                  v-model="groupInfo.welcome_message"
                  rows="4"
                  class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors resize-none"
                  :placeholder="t('welcome_message_placeholder')"
                ></textarea>
              </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div v-for="field in ['is_exit_hint', 'is_kick_hint', 'is_change_hint', 'is_right_hint', 'is_change_enter', 'is_mute_enter', 'is_change_message']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <span class="text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest">{{ t(field) }}</span>
                    <button 
                      @click="groupInfo[field] = !groupInfo[field]"
                      :class="groupInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                      class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5"
                    >
                      <div :class="groupInfo[field] ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                    </button>
                  </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 pt-4 border-t border-[var(--border-color)]">
                  <div class="space-y-2">
                    <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('is_accept_new_member') }}</label>
                    <select 
                      v-model.number="groupInfo.is_accept_new_member"
                      class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors appearance-none"
                    >
                      <option :value="1">{{ t('accept_auto') }}</option>
                      <option :value="0">{{ t('accept_none') }}</option>
                      <option :value="2">{{ t('accept_reject') }}</option>
                    </select>
                  </div>
                  <div class="space-y-2">
                    <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('reject_message') }}</label>
                    <input 
                      v-model="groupInfo.reject_message"
                      type="text" 
                      class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                      :placeholder="t('reject_message_placeholder')"
                    />
                  </div>
                </div>

                <div class="space-y-2 pt-4 border-t border-[var(--border-color)]">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('regex_request_join') }}</label>
                  <input 
                    v-model="groupInfo.regex_request_join"
                    type="text" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Keywords Tab -->
          <div v-if="activeTab === 'keywords'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Shield class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('keyword_management') }}
              </h3>

              <!-- 关键词矩阵 -->
              <div class="overflow-x-auto">
                <table class="w-full border-collapse">
                  <thead>
                    <tr>
                      <th class="p-2 text-left text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)]">{{ t('feature_type') }}</th>
                      <th v-for="type in keywordTypes" :key="type.key" class="p-2 text-center text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)]">
                        {{ t(type.name) }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="field in actionFields" :key="field.key" class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                      <td class="p-2 text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest border-b border-[var(--border-color)]">
                        {{ t(field.name) }}
                      </td>
                      <td v-for="type in keywordTypes" :key="type.key" class="p-2 text-center border-b border-[var(--border-color)]">
                        <button 
                          @click="toggleKeyword(field.key, type.name)"
                          :class="getKeywordStatus(field.key, type.name) ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'"
                          class="transition-colors"
                        >
                          <CheckCircle v-if="getKeywordStatus(field.key, type.name)" class="w-5 h-5 mx-auto" />
                          <div v-else class="w-5 h-5 rounded-full border-2 border-current mx-auto opacity-20"></div>
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="space-y-4 pt-6 border-t border-[var(--border-color)]">
                <div v-for="field in keywordFields" :key="field.key" class="space-y-2">
                  <div class="flex items-center justify-between">
                    <label :class="['text-[10px] font-black uppercase tracking-widest', field.color]">{{ field.name }}</label>
                  </div>
                  <div class="flex gap-2">
                    <input 
                      v-model="groupInfo[field.key]"
                      type="text" 
                      class="flex-1 p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                      :placeholder="t('keyword_placeholder')"
                    />
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-6 border-t border-[var(--border-color)]">
                <div v-for="field in ['is_recall', 'is_warn', 'is_block', 'is_white']" :key="field" class="flex flex-col items-center gap-2 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button 
                    @click="groupInfo[field] = !groupInfo[field]"
                    :class="groupInfo[field] ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'"
                    class="transition-colors"
                  >
                    <CheckCircle v-if="groupInfo[field]" class="w-6 h-6" />
                    <div v-else class="w-6 h-6 rounded-full border-2 border-current"></div>
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 pt-6 border-t border-[var(--border-color)]">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('recall_time') }}</label>
                  <input 
                    v-model.number="groupInfo.recall_time"
                    type="number" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- AI Tab -->
          <div v-if="activeTab === 'ai'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <div class="flex items-center justify-between">
                <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                  <Bot class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('ai_assistant_settings') }}
                </h3>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)]">{{ t('enable_ai_reply') }}</span>
                  <button 
                    @click="groupInfo.is_ai = !groupInfo.is_ai"
                    :class="groupInfo.is_ai ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                    class="relative w-10 h-5 rounded-full transition-colors flex items-center px-1"
                  >
                    <div :class="groupInfo.is_ai ? 'translate-x-5 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="space-y-4">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('ai_system_prompt') }}</label>
                  <textarea 
                    v-model="groupInfo.system_prompt"
                    rows="6"
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors resize-none"
                    :placeholder="t('ai_system_prompt_placeholder')"
                  ></textarea>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <div class="space-y-1">
                      <span class="text-xs font-bold text-[var(--text-main)]">{{ t('is_ai_knowledge') }}</span>
                      <p class="text-[10px] text-[var(--text-muted)]">{{ groupInfo.is_use_knowledgebase ? t('status_enabled') : t('status_disabled') }}</p>
                    </div>
                    <button 
                      @click="groupInfo.is_use_knowledgebase = !groupInfo.is_use_knowledgebase"
                      :class="groupInfo.is_use_knowledgebase ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                      class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5"
                    >
                      <div :class="groupInfo.is_use_knowledgebase ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                    </button>
                  </div>
                  <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <div class="space-y-1">
                      <span class="text-xs font-bold text-[var(--text-main)]">{{ t('is_ai_history') }}</span>
                      <p class="text-[10px] text-[var(--text-muted)]">{{ groupInfo.is_mult_ai ? t('status_enabled') : t('status_disabled') }}</p>
                    </div>
                    <button 
                      @click="groupInfo.is_mult_ai = !groupInfo.is_mult_ai"
                      :class="groupInfo.is_mult_ai ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                      class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5"
                    >
                      <div :class="groupInfo.is_mult_ai ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                    </button>
                  </div>
                </div>

                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('ai_history_count') }}</label>
                  <input 
                    v-model.number="groupInfo.context_count"
                    type="number" 
                    class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Advanced Tab -->
          <div v-if="activeTab === 'advanced'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
                <Zap class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('group_advanced_settings') }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div v-for="field in ['is_prop', 'is_pet', 'is_credit', 'is_credit_system', 'is_auto_signin', 'is_owner_pay', 'is_send_help_info', 'is_confirm_new', 'is_invite', 'is_reply_image', 'is_reply_recall', 'is_voice_reply', 'is_mute_refresh', 'is_black_refresh', 'is_block', 'is_white', 'is_warn', 'is_close_manager', 'is_black_exit', 'is_black_kick', 'is_black_share', 'is_hint_close']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-[10px] font-black text-[var(--text-main)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button 
                    @click="groupInfo[field] = !groupInfo[field]"
                    :class="groupInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'"
                    class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5"
                  >
                    <div :class="groupInfo[field] ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-2 sm:grid-cols-5 gap-4">
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('mute_enter_count') }}</label>
                  <input v-model.number="groupInfo.mute_enter_count" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('mute_keyword_count') }}</label>
                  <input v-model.number="groupInfo.mute_keyword_count" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('kick_count') }}</label>
                  <input v-model.number="groupInfo.kick_count" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('black_count') }}</label>
                  <input v-model.number="groupInfo.black_count" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('invite_credit') }}</label>
                  <input v-model.number="groupInfo.invite_credit" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('mute_refresh_count') }}</label>
                  <input v-model.number="groupInfo.mute_refresh_count" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('parent_group') }}</label>
                  <input v-model.number="groupInfo.parent_group" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('block_min') }}</label>
                  <input v-model.number="groupInfo.block_min" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('city_name') }}</label>
                  <input v-model="groupInfo.city_name" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('fans_name') }}</label>
                  <input v-model="groupInfo.fans_name" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('voice_id') }}</label>
                  <input v-model="groupInfo.voice_id" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] space-y-4">
                <h4 class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('group_name_auto_prefix') }}</h4>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div class="space-y-2">
                    <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('boy_prefix') }}</label>
                    <input v-model="groupInfo.card_name_prefix_boy" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-xs font-bold" />
                  </div>
                  <div class="space-y-2">
                    <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('girl_prefix') }}</label>
                    <input v-model="groupInfo.card_name_prefix_girl" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-xs font-bold" />
                  </div>
                  <div class="space-y-2">
                    <label class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('admin_prefix') }}</label>
                    <input v-model="groupInfo.card_name_prefix_manager" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-xs font-bold" />
                  </div>
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
