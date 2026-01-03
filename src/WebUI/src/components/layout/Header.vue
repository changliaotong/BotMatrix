<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useSystemStore, type Style } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { useBotStore } from '@/stores/bot';
import { useRoute, useRouter } from 'vue-router';
import { Menu, Github, Sun, Moon, LogOut, Palette, Check, Languages } from 'lucide-vue-next';
import { type Language } from '@/utils/i18n';

const systemStore = useSystemStore();
const authStore = useAuthStore();
const botStore = useBotStore();
const route = useRoute();
const router = useRouter();

const showStylePicker = ref(false);
const stylePickerRef = ref<HTMLElement | null>(null);
const showLangPicker = ref(false);
const langPickerRef = ref<HTMLElement | null>(null);

// Calculate uptime locally to ensure it updates every second
const localUptime = ref('0s');
let uptimeTimer: number | null = null;

const updateUptime = () => {
  const startTime = botStore.stats?.start_time;
  if (!startTime) {
    localUptime.value = '0s';
    return;
  }
  
  const now = Math.floor(Date.now() / 1000);
  const diff = now - startTime;
  
  if (diff < 60) {
    localUptime.value = `${diff}s`;
  } else if (diff < 3600) {
    localUptime.value = `${Math.floor(diff / 60)}m ${diff % 60}s`;
  } else if (diff < 86400) {
    localUptime.value = `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`;
  } else {
    localUptime.value = `${Math.floor(diff / 86400)}d ${Math.floor((diff % 86400) / 3600)}h`;
  }
};

// Map route paths back to translation keys
const routeTitleMap: Record<string, string> = {
  '/': 'dashboard',
  '/bots': 'bots',
  '/workers': 'workers',
  '/contacts': 'contacts',
  '/nexus': 'nexus',
  '/tasks': 'tasks',
  '/fission': 'fission',
  '/docker': 'docker',
  '/routing': 'routing',
  '/users': 'users',
  '/settings': 'settings',
  '/logs': 'logs',
  '/manual': 'manual',
  '/monitor': 'monitor'
};

const t = (key: string) => systemStore.t(key);

const langShortNameMap: Record<string, string> = {
  'zh-CN': 'lang_zh_cn_short',
  'zh-TW': 'lang_zh_tw_short',
  'en-US': 'lang_en_us_short',
  'ja-JP': 'lang_ja_jp_short'
};

const styleIconMap: Record<string, string> = {
  'classic': 'CL',
  'matrix': 'MX',
  'xp': 'XP',
  'ios': 'iOS',
  'kawaii': 'KA'
};

const styles: { id: Style; nameKey: string; colors: { light: any; dark: any } }[] = [
  { 
    id: 'classic', 
    nameKey: 'style_classic',
    colors: {
      light: { bg: '#f3f4f6', sidebar: '#ffffff', header: '#ffffff', accent: '#3b82f6', text: '#111827', border: '#e5e7eb' },
      dark: { bg: '#0f172a', sidebar: '#0f172a', header: '#0f172a', accent: '#f59e0b', text: '#f8fafc', border: '#1e293b' }
    }
  },
  { 
    id: 'matrix', 
    nameKey: 'style_matrix',
    colors: {
      light: { bg: '#f0fff4', sidebar: '#ffffff', header: '#ffffff', accent: '#059669', text: '#064e3b', border: '#d1fae5' },
      dark: { bg: '#000000', sidebar: '#000000', header: '#000000', accent: '#00ff41', text: '#00ff41', border: '#003b00' }
    }
  },
  { 
    id: 'xp', 
    nameKey: 'style_xp',
    colors: {
      light: { bg: '#ece9d8', sidebar: '#d6dff7', header: '#0058e6', accent: '#24a124', text: '#000000', border: '#0054e3' },
      dark: { bg: '#1c1c1c', sidebar: '#1a1a1a', header: '#003399', accent: '#33cc33', text: '#ffffff', border: '#003399' }
    }
  },
  { 
    id: 'ios', 
    nameKey: 'style_ios',
    colors: {
      light: { bg: '#ffffff', sidebar: 'rgba(255,255,255,0.7)', header: 'rgba(255,255,255,0.8)', accent: '#007aff', text: '#000000', border: 'rgba(0,0,0,0.1)' },
      dark: { bg: '#000000', sidebar: 'rgba(28,28,30,0.7)', header: 'rgba(0,0,0,0.8)', accent: '#0a84ff', text: '#ffffff', border: 'rgba(255,255,255,0.15)' }
    }
  },
  {
    id: 'kawaii',
    nameKey: 'style_kawaii',
    colors: {
      light: { bg: '#fff0f6', sidebar: '#ffffff', header: 'rgba(255,240,246,0.8)', accent: '#f06595', text: '#d6336c', border: '#ffdeeb' },
      dark: { bg: '#2b0b1a', sidebar: '#1a050f', header: 'rgba(40,10,25,0.8)', accent: '#ff85b3', text: '#ff85b3', border: '#ff85b3' }
    }
  }
];

