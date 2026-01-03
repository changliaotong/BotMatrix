<template>
  <div class="p-4 sm:p-8 space-y-4 sm:space-y-8">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-[var(--matrix-color)]/10 flex items-center justify-center">
          <MessageSquare class="w-5 h-5 sm:w-6 sm:h-6 text-[var(--matrix-color)]" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-black text-[var(--text-main)] tracking-tight uppercase italic">{{ t('nav_messages') }}</h1>
          <p class="text-[var(--text-muted)] text-[10px] sm:text-xs font-bold tracking-widest uppercase">{{ t('messages_description') }}</p>
        </div>
      </div>
      
      <div class="flex items-center gap-3">
        <button 
          @click="refreshMessages" 
          class="flex items-center gap-2 px-6 py-2 rounded-xl bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest hover:opacity-90 transition-all shadow-lg shadow-[var(--matrix-color)]/20 disabled:opacity-50"
          :disabled="loading"
        >
          <RefreshCw :class="{ 'animate-spin': loading }" class="w-4 h-4" />
          {{ t('refresh') }}
        </button>
      </div>
    </div>

    <!-- Messages List -->
    <div class="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] overflow-hidden shadow-sm flex flex-col h-[calc(100vh-280px)] sm:h-[calc(100vh-220px)] transition-colors duration-300">
      <div class="overflow-x-auto custom-scrollbar">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-black/5 dark:bg-white/5">
              <th class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)] border-b border-[var(--border-color)]">
                {{ t('time') }}
              </th>
              <th class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)] border-b border-[var(--border-color)]">
                {{ t('bot_id') }}
              </th>
              <th class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)] border-b border-[var(--border-color)]">
                {{ t('sender') }}
              </th>
              <th class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)] border-b border-[var(--border-color)]">
                {{ t('group') }}
              </th>
              <th class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-[var(--text-muted)] border-b border-[var(--border-color)]">
                {{ t('content') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--border-color)]">
            <tr v-if="!authStore.isAdmin" class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <td colspan="5" class="px-6 py-20 text-center">
                <div class="flex flex-col items-center justify-center gap-4 text-[var(--text-muted)] opacity-30">
                  <Activity class="w-12 h-12" />
                  <span class="text-[10px] font-black uppercase tracking-widest">{{ t('admin_required') }}</span>
                </div>
              </td>
            </tr>
            <tr v-else-if="messages.length === 0" class="hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
              <td colspan="5" class="px-6 py-20 text-center">
                <div v-if="loading" class="flex flex-col items-center justify-center gap-4 text-[var(--matrix-color)]/50">
                  <RefreshCw class="w-8 h-8 animate-spin" />
                  <span class="text-[10px] font-black uppercase tracking-widest">{{ t('loading') }}</span>
                </div>
                <div v-else class="flex flex-col items-center justify-center gap-4 text-[var(--text-muted)] opacity-30">
                  <MessageSquare class="w-12 h-12" />
                  <span class="text-[10px] font-black uppercase tracking-widest">{{ t('no_messages') }}</span>
                </div>
              </td>
            </tr>
            <tr v-for="msg in messages" :key="msg.id" class="group hover:bg-[var(--matrix-color)]/5 transition-all">
              <td class="px-6 py-4 whitespace-nowrap text-[10px] font-mono font-black text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors">
                {{ msg.created_at }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[8px] font-black uppercase tracking-widest bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] border border-[var(--matrix-color)]/20">
                  {{ msg.bot_id }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 flex items-center justify-center text-[var(--text-main)] font-black text-xs">
                    {{ (msg.user_name || msg.user_id || '?').substring(0, 1).toUpperCase() }}
                  </div>
                  <div>
                    <div class="text-xs font-black text-[var(--text-main)]">{{ msg.user_name }}</div>
                    <div class="text-[8px] font-mono text-[var(--text-muted)] uppercase tracking-widest">{{ msg.user_id }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div v-if="msg.group_id && msg.group_id !== '0'" class="flex items-center gap-2">
                  <span class="px-2 py-0.5 rounded-md text-[8px] font-black uppercase tracking-widest bg-blue-500/10 text-blue-500 border border-blue-500/20">
                    {{ msg.group_name || msg.group_id }}
                  </span>
                </div>
                <span v-else class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest opacity-40">
                  {{ t('private_chat') }}
                </span>
              </td>
              <td class="px-6 py-4 text-xs font-medium text-[var(--text-main)]/80 max-w-md truncate">
                {{ msg.content }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { MessageSquare, RefreshCw, Activity } from 'lucide-vue-next';
import { useBotStore } from '../stores/bot';
import { useSystemStore } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';

const systemStore = useSystemStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);
const botStore = useBotStore();
const loading = ref(false);
const messages = computed(() => botStore.messages.slice().reverse());

const refreshMessages = async () => {
  if (!authStore.isAdmin) return;
  loading.value = true;
  try {
    await botStore.fetchMessages(100);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  refreshMessages();
});
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
  height: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 10px;
}
</style>
