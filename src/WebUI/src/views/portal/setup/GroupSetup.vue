<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { useAuthStore } from '@/stores/auth';
import PortalHeader from '@/components/layout/PortalHeader.vue';
import PortalFooter from '@/components/layout/PortalFooter.vue';
import { 
  Settings, 
  Save, 
  ChevronLeft, 
  MessageSquare, 
  Shield, 
  Zap, 
  Bot, 
  Users,
  CheckCircle,
  User,
  Search
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const systemStore = useSystemStore();
const botStore = useBotStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);

const isBlankLayout = computed(() => route.meta.layout === 'blank');

const loading = ref(true);
const saving = ref(false);
const activeTab = ref('basic');

const groupInfo = ref<any>({
  id: 0,
  group_name: '',
  group_memo: '',
  group_owner: '',
  robot_owner: '',
  robot_owner_name: '',
  group_owner_name: '',
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
  use_right: 1,
  teach_right: 1,
  admin_right: 1,
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
  is_credit: false,
  is_credit_system: true,
  is_auto_signin: true,
  is_owner_pay: false,
  is_send_help_info: true,
  is_confirm_new: false,
  is_invite: false,
  is_reply_image: false,
  is_reply_recall: false,
  is_voice_reply: false,
  is_mute_refresh: false,
  is_black_refresh: false,
  is_block: false,
  is_white: false,
  is_warn: false,
  is_close_manager: false,
  is_black_exit: false,
  is_black_kick: false,
  is_black_share: false,
  is_hint_close: false,
  is_accept_new_member: 1,
  reject_message: '',
  regex_request_join: '',
  mute_enter_count: 0,
  mute_keyword_count: 0,
  kick_count: 0,
  black_count: 0,
  invite_credit: 0,
  mute_refresh_count: 0,
  parent_group: 0,
  block_min: 0,
  city_name: '',
  fans_name: '',
  voice_id: '',
  card_name_prefix_boy: '',
  card_name_prefix_girl: '',
  card_name_prefix_manager: '',
  is_require_prefix: false,
  is_sz84: false,
  is_cloud_black: false,
  is_cloud_answer: 2,
  is_change_enter: false,
  is_mute_enter: false,
  is_change_message: false,
});

const relatedGroups = ref<any[]>([]);

const tabs = [
  { id: 'basic', name: t('group_basic_settings'), icon: Settings },
  { id: 'message', name: t('group_message_settings'), icon: MessageSquare },
  { id: 'keywords', name: t('group_keyword_settings'), icon: Shield },
  { id: 'ai', name: t('group_ai_settings'), icon: Bot },
  { id: 'advanced', name: t('group_advanced_settings'), icon: Zap },
];

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

const groupSearchQuery = ref('');

const filteredGroups = computed(() => {
  if (!groupSearchQuery.value) return relatedGroups.value;
  const q = groupSearchQuery.value.toLowerCase();
  return relatedGroups.value.filter(g => 
    (g.group_name && g.group_name.toLowerCase().includes(q)) || 
    (g.id && g.id.toString().includes(q))
  );
});

const fetchGroupData = async () => {
  const groupId = route.query.id;
  // Use personal QQ from auth store as robot owner for filtering
  const robotOwner = authStore.user?.qq || (route.query.robot_owner as string);
  const groupOwner = undefined; 
  
  loading.value = true;
  try {
    // Call with robotOwner and groupOwner (personal QQ)
    const data = await botStore.fetchGroupSetup(robotOwner, groupOwner);
    if (data.success && data.data && data.data.groups) {
      const groups = data.data.groups;
      relatedGroups.value = groups;
      
      if (groupId) {
        const current = groups.find((g: any) => g.id.toString() === groupId.toString());
        if (current) {
          groupInfo.value = { ...groupInfo.value, ...current };
        } else if (groups.length > 0) {
          groupInfo.value = { ...groupInfo.value, ...groups[0] };
          router.replace({ query: { ...route.query, id: groups[0].id } });
        }
      } else if (groups.length > 0) {
        groupInfo.value = { ...groupInfo.value, ...groups[0] };
        router.replace({ query: { ...route.query, id: groups[0].id } });
      }
    }
  } catch (error) {
    console.error('Fetch error:', error);
  } finally {
    loading.value = false;
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    const res = await botStore.updateGroupSetup(groupInfo.value);
    if (res.success) {
      // Success toast
    }
  } catch (error) {
    console.error('Save error:', error);
  } finally {
    saving.value = false;
  }
};