const languages: { id: Language; nameKey: string }[] = [
  { id: 'zh-CN', nameKey: 'lang_zh_cn' },
  { id: 'zh-TW', nameKey: 'lang_zh_tw' },
  { id: 'en-US', nameKey: 'lang_en_us' },
  { id: 'ja-JP', nameKey: 'lang_ja_jp' }
];

const toggleStylePicker = () => {
  showStylePicker.value = !showStylePicker.value;
  showLangPicker.value = false;
};

const toggleLangPicker = () => {
  showLangPicker.value = !showLangPicker.value;
  showStylePicker.value = false;
};

const selectStyle = (style: Style) => {
  systemStore.setStyle(style);
  showStylePicker.value = false;
};

const selectLang = (lang: Language) => {
  systemStore.setLang(lang);
  showLangPicker.value = false;
};

const handleLogout = () => {
  authStore.logout();
  botStore.reset();
  router.push('/login');
};

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node;
  if (stylePickerRef.value && !stylePickerRef.value.contains(target)) {
    showStylePicker.value = false;
  }
  if (langPickerRef.value && !langPickerRef.value.contains(target)) {
    showLangPicker.value = false;
  }
};

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside);
  
  // Update uptime every second
  updateUptime();
  uptimeTimer = window.setInterval(updateUptime, 1000);
  
  // Fetch initial stats if not already loaded to get start_time
  if (!botStore.stats?.start_time) {
    botStore.fetchStats();
  }
});

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside);
  if (uptimeTimer) clearInterval(uptimeTimer);
});
</script>

