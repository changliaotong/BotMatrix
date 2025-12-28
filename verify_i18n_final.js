const fs = require('fs');
const content = fs.readFileSync('src/WebUI/src/utils/i18n.ts', 'utf8');

const langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP'];
const langKeys = {};

langs.forEach(lang => {
    const regex = new RegExp(`'${lang}': \\{([\\s\\S]+?)\\},`, 'm');
    const match = content.match(regex);
    if (!match) {
        console.log(`${lang} block not found`);
        return;
    }
    const block = match[1];
    const keyRegex = /'([^']+)':/g;
    const keys = new Set();
    let m;
    while ((m = keyRegex.exec(block)) !== null) {
        keys.add(m[1]);
    }
    langKeys[lang] = keys;
    console.log(`${lang} has ${keys.size} keys.`);
});

// Find all unique keys across all languages
const allKeys = new Set();
langs.forEach(lang => {
    if (langKeys[lang]) {
        langKeys[lang].forEach(key => allKeys.add(key));
    }
});

console.log(`Total unique keys: ${allKeys.size}`);

langs.forEach(lang => {
    const missing = [];
    allKeys.forEach(key => {
        if (!langKeys[lang].has(key)) {
            missing.push(key);
        }
    });
    if (missing.length > 0) {
        console.log(`${lang} is missing ${missing.length} keys:`, missing.slice(0, 10), missing.length > 10 ? '...' : '');
    } else {
        console.log(`${lang} has all keys.`);
    }
});
