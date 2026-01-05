<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useSystemStore } from '@/stores/system';
import { Bot, Lock, User, ArrowRight, Loader2, Languages, Check } from 'lucide-vue-next';
import { type Language } from '@/utils/i18n';

const router = useRouter();
const authStore = useAuthStore();
const systemStore = useSystemStore();

const t = (key: string) => systemStore.t(key);

const username = ref('');
const password = ref('');
const loading = ref(false);
const error = ref('');

const showLangPicker = ref(false);
const langPickerRef = ref<HTMLElement | null>(null);

const languages: { id: Language; nameKey: string }[] = [
  { id: 'zh-CN', nameKey: 'lang_zh_cn' },
  { id: 'zh-TW', nameKey: 'lang_zh_tw' },
  { id: 'en-US', nameKey: 'lang_en_us' },
  { id: 'ja-JP', nameKey: 'lang_ja_jp' }
];

const toggleLangPicker = () => {
  showLangPicker.value = !showLangPicker.value;
};

const selectLang = (lang: Language) => {
  systemStore.setLang(lang);
  showLangPicker.value = false;
};

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node;
  if (langPickerRef.value && !langPickerRef.value.contains(target)) {
    showLangPicker.value = false;
  }
};

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside);
});

const handleLogin = async () => {
  const trimmedUsername = username.value.trim();
  const trimmedPassword = password.value.trim();
  if (!trimmedUsername || !trimmedPassword) return;
  
  loading.value = true;
  error.value = '';
  
  try {
    const success = await authStore.login(trimmedUsername, trimmedPassword);
    if (success) {
      router.push('/console');
    } else {
      error.value = t('login_error_auth');
    }
  } catch (err: any) {
    error.value = err.message || t('login_error_generic');
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-black flex items-center justify-center p-4 relative overflow-hidden">
    <!-- Matrix Rain Background Placeholder -->
    <div class="absolute inset-0 opacity-10 pointer-events-none">
      <div class="absolute inset-0 bg-gradient-to-b from-transparent via-matrix/20 to-transparent animate-pulse"></div>
    </div>

    <div class="w-full max-w-md relative">
      <div class="p-8 sm:p-12 rounded-[2.5rem] bg-white dark:bg-zinc-900 border border-black/5 dark:border-white/5 shadow-2xl space-y-8 relative">
        <!-- i18n Selector -->
        <div class="absolute right-6 top-6" ref="langPickerRef">
          <button 
            @click="toggleLangPicker"
            class="p-2 rounded-xl hover:bg-gray-100 dark:hover:bg-white/5 text-gray-400 hover:text-matrix transition-all flex items-center gap-2 group"
          >
            <Languages class="w-5 h-5 group-hover:rotate-12 transition-transform" />
            <span class="text-[10px] font-bold uppercase tracking-widest hidden sm:block">{{ systemStore.currentLang }}</span>
          </button>

          <!-- Dropdown -->
          <div 
            v-if="showLangPicker"
            class="absolute right-0 mt-2 w-48 bg-white dark:bg-zinc-800 rounded-2xl shadow-2xl border border-black/5 dark:border-white/5 overflow-hidden z-50 animate-in fade-in slide-in-from-top-2 duration-200"
          >
            <div class="p-2">
              <button 
                v-for="lang in languages" 
                :key="lang.id"
                @click="selectLang(lang.id)"
                class="w-full flex items-center justify-between px-4 py-3 rounded-xl text-xs font-bold transition-all"
                :class="[
                  systemStore.currentLang === lang.id 
                    ? 'bg-matrix/10 text-matrix' 
                    : 'text-gray-500 hover:bg-gray-50 dark:hover:bg-white/5 dark:text-gray-400'
                ]"
              >
                {{ t(lang.nameKey) }}
                <Check v-if="systemStore.currentLang === lang.id" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Logo -->
        <div class="text-center space-y-4">
          <div class="inline-flex p-5 rounded-3xl bg-matrix/10 text-matrix animate-bounce-slow">
            <Bot class="w-10 h-10" />
          </div>
          <div class="space-y-1">
            <h1 class="text-3xl font-black text-[var(--text-main)] tracking-tighter uppercase italic">{{ t('botmatrix').substring(0, 3) }}<span class="text-matrix">{{ t('botmatrix').substring(3) }}</span></h1>
            <p class="text-xs font-bold text-gray-400 uppercase tracking-[0.2em]">{{ t('system_desc') }}</p>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleLogin" class="space-y-6">
          <div class="space-y-4">
            <div class="relative group">
              <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-matrix transition-colors">
                <User class="w-5 h-5" />
              </div>
              <input 
                v-model="username"
                type="text" 
                :placeholder="t('username')" 
                class="w-full bg-gray-50 dark:bg-black border border-black/5 dark:border-white/10 rounded-2xl pl-12 pr-4 py-4 focus:outline-none focus:border-matrix transition-all text-[var(--text-main)] font-bold placeholder:text-gray-400"
              />
            </div>
            <div class="relative group">
              <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-matrix transition-colors">
                <Lock class="w-5 h-5" />
              </div>
              <input 
                v-model="password"
                type="password" 
                :placeholder="t('password')" 
                class="w-full bg-gray-50 dark:bg-black border border-black/5 dark:border-white/10 rounded-2xl pl-12 pr-4 py-4 focus:outline-none focus:border-matrix transition-all text-[var(--text-main)] font-bold placeholder:text-gray-400"
              />
            </div>
          </div>

          <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold text-center">
            {{ error }}
          </div>

          <button 
            type="submit" 
            :disabled="loading"
            class="w-full bg-matrix hover:bg-matrix/90 disabled:opacity-50 text-black font-black py-4 rounded-2xl flex items-center justify-center gap-2 transition-all group active:scale-95 shadow-lg shadow-matrix/20 uppercase tracking-widest"
          >
            <template v-if="loading">
              <Loader2 class="w-5 h-5 animate-spin" /> {{ t('verifying') }}
            </template>
            <template v-else>
              {{ t('enter_matrix') }} <ArrowRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </template>
          </button>
        </form>

        <div class="text-center">
          <p class="text-[10px] font-bold text-gray-500 uppercase tracking-widest">
            {{ t('copyright') }}
          </p>
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
.animate-bounce-slow {
  animation: bounce 3s infinite;
}
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
</style>
