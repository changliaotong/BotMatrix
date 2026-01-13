<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useSystemStore, type Style } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { useBotStore } from '@/stores/bot';
import { useRoute, useRouter } from 'vue-router';
import { 
  Menu, 
  Github, 
  Sun, 
  Moon, 
  LogOut, 
  Palette, 
  Check, 
  Languages, 
  Globe,
  User,
  Settings,
  Shield,
  HelpCircle,
  LayoutDashboard,
  Bell,
  ChevronDown
} from 'lucide-vue-next';
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
const showUserMenu = ref(false);
const userMenuRef = ref<HTMLElement | null>(null);

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
  '/console': 'dashboard',
  '/console/bots': 'bots',
  '/console/contacts': 'contacts',
  '/console/messages': 'messages',
  '/console/tasks': 'tasks',
  '/console/fission': 'fission',
  '/console/manual': 'manual',
  '/console/settings': 'settings',
  '/admin/workers': 'workers',
  '/admin/users': 'users',
  '/admin/logs': 'logs',
  '/admin/monitor': 'monitor',
  '/admin/nexus': 'nexus',
  '/admin/ai': 'ai_nexus',
  '/admin/routing': 'routing',
  '/admin/docker': 'docker',
  '/admin/plugins': 'plugins',
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
  'matrix': 'MX'
};

