<template>
  <div class="fixed bottom-6 right-6 z-[100] flex flex-col items-end gap-4">
    <!-- Chat Window -->
    <transition name="chat-slide">
      <div 
        v-if="isOpen" 
        class="w-[350px] sm:w-[400px] h-[500px] sm:h-[600px] bg-[var(--bg-card)] border border-[var(--border-color)] rounded-[2rem] shadow-2xl flex flex-col overflow-hidden transition-colors duration-300"
      >
        <!-- Header -->
        <div class="p-4 bg-[var(--matrix-color)] text-black flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-xl bg-black/10 flex items-center justify-center">
              <MessageSquare class="w-4 h-4" />
            </div>
            <h3 class="font-black text-xs uppercase tracking-widest italic">{{ t('global_chat') }}</h3>
          </div>
          <button @click="isOpen = false" class="p-2 hover:bg-black/10 rounded-xl transition-colors">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Bot & Contact Selector -->
        <div class="p-4 border-b border-[var(--border-color)] space-y-3">
          <div class="flex gap-2">
            <select 
              v-model="selectedBotId" 
              class="flex-1 px-3 py-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest outline-none focus:border-[var(--matrix-color)] transition-all"
            >
              <option value="" disabled>{{ t('select_bot') }}</option>
              <option v-for="bot in botStore.bots" :key="bot.id" :value="bot.id">
                {{ bot.nickname || bot.id }} ({{ bot.platform }})
              </option>
            </select>
          </div>
          
          <div class="flex gap-2">
            <select 
              v-model="selectedContactId" 
              class="flex-1 px-3 py-2 rounded-xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] text-[10px] font-black uppercase tracking-widest outline-none focus:border-[var(--matrix-color)] transition-all"
              :disabled="!selectedBotId"
            >
              <option value="" disabled>{{ t('select_contact') }}</option>
              <optgroup :label="t('groups')">
                <option v-for="group in groups" :key="group.id" :value="group.id">
                  [G] {{ group.name || group.id }}
                </option>
              </optgroup>
              <optgroup :label="t('friends')">
                <option v-for="friend in friends" :key="friend.id" :value="friend.id">
                  [P] {{ friend.name || friend.nickname || friend.id }}
                </option>
              </optgroup>
            </select>
            <button 
              @click="refreshContacts" 
              class="p-2 rounded-xl bg-[var(--matrix-color)]/10 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)] hover:text-black transition-all"
              :disabled="!selectedBotId || loadingContacts"
            >
              <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loadingContacts }" />
            </button>
          </div>
        </div>

        <!-- Messages Area -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-black/5 dark:bg-white/5" ref="messageContainer">
          <div v-if="!selectedContactId" class="h-full flex flex-col items-center justify-center text-[var(--text-muted)] opacity-30 text-center space-y-2">
            <MessageSquare class="w-12 h-12" />
            <p class="text-[10px] font-black uppercase tracking-widest">{{ t('select_contact_to_start') }}</p>
          </div>
          <template v-else>
            <div 
              v-for="msg in filteredMessages" 
              :key="msg.id || msg.time" 
              class="flex flex-col"
              :class="msg.user_id === selectedBotId || msg.sender?.user_id === selectedBotId ? 'items-end' : 'items-start'"
            >
              <div class="flex items-center gap-2 mb-1">
                <span class="text-[8px] font-black text-[var(--text-muted)] uppercase tracking-widest">
                  {{ msg.user_name || msg.sender?.nickname || msg.user_id || msg.sender?.user_id }}
                </span>
                <span class="text-[8px] font-mono text-[var(--text-muted)] opacity-50">
                  {{ formatTime(msg.created_at || msg.time) }}
                </span>
              </div>
              <div 
                class="max-w-[85%] px-4 py-2 rounded-2xl text-xs font-medium break-words shadow-sm"
                :class="msg.user_id === selectedBotId || msg.sender?.user_id === selectedBotId 
                  ? 'bg-[var(--matrix-color)] text-black rounded-tr-none' 
                  : 'bg-[var(--bg-card)] border border-[var(--border-color)] text-[var(--text-main)] rounded-tl-none'"
              >
                {{ msg.content || msg.message || msg.raw_message }}
              </div>
            </div>
          </template>
        </div>

        <!-- Input Area -->
        <div class="p-4 border-t border-[var(--border-color)]">
          <div class="flex items-end gap-2 bg-black/5 dark:bg-white/5 p-2 rounded-2xl border border-[var(--border-color)] focus-within:border-[var(--matrix-color)] transition-all">
            <textarea 
              v-model="inputMessage" 
              @keydown.enter.prevent="sendMessage"
              :placeholder="t('type_message')"
              class="flex-1 bg-transparent border-none outline-none text-xs p-2 resize-none max-h-32 custom-scrollbar"
              rows="1"
              :disabled="!selectedContactId || sending"
            ></textarea>
            <button 
              @click="sendMessage" 
              class="p-3 rounded-xl bg-[var(--matrix-color)] text-black hover:opacity-90 transition-all disabled:opacity-50 shadow-lg shadow-[var(--matrix-color)]/20"
              :disabled="!inputMessage.trim() || !selectedContactId || sending"
            >
              <Send v-if="!sending" class="w-4 h-4" />
              <RefreshCw v-else class="w-4 h-4 animate-spin" />
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- Floating Button -->
    <button 
      @click="toggleChat" 
      class="w-14 h-14 rounded-2xl bg-[var(--matrix-color)] text-black shadow-2xl shadow-[var(--matrix-color)]/30 flex items-center justify-center hover:scale-110 active:scale-95 transition-all group relative overflow-hidden"
    >
      <div class="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
      <MessageSquare v-if="!isOpen" class="w-6 h-6 relative z-10" />
      <X v-else class="w-6 h-6 relative z-10" />

      <!-- Unread Badge -->
      <div v-if="unreadCount > 0 && !isOpen" class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-[10px] font-black rounded-full flex items-center justify-center border-2 border-[var(--bg-card)] z-20 animate-bounce">
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { MessageSquare, X, Send, RefreshCw, User, Users } from 'lucide-vue-next';
import { useBotStore } from '@/stores/bot';
import { useSystemStore } from '@/stores/system';

