import { defineStore } from 'pinia';
import { useBotStore } from './bot';
import { useSystemStore } from './system';
import { t } from '@/utils/i18n';
import api from '@/api';

export const useMeowStore = defineStore('earlymeow', {
  state: () => ({
    cabinMode: localStorage.getItem('meow_cabin_mode') || 'gentle',
    isCabinActive: localStorage.getItem('meow_cabin_active') === 'true',
    activityLogs: [] as any[],
    stats: {
      mood: 98,
      handledMessages: 0,
      focusHours: 0
    }
  }),
  actions: {
    async init() {
      try {
        const botStore = useBotStore();
        if (botStore.bots.length === 0) {
          // Add a timeout to fetchMemberSetup
          const fetchPromise = botStore.fetchMemberSetup();
          const timeoutPromise = new Promise((_, reject) => 
            setTimeout(() => reject(new Error('Fetch timeout')), 5000)
          );
          await Promise.race([fetchPromise, timeoutPromise]);
        }
        this.syncStats();
      } catch (err) {
        console.error('MeowStore init failed:', err);
        // Ensure syncStats runs anyway to provide some data
        this.syncStats();
      }
    },
    
    syncStats() {
      // Simulate mapping from backend data to meow stats
      this.stats.handledMessages = Math.floor(Math.random() * 500) + 1000;
      this.stats.focusHours = (Math.random() * 5 + 2).toFixed(1);
    },

    setMode(mode: string) {
      this.cabinMode = mode;
      localStorage.setItem('meow_cabin_mode', mode);
      const systemStore = useSystemStore();
      const modeName = this.getModeName(mode);
      const logMsg = t(systemStore.lang as any, 'earlymeow.store.log_switch_mode').replace('{mode}', modeName);
      this.addLog(logMsg, 'info');
    },

    toggleCabin() {
      this.isCabinActive = !this.isCabinActive;
      localStorage.setItem('meow_cabin_active', String(this.isCabinActive));
      const systemStore = useSystemStore();
      const logMsg = this.isCabinActive 
        ? t(systemStore.lang as any, 'earlymeow.store.log_cabin_active')
        : t(systemStore.lang as any, 'earlymeow.store.log_cabin_inactive');
      this.addLog(logMsg, this.isCabinActive ? 'success' : 'warning');
    },

    addLog(action: string, type: string = 'info') {
      const now = new Date();
      const time = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`;
      this.activityLogs.unshift({ time, action, type });
      if (this.activityLogs.length > 20) this.activityLogs.pop();
    },

    getModeName(mode: string) {
      const systemStore = useSystemStore();
      const modes: Record<string, string> = {
        gentle: t(systemStore.lang as any, 'earlymeow.store.mode_gentle'),
        focus: t(systemStore.lang as any, 'earlymeow.store.mode_focus'),
        sleep: t(systemStore.lang as any, 'earlymeow.store.mode_sleep')
      };
      return modes[mode] || mode;
    }
  }
});
