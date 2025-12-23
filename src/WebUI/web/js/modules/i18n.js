/**
 * Internationalization (i18n) management
 */
import { translations } from '../../locales.js?v=1.1.87';

export let currentLang = 'zh-CN';

export function initLanguage() {
    try {
        console.log('initLanguage starting...');
        const savedLang = localStorage.getItem('language') || 'zh-CN';
        setLanguage(savedLang);
        console.log('initLanguage completed, lang:', savedLang);
    } catch (e) {
        console.error('initLanguage failed:', e);
    }
}

export function setLanguage(lang) {
    currentLang = lang;
    localStorage.setItem('language', lang);
    
    const t = translations[lang] || translations['zh-CN'] || {};

    // Update document title
    if (t.app_title) {
        document.title = t.app_title;
    }

    // Update elements with data-i18n (innerText)
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (t[key]) {
             el.innerText = t[key];
        }
    });

    // Update elements with data-i18n-placeholder
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (t[key]) {
             el.placeholder = t[key];
        }
    });
    
     // Update elements with data-i18n-title
    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        const key = el.getAttribute('data-i18n-title');
        if (t[key]) {
             el.title = t[key];
        }
    });

    // Update Overmind iframe if loaded
    const overmindIframe = document.getElementById('overmind-iframe');
    if (overmindIframe && overmindIframe.src && overmindIframe.src !== 'about:blank' && overmindIframe.src.includes('/overmind/')) {
        const token = localStorage.getItem('wxbot_token');
        const langParam = lang === 'zh-CN' ? 'zh' : 'en';
        const newSrc = `/overmind/?lang=${langParam}&token=${encodeURIComponent(token)}`;
        if (!overmindIframe.src.includes(`lang=${langParam}`)) {
            overmindIframe.src = newSrc;
        }
    }

    // Update visualization if active
    if (window.visualizer && typeof window.visualizer.updateLanguage === 'function') {
        window.visualizer.updateLanguage();
    }

    // Refresh dynamic lists if they have update/render functions
    if (window.renderBots) window.renderBots();
    if (window.renderWorkers) window.renderWorkers();
    if (window.renderGroups) window.renderGroups();
    if (window.renderFriends) window.renderFriends();
}

/**
 * Get translation for a key
 */
export function t(key, replacements = {}) {
    const lang = localStorage.getItem('language') || 'zh-CN';
    const dict = translations[lang] || translations['zh-CN'] || {};
    let text = dict[key] || key;

    // Support basic placeholder replacement: {key}
    Object.keys(replacements).forEach(k => {
        text = text.replace(new RegExp(`\\{${k}\\}`, 'g'), replacements[k]);
    });

    return text;
}

// Expose to window for global access
window.t = t;

export { translations };

