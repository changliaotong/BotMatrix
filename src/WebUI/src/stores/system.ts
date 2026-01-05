import { defineStore } from 'pinia';
import { type Language, t } from '../utils/i18n';

export type Style = 'classic' | 'matrix' | 'xp' | 'ios' | 'kawaii' | 'custom-style';
export type Mode = 'light' | 'dark';

export interface CustomStyleConfig {
  '--custom-bg-body': string;
  '--custom-bg-card': string;
  '--custom-bg-header': string;
  '--custom-bg-sidebar': string;
  '--custom-text-main': string;
  '--custom-text-muted': string;
  '--custom-border-color': string;
  '--custom-matrix-color': string;
  '--custom-radius-main'?: string;
  '--custom-radius-card'?: string;
}

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

    const savedCustomStyle = localStorage.getItem('wxbot_custom_style_config');

    return {
      uptime: '0m',
      currentTime: new Date().toLocaleTimeString(),
      lang: getInitialLang(),
      style: (localStorage.getItem('wxbot_style') as Style) || 'matrix',
      mode: (localStorage.getItem('wxbot_mode') as Mode) || 'dark',
      customStyleConfig: savedCustomStyle ? JSON.parse(savedCustomStyle) as CustomStyleConfig : null,
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
      const local = t(state.lang, key);
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
    setCustomStyle(config: CustomStyleConfig) {
      this.customStyleConfig = config;
      localStorage.setItem('wxbot_custom_style_config', JSON.stringify(config));
      this.setStyle('custom-style');
    },
    async generateAIColors(primaryColor: string) {
      // Simulation of AI color generation
      // In a real app, this would call an LLM or a color palette API
      const isDark = this.mode === 'dark';
      
      const config: CustomStyleConfig = {
        '--custom-matrix-color': primaryColor,
        '--custom-bg-body': isDark ? '#000000' : '#ffffff',
        '--custom-bg-card': isDark ? 'rgba(30, 30, 30, 0.8)' : 'rgba(255, 255, 255, 0.8)',
        '--custom-bg-header': isDark ? 'rgba(20, 20, 20, 0.9)' : 'rgba(240, 240, 240, 0.9)',
        '--custom-bg-sidebar': isDark ? '#111111' : '#f8f8f8',
        '--custom-text-main': isDark ? '#ffffff' : '#000000',
        '--custom-text-muted': isDark ? '#888888' : '#666666',
        '--custom-border-color': isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
        '--custom-radius-main': '16px',
        '--custom-radius-card': '24px',
      };

      this.setCustomStyle(config);
    },
    applyTheme() {
      // Update DOM classes
      document.documentElement.classList.remove('classic', 'matrix', 'xp', 'ios', 'kawaii', 'custom-style', 'light', 'dark');
      document.documentElement.classList.add(this.style);
      document.documentElement.classList.add(this.mode);
      
      // If custom style, apply variables to root
      if (this.style === 'custom-style' && this.customStyleConfig) {
        Object.entries(this.customStyleConfig).forEach(([key, value]) => {
          document.documentElement.style.setProperty(key, value);
        });
      } else {
        // Clear custom variables if not using custom style
        const vars = [
          '--custom-bg-body', '--custom-bg-card', '--custom-bg-header', '--custom-bg-sidebar',
          '--custom-text-main', '--custom-text-muted', '--custom-border-color', '--custom-matrix-color',
          '--custom-radius-main', '--custom-radius-card'
        ];
        vars.forEach(v => document.documentElement.style.removeProperty(v));
      }

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
