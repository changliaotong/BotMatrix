<template>
  <nav class="fixed top-0 w-full z-50 transition-all duration-500 border-b"
    :class="[
      isScrolled 
        ? 'py-1.5 bg-[var(--bg-header)] backdrop-blur-2xl border-[var(--border-color)] shadow-2xl' 
        : 'py-3 bg-[var(--bg-header)]/40 backdrop-blur-sm border-[var(--border-color)]/50',
      isVisible ? 'translate-y-0' : '-translate-y-full'
    ]"
  >
    <div class="max-w-7xl mx-auto px-6 lg:px-8">
      <div class="flex justify-between h-12 items-center">
        <router-link to="/" class="flex items-center gap-3 group cursor-pointer">
          <div class="w-10 h-10 bg-[var(--matrix-color)] rounded-xl flex items-center justify-center font-black text-white group-hover:rotate-12 transition-transform shadow-[var(--matrix-glow)]">M</div>
          <div class="flex flex-col">
            <span class="text-xl font-black tracking-tighter text-[var(--text-main)] leading-none uppercase">{{ tt('common.project_name') }}</span>
            <span class="text-[10px] uppercase tracking-[0.4em] font-bold text-[var(--text-muted)]">{{ tt('common.nexus_os') }}</span>
          </div>
        </router-link>
        
        <!-- Desktop Nav -->
        <div class="hidden md:flex items-center gap-6 text-sm font-black uppercase tracking-[0.2em] text-[var(--text-muted)]">
          <router-link to="/" class="hover:text-[var(--matrix-color)] transition-colors relative group" :class="{ 'text-[var(--matrix-color)]': route.path === '/' }">
            {{ tt('common.nav_home') }}
            <div class="absolute -bottom-1 left-0 w-0 h-0.5 bg-[var(--matrix-color)] group-hover:w-full transition-all" :class="{ 'w-full': route.path === '/' }"></div>
          </router-link>
          
          <!-- Bots Dropdown -->
          <div class="relative group">
            <button class="flex items-center gap-1 hover:text-[var(--matrix-color)] transition-colors py-4" :class="{ 'text-[var(--matrix-color)]': route.path.startsWith('/bots') || route.path.startsWith('/meow') }">
              {{ tt('common.nav_official_bots') }}
              <ChevronDown class="w-3 h-3 group-hover:rotate-180 transition-transform" />
            </button>
            <div class="absolute top-full left-1/2 -translate-x-1/2 w-64 p-2 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-2xl shadow-2xl opacity-0 translate-y-4 pointer-events-none group-hover:opacity-100 group-hover:translate-y-0 group-hover:pointer-events-auto transition-all backdrop-blur-xl">
              <router-link to="/meow" class="flex items-center gap-3 px-4 py-3 hover:bg-[var(--matrix-color)]/5 rounded-xl transition-colors group/item">
                <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover/item:scale-110 transition-transform">
                  <Cat class="w-4 h-4" />
                </div>
                <div class="flex flex-col text-left">
                  <span class="text-sm font-black text-[var(--text-main)] uppercase">{{ tt('common.earlymeow', tt('botmatrix_showcase_earlymeow_title')) }}</span>
                  <span class="text-xs text-[var(--text-muted)] lowercase">{{ tt('common.nav_earlymeow_desc') }}</span>
                </div>
              </router-link>
              <router-link to="/bots/nexus-guard" class="flex items-center gap-3 px-4 py-3 hover:bg-[var(--matrix-color)]/5 rounded-xl transition-colors group/item">
                <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover/item:scale-110 transition-transform">
                  <Shield class="w-4 h-4" />
                </div>
                <div class="flex flex-col text-left">
                  <span class="text-sm font-black text-[var(--text-main)] uppercase">{{ tt('common.nav_nexus_guard') }}</span>
                  <span class="text-xs text-[var(--text-muted)] lowercase">{{ tt('common.nav_nexus_guard_desc') }}</span>
                </div>
              </router-link>
              <router-link to="/bots/digital-employee" class="flex items-center gap-3 px-4 py-3 hover:bg-[var(--matrix-color)]/5 rounded-xl transition-colors group/item border-t border-[var(--border-color)] mt-1 pt-3">
                <div class="w-8 h-8 rounded-lg bg-[var(--matrix-color)]/10 flex items-center justify-center text-[var(--matrix-color)] group-hover/item:scale-110 transition-transform">
                  <User class="w-4 h-4" />
                </div>
                <div class="flex flex-col text-left">
                  <span class="text-sm font-black text-[var(--text-main)] uppercase">{{ tt('common.digital_employee') }}</span>
                  <span class="text-xs text-[var(--text-muted)] lowercase">{{ tt('common.nav_digital_employee_desc') }}</span>
                </div>
              </router-link>
            </div>
          </div>

          <router-link to="/docs" class="hover:text-[var(--matrix-color)] transition-colors relative group" :class="{ 'text-[var(--matrix-color)]': route.path === '/docs' }">
            {{ tt('common.nav_docs') }}
            <div class="absolute -bottom-1 left-0 w-0 h-0.5 bg-[var(--matrix-color)] group-hover:w-full transition-all" :class="{ 'w-full': route.path === '/docs' }"></div>
          </router-link>
          <router-link to="/news" class="hover:text-[var(--matrix-color)] transition-colors relative group" :class="{ 'text-[var(--matrix-color)]': route.path === '/news' }">
            {{ tt('common.nav_news', tt('portal.更新')) }}
            <div class="absolute -bottom-1 left-0 w-0 h-0.5 bg-[var(--matrix-color)] group-hover:w-full transition-all" :class="{ 'w-full': route.path === '/news' }"></div>
          </router-link>
          <router-link to="/pricing" class="hover:text-[var(--matrix-color)] transition-colors relative group" :class="{ 'text-[var(--matrix-color)]': route.path === '/pricing' }">
            {{ tt('common.nav_pricing', tt('portal.方案')) }}
            <div class="absolute -bottom-1 left-0 w-0 h-0.5 bg-[var(--matrix-color)] group-hover:w-full transition-all" :class="{ 'w-full': route.path === '/pricing' }"></div>
          </router-link>
          <router-link to="/about" class="hover:text-[var(--matrix-color)] transition-colors relative group" :class="{ 'text-[var(--matrix-color)]': route.path === '/about' }">
            {{ tt('common.nav_about') }}
            <div class="absolute -bottom-1 left-0 w-0 h-0.5 bg-[var(--matrix-color)] group-hover:w-full transition-all" :class="{ 'w-full': route.path === '/about' }"></div>
          </router-link>
          
          <!-- System Controls -->
          <div class="flex items-center gap-1.5 ml-4">
            <!-- Style Picker -->
            <div class="relative" ref="stylePickerRef">
              <button 
                @click="showStylePicker = !showStylePicker"
                class="p-1.5 rounded-full hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all group border border-transparent hover:border-[var(--matrix-color)]/20"
                :title="tt('common.interface_style')"
              >
                <Palette class="w-3.5 h-3.5 transition-colors" />
              </button>
              
              <transition name="fade-slide">
                <div v-if="showStylePicker" class="absolute right-0 mt-4 w-40 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-2xl z-50 overflow-hidden backdrop-blur-xl">
                  <button 
                    v-for="s in styles" 
                    :key="s.id"
                    @click="selectStyle(s.id)"
                    class="w-full flex items-center justify-between px-4 py-2 hover:bg-[var(--matrix-color)]/5 transition-colors group text-left"
                    :class="{ 'text-[var(--matrix-color)]': systemStore.style === s.id }"
                  >
                    <span class="text-xs font-black uppercase tracking-widest">{{ tt(s.nameKey, s.id) }}</span>
                    <div v-if="systemStore.style === s.id" class="w-1.5 h-1.5 rounded-full bg-[var(--matrix-color)] shadow-[0_0_8px_var(--matrix-color)]"></div>
                  </button>
                </div>
              </transition>
            </div>

            <!-- Mode Toggle -->
            <button 
              @click="systemStore.toggleMode()"
              class="p-1.5 rounded-full hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all group border border-transparent hover:border-[var(--matrix-color)]/20"
              :title="systemStore.mode === 'dark' ? tt('common.light_mode', tt('earlymeow.theme.light')) : tt('common.dark_mode', tt('earlymeow.theme.dark'))"
            >
              <Sun v-if="systemStore.mode === 'dark'" class="w-3.5 h-3.5 transition-colors" />
              <Moon v-else class="w-3.5 h-3.5 transition-colors" />
            </button>

            <!-- Language Picker -->
            <div class="relative" ref="langPickerRef">
              <button 
                @click="showLangPicker = !showLangPicker"
                class="p-1.5 rounded-full hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all group border border-transparent hover:border-[var(--matrix-color)]/20"
              >
                <Globe class="w-3.5 h-3.5 transition-colors" />
              </button>

              <!-- Language Picker Panel -->
              <transition name="fade-slide">
                <div v-if="showLangPicker" class="absolute right-0 mt-4 w-40 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-2xl z-50 overflow-hidden backdrop-blur-xl">
                  <button 
                    v-for="l in languages" 
                    :key="l.id"
                    @click="selectLang(l.id)"
                    class="w-full flex items-center justify-between px-4 py-2 hover:bg-[var(--matrix-color)]/5 transition-colors group text-left"
                    :class="{ 'text-[var(--matrix-color)]': systemStore.lang === l.id }"
                  >
                    <span class="text-sm font-black text-[var(--text-main)]">{{ tt(l.nameKey, l.id) }}</span>
                    <Check v-if="systemStore.lang === l.id" class="w-3 h-3 text-[var(--matrix-color)]" />
                  </button>
                </div>
              </transition>
            </div>

            <!-- GitHub Link -->
            <a 
              href="https://github.com/changliaotong/BotMatrix" 
              target="_blank"
              class="p-1.5 rounded-full hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)] hover:text-[var(--matrix-color)] transition-all group border border-transparent hover:border-[var(--matrix-color)]/20"
              title="GitHub"
            >
              <Github class="w-3.5 h-3.5 transition-colors" />
            </a>
          </div>

          <!-- Login/User Menu -->
          <template v-if="!authStore.isAuthenticated">
            <router-link to="/login" class="px-6 py-2 bg-[var(--text-main)] hover:bg-[var(--matrix-color)] hover:text-white text-[var(--bg-body)] rounded-full transition-all font-black text-xs tracking-widest shadow-[0_0_20px_rgba(255,255,255,0.1)] hover:shadow-[var(--matrix-glow)]">
              {{ tt('common.start_now', tt('earlymeow.nav.start_now')) }}
            </router-link>
          </template>
          <template v-else>
            <div class="relative" ref="userMenuRef">
              <button 
                @click="toggleUserMenu"
                class="flex items-center gap-2 p-1 pr-4 rounded-full bg-[var(--matrix-color)]/5 border border-[var(--border-color)] hover:border-[var(--matrix-color)]/50 transition-all group"
              >
                <div class="w-8 h-8 rounded-full bg-[var(--matrix-color)] flex items-center justify-center text-white font-black group-hover:scale-105 transition-transform">
                  <User class="w-4 h-4" />
                </div>
                <span class="text-xs font-black text-[var(--text-main)] hidden lg:inline">{{ authStore.user?.username || 'Admin' }}</span>
                <ChevronDown class="w-3 h-3 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)] transition-colors" :class="{ 'rotate-180': showUserMenu }" />
              </button>

              <!-- User Menu Panel -->
              <transition name="fade-slide">
                <div v-if="showUserMenu" class="absolute right-0 mt-4 w-64 py-2 rounded-2xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-2xl z-50 overflow-hidden backdrop-blur-xl">
                  <div class="px-6 py-4 border-b border-[var(--border-color)] mb-1">
                    <div class="flex flex-col gap-1 text-left">
                      <span class="text-sm font-black text-[var(--text-main)]">{{ authStore.user?.username || 'Admin User' }}</span>
                      <span class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ authStore.user?.email || 'admin@botmatrix.ai' }}</span>
                    </div>
                  </div>

                  <div class="p-2">
                    <router-link to="/console" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-[var(--matrix-color)]/5 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all group" @click="showUserMenu = false">
                      <LayoutDashboard class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.control_center', tt('portal.控制中心')) }}</span>
                    </router-link>

                    <router-link to="/console/bot-setup" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-[var(--matrix-color)]/5 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all group" @click="showUserMenu = false">
                      <Bot class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.bot_setup', tt('portal.小号设置')) }}</span>
                    </router-link>

                    <router-link to="/console/group-setup" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-[var(--matrix-color)]/5 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all group" @click="showUserMenu = false">
                      <Users class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.group_setup', tt('portal.群聊设置')) }}</span>
                    </router-link>

                    <router-link to="/console/settings" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-[var(--matrix-color)]/5 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all group" @click="showUserMenu = false">
                      <User class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.profile', tt('personal_profile')) }}</span>
                    </router-link>

                    <router-link to="/console/settings" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-[var(--matrix-color)]/5 text-[var(--text-muted)] hover:text-[var(--text-main)] transition-all group" @click="showUserMenu = false">
                      <Settings class="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--matrix-color)]" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.settings', tt('common.system_settings')) }}</span>
                    </router-link>
                  </div>

                  <div class="h-px bg-[var(--border-color)] my-1"></div>

                  <div class="p-2">
                    <button @click="handleLogout" class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-red-500/10 text-red-500 transition-all group">
                      <LogOut class="w-4 h-4" />
                      <span class="text-sm font-black uppercase tracking-widest">{{ tt('common.logout', tt('portal.退出登录')) }}</span>
                    </button>
                  </div>
                </div>
              </transition>
            </div>
          </template>
        </div>

        <!-- Mobile Menu Toggle -->
        <div class="md:hidden">
          <button @click="isMobileMenuOpen = !isMobileMenuOpen" class="p-2 text-[var(--text-muted)]">
            <Menu v-if="!isMobileMenuOpen" class="w-6 h-6" />
            <X v-else class="w-6 h-6" />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile Menu -->
    <transition name="fade">
      <div v-if="isMobileMenuOpen" class="md:hidden bg-[var(--bg-body)]/95 backdrop-blur-3xl border-b border-[var(--border-color)] py-10 px-6 space-y-8 h-screen overflow-y-auto">
        <router-link to="/" class="block text-2xl font-black uppercase tracking-tighter hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">{{ tt('common.nav_home') }}</router-link>
        <div class="space-y-4">
          <div class="text-xs font-black text-[var(--text-muted)]/40 uppercase tracking-[0.4em]">{{ tt('common.nav_official_bots') }}</div>
          <router-link to="/meow" class="flex items-center gap-4 text-xl font-black hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">
            <Cat class="w-6 h-6 text-[var(--matrix-color)]" />
            {{ tt('common.earlymeow', tt('botmatrix_showcase_earlymeow_title')) }}
          </router-link>
          <router-link to="/bots/nexus-guard" class="flex items-center gap-4 text-xl font-black hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">
            <Shield class="w-6 h-6 text-[var(--matrix-color)]" />
            {{ tt('common.nav_nexus_guard') }}
          </router-link>
          <router-link to="/bots/digital-employee" class="flex items-center gap-4 text-xl font-black hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">
            <User class="w-6 h-6 text-[var(--matrix-color)]" />
            {{ tt('common.digital_employee') }}
          </router-link>
        </div>
        <div class="h-px bg-[var(--border-color)]"></div>
        <router-link to="/docs" class="block text-xl font-black uppercase tracking-tighter hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">{{ tt('common.nav_docs', tt('portal.文档中心')) }}</router-link>
        <router-link to="/news" class="block text-xl font-black uppercase tracking-tighter hover:text-[var(--matrix-color)] text-[var(--text-main)]" @click="isMobileMenuOpen = false">{{ tt('common.nav_news', tt('portal.动态更新')) }}</router-link>

        <router-link to="/pricing" class="block text-xl font-black uppercase tracking-tighter hover:text-[var(--matrix-color)]" @click="isMobileMenuOpen = false">{{ tt('common.nav_pricing', tt('portal.版本计划')) }}</router-link>
        <router-link to="/about" class="block text-xl font-black uppercase tracking-tighter hover:text-[var(--matrix-color)]" @click="isMobileMenuOpen = false">{{ tt('common.nav_about', tt('portal.关于我们')) }}</router-link>
        
        <router-link :to="authStore.isAuthenticated ? '/console' : '/login'" class="block w-full py-5 bg-white text-black text-center rounded-2xl font-black text-lg tracking-widest shadow-xl" @click="isMobileMenuOpen = false">
          {{ authStore.isAuthenticated ? tt('common.control_center', tt('earlymeow.nav.enter_console')) : tt('common.start_now', tt('portal.开始使用')) }}
        </router-link>
      </div>
    </transition>
  </nav>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useSystemStore } from '@/stores/system';
