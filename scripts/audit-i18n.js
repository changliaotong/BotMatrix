const fs = require('fs');
const path = require('path');

/**
 * BotMatrix I18N Audit Tool
 * This script checks:
 * 1. Key consistency across all languages in i18n.ts
 * 2. Usage of t() in Vue/TS files vs defined keys in i18n.ts
 */

const I18N_FILE = path.join(__dirname, '../src/WebUI/src/utils/i18n.ts');
const SCAN_DIR = path.join(__dirname, '../src/WebUI/src');

// 1. Check i18n.ts consistency
function checkI18nConsistency() {
    const content = fs.readFileSync(I18N_FILE, 'utf8');
    const langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP'];
    const results = {};

    langs.forEach(lang => {
        const startMarker = `  '${lang}': {`;
        const start = content.indexOf(startMarker);
        if (start === -1) return;
        
        let end = -1;
        let depth = 0;
        for (let i = start + startMarker.length - 1; i < content.length; i++) {
            if (content[i] === '{') depth++;
            if (content[i] === '}') {
                depth--;
                if (depth === 0) {
                    end = i;
                    break;
                }
            }
        }

        const section = content.substring(start + startMarker.length, end);
        const keyRegex = /^\s*'([^']+)':/gm;
        const keys = new Set();
        let match;
        while ((match = keyRegex.exec(section)) !== null) {
            keys.add(match[1]);
        }
        results[lang] = keys;
    });

    const allKeys = new Set();
    Object.values(results).forEach(keys => keys.forEach(k => allKeys.add(k)));

    console.log(`[I18N] Total unique keys: ${allKeys.size}`);
    
    let hasError = false;
    langs.forEach(lang => {
        const missing = [...allKeys].filter(k => !results[lang].has(k));
        if (missing.length > 0) {
            console.error(`[ERROR] ${lang} is missing ${missing.length} keys: ${missing.join(', ')}`);
            hasError = true;
        }
    });

    return { allKeys, hasError };
}

// 2. Check usages
function checkUsages(validKeys) {
    function getFiles(dir, allFiles = []) {
        const files = fs.readdirSync(dir);
        files.forEach(file => {
            const name = path.join(dir, file);
            if (fs.statSync(name).isDirectory()) {
                getFiles(name, allFiles);
            } else if (file.endsWith('.vue') || (file.endsWith('.ts') && !file.includes('i18n.ts'))) {
                allFiles.push(name);
            }
        });
        return allFiles;
    }

    const files = getFiles(SCAN_DIR);
    const tRegex = /\bt\(['"]([^'"]+)['"]\)/g;
    let totalMissing = 0;

    files.forEach(file => {
        const content = fs.readFileSync(file, 'utf8');
        let match;
        const missingInFile = new Set();
        while ((match = tRegex.exec(content)) !== null) {
            const key = match[1];
            if (!validKeys.has(key)) {
                missingInFile.add(key);
            }
        }
        
        if (missingInFile.size > 0) {
            const relativePath = path.relative(path.join(__dirname, '..'), file);
            console.error(`[ERROR] Missing keys in ${relativePath}:`);
            missingInFile.forEach(k => console.error(`  - ${k}`));
            totalMissing += missingInFile.size;
        }
    });

    return totalMissing;
}

// Main execution
console.log('--- BotMatrix I18N Audit Start ---');
const { allKeys, hasError: consistencyError } = checkI18nConsistency();
const usageErrors = checkUsages(allKeys);

console.log('--- Audit Summary ---');
if (!consistencyError && usageErrors === 0) {
    console.log('[SUCCESS] All i18n checks passed!');
    process.exit(0);
} else {
    console.error(`[FAILED] Found consistency errors: ${consistencyError}, Missing usage keys: ${usageErrors}`);
    process.exit(1);
}
