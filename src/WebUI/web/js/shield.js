/**
 * Ultimate System Shield - Standalone Early Version
 * Combined protection from global and module versions.
 * Protects against MutationObserver errors, filters extension noise, and maintains console history.
 */
(function(window) {
    const TAG = '[Shield]';
    if (window.__shield_active) {
        console.log(TAG, 'Already active, skipping re-init');
        return;
    }
    window.__shield_active = true;

    console.log(TAG, 'Early protection layer activating...');

    // 1. MutationObserver Protection (Aggressive Hijack)
    const moNames = ['MutationObserver', 'WebKitMutationObserver', 'MozMutationObserver'];
    
    const isSafeNode = (target) => {
        try {
            if (!target) return false;
            if (typeof Node !== 'undefined' && target instanceof Node) return true;
            // Cross-context / duck-typing check
            return (typeof target === 'object' && 
                    typeof target.nodeType === 'number' && 
                    typeof target.nodeName === 'string');
        } catch (e) { return false; }
    };

    const createSafeObserve = (nativeObserve) => {
        return function(target, options) {
            try {
                let effectiveTarget = target;
                if (typeof effectiveTarget === 'string') {
                    try {
                        effectiveTarget = document.querySelector(effectiveTarget);
                    } catch (e) { effectiveTarget = null; }
                }

                if (!isSafeNode(effectiveTarget)) {
                    // console.debug(TAG, 'Blocked invalid observe call');
                    return; 
                }

                const safeOptions = options || { childList: true, subtree: true };
                return nativeObserve.call(this, effectiveTarget, safeOptions);
            } catch (e) {
                // console.debug(TAG, 'Caught native observe error:', e.message);
                return; 
            }
        };
    };

    moNames.forEach(name => {
        const OriginalMO = window[name];
        if (!OriginalMO || !OriginalMO.prototype) return;

        const nativeObserve = OriginalMO.prototype.observe;
        const safeObserve = createSafeObserve(nativeObserve);

        // Proxy constructor to handle new instances and prototype access
        const ProxyMO = new Proxy(OriginalMO, {
            construct(target, args) {
                const instance = Reflect.construct(target, args);
                try {
                    Object.defineProperty(instance, 'observe', { 
                        value: safeObserve, 
                        writable: true, 
                        configurable: true 
                    });
                } catch (e) {}
                return instance;
            },
            get(target, prop) {
                if (prop === 'prototype') return target.prototype;
                if (prop === 'observe') return safeObserve;
                const val = Reflect.get(target, prop);
                return typeof val === 'function' ? val.bind(target) : val;
            }
        });

        // Hijack prototype and global
        try {
            Object.defineProperties(OriginalMO.prototype, {
                'observe': { 
                    get() { return safeObserve; }, 
                    set(v) {}, 
                    configurable: true 
                }
            });
            Object.defineProperty(window, name, { 
                value: ProxyMO, 
                configurable: true, 
                writable: true 
            });
        } catch (e) {
            window[name] = ProxyMO;
        }
    });

    // 2. Iframe Protection (Intercept dynamically created iframes)
    const applyProtectionToWindow = (win) => {
        try {
            if (!win || win === window) return;
            
            // Apply MutationObserver protection to iframe window
            moNames.forEach(name => {
                const MO = win[name];
                if (MO && MO.prototype) {
                    const original = MO.prototype.observe;
                    MO.prototype.observe = createSafeObserve(original);
                }
            });

            // Apply error filtering to iframe window
            win.addEventListener('error', (e) => {
                if (isNoisyError(e.message, e.filename, e.error)) {
                    e.preventDefault(); e.stopImmediatePropagation();
                }
            }, true);
        } catch (e) {}
    };

    try {
        const originalAppendChild = Element.prototype.appendChild;
        Element.prototype.appendChild = function(node) {
            const res = originalAppendChild.apply(this, arguments);
            if (node && node.tagName === 'IFRAME' && node.contentWindow) {
                applyProtectionToWindow(node.contentWindow);
            }
            return res;
        };
    } catch (e) {}

    // 3. Error & Console Filtering
    const noisyStrings = [
        "MutationObserver",
        "parameter 1 is not of type 'Node'",
        "observe' on 'MutationObserver",
        "chrome-extension://",
        "dynamically imported module",
        "index.ts-e1d874e5.js",
        "index.ts-loader3.js",
        "Duplicate export of 'translations'",
        "setting 'innerHTML'",
        "Failed to execute 'observe'",
        "index.ts-",
        "not of type 'node'",
        "failed to execute 'observe'",
        "cdn.tailwindcss.com should not be used in production",
        "Failed to execute 'observe' on 'MutationObserver'",
        "parameter 1 is not of type 'Node'",
        "parameter 1 is not of type",
        "ResizeObserver loop limit exceeded",
        "three.js r128",
        "THREE.WebGLRenderer"
    ];

    const isNoisyError = (msg, source, errorObj) => {
        if (!msg) return false;
        let combined = (String(msg) + " " + (source || "")).toLowerCase();
        if (errorObj && errorObj.stack) {
            combined += " " + String(errorObj.stack).toLowerCase();
        }
        
        // Custom visualization error filter
        if (combined.includes('visualization.js') && combined.includes('innerhtml')) return true;

        // Specific MutationObserver error filtering
        if (combined.includes('mutationobserver') && (combined.includes('node') || combined.includes('parameter 1'))) return true;

        return noisyStrings.some(s => combined.includes(s.toLowerCase()));
    };

    // Global Error Listeners
    window.addEventListener('error', (e) => {
        if (isNoisyError(e.message, e.filename, e.error)) {
            e.preventDefault(); 
            e.stopImmediatePropagation();
        }
    }, true);

    window.addEventListener('unhandledrejection', (e) => {
        const reason = e.reason || {};
        const msg = reason.message || String(reason);
        const stack = reason.stack || "";
        if (isNoisyError(msg, "", { stack })) {
            e.preventDefault(); 
            e.stopImmediatePropagation();
        }
    }, true);

    // 4. Console Hijacking (History + Filtering)
    const originalConsole = {
        log: console.log,
        error: console.error,
        warn: console.warn,
        info: console.info,
        debug: console.debug
    };

    window.console.history = window.console.history || [];
    const addToHistory = (type, args) => {
        try {
            const msg = `[${type}] ${new Date().toLocaleTimeString()} - ${Array.from(args).map(a => {
                try {
                    if (a instanceof Error) return a.message + (a.stack ? " " + a.stack : "");
                    if (typeof a === 'object') {
                        // Avoid circular refs or huge objects
                        if (a === window || a === document) return '[Global Object]';
                        return JSON.stringify(a).substring(0, 500);
                    }
                    return String(a);
                } catch(e) { return String(a); }
            }).join(' ')}`;
            window.console.history.push(msg);
            if (window.console.history.length > 500) window.console.history.shift();
        } catch (e) {}
    };

    console.log = function() {
        addToHistory('LOG', arguments);
        originalConsole.log.apply(console, arguments);
    };
    console.warn = function() {
        const args = Array.from(arguments);
        const msg = args.map(a => String(a)).join(' ');
        if (isNoisyError(msg, "")) return;
        addToHistory('WARN', arguments);
        originalConsole.warn.apply(console, arguments);
    };
    console.info = function() {
        addToHistory('INFO', arguments);
        originalConsole.info.apply(console, arguments);
    };
    console.debug = function() {
        addToHistory('DEBUG', arguments);
        originalConsole.debug.apply(console, arguments);
    };

    console.error = function() {
        const args = Array.from(arguments);
        const msg = args.map(a => {
            if (a instanceof Error) return a.message + " " + (a.stack || "");
            if (a && a.stack) return String(a) + " " + a.stack;
            return String(a);
        }).join(' ');
        
        if (isNoisyError(msg, "")) return;
        
        addToHistory('ERROR', arguments);
        originalConsole.error.apply(console, arguments);
    };

    console.log(TAG, 'Ultimate Early protection active.');
})(window);
