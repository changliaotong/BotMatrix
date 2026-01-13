import { defineStore } from 'pinia';
import { type Language, t, default as i18n } from '../utils/i18n';

export type Style = 'classic' | 'matrix' | 'industrial';
export type Mode = 'light' | 'dark';

export interface MenuItem {
  id: string;
  icon: string;
  titleKey: string;
  adminOnly?: boolean;
}

export interface MenuGroup {
  id: string;
  titleKey: string;
  items: MenuItem[];
  adminOnly?: boolean;
}

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

    const getInitialStyle = (): Style => {
      const saved = localStorage.getItem('wxbot_style') as Style;
      return (saved === 'classic' || saved === 'matrix' || saved === 'industrial') ? saved : 'industrial';
    };

    return {
      uptime: '0m',
      currentTime: new Date().toLocaleTimeString(),
      lang: getInitialLang(),
      style: getInitialStyle(),
      mode: (localStorage.getItem('wxbot_mode') as Mode) || 'dark',
      neuralLinkActive: true,
      isSidebarCollapsed: localStorage.getItem('wxbot_sidebar_collapsed') === 'true',
      showMobileMenu: false,
      aiTranslations: {} as Record<string, string>,
      rawMenuGroups: [
        {
          id: 'console',
          titleKey: 'console_menu',
          items: [
            { id: 'dashboard', icon: 'LayoutDashboard', titleKey: 'dashboard' },
            { id: 'bots', icon: 'Bot', titleKey: 'bots' },
            { id: 'bot-setup', icon: 'Wrench', titleKey: 'bot_setup' },
            { id: 'group-setup', icon: 'Settings2', titleKey: 'group_setup' },
            { id: 'contacts', icon: 'Users', titleKey: 'contacts' },
            { id: 'messages', icon: 'MessageSquare', titleKey: 'messages' },
            { id: 'tasks', icon: 'ListTodo', titleKey: 'tasks' },
            { id: 'fission', icon: 'Share2', titleKey: 'fission' },
            { id: 'settings', icon: 'Settings', titleKey: 'settings' },
          ]
        },
        {
          id: 'admin',
          titleKey: 'admin_menu',
          adminOnly: true,
          items: [
            { id: 'workers', icon: 'Cpu', titleKey: 'workers' },
            { id: 'users', icon: 'UserCog', titleKey: 'users' },
            { id: 'logs', icon: 'Terminal', titleKey: 'logs' },
            { id: 'monitor', icon: 'Activity', titleKey: 'sidebar_monitor' },
            { id: 'nexus', icon: 'Network', titleKey: 'nexus' },
            { id: 'ai', icon: 'Sparkles', titleKey: 'ai_nexus' },
            { id: 'routing', icon: 'Route', titleKey: 'routing' },
            { id: 'docker', icon: 'Box', titleKey: 'docker' },
            { id: 'plugins', icon: 'Box', titleKey: 'plugins' },
          ]
        },
        {
          id: 'help',
          titleKey: 'help_menu',
          items: [
            { id: 'manual', icon: 'BookOpen', titleKey: 'manual' },
          ]
        }
      ] as MenuGroup[]
    };
  },
  getters: {
    isDark: (state) => state.mode === 'dark',
    t: (state) => (key: string) => {
      const local = t(key);
      if (local !== key) return local;
      return state.aiTranslations[`${state.lang}:${key}`] || key;
    },
    menuGroups: (state) => {
      const authStore = (window as any).authStore; // We'll need a way to access authStore here
      // Alternatively, we can pass it from Sidebar.vue or use a reactive approach
      // But for now, let's assume Sidebar.vue will use a filtered version.
      return state.rawMenuGroups;
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
      if (i18n && i18n.global) {
        (i18n.global.locale as any).value = lang;
      }
    },
    setStyle(style: Style) {
      this.style = style;
      localStorage.setItem('wxbot_style', style);
      this.applyTheme();
    },
    setMode(mode: Mode) {
      this.mode = mode;
      localStorage.setItem('theme', mode); // Sync with EarlyMeow's key
      localStorage.setItem('wxbot_mode', mode);
      this.applyTheme();
    },
    applyTheme() {
      // Update DOM classes
      document.documentElement.classList.remove('classic', 'matrix', 'industrial', 'light', 'dark');
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