import { useBotStore } from '@/stores/bot';
import { 
  ChevronDown, 
  Menu, 
  X, 
  Check, 
  Globe, 
  User, 
  LogOut, 
  Settings, 
  LayoutDashboard, 
  Cat,
  Shield,
  Palette,
  Layout,
  Bot,
  Users,
  Sun,
  Moon,
  Github
} from 'lucide-vue-next';
import { type Language, useI18n } from '@/utils/i18n';
import type { Style } from '@/stores/system';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const systemStore = useSystemStore();
const botStore = useBotStore();
const { t: tt } = useI18n();

const isMobileMenuOpen = ref(false);
const showLangPicker = ref(false);
const langPickerRef = ref<HTMLElement | null>(null);
const showStylePicker = ref(false);
const stylePickerRef = ref<HTMLElement | null>(null);
const showUserMenu = ref(false);
const userMenuRef = ref<HTMLElement | null>(null);
const isScrolled = ref(false);
const isVisible = ref(true);
let lastScrollY = 0;

const styles: { id: Style; nameKey: string; colors: { light: any; dark: any } }[] = [
  { 
    id: 'classic', 
    nameKey: 'common.style_classic',
    colors: {
      light: { bg: '#fdfaff', sidebar: '#ffffff', header: '#ffffff', accent: '#9333ea', text: '#1e1b4b', border: 'rgba(147, 51, 234, 0.1)' },
      dark: { bg: '#020617', sidebar: '#020617', header: '#020617', accent: '#a855f7', text: '#f8fafc', border: 'rgba(168, 85, 247, 0.15)' }
    }
  },
  { 
    id: 'matrix', 
    nameKey: 'common.style_matrix',
    colors: {
      light: { bg: '#f0fff4', sidebar: '#ffffff', header: '#ffffff', accent: '#059669', text: '#064e3b', border: '#d1fae5' },
      dark: { bg: '#000000', sidebar: '#000000', header: '#000000', accent: '#00ff41', text: '#00ff41', border: '#003b00' }
    }
  },
  { 
    id: 'industrial', 
    nameKey: 'common.style_industrial',
    colors: {
      light: { bg: '#f5f5f5', sidebar: '#e5e5e5', header: '#f5f5f5', accent: '#c2410c', text: '#171717', border: '#d4d4d4' },
      dark: { bg: '#1a1a1a', sidebar: '#141414', header: '#1a1a1a', accent: '#f97316', text: '#e5e5e5', border: '#404040' }
    }
  }
];

