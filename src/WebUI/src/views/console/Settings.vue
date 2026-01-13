<script setup lang="ts">
import { ref } from 'vue';
import { useSystemStore, type Style, type Mode } from '@/stores/system';
import { useAuthStore } from '@/stores/auth';
import { type Language } from '@/utils/i18n';
import { Settings, Shield, Bell, Globe, Save, Palette, Languages, Sun, Moon } from 'lucide-vue-next';

const systemStore = useSystemStore();
const authStore = useAuthStore();
const t = (key: string) => systemStore.t(key);

const settings = ref({
  systemName: 'BotMatrix',
  notifications: true,
});

const activeTab = ref('general');

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
  },
  { 
    id: 'industrial', 
    nameKey: 'style_industrial',
    colors: {
      light: { bg: '#ffffff', sidebar: '#f8fafc', header: '#ffffff', accent: '#2563eb', text: '#0f172a', border: 'rgba(148, 163, 184, 0.15)' },
      dark: { bg: '#0f172a', sidebar: '#070a14', header: '#0f172a', accent: '#38bdf8', text: '#f8fafc', border: '#38bdf8' }
    }
  }
];

const modes: { id: Mode; nameKey: string; icon: any }[] = [
  { id: 'light', nameKey: 'mode_light', icon: Sun },
  { id: 'dark', nameKey: 'mode_dark', icon: Moon }
];

const languages: { id: Language; nameKey: string }[] = [
  { id: 'zh-CN', nameKey: 'lang_zh_cn' },
  { id: 'zh-TW', nameKey: 'lang_zh_tw' },
  { id: 'en-US', nameKey: 'lang_en_us' },
  { id: 'ja-JP', nameKey: 'lang_ja_jp' }
];
</script>

