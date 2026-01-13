import { createI18n, useI18n } from 'vue-i18n';
import messages from '../locales';

export type Language = 'zh-CN' | 'zh-TW' | 'en-US' | 'ja-JP';

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('wxbot_lang') || 'zh-CN',
  fallbackLocale: {
    'zh-TW': ['zh-CN', 'en-US'],
    'ja-JP': ['en-US'],
    'default': 'en-US'
  },
  messages,
});

export default i18n;

export { useI18n };

export const t = (key: string, ...args: any[]) => {
  return i18n.global.t(key, ...args);
};

export const tt = (key: string, ...args: any[]) => {
  return i18n.global.t(key, ...args);
};
