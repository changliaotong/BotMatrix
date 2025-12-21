/**
 * Utilities and global handlers
 */

import { currentLang, translations } from './i18n.js';

// Console history for diagnostic purposes
console.history = [];
(function() {
    const originalLog = console.log;
    const originalError = console.error;
    const originalWarn = console.warn;
    
    console.log = function() {
        console.history.push('[LOG] ' + Array.from(arguments).join(' '));
        originalLog.apply(console, arguments);
    };
    console.error = function() {
        console.history.push('[ERR] ' + Array.from(arguments).join(' '));
        originalError.apply(console, arguments);
    };
    console.warn = function() {
        console.history.push('[WRN] ' + Array.from(arguments).join(' '));
        originalWarn.apply(console, arguments);
    };
})();

// Global Error Handler
window.onerror = function(message, source, lineno, colno, error) {
    console.error('Global Error Caught:', { message, source, lineno, colno, error });
    // Don't show alert for known external library errors that don't break the app
    if (source && (source.includes('chart.js') || source.includes('three.js'))) return;
    return false;
};

/**
 * Find bot information by ID
 */
export const findBotInfo = (botId, platformFromEvent) => {
    const currentBots = window.currentBots || [];
    const bot = currentBots.find(b => b.self_id === botId || b.id === botId);
    if (bot) {
        let avatarUrl = null;
        const platform = (platformFromEvent || bot.platform || '').toUpperCase();
        if (platform === 'QQ') {
            avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${bot.self_id}&s=640`;
        } else if (platform === 'WECHAT' || platform === 'WX') {
            avatarUrl = '/static/avatars/wechat_default.png';
        } else if (bot.user_avatar) {
            avatarUrl = bot.user_avatar;
        }
        return { nickname: bot.nickname || bot.name || botId, avatar: avatarUrl };
    }
    
    if (platformFromEvent) {
        const platform = platformFromEvent.toUpperCase();
        if (platform === 'QQ') {
            return { nickname: botId, avatar: `https://q1.qlogo.cn/g?b=qq&nk=${botId}&s=640` };
        }
    }
    return { nickname: botId, avatar: null };
};

/**
 * Format user avatar URL
 */
export const formatUserAvatar = (avatar, source, platform) => {
    if (avatar && avatar.startsWith('http')) return avatar;
    if (platform && platform.toUpperCase().includes('QQ')) {
        const qq = avatar || source;
        if (/^\d+$/.test(qq)) {
            return `https://q1.qlogo.cn/g?b=qq&nk=${qq}&s=640`;
        }
    }
    return avatar;
};

/**
 * Format bytes to human readable string (e.g., "1.2 MB")
 */
export function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    if (i < 0 || isNaN(i)) return '0 B';
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Relative time formatter
 */
export function timeAgo(date) {
    if (!date) return 'N/A';
    // Use translations if available, else fallback to default
    const t = (translations && currentLang) ? 
              (translations[currentLang] || translations['zh-CN']) : 
              {};
              
    const seconds = Math.floor((new Date() - new Date(date)) / 1000);
    let interval = seconds / 31536000;

    if (interval > 1) return Math.floor(interval) + (t.years_ago || " 年前");
    interval = seconds / 2592000;
    if (interval > 1) return Math.floor(interval) + (t.months_ago || " 个月前");
    interval = seconds / 86400;
    if (interval > 1) return Math.floor(interval) + (t.days_ago || " 天前");
    interval = seconds / 3600;
    if (interval > 1) return Math.floor(interval) + (t.hours_ago || " 小时前");
    interval = seconds / 60;
    if (interval > 1) return Math.floor(interval) + (t.minutes_ago || " 分钟前");
    return Math.floor(seconds) + (t.seconds_ago || " 秒前");
}

/**
 * Show toast notification
 */
export function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        const container = document.createElement('div');
        container.id = 'toast-container';
        container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 10000;';
        document.body.appendChild(container);
    }
    
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} fade show`;
    toast.style.cssText = 'min-width: 200px; margin-bottom: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);';
    toast.innerHTML = message;
    
    document.getElementById('toast-container').appendChild(toast);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 500);
    }, 3000);
}

// Expose to window for legacy support
window.showToast = showToast;
window.findBotInfo = findBotInfo;
window.formatUserAvatar = formatUserAvatar;
window.formatBytes = formatBytes;
window.timeAgo = timeAgo;
