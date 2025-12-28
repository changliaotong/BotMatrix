import { defineStore } from 'pinia';
import { type Language, t } from '../utils/i18n';

export type Style = 'classic' | 'matrix' | 'xp' | 'ios' | 'kawaii';
export type Mode = 'light' | 'dark';

export const useSystemStore = defineStore('system', {
  state: () => {
    const getInitialLang = (): Language => {
      const saved = localStorage.getItem('wxbot_lang') as Language;
      if (saved) return saved;
      
      if (typeof navigator !== 'undefined') {
        const browserLang = navigator.language;
        if (browserLang.startsWith('zh-TW') || browserLang.startsWith('zh-HK')) return 'zh-TW';
        if (browserLang.startsWith('zh')) return 'zh-CN';
        if (browserLang.startsWith('ja')) return 'ja-JP';
      }
      return 'en-US';
    };

    return {
      uptime: '0m',
      currentTime: new Date().toLocaleTimeString(),
      lang: getInitialLang(),
      style: (localStorage.getItem('wxbot_style') as Style) || 'matrix',
      mode: (localStorage.getItem('wxbot_mode') as Mode) || 'dark',
      neuralLinkActive: true,
      isSidebarCollapsed: localStorage.getItem('wxbot_sidebar_collapsed') === 'true',
      showMobileMenu: false,
      aiTranslations: {} as Record<string, string>,
      menuGroups: [
        {
          id: 'main',
          titleKey: 'main_menu',
          items: [
            { id: 'dashboard', icon: 'LayoutDashboard', titleKey: 'dashboard' },
            { id: 'bots', icon: 'Bot', titleKey: 'bots' },
            { id: 'messages', icon: 'MessageSquare', titleKey: 'messages' },
            { id: 'workers', icon: 'Cpu', titleKey: 'workers' },
            { id: 'contacts', icon: 'Users', titleKey: 'contacts' },
            { id: 'visualization', icon: 'Share2', titleKey: 'sidebar_visualization' },
          ]
        },
        {
          id: 'automation',
          titleKey: 'automation_menu',
          items: [
            { id: 'tasks', icon: 'ListTodo', titleKey: 'tasks' },
            { id: 'fission', icon: 'Share2', titleKey: 'fission' },
          ]
        },
        {
          id: 'infrastructure',
          titleKey: 'infrastructure_menu',
          items: [
            { id: 'docker', icon: 'Box', titleKey: 'docker' },
            { id: 'routing', icon: 'Route', titleKey: 'routing' },
            { id: 'monitor', icon: 'Activity', titleKey: 'sidebar_monitor' },
          ]
        },
        {
          id: 'system',
          titleKey: 'system_menu',
          items: [
            { id: 'users', icon: 'UserCog', titleKey: 'users' },
            { id: 'logs', icon: 'Terminal', titleKey: 'logs' },
            { id: 'manual', icon: 'BookOpen', titleKey: 'manual' },
            { id: 'settings', icon: 'Settings', titleKey: 'settings' },
          ]
        }
      ]
    };
  },
  getters: {
    isDark: (state) => state.mode === 'dark',
    t: (state) => (key: string) => {
      const local = t(state.lang, key);
      if (local !== key) return local;
      return state.aiTranslations[`${state.lang}:${key}`] || key;
    }
  },
  actions: {
    initTheme() {
      this.applyTheme();
      // Start clock updates
      setInterval(() => {
        this.updateTime();
      }, 1000);
      
      // Listen for AI translation update events
      if (typeof window !== 'undefined') {
        window.addEventListener('ai-translation-updated', ((e: CustomEvent) => {
          const { lang, key } = e.detail;
          this.aiTranslations[`${lang}:${key}`] = (window as any).translatedKeys[lang][key];
        }) as EventListener);
      }
    },
    toggleSidebar() {
      this.isSidebarCollapsed = !this.isSidebarCollapsed;
      localStorage.setItem('wxbot_sidebar_collapsed', String(this.isSidebarCollapsed));
    },
    toggleMobileMenu() {
      this.showMobileMenu = !this.showMobileMenu;
    },
    updateTime() {
      this.currentTime = new Date().toLocaleTimeString();
    },
    setUptime(uptime: string) {
      this.uptime = uptime;
    },
    setLang(lang: Language) {
      this.lang = lang;
      localStorage.setItem('wxbot_lang', lang);
    },
    setStyle(style: Style) {
      this.style = style;
      localStorage.setItem('wxbot_style', style);
      this.applyTheme();
    },
    setMode(mode: Mode) {
      this.mode = mode;
      localStorage.setItem('wxbot_mode', mode);
      this.applyTheme();
    },
    applyTheme() {
      // Update DOM classes
      document.documentElement.classList.remove('classic', 'matrix', 'xp', 'ios', 'kawaii', 'light', 'dark');
      document.documentElement.classList.add(this.style);
      document.documentElement.classList.add(this.mode);
      
      // Also add 'dark' class if mode is dark for Tailwind
      if (this.mode === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
    },
    toggleMode() {
      this.setMode(this.mode === 'light' ? 'dark' : 'light');
    }
  }
});
