/**
 * Theme Management Module
 */

let themeListener = null;

export function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'auto';
    applyTheme(savedTheme);
    
    // Sync with settings dropdown if exists
    const themeSelect = document.querySelector('select[onchange="setTheme(this.value)"]');
    if (themeSelect) themeSelect.value = savedTheme;
}

export function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
    const next = current === 'dark' ? 'light' : 'dark';
    applyTheme(next);
}

export function applyTheme(theme) {
    // Remove existing listener if any
    if (themeListener) {
        window.matchMedia('(prefers-color-scheme: dark)').removeEventListener('change', themeListener);
        themeListener = null;
    }

    if (theme === 'auto') {
        const mq = window.matchMedia('(prefers-color-scheme: dark)');
        
        // Define listener
        themeListener = (e) => {
             if (e.matches) {
                document.documentElement.setAttribute('data-theme', 'dark');
            } else {
                document.documentElement.removeAttribute('data-theme');
            }
            updateThemeIcon();
        };
        
        // Add listener
        mq.addEventListener('change', themeListener);
        
        // Apply current
        themeListener(mq);
        
    } else if (theme === 'dark') {
        document.documentElement.setAttribute('data-theme', 'dark');
        updateThemeIcon();
    } else {
        document.documentElement.removeAttribute('data-theme');
        updateThemeIcon();
    }
    
    localStorage.setItem('theme', theme);
}

export function updateThemeIcon() {
    const btn = document.querySelector('.theme-toggle-btn-sidebar i');
    if (btn) {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        btn.className = isDark ? 'bi bi-sun-fill' : 'bi bi-moon-stars';
    }
}

export function setTheme(theme) {
    applyTheme(theme);
    // 同步更新设置页面的下拉框
    const select = document.querySelector('select[onchange="setTheme(this.value)"]');
    if (select) select.value = theme;
}

// Global exposure for legacy compatibility
window.initTheme = initTheme;
window.toggleTheme = toggleTheme;
window.applyTheme = applyTheme;
window.updateThemeIcon = updateThemeIcon;
window.setTheme = setTheme;