const styles: { id: Style; nameKey: string; colors: { light: any; dark: any } }[] = [
  { 
    id: 'classic', 
    nameKey: 'style_classic',
    colors: {
      light: { bg: '#fdfaff', sidebar: '#ffffff', header: '#ffffff', accent: '#9333ea', text: '#1e1b4b', border: 'rgba(147, 51, 234, 0.1)' },
      dark: { bg: '#020617', sidebar: '#020617', header: '#020617', accent: '#a855f7', text: '#f8fafc', border: 'rgba(168, 85, 247, 0.15)' }
    }
  },
  { 
    id: 'matrix', 
    nameKey: 'style_matrix',
    colors: {
      light: { bg: '#f0fff4', sidebar: '#ffffff', header: '#ffffff', accent: '#059669', text: '#064e3b', border: '#d1fae5' },
      dark: { bg: '#000000', sidebar: '#000000', header: '#000000', accent: '#00ff41', text: '#00ff41', border: '#003b00' }
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
  showUserMenu.value = false;
};

const toggleLangPicker = () => {
  showLangPicker.value = !showLangPicker.value;
  showStylePicker.value = false;
  showUserMenu.value = false;
};

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value;
  showLangPicker.value = false;
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
  showUserMenu.value = false;
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
  if (userMenuRef.value && !userMenuRef.value.contains(target)) {
    showUserMenu.value = false;
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
      <button @click="systemStore.toggleMobileMenu()" class="p-2 text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors md:hidden">
        <Menu class="w-5 h-5" />
      </button>
      <button @click="systemStore.toggleSidebar()" class="hidden md:block p-2 text-[var(--text-muted)] hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">
        <Menu class="w-5 h-5" />
      </button>
      <h2 class="text-base sm:text-lg font-bold tracking-tight text-[var(--text-main)] truncate max-w-[120px] sm:max-w-none">{{ t(routeTitleMap[route.path] || 'dashboard') }}</h2>
    </div>
    
    <div class="flex items-center gap-2 sm:gap-4">
      <!-- Portal Link -->
      <router-link to="/" class="hidden md:flex items-center gap-2 px-3 py-1.5 text-xs font-bold text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-colors">
        <Globe class="w-4 h-4" />
        {{ t('nav_portal') }}
      </router-link>

      <!-- Uptime & Time (Hidden on small mobile) -->
      <div class="hidden sm:flex items-center gap-2 sm:gap-6 px-2 sm:px-4 py-1 sm:py-2 rounded-xl sm:rounded-2xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5">
        <div class="flex flex-col">
          <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] leading-tight">{{ t('system_uptime') }}</span>
          <span class="text-xs sm:text-sm font-bold text-[var(--matrix-color)] mono">{{ localUptime }}</span>
        </div>
        <div class="h-4 sm:h-6 w-px bg-black/10 dark:bg-white/10"></div>
        <div class="flex flex-col text-right">
          <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-[0.2em] leading-tight">{{ t('current_time') }}</span>
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
            <span class="text-xs font-bold">{{ t(langShortNameMap[systemStore.lang]) }}</span>
          </button>

          <!-- Language Picker Panel -->
          <transition name="fade-slide">
            <div v-if="showLangPicker" class="absolute right-0 mt-2 w-40 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-xl z-50 overflow-hidden">
              <div class="px-3 py-2 border-b border-[var(--border-color)] mb-1">
                <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('language_region') }}</span>
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
            <span class="text-xs font-bold uppercase">{{ styleIconMap[systemStore.style] }}</span>
          </button>

          <!-- Style Picker Panel -->
          <transition name="fade-slide">
            <div v-if="showStylePicker" class="absolute right-0 mt-2 w-64 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-xl z-50 overflow-hidden">
              <div class="px-3 py-2 border-b border-[var(--border-color)] mb-1">
                <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('interface_theme') }}</span>
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
                      <span class="text-[10px] uppercase text-[var(--text-muted)] tracking-tighter">{{ s.id }}</span>
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
      
      <!-- User Avatar & Menu -->
      <div class="relative" ref="userMenuRef">
        <button 
          @click="toggleUserMenu"
          class="flex items-center gap-2 p-1 pr-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-all group"
          :class="{ 'bg-black/5 dark:bg-white/5': showUserMenu }"
        >
          <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)]/10 border border-[var(--matrix-color)]/20 flex items-center justify-center text-[var(--matrix-color)] group-hover:scale-105 transition-transform">
            <User class="w-5 h-5" />
          </div>
          <div class="hidden sm:flex flex-col items-start text-left">
            <span class="text-xs font-bold text-[var(--text-main)] leading-none">{{ authStore.user?.username || 'Admin' }}</span>
            <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-tighter">{{ authStore.role }}</span>
          </div>
          <ChevronDown class="w-3.5 h-3.5 text-[var(--text-muted)] group-hover:text-[var(--text-main)] transition-colors" :class="{ 'rotate-180': showUserMenu }" />
        </button>

        <!-- User Menu Panel -->
        <transition name="fade-slide">
          <div v-if="showUserMenu" class="absolute right-0 mt-2 w-56 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-xl z-50 overflow-hidden">
            <!-- User Info Header -->
            <div class="px-4 py-3 border-b border-[var(--border-color)] mb-1">
              <div class="flex flex-col gap-0.5">
                <span class="text-xs font-bold text-[var(--text-main)]">{{ authStore.user?.username || 'Admin User' }}</span>
                <span class="text-xs text-[var(--text-muted)]">{{ authStore.user?.email || 'admin@botmatrix.ai' }}</span>
              </div>
            </div>

            <div class="p-1">
              <router-link to="/console/settings" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] transition-colors group" @click="showUserMenu = false">
                <User class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                <span class="text-xs font-bold">{{ t('personal_profile') }}</span>
              </router-link>

              <router-link to="/console/bots" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] transition-colors group" @click="showUserMenu = false">
                <LayoutDashboard class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                <span class="text-xs font-bold">{{ t('my_bots') }}</span>
              </router-link>

              <router-link to="/console/system" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] transition-colors group" @click="showUserMenu = false">
                <Settings class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                <span class="text-xs font-bold">{{ t('settings') }}</span>
              </router-link>

              <a href="https://docs.botmatrix.ai" target="_blank" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-[var(--matrix-color)]/10 text-[var(--text-main)] transition-colors group">
                <HelpCircle class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                <span class="text-xs font-bold">{{ t('help_center') }}</span>
              </a>
            </div>

            <div class="h-px bg-[var(--border-color)] my-1"></div>

            <div class="p-1">
              <button @click="handleLogout" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-red-500/10 text-red-500 transition-colors group">
                <LogOut class="w-4 h-4" />
                <span class="text-xs font-bold">{{ t('logout') }}</span>
              </button>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </header>
</template>

<style scoped>
.text-matrix {
  color: var(--matrix-color);
}
.bg-matrix\/10 {
  background-color: color-mix(in srgb, var(--matrix-color) 10%, transparent);
}
.border-matrix\/20 {
  border-color: color-mix(in srgb, var(--matrix-color) 20%, transparent);
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
