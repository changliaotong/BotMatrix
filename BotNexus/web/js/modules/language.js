/**
 * Language and i18n management
 */

let currentLang = 'zh-CN';

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
    
    if (typeof translations === 'undefined') {
        console.warn('Translations not loaded yet. Waiting...');
        setTimeout(() => setLanguage(lang), 100);
        return;
    }

    // Fallback for translations
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
        const token = window.authToken || localStorage.getItem('wxbot_token');
        const langParam = lang === 'zh-CN' ? 'zh' : 'en';
        const newSrc = `/overmind/?lang=${langParam}&token=${encodeURIComponent(token)}`;
        // Only update if language actually changed in the URL
        if (!overmindIframe.src.includes(`lang=${langParam}`)) {
            overmindIframe.src = newSrc;
        }
    }

    // Update visualization if active
    if (window.visualizer && typeof window.visualizer.updateLanguage === 'function') {
        window.visualizer.updateLanguage();
    }
}

export function getCurrentLang() {
    return currentLang;
}

// Expose to window for legacy compatibility
window.initLanguage = initLanguage;
window.setLanguage = setLanguage;
window.getCurrentLang = getCurrentLang;
