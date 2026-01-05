<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { getPlatformIcon, getPlatformColor, isPlatformAvatar, getPlatformFromAvatar } from '@/utils/avatar';
import { 
  Users, 
  Search, 
  RefreshCw, 
  User, 
  MessageSquare, 
  Filter,
  MoreVertical,
  Shield,
  Bot
} from 'lucide-vue-next';

const systemStore = useSystemStore();
const botStore = useBotStore();
const t = (key: string) => systemStore.t(key);

const contacts = ref<any[]>([]);
const loading = ref(true);
const searchQuery = ref('');
const activeTab = ref('all'); // all, friend, group

const fetchContacts = async () => {
  loading.value = true;
  try {
    const data = await botStore.fetchContacts();
    if (data.success && data.data) {
      contacts.value = data.data.contacts || [];
    }
  } finally {
    loading.value = false;
  }
};

onMounted(fetchContacts);

const filteredContacts = computed(() => {
  return contacts.value.filter(c => {
    const matchesSearch = (c.name || c.id || '').toLowerCase().includes(searchQuery.value.toLowerCase());
    const matchesTab = activeTab.value === 'all' || 
                       (activeTab.value === 'friend' && c.type === 'private') || 
                       (activeTab.value === 'group' && c.type === 'group');
    return matchesSearch && matchesTab;
  });
});

const syncAll = async () => {
  // Sync contacts for all bots
  loading.value = true;
  try {
    for (const bot of botStore.bots) {
      await botStore.syncContacts(bot.id);
    }
    await fetchContacts();
  } finally {
    loading.value = false;
  }
};

</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-[var(--text-main)] tracking-tight">{{ t('contacts') }}</h1>
        <p class="text-sm font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('contacts_desc') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative flex-1 md:w-64">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input 
            v-model="searchQuery"
            type="text" 
            :placeholder="t('search_contacts')"
            class="w-full pl-10 pr-4 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] text-sm focus:border-[var(--matrix-color)] outline-none transition-all"
          />
        </div>
        <button 
          @click="syncAll"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-opacity"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          {{ t('sync_all') }}
        </button>
      </div>
    </div>

    <!-- Tabs & Stats -->
    <div class="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]">
      <div class="flex p-1 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)]">
        <button 
          v-for="tab in ['all', 'friend', 'group']" 
          :key="tab"
          @click="activeTab = tab"
          class="px-6 py-2 rounded-xl text-xs font-black uppercase tracking-widest transition-all"
          :class="activeTab === tab ? 'bg-[var(--bg-card)] text-[var(--matrix-color)] shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
        >
          {{ t(tab) }}
        </button>
      </div>
      <div class="flex items-center gap-6 px-4">
        <div class="text-center">
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('total') }}</p>
          <p class="text-lg font-black text-[var(--text-main)]">{{ contacts.length }}</p>
        </div>
        <div class="w-px h-8 bg-[var(--border-color)]"></div>
        <div class="text-center">
          <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('groups') }}</p>
          <p class="text-lg font-black text-blue-500">{{ contacts.filter(c => c.type === 'group').length }}</p>
        </div>
      </div>
    </div>

    <!-- Contacts Grid -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 animate-pulse">
      <div v-for="i in 8" :key="i" class="h-48 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)]"></div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      <div 
        v-for="contact in filteredContacts" 
        :key="contact.id"
        class="group p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] hover:border-[var(--matrix-color)]/30 transition-all duration-500 relative overflow-hidden"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="relative">
            <div class="w-16 h-16 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] flex items-center justify-center overflow-hidden">
              <template v-if="contact.avatar && !isPlatformAvatar(contact.avatar)">
                <img :src="contact.avatar" class="w-full h-full object-cover" />
              </template>
              <template v-else>
                <component 
                  :is="isPlatformAvatar(contact.avatar) ? getPlatformIcon(getPlatformFromAvatar(contact.avatar)) : (contact.type === 'group' ? Users : User)" 
                  :class="['w-8 h-8', isPlatformAvatar(contact.avatar) ? getPlatformColor(getPlatformFromAvatar(contact.avatar)) : 'text-[var(--text-muted)]']" 
                />
              </template>
            </div>
            <div class="absolute -bottom-1 -right-1 p-1 rounded-lg bg-[var(--bg-card)] border border-[var(--border-color)]">
              <Bot class="w-3 h-3 text-[var(--matrix-color)]" />
            </div>
          </div>
          <button class="p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
            <MoreVertical class="w-4 h-4 text-[var(--text-muted)]" />
          </button>
        </div>

        <div class="space-y-1">
          <h3 class="font-black text-[var(--text-main)] truncate">{{ contact.name || contact.nickname || contact.id || t('unknown') }}</h3>
          <p v-if="contact.nickname && contact.name && contact.name !== contact.nickname" class="text-[10px] font-bold text-[var(--text-muted)] truncate">{{ contact.nickname }}</p>
          <p class="text-[10px] font-mono text-[var(--text-muted)] uppercase tracking-widest">{{ contact.id }}</p>
        </div>

        <div class="mt-4 pt-4 border-t border-[var(--border-color)] flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="px-2 py-0.5 rounded-md text-[8px] font-black uppercase tracking-widest border" :class="contact.type === 'group' ? 'text-blue-500 border-blue-500/20 bg-blue-500/5' : 'text-purple-500 border-purple-500/20 bg-purple-500/5'">
              {{ contact.type === 'group' ? t('group') : t('friend') }}
            </span>
            <span v-if="contact.source" class="text-[8px] font-bold text-[var(--text-muted)] uppercase">{{ t(contact.source) }}</span>
          </div>
          <button class="p-2 rounded-xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)] hover:text-black transition-all">
            <MessageSquare class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && filteredContacts.length === 0" class="flex flex-col items-center justify-center py-20 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-3xl">
      <Users class="w-16 h-16 text-[var(--text-muted)] mb-4 opacity-20" />
      <h2 class="text-xl font-black text-[var(--text-main)] uppercase tracking-tight">{{ t('no_contacts_found') }}</h2>
      <p class="text-[var(--text-muted)] text-sm font-bold uppercase tracking-widest mt-2">{{ t('search_sync_desc') }}</p>
    </div>
  </div>
</template>