const handleScroll = () => {
  const currentScrollY = window.scrollY;
  isScrolled.value = currentScrollY > 20;
  
  // Keep header visible as per user request (confused by auto-hide)
  isVisible.value = true;
  
  lastScrollY = currentScrollY;
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});

const languages: { id: Language; nameKey: string }[] = [
  { id: 'zh-CN', nameKey: 'lang_zh_cn' },
  { id: 'zh-TW', nameKey: 'lang_zh_tw' },
  { id: 'en-US', nameKey: 'lang_en_us' },
  { id: 'ja-JP', nameKey: 'lang_ja_jp' }
];

const toggleLangPicker = () => {
  showLangPicker.value = !showLangPicker.value;
  showUserMenu.value = false;
  showStylePicker.value = false;
};

const toggleStylePicker = () => {
  showStylePicker.value = !showStylePicker.value;
  showUserMenu.value = false;
  showLangPicker.value = false;
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
  if (langPickerRef.value && !langPickerRef.value.contains(target)) {
    showLangPicker.value = false;
  }
  if (stylePickerRef.value && !stylePickerRef.value.contains(target)) {
    showStylePicker.value = false;
  }
  if (userMenuRef.value && !userMenuRef.value.contains(target)) {
    showUserMenu.value = false;
  }
};

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside);
});
</script>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
.fade-slide-enter-active {
  transition: all 0.3s ease-out;
}
.fade-slide-leave-active {
  transition: all 0.2s ease-in;
}
.fade-slide-enter-from,
.fade-slide-leave-to {
  transform: translateY(-10px);
  opacity: 0;
}
</style>
