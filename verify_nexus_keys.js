const fs = require('fs');
const nexus = fs.readFileSync('src/WebUI/src/views/Nexus.vue', 'utf8');
const i18n = fs.readFileSync('src/WebUI/src/utils/i18n.ts', 'utf8');

const keys = new Set();
const re = /t\('(.+?)'\)/g;
let m;
while (m = re.exec(nexus)) {
    keys.add(m[1]);
}

console.log('Found keys in Nexus.vue:', Array.from(keys));

const langs = ['zh-CN', 'zh-TW', 'en-US', 'ja-JP'];
langs.forEach(lang => {
    // Find the start of the language object
    const langStartMatch = i18n.match(new RegExp(`'${lang}': \\{`));
    if (!langStartMatch) {
        console.log(`Language ${lang} not found`);
        return;
    }
    
    // Find the matching closing brace (simple version)
    const startIdx = langStartMatch.index;
    let braceCount = 0;
    let endIdx = -1;
    for (let i = startIdx; i < i18n.length; i++) {
        if (i18n[i] === '{') braceCount++;
        if (i18n[i] === '}') {
            braceCount--;
            if (braceCount === 0) {
                endIdx = i;
                break;
            }
        }
    }
    
    const section = i18n.substring(startIdx, endIdx);
    keys.forEach(key => {
        if (!section.includes(`'${key}':`)) {
            console.log(`Missing key in ${lang}: ${key}`);
        }
    });
});