<template>
  <div class="p-4 sm:p-6 max-w-4xl mx-auto space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-black text-[var(--text-main)] tracking-tight flex items-center gap-3">
          <Settings class="w-8 h-8 text-[var(--matrix-color)]" /> {{ t('settings') }}
        </h1>
        <p class="text-xs font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('system_settings_desc') }}</p>
      </div>
      <button class="w-full sm:w-auto px-6 py-2 bg-[var(--matrix-color)] text-black font-black text-xs uppercase tracking-widest rounded-xl hover:opacity-90 transition-opacity flex items-center justify-center gap-2 shadow-lg shadow-[var(--matrix-color)]/20">
        <Save class="w-4 h-4" /> {{ t('save_changes') }}
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- Sidebar Tabs -->
      <div class="flex md:flex-col overflow-x-auto pb-2 md:pb-0 gap-2 md:col-span-1 no-scrollbar">
        <button 
          @click="activeTab = 'general'"
          :class="activeTab === 'general' ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap">
          <Settings class="w-5 h-5" /> {{ t('general') }}
        </button>
        <button 
          @click="activeTab = 'security'"
          :class="activeTab === 'security' ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap">
          <Shield class="w-5 h-5" /> {{ t('security') }}
        </button>
        <button 
          @click="activeTab = 'notifications'"
          :class="activeTab === 'notifications' ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap">
          <Bell class="w-5 h-5" /> {{ t('notifications') }}
        </button>
        <button 
          @click="activeTab = 'language'"
          :class="activeTab === 'language' ? 'bg-[var(--matrix-color)] text-black' : 'hover:bg-[var(--matrix-color)]/10 text-[var(--text-muted)]'"
          class="flex-shrink-0 md:w-full flex items-center gap-3 p-3 sm:p-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all whitespace-nowrap">
          <Globe class="w-5 h-5" /> {{ t('language_region') }}
        </button>
      </div>

      <!-- Content Area -->
      <div class="md:col-span-3 space-y-6">
        <div v-if="activeTab === 'general'" class="space-y-6">
          <div v-if="authStore.isAdmin" class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-4">
            <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
              <Settings class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('basic_config') }}
            </h3>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-[var(--text-muted)] uppercase tracking-widest">{{ t('system_name') }}</label>
              <input 
                v-model="settings.systemName"
                type="text" 
                class="w-full p-3 sm:p-4 rounded-2xl bg-black/5 dark:bg-white/5 border border-[var(--border-color)] focus:border-[var(--matrix-color)] outline-none text-xs font-bold text-[var(--text-main)] transition-colors"
              />
            </div>
          </div>

          <div class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-4">
            <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
              <Sun class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('mode_selection') }}
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <button 
                v-for="mode in modes"
                :key="mode.id"
                @click="systemStore.setMode(mode.id)"
                :class="systemStore.mode === mode.id ? 'border-[var(--matrix-color)] ring-2 ring-[var(--matrix-color)]/20' : 'border-[var(--border-color)] hover:border-[var(--matrix-color)]/50'"
                class="group relative p-0 rounded-2xl border-2 transition-all overflow-hidden flex flex-col bg-[var(--bg-card)]"
              >
                <!-- Mini Preview -->
                <div class="w-full aspect-[21/9] relative overflow-hidden transition-transform group-hover:scale-[1.02] duration-500">
                  <div class="absolute inset-0 flex" :style="{ backgroundColor: mode.id === 'light' ? '#f3f4f6' : '#111827' }">
                    <div class="w-1/3 h-full" :style="{ backgroundColor: mode.id === 'light' ? '#ffffff' : '#1f2937' }"></div>
                    <div class="flex-1 p-2 flex flex-col gap-2">
                      <div class="h-2 w-1/2 rounded-full" :style="{ backgroundColor: mode.id === 'light' ? '#e5e7eb' : '#374151' }"></div>
                      <div class="flex-1 rounded border border-dashed" :style="{ borderColor: mode.id === 'light' ? '#d1d5db' : '#4b5563' }"></div>
                    </div>
                  </div>
                  
                  <!-- Icon Overlay -->
                  <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                    <div class="p-2 rounded-full bg-white/10 backdrop-blur-md border border-white/20 text-[var(--sidebar-text)] shadow-xl transform transition-transform group-hover:rotate-12">
                      <component :is="mode.icon" class="w-6 h-6" />
                    </div>
                  </div>
                </div>
                
                <!-- Label -->
                <div class="p-3 flex items-center justify-center border-t border-[var(--border-color)]">
                  <span class="font-black text-[10px] uppercase tracking-widest" :class="systemStore.mode === mode.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-main)]'">
                    {{ t(mode.nameKey) }}
                  </span>
                </div>
              </button>
            </div>
          </div>

          <div class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-4">
            <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
              <Palette class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('interface_theme') }}
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <button 
                v-for="style in styles"
                :key="style.id"
                @click="systemStore.setStyle(style.id)"
                :class="systemStore.style === style.id ? 'border-[var(--matrix-color)] ring-2 ring-[var(--matrix-color)]/20' : 'border-[var(--border-color)] hover:border-[var(--matrix-color)]/50'"
                class="group relative p-0 rounded-2xl border-2 transition-all overflow-hidden flex flex-col bg-[var(--bg-card)]"
              >
                <!-- Mini Preview -->
                <div class="w-full aspect-[16/9] relative overflow-hidden transition-transform group-hover:scale-[1.02] duration-500">
                  <div class="absolute inset-0 flex" :style="{ backgroundColor: style.colors[systemStore.mode].bg }">
                    <!-- Sidebar -->
                    <div class="w-1/4 h-full border-r" :style="{ backgroundColor: style.colors[systemStore.mode].sidebar, borderColor: style.colors[systemStore.mode].border + '40' }">
                      <div class="p-1 space-y-1">
                        <div v-for="i in 3" :key="i" class="h-1 w-full rounded-full opacity-20" :style="{ backgroundColor: style.colors[systemStore.mode].text }"></div>
                      </div>
                    </div>
                    <!-- Main -->
                    <div class="flex-1 flex flex-col">
                      <!-- Header -->
                      <div class="h-1/4 w-full border-b flex items-center px-2 justify-between" :style="{ backgroundColor: style.colors[systemStore.mode].header, borderColor: style.colors[systemStore.mode].border + '40' }">
                        <div class="h-1 w-4 rounded-full opacity-30" :style="{ backgroundColor: style.colors[systemStore.mode].text }"></div>
                        <div class="h-2 w-2 rounded-full" :style="{ backgroundColor: style.colors[systemStore.mode].accent }"></div>
                      </div>
                      <!-- Content -->
                      <div class="p-2 flex-1">
                        <div class="w-full h-full rounded border-2 border-dashed opacity-20" :style="{ borderColor: style.colors[systemStore.mode].accent }"></div>
                      </div>
                    </div>
                  </div>
                  
                  <!-- Selected Overlay -->
                  <div v-if="systemStore.style === style.id" class="absolute inset-0 bg-[var(--matrix-color)]/10 flex items-center justify-center">
                    <div class="bg-[var(--matrix-color)] text-black p-1.5 rounded-full shadow-lg">
                      <Save class="w-4 h-4" />
                    </div>
                  </div>
                </div>
                
                <!-- Label -->
                <div class="p-3 flex items-center justify-between border-t border-[var(--border-color)]">
                  <span class="font-black text-[10px] uppercase tracking-widest" :class="systemStore.style === style.id ? 'text-[var(--matrix-color)]' : 'text-[var(--text-main)]'">
                    {{ t(style.nameKey) }}
                  </span>
                  <div 
                    class="w-2 h-2 rounded-full"
                    :style="{ backgroundColor: style.colors[systemStore.mode].accent, boxShadow: `0 0 10px ${style.colors[systemStore.mode].accent}` }"
                  ></div>
                </div>
              </button>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'language'" class="space-y-6">
          <div class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-4">
            <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
              <Languages class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('language_region') }}
            </h3>
            <div class="grid grid-cols-1 gap-3">
              <button 
                v-for="lang in languages"
                :key="lang.id"
                @click="systemStore.setLang(lang.id)"
                :class="systemStore.lang === lang.id ? 'border-[var(--matrix-color)] bg-[var(--matrix-color)]/10' : 'border-[var(--border-color)] hover:border-[var(--matrix-color)]/50'"
                class="flex items-center justify-between p-4 rounded-2xl border-2 transition-all"
              >
                <span class="font-black text-xs uppercase tracking-widest">{{ t(lang.nameKey) }}</span>
                <div v-if="systemStore.lang === lang.id" class="w-2 h-2 rounded-full bg-[var(--matrix-color)]"></div>
              </button>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'notifications'" class="space-y-6">
          <div class="p-4 sm:p-6 rounded-3xl bg-[var(--bg-card)] border border-[var(--border-color)] shadow-sm space-y-4">
            <h3 class="text-sm font-black uppercase tracking-widest flex items-center gap-2">
              <Bell class="w-5 h-5 text-[var(--matrix-color)]" /> {{ t('notifications') }}
            </h3>
            <div class="flex items-center justify-between p-4 rounded-2xl bg-black/5 dark:bg-white/5">
              <div class="space-y-1">
                <p class="font-black text-xs uppercase tracking-widest">{{ t('system_notifications') }}</p>
                <p class="text-[10px] font-bold text-[var(--text-muted)] uppercase tracking-widest">{{ t('notification_desc') }}</p>
              </div>
              <button 
                @click="settings.notifications = !settings.notifications"
                :class="settings.notifications ? 'bg-[var(--matrix-color)]' : 'bg-gray-400'"
                class="w-10 sm:w-12 h-5 sm:h-6 rounded-full relative transition-colors flex-shrink-0"
              >
                <div 
                  :class="settings.notifications ? 'translate-x-6 sm:translate-x-7' : 'translate-x-1'"
                  class="absolute top-1 left-0 w-3 h-3 sm:w-4 sm:h-4 bg-white rounded-full transition-transform"
                ></div>
              </button>
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