onMounted(fetchGroupData);

watch(() => route.query.id, (newId) => {
  if (newId) {
    const current = relatedGroups.value.find((g: any) => g.id.toString() === newId.toString());
    if (current) {
      groupInfo.value = { ...groupInfo.value, ...current };
    }
  }
});

watch(() => route.query.admin_id, () => {
  fetchGroupData();
});
</script>

<template>
  <div class="min-h-screen bg-[var(--bg-body)] text-[var(--text-main)] selection:bg-[var(--matrix-color)]/30 overflow-x-hidden" :class="[systemStore.style]">
    <div class="fixed inset-0 pointer-events-none bg-[radial-gradient(circle_at_50%_-20%,var(--matrix-color),transparent_50%)] opacity-10"></div>
    <PortalHeader v-if="isBlankLayout" />
    
    <div :class="[isBlankLayout ? 'pt-40 pb-32' : 'py-6']" class="p-4 sm:p-6 max-w-7xl mx-auto space-y-6">
      <!-- Header -->
      <div class="sticky top-[72px] z-30 -mx-4 px-4 sm:-mx-6 sm:px-6 py-2 bg-[var(--bg-body)]/80 backdrop-blur-md border-b border-[var(--border-color)] mb-8 transition-all duration-300">
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <button @click="router.back()" class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <ChevronLeft class="w-5 h-5 text-[var(--text-muted)]" />
            </button>
            <div class="flex flex-col sm:flex-row sm:items-baseline sm:gap-3">
              <h1 class="text-base sm:text-xl font-black text-[var(--text-main)] tracking-tight flex items-center gap-2">
                <Users class="w-5 h-5 text-[var(--matrix-color)]" /> {{ groupInfo.group_name || t('group_setup') }}
              </h1>
              <div class="flex items-center gap-2">
                <p class="text-[11px] font-mono text-[var(--text-muted)] uppercase tracking-widest opacity-60">ID: {{ groupInfo.id }}</p>
                <div v-if="authStore.user?.qq" class="flex items-center gap-1.5 px-2 py-0.5 bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 rounded-full">
                  <User class="w-3 h-3 text-[var(--matrix-color)]" />
                  <span class="text-[9px] font-black text-[var(--matrix-color)] uppercase tracking-widest">{{ authStore.user.qq }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <button 
              @click="handleSave" 
              :disabled="saving"
              class="px-4 py-1.5 bg-[var(--matrix-color)] text-black font-black text-xs sm:text-sm uppercase tracking-widest rounded-lg hover:opacity-90 transition-all flex items-center justify-center gap-2 shadow-lg shadow-[var(--matrix-color)]/20 disabled:opacity-50"
            >
              <Save v-if="!saving" class="w-3.5 h-3.5" />
              <div v-else class="w-3.5 h-3.5 border-2 border-black border-t-transparent rounded-full animate-spin"></div>
              <span class="hidden sm:inline">{{ saving ? t('saving') : t('save_settings') }}</span>
              <span class="sm:hidden">{{ saving ? '...' : t('save') }}</span>
            </button>
          </div>
        </div>
      </div>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- Sidebar Tabs -->
      <div class="md:sticky md:top-[140px] h-fit flex md:flex-col overflow-x-auto pb-2 md:pb-0 gap-2 md:col-span-1 no-scrollbar z-20">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="activeTab === tab.id ? 'bg-[var(--matrix-color)] text-black shadow-lg shadow-[var(--matrix-color)]/20' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-base uppercase tracking-widest transition-all whitespace-nowrap"
        >
          <component :is="tab.icon" class="w-5 h-5" /> {{ tab.name }}
        </button>

        <div class="hidden md:block mt-6 pt-6 border-t border-[var(--border-color)]">
          <h4 class="text-base font-black text-[var(--text-muted)] uppercase tracking-widest mb-4 px-4">{{ t('linked_groups') }}</h4>
          
          <!-- Group Search -->
          <div class="px-2 mb-4">
            <div class="relative group">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)] group-focus-within:text-[var(--matrix-color)] transition-colors" />
              <input 
                v-model="groupSearchQuery"
                type="text" 
                :placeholder="t('search_groups')"
                class="w-full pl-9 pr-4 py-2 bg-black/5 dark:bg-white/5 border border-[var(--border-color)] rounded-xl text-xs font-bold focus:outline-none focus:border-[var(--matrix-color)]/50 transition-all"
              >
            </div>
          </div>

          <div class="space-y-2 max-h-[400px] overflow-y-auto custom-scrollbar px-1">
            <button 
              v-for="related in filteredGroups" 
              :key="related.id"
              @click="router.push({ name: 'portal-group-setup', query: { ...route.query, id: related.id } })"
              class="w-full flex items-center gap-3 p-3 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-all text-left group/item"
              :class="{ 'bg-[var(--matrix-color)]/5 border border-[var(--matrix-color)]/20': String(related.id) === String(groupInfo.id) }"
            >
              <div class="w-10 h-10 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center text-base font-black shrink-0 group-hover/item:bg-[var(--matrix-color)]/10 transition-colors">
                {{ (related.group_name || 'G').substring(0, 1) }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-black text-[var(--text-main)] truncate group-hover/item:text-[var(--matrix-color)] transition-colors">{{ related.group_name || 'Unnamed Group' }}</p>
                <p class="text-[10px] font-mono text-[var(--text-muted)]">{{ related.id }}</p>
              </div>
            </button>
            
            <div v-if="filteredGroups.length === 0" class="py-10 text-center">
              <p class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('no_groups_found') }}</p>
            </div>
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
              <div class="flex items-center justify-between">
                <h3 class="text-lg font-black uppercase tracking-widest flex items-center gap-2">
                  <Settings class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('core_config') }}
                </h3>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-black uppercase tracking-widest text-[var(--text-muted)]">{{ t('enable_bot_service') }}</span>
                  <button @click="groupInfo.is_open = !groupInfo.is_open" :class="groupInfo.is_open ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-12 h-6 rounded-full transition-colors flex items-center px-1">
                    <div :class="groupInfo.is_open ? 'translate-x-6 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-4 h-4 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div v-for="field in ['group_name', 'group_memo', 'robot_owner', 'group_owner_name', 'group_owner']" :key="field" class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <input v-model="groupInfo[field]" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 pt-4 border-t border-[var(--border-color)]">
                <div v-for="field in ['use_right', 'teach_right', 'admin_right']" :key="field" class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <select v-model.number="groupInfo[field]" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors appearance-none">
                    <option :value="1">{{ t('right_everyone') }}</option>
                    <option :value="2">{{ t('right_admin') }}</option>
                    <option :value="3">{{ t('right_white') }}</option>
                    <option :value="4">{{ t('right_owner') }}</option>
                  </select>
                </div>
              </div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-4 border-t border-[var(--border-color)]">
        <div v-for="field in ['is_power_on', 'is_require_prefix', 'is_sz84', 'is_cloud_black']" :key="field" class="flex flex-col items-center gap-2 p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
          <span class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
          <button @click="groupInfo[field] = !groupInfo[field]" :class="groupInfo[field] ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" class="transition-colors">
            <CheckCircle v-if="groupInfo[field]" class="w-8 h-8" />
            <div v-else class="w-8 h-8 rounded-full border-2 border-current"></div>
          </button>
        </div>
      </div>

              <div class="space-y-2 pt-4 border-t border-[var(--border-color)]">
                <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('is_cloud_answer') }}</label>
                <select v-model.number="groupInfo.is_cloud_answer" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors appearance-none">
                  <option :value="1">{{ t('common.open_full_auto') }}</option>
                  <option :value="2">{{ t('common.close') }}</option>
                  <option :value="3">{{ t('groupsetup.keyword_only') }}</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Message Tab -->
          <div v-if="activeTab === 'message'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-base font-black uppercase tracking-widest flex items-center gap-2">
                <MessageSquare class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('welcome_exit_message') }}
              </h3>
              
              <div class="space-y-4">
                <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <div class="space-y-1">
                    <span class="text-sm font-black text-[var(--text-main)] uppercase tracking-widest">{{ t('welcome_message_title') }}</span>
                    <p class="text-[10px] text-[var(--text-muted)] uppercase tracking-widest font-bold">{{ groupInfo.is_welcome_hint ? t('status_enabled') : t('status_disabled') }}</p>
                  </div>
                  <button @click="groupInfo.is_welcome_hint = !groupInfo.is_welcome_hint" :class="groupInfo.is_welcome_hint ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-12 h-6 rounded-full transition-colors flex items-center px-1">
                    <div :class="groupInfo.is_welcome_hint ? 'translate-x-6 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-4 h-4 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
                
                <textarea v-model="groupInfo.welcome_message" rows="4" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors resize-none" :placeholder="t('welcome_message_placeholder')"></textarea>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div v-for="field in ['is_exit_hint', 'is_kick_hint', 'is_change_hint', 'is_right_hint', 'is_change_enter', 'is_mute_enter', 'is_change_message']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-sm font-black text-[var(--text-main)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button @click="groupInfo[field] = !groupInfo[field]" :class="groupInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-10 h-5 rounded-full transition-colors flex items-center px-0.5">
                    <div :class="groupInfo[field] ? 'translate-x-5 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-4 h-4 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 pt-4 border-t border-[var(--border-color)]">
                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('is_accept_new_member') }}</label>
                  <select v-model.number="groupInfo.is_accept_new_member" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors appearance-none">
                    <option :value="1">{{ t('accept_auto') }}</option>
                    <option :value="0">{{ t('accept_none') }}</option>
                    <option :value="2">{{ t('accept_reject') }}</option>
                  </select>
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('reject_message') }}</label>
                  <input v-model="groupInfo.reject_message" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" :placeholder="t('reject_message_placeholder')" />
                </div>
              </div>

              <div class="space-y-2 pt-4 border-t border-[var(--border-color)]">
                <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('regex_request_join') }}</label>
                <input v-model="groupInfo.regex_request_join" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border(--matrix-color) outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
              </div>
            </div>
          </div>

          <!-- Keywords Tab -->
          <div v-if="activeTab === 'keywords'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-base font-black uppercase tracking-widest flex items-center gap-2">
                <Shield class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('keyword_management') }}
              </h3>

              <div class="overflow-x-auto">
                <table class="w-full border-collapse">
                  <thead>
                    <tr>
                      <th class="p-2 text-left text-sm font-black text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)]">{{ t('feature_type') }}</th>
                      <th v-for="type in keywordTypes" :key="type.key" class="p-2 text-center text-sm font-black text-[var(--text-muted)] uppercase tracking-widest border-b border-[var(--border-color)]">{{ t(type.name) }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="field in actionFields" :key="field.key" class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
                      <td class="p-2 text-sm font-black text-[var(--text-main)] uppercase tracking-widest border-b border-[var(--border-color)]">{{ t(field.name) }}</td>
                      <td v-for="type in keywordTypes" :key="type.key" class="p-2 text-center border-b border-[var(--border-color)]">
                        <button @click="toggleKeyword(field.key, type.name)" :class="getKeywordStatus(field.key, type.name) ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" class="transition-colors">
                          <CheckCircle v-if="getKeywordStatus(field.key, type.name)" class="w-5 h-5 mx-auto" />
                          <div v-else class="w-5 h-5 rounded-full border-2 border-current mx-auto opacity-20"></div>
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="space-y-4 pt-6 border-t border-[var(--border-color)]">
                <div v-for="field in [
                  { key: 'recall_keyword', name: t('recall_keyword'), color: 'text-orange-500' },
                  { key: 'warn_keyword', name: t('warn_keyword'), color: 'text-yellow-500' },
                  { key: 'mute_keyword', name: t('mute_keyword'), color: 'text-purple-500' },
                  { key: 'kick_keyword', name: t('kick_keyword'), color: 'text-red-500' },
                  { key: 'black_keyword', name: t('black_keyword'), color: 'text-gray-800' },
                  { key: 'white_keyword', name: t('white_keyword'), color: 'text-green-500' },
                  { key: 'credit_keyword', name: t('credit_keyword'), color: 'text-blue-500' }
                ]" :key="field.key" class="space-y-2">
                  <label :class="['text-sm font-black uppercase tracking-widest', field.color]">{{ field.name }}</label>
                  <input v-model="groupInfo[field.key]" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" :placeholder="t('keyword_placeholder')" />
                </div>
              </div>

              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-6 border-t border-[var(--border-color)]">
                <div v-for="field in ['is_recall', 'is_warn', 'is_block', 'is_white']" :key="field" class="flex flex-col items-center gap-2 p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button @click="groupInfo[field] = !groupInfo[field]" :class="groupInfo[field] ? 'text-[var(--matrix-color)]' : 'text-[var(--text-muted)]'" class="transition-colors">
                    <CheckCircle v-if="groupInfo[field]" class="w-6 h-6" />
                    <div v-else class="w-6 h-6 rounded-full border-2 border-current"></div>
                  </button>
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)]">
                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('recall_time') }}</label>
                  <input v-model.number="groupInfo.recall_time" type="number" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>
            </div>
          </div>

          <!-- AI Tab -->
          <div v-if="activeTab === 'ai'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <div class="flex items-center justify-between">
                <h3 class="text-base font-black uppercase tracking-widest flex items-center gap-2">
                  <Bot class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('ai_assistant_settings') }}
                </h3>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-black uppercase tracking-widest text-[var(--text-muted)]">{{ t('enable_ai_reply') }}</span>
                  <button @click="groupInfo.is_ai = !groupInfo.is_ai" :class="groupInfo.is_ai ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-10 h-5 rounded-full transition-colors flex items-center px-1">
                    <div :class="groupInfo.is_ai ? 'translate-x-5 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="space-y-4">
                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('ai_system_prompt') }}</label>
                  <textarea v-model="groupInfo.system_prompt" rows="6" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors resize-none" :placeholder="t('ai_system_prompt_placeholder')"></textarea>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div v-for="field in ['is_use_knowledgebase', 'is_mult_ai']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                    <div class="space-y-1">
                      <span class="text-sm font-black text-[var(--text-main)] uppercase tracking-widest">{{ t(field) }}</span>
                      <p class="text-xs text-[var(--text-muted)] font-bold uppercase tracking-widest opacity-60">{{ groupInfo[field] ? t('status_enabled') : t('status_disabled') }}</p>
                    </div>
                    <button @click="groupInfo[field] = !groupInfo[field]" :class="groupInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5">
                      <div :class="groupInfo[field] ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                    </button>
                  </div>
                </div>

                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('ai_history_count') }}</label>
                  <input v-model.number="groupInfo.context_count" type="number" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>
            </div>
          </div>

          <!-- Advanced Tab -->
          <div v-if="activeTab === 'advanced'" class="space-y-6">
            <div class="p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-6">
              <h3 class="text-base font-black uppercase tracking-widest flex items-center gap-2">
                <Zap class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('group_advanced_settings') }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div v-for="field in ['is_prop', 'is_pet', 'is_credit', 'is_credit_system', 'is_auto_signin', 'is_owner_pay', 'is_send_help_info', 'is_confirm_new', 'is_invite', 'is_reply_image', 'is_reply_recall', 'is_voice_reply', 'is_mute_refresh', 'is_black_refresh', 'is_block', 'is_white', 'is_warn', 'is_close_manager', 'is_black_exit', 'is_black_kick', 'is_black_share', 'is_hint_close']" :key="field" class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
                  <span class="text-sm font-black text-[var(--text-main)] uppercase tracking-widest">{{ t(field) }}</span>
                  <button @click="groupInfo[field] = !groupInfo[field]" :class="groupInfo[field] ? 'bg-[var(--matrix-color)]' : 'bg-gray-500/20'" class="relative w-8 h-4 rounded-full transition-colors flex items-center px-0.5">
                    <div :class="groupInfo[field] ? 'translate-x-4 bg-black' : 'translate-x-0 bg-[var(--text-muted)]'" class="w-3 h-3 rounded-full transition-transform shadow-sm"></div>
                  </button>
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-2 sm:grid-cols-5 gap-4">
                <div v-for="field in ['mute_enter_count', 'mute_keyword_count', 'kick_count', 'black_count', 'invite_credit']" :key="field" class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <input v-model.number="groupInfo[field]" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div v-for="field in ['mute_refresh_count', 'parent_group', 'block_min']" :key="field" class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <input v-model.number="groupInfo[field]" type="number" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('city_name') }}</label>
                  <input v-model="groupInfo.city_name" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div v-for="field in ['fans_name', 'voice_id']" :key="field" class="space-y-2">
                  <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field) }}</label>
                  <input v-model="groupInfo[field]" type="text" class="w-full p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-base font-bold text-[var(--text-main)] transition-colors" />
                </div>
              </div>

              <div class="pt-6 border-t border-[var(--border-color)] space-y-4">
                <h4 class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('group_name_auto_prefix') }}</h4>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div v-for="field in ['card_name_prefix_boy', 'card_name_prefix_girl', 'card_name_prefix_manager']" :key="field" class="space-y-2">
                    <label class="text-sm font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t(field.replace('card_name_prefix_', '') + '_prefix') }}</label>
                    <input v-model="groupInfo[field]" type="text" class="w-full p-3 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-base font-bold" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
    <PortalFooter v-if="isBlankLayout" />
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