<template>
  <header class="h-16 flex-shrink-0 flex items-center justify-between px-4 sm:px-6 bg-[var(--bg-header)] border-b border-[var(--border-color)] z-40 transition-colors duration-300">
    <div class="flex items-center gap-2 sm:gap-4">
      <button @click="systemStore.toggleMobileMenu()" class="p-2 text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors lg:hidden">
        <Menu class="w-5 h-5" />
      </button>
      <button @click="systemStore.toggleSidebar()" class="hidden lg:block p-2 text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
        <Menu class="w-5 h-5" />
      </button>
      <h2 class="text-base sm:text-lg font-bold tracking-tight text-[var(--text-main)] truncate max-w-[120px] sm:max-w-none">{{ t(routeTitleMap[route.path] || 'dashboard') }}</h2>
    </div>
    
    <div class="flex items-center gap-2 sm:gap-4">
      <!-- Uptime & Time (Hidden on small mobile) -->
      <div class="hidden sm:flex items-center gap-2 sm:gap-6 px-2 sm:px-4 py-1 sm:py-2 rounded-xl sm:rounded-2xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5">
        <div class="flex flex-col">
          <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] leading-tight">{{ t('system_uptime') }}</span>
          <span class="text-xs sm:text-sm font-bold text-[var(--matrix-color)] mono">{{ localUptime }}</span>
        </div>
        <div class="h-4 sm:h-6 w-px bg-black/10 dark:bg-white/10"></div>
        <div class="flex flex-col text-right">
          <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] leading-tight">{{ t('current_time') }}</span>
          <span class="text-xs sm:text-sm font-bold text-[var(--text-main)] mono">{{ systemStore.currentTime }}</span>
        </div>
      </div>

      <div class="h-6 w-px bg-black/5 dark:bg-white/5 hidden sm:block"></div>
      
      <div class="flex items-center gap-1">
        <a href="https://github.com/changliaotong/BotMatrix" target="_blank" class="hidden sm:flex items-center justify-center w-8 h-8 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-gray-400 transition-colors" title="GitHub">
          <Github class="w-4 h-4" />
        </a>

        <!-- Language Picker -->
        <div class="relative" ref="langPickerRef">
          <button 
            @click="toggleLangPicker" 
            class="flex items-center justify-center px-2 sm:px-3 h-8 rounded-lg transition-all border"
            :class="[
              showLangPicker 
                ? 'bg-[var(--matrix-color)] text-[var(--sidebar-text-active)] border-[var(--matrix-color)]' 
                : 'bg-black/5 dark:bg-white/5 text-[var(--matrix-color)] hover:bg-[var(--matrix-color)] hover:text-[var(--sidebar-text-active)] border-[var(--matrix-color)]/20'
            ]"
            :title="t('switch_lang')"
          >
            <span class="text-[10px] sm:text-xs font-bold">{{ t(langShortNameMap[systemStore.lang]) }}</span>
          </button>

          <!-- Language Picker Panel -->
          <transition name="fade-slide">
            <div v-if="showLangPicker" class="absolute right-0 mt-2 w-40 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-xl z-50 overflow-hidden">
              <div class="px-3 py-2 border-b border-[var(--border-color)] mb-1">
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('language_region') }}</span>
              </div>
              <button 
                v-for="l in languages" 
                :key="l.id"
                @click="selectLang(l.id)"
                class="w-full flex items-center justify-between px-3 py-2 hover:bg-[var(--matrix-color)]/10 transition-colors group"
                :class="{ 'text-[var(--matrix-color)]': systemStore.lang === l.id }"
              >
                <span class="text-xs font-bold">{{ t(l.nameKey) }}</span>
                <Check v-if="systemStore.lang === l.id" class="w-3 h-3" />
              </button>
            </div>
          </transition>
        </div>
        
        <!-- Style Toggle -->
        <div class="relative" ref="stylePickerRef">
          <button 
            @click="toggleStylePicker" 
            class="flex items-center justify-center px-1.5 sm:px-2 h-8 rounded-lg transition-all border"
            :class="[
              showStylePicker 
                ? 'bg-[var(--matrix-color)] text-black border-[var(--matrix-color)]' 
                : 'bg-black/5 dark:bg-white/5 text-gray-400 hover:text-[var(--matrix-color)] border-transparent hover:border-[var(--matrix-color)]/20'
            ]"
            :title="t('style_' + systemStore.style)"
          >
            <Palette class="w-3.5 h-3.5 sm:w-4 sm:h-4 mr-0.5 sm:mr-1" />
            <span class="text-[10px] font-bold uppercase">{{ styleIconMap[systemStore.style] }}</span>
          </button>

          <!-- Style Picker Panel -->
          <transition name="fade-slide">
            <div v-if="showStylePicker" class="absolute right-0 mt-2 w-64 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-xl z-50 overflow-hidden">
              <div class="px-3 py-2 border-b border-[var(--border-color)] mb-1">
                <span class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('interface_theme') }}</span>
              </div>
              <div class="grid grid-cols-1 gap-1 p-1">
                <button 
                  v-for="s in styles" 
                  :key="s.id"
                  @click="selectStyle(s.id)"
                  class="w-full flex items-center gap-3 p-2 rounded-xl hover:bg-[var(--matrix-color)]/10 transition-all group relative"
                  :class="{ 'bg-[var(--matrix-color)]/5': systemStore.style === s.id }"
                >
                  <!-- Mini Preview -->
                  <div class="w-16 h-10 rounded-lg overflow-hidden border border-[var(--border-color)] relative flex-shrink-0">
                    <div class="absolute inset-0 flex" :style="{ backgroundColor: s.colors[systemStore.mode].bg }">
                      <div class="w-1/3 h-full" :style="{ backgroundColor: s.colors[systemStore.mode].sidebar }"></div>
                      <div class="flex-1 flex flex-col">
                        <div class="h-1/3 w-full" :style="{ backgroundColor: s.colors[systemStore.mode].header }"></div>
                        <div class="flex-1 p-1">
                          <div class="w-full h-full rounded-[2px] border border-dashed opacity-30" :style="{ borderColor: s.colors[systemStore.mode].accent }"></div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="flex flex-col items-start gap-0.5">
                    <span class="text-xs font-bold" :class="systemStore.style === s.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-main)]'">
                      {{ t(s.nameKey) }}
                    </span>
                    <div class="flex items-center gap-1">
                      <div class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: s.colors[systemStore.mode].accent }"></div>
                      <span class="text-[8px] uppercase text-[var(--text-muted)] tracking-tighter">{{ s.id }}</span>
                    </div>
                  </div>

                  <Check v-if="systemStore.style === s.id" class="w-3 h-3 ml-auto text-[var(--matrix-color)]" />
                </button>
              </div>
            </div>
          </transition>
        </div>

        <!-- Mode Toggle -->
        <button @click="systemStore.toggleMode()" class="flex items-center justify-center w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 text-gray-400 hover:text-[var(--matrix-color)] transition-colors" :title="t('mode_' + systemStore.mode)">
          <Sun v-if="systemStore.mode === 'dark'" class="w-4 h-4" />
          <Moon v-else class="w-4 h-4" />
        </button>
      </div>
      
      <div class="h-6 w-px bg-black/5 dark:bg-white/5 hidden xs:block"></div>
      
      <button @click="handleLogout" class="flex items-center gap-2 px-2 sm:px-3 py-1.5 rounded-lg text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
        <LogOut class="w-4 h-4" />
        <span class="text-xs font-bold hidden md:inline">{{ t('logout') }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.text-matrix {
  color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: rgba(0, 255, 65, 0.1);
}
.border-matrix\/20 {
  border-color: rgba(0, 255, 65, 0.2);
}

/* Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