const botStore = useBotStore();
const systemStore = useSystemStore();
const t = (key: string) => systemStore.t(key);

const isOpen = ref(false);
const selectedBotId = ref(botStore.currentBotId || '');
const selectedContactId = ref('');
const inputMessage = ref('');
const sending = ref(false);
const unreadCount = ref(0);
const lastReadMessageId = ref<string | null>(null);
const contacts = ref<any[]>([]);
const loadingContacts = ref(false);
const messageContainer = ref<HTMLElement | null>(null);

// Watch for bot selection change to refresh contacts
watch(selectedBotId, (newId) => {
  if (newId) {
    botStore.setCurrentBotId(newId);
    refreshContacts();
  }
  selectedContactId.value = '';
});

// Watch for new messages to increment unread count if window is closed
watch(() => botStore.messages, (newMessages) => {
  if (!isOpen.value && newMessages.length > 0) {
    const lastMsg = newMessages[newMessages.length - 1];
    // Only count if it's not from the bot itself
    if (lastMsg.self_id !== lastMsg.user_id) {
      unreadCount.value++;
    }
  }
}, { deep: true });

const groups = computed(() => contacts.value.filter(c => c.type === 'group'));
const friends = computed(() => contacts.value.filter(c => c.type === 'private'));

const refreshContacts = async () => {
  if (!selectedBotId.value) return;
  loadingContacts.value = true;
  try {
    const data = await botStore.fetchContacts(selectedBotId.value);
    if (data.success && data.data) {
      contacts.value = data.data.contacts || [];
    }
  } catch (err) {
    console.error('Failed to fetch contacts:', err);
  } finally {
    loadingContacts.value = false;
  }
};

const toggleChat = () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    unreadCount.value = 0;
    if (botStore.bots.length === 0) {
      botStore.fetchBots();
    }
    if (selectedBotId.value && contacts.value.length === 0) {
      refreshContacts();
    }
  }
};

const filteredMessages = computed(() => {
  if (!selectedContactId.value) return [];
  const contact = contacts.value.find(c => c.id === selectedContactId.value);
  if (!contact) return [];

  return botStore.messages.filter(msg => {
    // Check if message belongs to this bot
    if (msg.bot_id !== selectedBotId.value && msg.self_id !== selectedBotId.value) return false;

    if (contact.type === 'group') {
      return msg.group_id === contact.id;
    } else {
      // Private message
      return (msg.user_id === contact.id || msg.target_id === contact.id) && !msg.group_id;
    }
  }).slice(-50); // Show last 50 messages
});

// Auto scroll to bottom
watch(filteredMessages, () => {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
    }
  });
}, { deep: true });

const formatTime = (time: any) => {
  if (!time) return '';
  if (typeof time === 'string') return time.split(' ')[1] || time;
  const date = new Date(time * 1000);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

const sendMessage = async () => {
  if (!inputMessage.value.trim() || !selectedContactId.value || sending.value) return;

  const contact = contacts.value.find(c => c.id === selectedContactId.value);
  if (!contact) return;

  sending.value = true;
  try {
    const action = contact.type === 'group' ? 'send_group_msg' : 'send_private_msg';
    const params: any = {
      message: inputMessage.value.trim()
    };
    if (contact.type === 'group') {
      params.group_id = contact.id;
    } else {
      params.user_id = contact.id;
    }

    await botStore.callBotApi(action, params, selectedBotId.value);
    inputMessage.value = '';
    
    // Refresh messages
    await botStore.fetchMessages(100);
  } catch (err) {
    console.error('Failed to send message:', err);
  } finally {
    sending.value = false;
  }
};

onMounted(() => {
  if (botStore.bots.length === 0) {
    botStore.fetchBots();
  }
});
</script>

<style scoped>
.chat-slide-enter-active,
.chat-slide-leave-active {
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.chat-slide-enter-from,
.chat-slide-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
  filter: blur(10px);
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 10px;
}

textarea {
  scrollbar-width: thin;
}
</style>
