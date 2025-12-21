/**
 * 可视化引擎模块
 */
import { currentLang, translations, setLanguage } from './i18n.js';
import { authToken, authRole } from './auth.js';
import { addEventLog } from './logs.js';
import { renderBots, updateGlobalBotSelectors, currentBots } from './bots.js';
import { refreshGroupList } from './groups.js';
import { refreshFriendList } from './friends.js';
import { updateStats } from './stats.js';

// --- Visualization Engine ---
export class RoutingVisualizer {
    constructor() {
        this.container = document.getElementById('visualization-container');
        if (!this.container) return;

        if (typeof THREE === 'undefined') {
            console.error('THREE.js is not loaded');
            this.container.innerHTML = '<div class="alert alert-danger m-3">THREE.js 库加载失败，无法显示可视化。</div>';
            return;
        }

        // 3D Scene Setup
        this.scene = new THREE.Scene();
        this.camera = new THREE.PerspectiveCamera(75, this.container.clientWidth / this.container.clientHeight, 0.1, 100000);
        this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
        this.renderer.setSize(this.container.clientWidth, this.container.clientHeight);
        this.renderer.setPixelRatio(window.devicePixelRatio);
        this.container.appendChild(this.renderer.domElement);

        // Controls
        if (THREE.OrbitControls) {
            this.controls = new THREE.OrbitControls(this.camera, this.renderer.domElement);
            this.controls.enableDamping = true;
            this.controls.dampingFactor = 0.05;
            this.controls.minDistance = 50;
            this.controls.maxDistance = 50000;
        } else {
            console.warn('THREE.OrbitControls not found, skipping controls initialization');
        }

        // Camera Position
        this.camera.position.z = 500;

        // Lights
        const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
        this.scene.add(ambientLight);
        const pointLight = new THREE.PointLight(0x00ff41, 1, 1000);
        pointLight.position.set(0, 0, 100);
        this.scene.add(pointLight);

        this.nodes = new Map(); // ID -> {mesh, type, label, lastActive, floatingOffset}
        this.particles = [];
        this.running = true;
        this.totalMessages = 0;

        // Configuration for visualization (can be adjusted by user)
        this.config = {
            botRadius: 1750,
            botGroupMultiplier: 200,
            groupRadius: 15000,
            groupCountMultiplier: 2000,
            groupSpread: 8000,
            userRadius: 8000,
            userCountMultiplier: 500,
            userSpread: 4000,
            wanderScale: 1.0,
            verticalSpread: 1.0
        };
        this.loadConfig();

        // Add Stars
        this.addStars();
        
        // Stats UI
        this.statsEl = document.createElement('div');
        this.statsEl.className = 'viz-stats';
        this.container.appendChild(this.statsEl);

        // Add Nexus (Central Node)
        const t = translations[currentLang] || translations['zh-CN'] || {};
        this.getOrCreateNode('nexus', 'nexus', t.viz_nexus || 'Nexus', null, null, null, t.viz_nexus || 'Nexus');

        this.updateStatsUI();

        // Settings UI
        this.createSettingsUI();

        this.raycaster = new THREE.Raycaster();
        this.mouse = new THREE.Vector2();
        this.container.addEventListener('dblclick', (e) => this.onDoubleClick(e));

        window.addEventListener('resize', () => this.resize());
        
        // Start Animation Loop
        this.animate();

        // Periodic cleanup
        this.cleanupInterval = setInterval(() => {
            this.cleanupNodes();
        }, 10000);
    }

    addStars() {
        const starsGeometry = new THREE.BufferGeometry();
        const starsMaterial = new THREE.PointsMaterial({ color: 0xffffff, size: 1, transparent: true });
        const starsVertices = [];
        for (let i = 0; i < 5000; i++) { // Reduced star count
            const x = (Math.random() - 0.5) * 100000;
            const y = (Math.random() - 0.5) * 100000;
            const z = (Math.random() - 0.5) * 100000;
            starsVertices.push(x, y, z);
        }
        starsGeometry.setAttribute('position', new THREE.Float32BufferAttribute(starsVertices, 3));
        this.stars = new THREE.Points(starsGeometry, starsMaterial);
        this.scene.add(this.stars);
    }

    updateStatsUI() {
        if (!this.statsEl) return;
        const t = translations[currentLang] || translations['zh-CN'] || {};
        const activeNodes = this.nodes.size;
        const activeParticles = this.particles.length;
        
        this.statsEl.innerHTML = `
            <div style="font-size: 1.1rem; border-bottom: 1px solid rgba(0,255,65,0.3); margin-bottom: 5px; padding-bottom: 2px;">
                ${t.viz_stats_title || 'SYSTEM STATUS'}
            </div>
            <div>${t.total_messages || '消息总量'}: <span style="color: #fff; text-shadow: 0 0 5px #00ff41;">${this.totalMessages}</span></div>
            <div>${t.viz_active_nodes || '活动节点'}: ${activeNodes}</div>
            <div>${t.viz_active_tasks || '处理中任务'}: ${activeParticles}</div>
        `;
    }

    loadConfig() {
        const saved = localStorage.getItem('botmatrix_viz_config');
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                this.config = { ...this.config, ...parsed };
            } catch (e) { console.error("Failed to load viz config", e); }
        }
    }

    saveConfig() {
        localStorage.setItem('botmatrix_viz_config', JSON.stringify(this.config));
        // Recalculate all positions
        this.nodes.forEach(node => {
            if (node.type === 'bot') {
                // Keep current angle, but update radius
                const angle = Math.atan2(node.targetPos.z, node.targetPos.x);
                let groupCount = 0;
                this.nodes.forEach(n => { if (n.type === 'group' && n.botId === node.id) groupCount++; });
                const radius = this.config.botRadius + (groupCount * this.config.botGroupMultiplier);
                node.targetPos.set(Math.cos(angle) * radius, (Math.random() - 0.5) * 1500 * this.config.verticalSpread, Math.sin(angle) * radius);
            } else if (node.type === 'group') {
                this.resetGroupPosition(node);
            } else if (node.type === 'user') {
                this.resetUserPosition(node);
            }
        });
    }

    toggleSettings() {
        const el = document.getElementById('viz-settings-panel');
        if (el) {
            const isHidden = (el.style.display === 'none' || el.style.display === '');
            el.style.display = isHidden ? 'block' : 'none';
            console.log("[Visualizer] Settings panel toggled. Now visible:", !isHidden);
        } else {
            console.error("[Visualizer] Settings panel element not found!");
        }
    }

    createSettingsUI() {
        // 1. Create Settings Trigger Button (Gear icon)
        if (!document.getElementById('viz-settings-trigger')) {
            const trigger = document.createElement('div');
            trigger.id = 'viz-settings-trigger';
            trigger.innerHTML = `<i class="bi bi-gear-fill"></i>`;
            trigger.style.cssText = `
                position: absolute;
                bottom: 20px;
                right: 20px;
                width: 44px;
                height: 44px;
                background: rgba(0, 255, 65, 0.2);
                border: 1px solid #00ff41;
                color: #00ff41;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
                z-index: 1010;
                transition: all 0.3s;
                box-shadow: 0 0 15px rgba(0,255,65,0.4);
                font-size: 1.2rem;
            `;
            trigger.onmouseover = () => {
                trigger.style.background = 'rgba(0, 255, 65, 0.4)';
                trigger.style.transform = 'rotate(30deg) scale(1.1)';
            };
            trigger.onmouseout = () => {
                trigger.style.background = 'rgba(0, 255, 65, 0.2)';
                trigger.style.transform = 'rotate(0) scale(1)';
            };
            trigger.onclick = (e) => {
                e.stopPropagation();
                this.toggleSettings();
            };
            this.container.appendChild(trigger);
        }

        // 2. Create Settings Panel if not exists
        let panel = document.getElementById('viz-settings-panel');
        if (!panel) {
            panel = document.createElement('div');
            panel.id = 'viz-settings-panel';
            panel.style.cssText = `
                position: absolute;
                top: 80px;
                right: 20px;
                width: 280px;
                background: rgba(0, 10, 5, 0.9);
                backdrop-filter: blur(15px);
                border: 1px solid rgba(0, 255, 65, 0.4);
                border-radius: 12px;
                padding: 18px;
                color: #00ff41;
                font-family: 'Courier New', Courier, monospace;
                font-size: 0.85rem;
                z-index: 1011;
                display: none;
                box-shadow: 0 0 30px rgba(0,0,0,0.8), inset 0 0 10px rgba(0,255,65,0.1);
                max-height: calc(100% - 100px);
                overflow-y: auto;
            `;
            this.container.appendChild(panel);
        }

        this.updateSettingsUIContent();
    }

    updateSettingsUIContent() {
        const panel = document.getElementById('viz-settings-panel');
        if (!panel) return;
        
        const t = translations[currentLang] || translations['zh-CN'] || {};
        
        const createSlider = (label, key, min, max, step = 1) => `
            <div style="margin-bottom: 12px;">
                <div style="display: flex; justify-content: space-between; margin-bottom: 4px;">
                    <span>${label}</span>
                    <span id="val-${key}">${this.config[key]}</span>
                </div>
                <input type="range" min="${min}" max="${max}" step="${step}" value="${this.config[key]}" 
                    style="width: 100%; accent-color: #00ff41; cursor: pointer;"
                    oninput="window.visualizer.updateConfig('${key}', this.value)">
            </div>
        `;

        panel.innerHTML = `
            <div style="font-weight: bold; border-bottom: 1px solid rgba(0,255,65,0.3); margin-bottom: 15px; padding-bottom: 5px; display: flex; justify-content: space-between;">
                <span>${t.viz_settings_title || 'VISUALIZATION CONFIG'}</span>
                <i class="bi bi-x-lg" onclick="window.visualizer.toggleSettings()" style="cursor: pointer;"></i>
            </div>
            ${createSlider(t.viz_bot_dist || '机器人距离', 'botRadius', 500, 10000, 50)}
            ${createSlider(t.viz_bot_mult || '机器人间距系数', 'botGroupMultiplier', 0, 1000, 10)}
            ${createSlider(t.viz_group_dist || '群组距离', 'groupRadius', 1000, 40000, 100)}
            ${createSlider(t.viz_group_mult || '群组间距系数', 'groupCountMultiplier', 0, 5000, 100)}
            ${createSlider(t.viz_group_spread || '群组离散度', 'groupSpread', 0, 20000, 100)}
            ${createSlider(t.viz_user_dist || '用户距离', 'userRadius', 1000, 30000, 100)}
            ${createSlider(t.viz_user_spread || '用户离散度', 'userSpread', 0, 10000, 100)}
            ${createSlider(t.viz_wander || '浮动幅度', 'wanderScale', 0, 5, 0.1)}
            ${createSlider(t.viz_vertical || '垂直分布', 'verticalSpread', 0, 3, 0.1)}
            <div style="margin-top: 15px; display: flex; gap: 10px;">
                <button onclick="window.visualizer.resetConfig()" style="flex: 1; background: rgba(255,0,0,0.2); border: 1px solid #ff4444; color: #ff4444; padding: 5px; cursor: pointer; border-radius: 4px;">${t.reset || '重置'}</button>
                <button onclick="window.visualizer.saveConfigToDisk()" style="flex: 1; background: rgba(0,255,65,0.2); border: 1px solid #00ff41; color: #00ff41; padding: 5px; cursor: pointer; border-radius: 4px;">${t.save || '保存'}</button>
            </div>
        `;
    }

    updateLanguage() {
        this.updateStatsUI();
        this.updateSettingsUIContent();

        const t = translations[currentLang] || translations['zh-CN'] || {};

        // Update Nexus Label
        const nexus = this.nodes.get('nexus');
        if (nexus) {
            nexus.label = t.viz_nexus || 'Nexus';
            nexus.typeLabel = t.viz_nexus || 'Nexus';
        }

        // Update all node textures and type labels to reflect language change
        this.nodes.forEach(node => {
            if (node.id === 'nexus') return;
            
            // Update type label based on type
            if (node.type === 'bot') node.typeLabel = t.viz_bot || 'Bot';
            else if (node.type === 'group') node.typeLabel = t.viz_group || 'Group';
            else if (node.type === 'user') node.typeLabel = t.viz_user || 'User';
            else if (node.type === 'worker') node.typeLabel = t.viz_worker || 'Worker';
            else if (node.type === 'member') node.typeLabel = t.viz_member || 'Member';

            this.updateNodeTexture(node);
        });
    }

    updateConfig(key, value) {
        this.config[key] = parseFloat(value);
        const valEl = document.getElementById(`val-${key}`);
        if (valEl) valEl.innerText = value;
        this.saveConfig(); // Real-time update in scene
    }

    resetConfig() {
        this.config = {
            botRadius: 1750,
            botGroupMultiplier: 200,
            groupRadius: 15000,
            groupCountMultiplier: 2000,
            groupSpread: 8000,
            userRadius: 8000,
            userCountMultiplier: 500,
            userSpread: 4000,
            wanderScale: 1.0,
            verticalSpread: 1.0
        };
        localStorage.removeItem('botmatrix_viz_config');
        // Update UI sliders
        Object.keys(this.config).forEach(key => {
            const el = document.querySelector(`input[oninput*="'${key}'"]`);
            if (el) el.value = this.config[key];
            const valEl = document.getElementById(`val-${key}`);
            if (valEl) valEl.innerText = this.config[key];
        });
        this.saveConfig();
    }

    saveConfigToDisk() {
        this.saveConfig();
        const t = translations[currentLang] || translations['zh-CN'] || {};
        alert(t.viz_save_success || 'Settings saved to local storage');
        this.toggleSettings();
    }

    resize() {
        if (!this.container || !this.renderer) return;
        const width = this.container.clientWidth;
        const height = this.container.clientHeight;
        if (width <= 0 || height <= 0) return;
        this.camera.aspect = width / height;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(width, height);
    }

    getOrCreateNode(id, type, label, avatar, groupId = null, botId = null, typeLabel = null) {
        if (this.nodes.has(id)) {
            const node = this.nodes.get(id);
            node.lastActive = Date.now();
            
            let needsUpdate = false;
            // Update avatar if changed
            if (avatar && node.avatar !== avatar) {
                node.avatar = avatar;
                needsUpdate = true;
            }
            // Update label if changed
            if (label && node.label !== label) {
                node.label = label;
                needsUpdate = true;
            }
            // Update typeLabel if provided and changed
            if (typeLabel && node.typeLabel !== typeLabel) {
                node.typeLabel = typeLabel;
                needsUpdate = true;
            }
            // Update groupId if provided and changed
            if (groupId && node.groupId !== groupId) {
                node.groupId = groupId;
                // For users, reset their position to cluster around new group
                if (node.type === 'user') {
                    this.resetUserPosition(node);
                }
            }
            // Update botId if provided and changed
            if (botId && node.botId !== botId) {
                node.botId = botId;
                if (node.type === 'group') {
                    this.resetGroupPosition(node);
                }
            }

            if (needsUpdate) {
                this.updateNodeTexture(node);
            }
            return node;
        }

        const node = {
            id, type, label, avatar, groupId, botId,
            typeLabel: typeLabel || type,
            mesh: null,
            orbs: [], // Color orbiting orbs
            nexusLine: null, // Virtual line to Nexus
            lastActive: Date.now(),
            floatingOffset: Math.random() * Math.PI * 2,
            targetPos: new THREE.Vector3(0, 0, 0),
            clusterOffset: new THREE.Vector3(0, 0, 0) // Offset relative to center
        };

        // 3D Space Partitioning (Hierarchy Layout)
        if (id === 'nexus') {
            node.targetPos.set(0, 0, 0);
        } else if (type === 'worker') {
            // Workers: Closest to Nexus
            const angle = Math.random() * Math.PI * 2;
            const radius = 1000 + Math.random() * 500; // Increased 10x
            node.targetPos.set(
                Math.cos(angle) * radius,
                (Math.random() - 0.5) * 800, // Increased 10x
                Math.sin(angle) * radius
            );
        } else if (type === 'bot') {
            // Count how many groups this bot has
            let groupCount = 0;
            this.nodes.forEach(n => {
                if (n.type === 'group' && n.botId === id) groupCount++;
            });

            // Bots: Secondary orbit
            const angle = Math.random() * Math.PI * 2;
            const radius = this.config.botRadius + (groupCount * this.config.botGroupMultiplier);
            node.targetPos.set(
                Math.cos(angle) * radius,
                (Math.random() - 0.5) * 1500 * this.config.verticalSpread,
                Math.sin(angle) * radius
            );
        } else if (type === 'group') {
            this.resetGroupPosition(node);
        } else if (type === 'user') {
            this.resetUserPosition(node);
        }

        // Create Sprite Mesh
        this.updateNodeTexture(node);
        this.scene.add(node.mesh);
        this.nodes.set(id, node);
        this.updateStatsUI();

        return node;
    }

    resetGroupPosition(node) {
        if (node.botId && this.nodes.has(node.botId)) {
            // Group with bot: Cluster around bot
            const spread = this.config.groupSpread;
            node.clusterOffset.set(
                (Math.random() - 0.5) * spread,
                (Math.random() - 0.5) * spread * 0.5 * this.config.verticalSpread,
                (Math.random() - 0.5) * spread
            );
            const bot = this.nodes.get(node.botId);
            node.targetPos.copy(bot.targetPos).add(node.clusterOffset);
        } else {
            // Lone group: Outer orbit
            const angle = Math.random() * Math.PI * 2;
            const radius = this.config.groupRadius + Math.random() * 5000;
            node.targetPos.set(
                Math.cos(angle) * radius,
                (Math.random() - 0.5) * 4000 * this.config.verticalSpread,
                Math.sin(angle) * radius
            );
        }
    }

    resetUserPosition(node) {
        if (node.groupId && this.nodes.has(node.groupId)) {
            // User in group: Cluster around group
            const spread = this.config.userSpread;
            node.clusterOffset.set(
                (Math.random() - 0.5) * spread,
                (Math.random() - 0.5) * spread * 0.5 * this.config.verticalSpread,
                (Math.random() - 0.5) * spread
            );
            const group = this.nodes.get(node.groupId);
            node.targetPos.copy(group.targetPos).add(node.clusterOffset);
        } else {
            // Lone user: Farthest orbit
            const angle = Math.random() * Math.PI * 2;
            const radius = this.config.userRadius + Math.random() * 5000;
            node.targetPos.set(
                Math.cos(angle) * radius,
                (Math.random() - 0.5) * 6000 * this.config.verticalSpread,
                Math.sin(angle) * radius
            );
        }
    }

    updateNodeTexture(node) {
        const canvas = document.createElement('canvas');
        canvas.width = 256;
        canvas.height = 256;
        const ctx = canvas.getContext('2d');

        // Draw Glow
        const gradient = ctx.createRadialGradient(128, 128, 0, 128, 128, 128);
        let color = '#00ff41';
        if (node.id === 'nexus') color = '#00ff41';
        else if (node.type === 'worker') color = '#00d2ff';
        else if (node.type === 'bot') color = '#33ccff';
        else if (node.type === 'group') color = '#ffcc00';
        else color = '#ffffff';

        gradient.addColorStop(0, color + 'aa');
        gradient.addColorStop(0.5, color + '33');
        gradient.addColorStop(1, 'transparent');
        ctx.fillStyle = gradient;
        ctx.fillRect(0, 0, 256, 256);

        // Draw Avatar or Icon
        ctx.save();
        ctx.beginPath();
        ctx.arc(128, 128, 64, 0, Math.PI * 2);
        ctx.clip();
        
        if (node.avatar) {
            const img = new Image();
            // Use proxy to avoid CORS and Referer issues
            let avatarUrl = node.avatar;
            if (avatarUrl.includes('qlogo.cn') || avatarUrl.includes('http')) {
                avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
            }
            img.crossOrigin = "anonymous";
            img.src = avatarUrl;
            img.onload = () => {
                ctx.drawImage(img, 64, 64, 128, 128);
                this.finalizeNodeTexture(node, canvas);
            };
            // Fallback if avatar fails to load
            img.onerror = () => {
                this.drawNodeIcon(ctx, node, color);
                this.finalizeNodeTexture(node, canvas);
            };
        } else {
            this.drawNodeIcon(ctx, node, color);
            this.finalizeNodeTexture(node, canvas);
        }
        ctx.restore();
    }

    drawNodeIcon(ctx, node, color) {
        ctx.fillStyle = color;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.font = 'bold 80px "Bootstrap-Icons", "Segoe UI Symbol"';
        
        let icon = '?';
        if (node.id === 'nexus') icon = 'N';
        else if (node.type === 'worker') icon = 'W';
        else if (node.type === 'bot') icon = 'B';
        else if (node.type === 'group') icon = 'G';
        else icon = 'U';
        
        ctx.fillText(icon, 128, 128);
    }

    finalizeNodeTexture(node, canvas) {
        const ctx = canvas.getContext('2d');
        const t = translations[currentLang] || translations['zh-CN'] || {};
        
        // Draw Label
        ctx.fillStyle = '#ffffff';
        ctx.font = 'bold 24px "PingFang SC", "Microsoft YaHei", sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText(node.label || node.id, 128, 220);
        
        // Draw Type Label
        ctx.fillStyle = 'rgba(255,255,255,0.6)';
        ctx.font = '18px "PingFang SC", "Microsoft YaHei", sans-serif';
        ctx.fillText(`[${node.typeLabel}]`, 128, 245);

        const texture = new THREE.CanvasTexture(canvas);
        if (node.mesh) {
            node.mesh.material.map = texture;
            node.mesh.material.needsUpdate = true;
        } else {
            const material = new THREE.SpriteMaterial({ map: texture, transparent: true });
            node.mesh = new THREE.Sprite(material);
            const scale = node.type === 'user' ? 700 : 900; // Increased 10x
            node.mesh.scale.set(scale, scale, 1);
            node.mesh.position.copy(node.targetPos);
        }
    }

    addEvent(event) {
        if (event.total_messages !== undefined) {
            this.totalMessages = event.total_messages;
            this.updateStatsUI();
        }
        const t = translations[currentLang] || translations['zh-CN'] || {};
        const nexusLabel = t.viz_nexus || 'Nexus';
        const nexus = this.getOrCreateNode('nexus', 'nexus', nexusLabel);
        let startNode, endNode;

        // Try to find bot info from currentBots or event platform
        const findBotInfo = (botId, platformFromEvent) => {
            if (currentBots && Array.isArray(currentBots)) {
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
            }
            // Fallback using platform from event
            if (platformFromEvent) {
                const platform = platformFromEvent.toUpperCase();
                if (platform === 'QQ') {
                    return { nickname: botId, avatar: `https://q1.qlogo.cn/g?b=qq&nk=${botId}&s=640` };
                }
            }
            return { nickname: botId, avatar: null };
        };

        const formatUserAvatar = (avatar, source, platform) => {
            if (avatar && avatar.startsWith('http')) return avatar;
            if (platform && platform.toUpperCase().includes('QQ')) {
                const qq = avatar || source;
                if (/^\d+$/.test(qq)) {
                    return `https://q1.qlogo.cn/g?b=qq&nk=${qq}&s=640`;
                }
            }
            return avatar;
        };

        const formatGroupAvatar = (groupId, platform) => {
            if (platform && platform.toUpperCase().includes('QQ')) {
                if (/^\d+$/.test(groupId)) {
                    return `https://p.qlogo.cn/gh/${groupId}/${groupId}/640/`;
                }
            }
            return null;
        };

        if (event.direction === 'user_to_bot') {
            const userAvatar = formatUserAvatar(event.user_avatar, event.source, event.platform);
            const userNode = this.getOrCreateNode(event.source, 'user', event.user_name || event.source, userAvatar, event.group_id, null, t.viz_user || 'User');
            const botInfo = findBotInfo(event.target, event.platform);
            const botNode = this.getOrCreateNode(event.target, 'bot', botInfo.nickname, botInfo.avatar, event.group_id, null, t.viz_bot || 'Bot');

            if (event.group_id) {
                const groupAvatar = formatGroupAvatar(event.group_id, event.platform);
                const groupNode = this.getOrCreateNode(event.group_id, 'group', event.group_name || `Group ${event.group_id}`, groupAvatar, null, event.target, t.viz_group || 'Group');
                // Routing: User -> Group -> Bot -> Nexus
                this.createParticle(userNode, groupNode, event.msg_type, event.content);
                setTimeout(() => {
                    this.createParticle(groupNode, botNode, event.msg_type);
                    setTimeout(() => {
                        this.createParticle(botNode, nexus, event.msg_type);
                    }, 400);
                }, 400); // Slight delay for second leg
            } else {
                this.createParticle(userNode, botNode, event.msg_type, event.content);
                setTimeout(() => {
                    this.createParticle(botNode, nexus, event.msg_type);
                }, 400);
            }
            return;
        } else if (event.direction === 'bot_to_user') {
            const botInfo = findBotInfo(event.source, event.platform);
            const botNode = this.getOrCreateNode(event.source, 'bot', botInfo.nickname, botInfo.avatar, event.group_id, null, t.viz_bot || 'Bot');
            const userAvatar = formatUserAvatar(event.user_avatar, event.target, event.platform);
            const userNode = this.getOrCreateNode(event.target, 'user', event.user_name || event.target, userAvatar, event.group_id, null, t.viz_user || 'User');

            if (event.group_id) {
                const groupAvatar = formatGroupAvatar(event.group_id, event.platform);
                const groupNode = this.getOrCreateNode(event.group_id, 'group', event.group_name || `Group ${event.group_id}`, groupAvatar, null, event.source, t.viz_group || 'Group');
                // Routing: Nexus -> Bot -> Group -> User
                this.createParticle(nexus, botNode, event.msg_type, event.content);
                setTimeout(() => {
                    this.createParticle(botNode, groupNode, event.msg_type);
                    setTimeout(() => {
                        this.createParticle(groupNode, userNode, event.msg_type);
                    }, 400);
                }, 400);
            } else {
                this.createParticle(nexus, botNode, event.msg_type, event.content);
                setTimeout(() => {
                    this.createParticle(botNode, userNode, event.msg_type);
                }, 400);
            }
            return;
        } else if (event.direction === 'bot_to_nexus') {
            const botInfo = findBotInfo(event.source, event.platform);
            startNode = this.getOrCreateNode(event.source, 'bot', botInfo.nickname, botInfo.avatar, null, null, t.viz_bot || 'Bot');
            endNode = nexus;
        } else if (event.direction === 'nexus_to_worker') {
            startNode = nexus;
            endNode = this.getOrCreateNode(event.target, 'worker', event.target, null, null, null, t.viz_worker || 'Worker');
        } else if (event.direction === 'worker_to_nexus') {
            startNode = this.getOrCreateNode(event.source, 'worker', event.source, null, null, null, t.viz_worker || 'Worker');
            endNode = nexus;
        } else if (event.direction === 'nexus_to_bot') {
            const botInfo = findBotInfo(event.target, event.platform);
            const botNode = this.getOrCreateNode(event.target, 'bot', botInfo.nickname, botInfo.avatar, event.group_id, null, t.viz_bot || 'Bot');
            
            // Routing: Nexus -> Bot -> (Group -> User)
            this.createParticle(nexus, botNode, event.msg_type, event.content);
            if (event.group_id && event.user_id) {
                setTimeout(() => {
                    const groupNode = this.getOrCreateNode(event.group_id, 'group', event.group_name || `Group ${event.group_id}`, null, null, event.target, t.viz_group || 'Group');
                    const userAvatar = formatUserAvatar(event.user_avatar, event.user_id, event.platform);
                    const userNode = this.getOrCreateNode(event.user_id, 'user', event.user_name || event.user_id, userAvatar, event.group_id, null, t.viz_user || 'User');
                    
                    this.createParticle(botNode, groupNode, event.msg_type);
                    setTimeout(() => {
                        this.createParticle(groupNode, userNode, event.msg_type);
                    }, 400);
                }, 400);
            }
            return;
        } else if (event.direction === 'worker_to_bot') {
            const workerNode = this.getOrCreateNode(event.source, 'worker', event.source, null, null, null, t.viz_worker || 'Worker');
            const botInfo = findBotInfo(event.target, event.platform);
            const botNode = this.getOrCreateNode(event.target, 'bot', botInfo.nickname, botInfo.avatar, event.group_id, null, t.viz_bot || 'Bot');
            
            // Full path: Worker -> Nexus -> Bot
            this.createParticle(workerNode, nexus, event.msg_type, event.content);
            setTimeout(() => {
                this.createParticle(nexus, botNode, event.msg_type);
                
                // If it's a group message, continue the path: Bot -> Group -> User
                if (event.group_id && event.user_id) {
                    setTimeout(() => {
                        const groupNode = this.getOrCreateNode(event.group_id, 'group', event.group_name || `Group ${event.group_id}`, null, null, event.target, t.viz_group || 'Group');
                        const userAvatar = formatUserAvatar(event.user_avatar, event.user_id, event.platform);
                        const userNode = this.getOrCreateNode(event.user_id, 'user', event.user_name || event.user_id, userAvatar, event.group_id, null, t.viz_user || 'User');
                        
                        this.createParticle(botNode, groupNode, event.msg_type);
                        setTimeout(() => {
                            this.createParticle(groupNode, userNode, event.msg_type);
                        }, 400);
                    }, 400);
                }
            }, 400);
            return;
        }

        if (startNode && endNode) {
            this.createParticle(startNode, endNode, event.msg_type, event.content);
        }
    }

    createParticle(start, end, msgType, content) {
        const colors = {
            'message': 0x00ff41,
            'request': 0x00d2ff,
            'response': 0xff00ff,
            'image': 0xffff00
        };
        const color = colors[msgType] || 0x00ff41;

        // 1. Create fast traveling sphere
        const geometry = new THREE.SphereGeometry(80, 16, 16); // Increased 10x
        const material = new THREE.MeshBasicMaterial({ color: color });
        const ballMesh = new THREE.Mesh(geometry, material);
        
        // Add Glow to ball
        const light = new THREE.PointLight(color, 3, 1000); // Increased 10x
        ballMesh.add(light);
        ballMesh.position.copy(start.mesh.position);
        this.scene.add(ballMesh);

        // 2. Create dashed connection
        const startPos = start.mesh.position.clone();
        const endPos = end.mesh.position.clone();
        
        // Create a dashed line connection
        const lineGeom = new THREE.BufferGeometry().setFromPoints([startPos, endPos]);
        const lineMat = new THREE.LineDashedMaterial({ 
            color: color, 
            transparent: true, 
            opacity: 0.6,
            dashSize: 150, // Increased 10x
            gapSize: 80,   // Increased 10x
            blending: THREE.AdditiveBlending
        });
        const lineMesh = new THREE.Line(lineGeom, lineMat);
        lineMesh.computeLineDistances(); // Required for dashed lines
        this.scene.add(lineMesh);

        // 3. Create static message hint at midpoint
        let hintMesh = null;
        const midpoint = new THREE.Vector3().lerpVectors(start.mesh.position, end.mesh.position, 0.5);
        if (content) {
            hintMesh = this.createMessageHint(midpoint, content, color);
            // Move it slightly above the line
            hintMesh.position.y += 400; // Increased 10x
        }

        this.particles.push({
            mesh: ballMesh,
            lineMesh: lineMesh,
            hintMesh: hintMesh,
            startPos: start.mesh.position.clone(),
            endPos: end.mesh.position.clone(),
            startNode: start, // Keep reference to track node movement
            endNode: end,     // Keep reference to track node movement
            progress: 0,
            speed: 0.12, // Faster speed for better performance (less concurrent particles)
            ttl: 5000, // Messages and lines vanish after 5 seconds to keep scene clean
            createdAt: Date.now(),
            color: color
        });
    }

    createMessageHint(pos, content, color) {
        const canvas = document.createElement('canvas');
        // Increase resolution for clearer text
        const resolutionScale = 2;
        canvas.width = 512 * resolutionScale;
        canvas.height = 128 * resolutionScale;
        const ctx = canvas.getContext('2d');
        ctx.scale(resolutionScale, resolutionScale);
        
        // Holographic grid background
        ctx.strokeStyle = 'rgba(0, 255, 65, 0.1)';
        ctx.lineWidth = 1;
        for (let i = 0; i < 512; i += 20) {
            ctx.beginPath();
            ctx.moveTo(i, 0);
            ctx.lineTo(i, 128);
            ctx.stroke();
        }
        for (let i = 0; i < 128; i += 20) {
            ctx.beginPath();
            ctx.moveTo(0, i);
            ctx.lineTo(512, i);
            ctx.stroke();
        }

        // Draw background with holographic glow
        ctx.fillStyle = 'rgba(0, 15, 0, 0.9)';
        ctx.beginPath();
        ctx.roundRect(10, 10, 492, 108, 15);
        ctx.fill();
        
        // Double border for high-tech look
        const colorStr = '#' + color.toString(16).padStart(6, '0');
        ctx.strokeStyle = colorStr;
        ctx.lineWidth = 3; // Thicker border
        ctx.stroke();
        
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.4)';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.roundRect(15, 15, 482, 98, 12);
        ctx.stroke();

        // Text styling - use better fonts and clear rendering
        ctx.fillStyle = '#ffffff';
        ctx.font = 'bold 28px "PingFang SC", "Microsoft YaHei", "Segoe UI", monospace';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        
        let text = content;
        if (text.length > 40) text = text.substring(0, 37) + '...';
        
        // Shadow for depth instead of aberration which can blur
        ctx.shadowColor = 'rgba(0, 0, 0, 0.5)';
        ctx.shadowBlur = 4;
        ctx.shadowOffsetX = 2;
        ctx.shadowOffsetY = 2;

        if (text.length > 20) {
            ctx.fillText(text.substring(0, 20), 256, 45);
            ctx.fillText(text.substring(20), 256, 85);
        } else {
            ctx.fillText(text, 256, 64);
        }

        const texture = new THREE.CanvasTexture(canvas);
        // Important for clarity
        texture.minFilter = THREE.LinearFilter;
        texture.magFilter = THREE.LinearFilter;
        texture.generateMipmaps = false;
        
        const material = new THREE.SpriteMaterial({ 
            map: texture, 
            transparent: true, 
            blending: THREE.AdditiveBlending,
            opacity: 0.9
        });
        const sprite = new THREE.Sprite(material);
        
        sprite.position.copy(pos);
        sprite.position.y += 800; // Increased 10x
        sprite.scale.set(1600, 400, 1); // Increased 10x
        
        this.scene.add(sprite);
        return sprite;
    }

    animate() {
        if (!this.running) return;
        requestAnimationFrame(() => this.animate());

        const time = Date.now() * 0.001;

        // Rotate stars slowly
        if (this.stars) {
            this.stars.rotation.y += 0.0002;
            this.stars.rotation.x += 0.0001;
        }

        // Update Nodes Animation
        this.nodes.forEach(node => {
            if (node.id === 'nexus') {
                // Nexus: Stable center but shaking
                const shakeFreq = 15;
                const shakeAmp = 2;
                node.mesh.position.set(
                    Math.sin(time * shakeFreq) * shakeAmp,
                    Math.cos(time * (shakeFreq + 1)) * shakeAmp,
                    Math.sin(time * (shakeFreq - 1)) * shakeAmp
                );
                
                const scale = 120 + Math.sin(time * 2) * 5;
                node.mesh.scale.set(scale, scale, 1);
                node.mesh.material.rotation = Math.sin(time * 0.5) * 0.1;
                return;
            }

            const offset = node.floatingOffset;
            const inactivity = (Date.now() - node.lastActive) / 1000;

            // Update target position based on inactivity and clustering
            if (node.type === 'user') {
                if (node.groupId && this.nodes.has(node.groupId)) {
                    // Cluster around group: Update targetPos relative to group
                    const group = this.nodes.get(node.groupId);
                    node.targetPos.copy(group.targetPos).add(node.clusterOffset);
                } else {
                    // Lone user: Outer rim drifting
                    const isActive = inactivity < 10;
                    const targetRadius = isActive ? this.config.userRadius * 2 : this.config.userRadius * 8;
                    
                    // Optimization: If it's a bot, update its targetPos based on group count periodically
                    if (node.type === 'bot') {
                        let groupCount = 0;
                        this.nodes.forEach(n => {
                            if (n.type === 'group' && n.botId === node.id) groupCount++;
                        });
                        const dynamicRadius = this.config.botRadius + (groupCount * this.config.botGroupMultiplier);
                        const currentRadius = node.targetPos.length();
                        if (currentRadius > 0.1) {
                            node.targetPos.multiplyScalar(dynamicRadius / currentRadius);
                        }
                    } else {
                        const currentRadius = node.targetPos.length();
                        const lerpFactor = isActive ? 0.08 : 0.002;
                        const nextRadius = currentRadius + (targetRadius - currentRadius) * lerpFactor;
                        if (currentRadius > 0.1) {
                            node.targetPos.multiplyScalar(nextRadius / currentRadius);
                        }
                    }
                }
            } else if (node.type === 'bot') {
                // Bots: Keep them stable
                let groupCount = 0;
                this.nodes.forEach(n => {
                    if (n.type === 'group' && n.botId === node.id) groupCount++;
                });
                const targetRadius = this.config.botRadius + (groupCount * this.config.botGroupMultiplier);
                const currentRadius = node.targetPos.length();
                const lerpFactor = 0.01;
                const nextRadius = currentRadius + (targetRadius - currentRadius) * lerpFactor;
                if (currentRadius > 0.1) {
                    node.targetPos.multiplyScalar(nextRadius / currentRadius);
                }
            } else if (node.type === 'group') {
                if (node.botId && this.nodes.has(node.botId)) {
                    // Cluster around bot: Update targetPos relative to bot
                    const bot = this.nodes.get(node.botId);
                    node.targetPos.copy(bot.targetPos).add(node.clusterOffset);
                } else {
                    // Lone group: Middle layer stability
                    const targetRadius = this.config.groupRadius * 2; 
                    const currentRadius = node.targetPos.length();
                    const lerpFactor = 0.005;
                    const nextRadius = currentRadius + (targetRadius - currentRadius) * lerpFactor;
                    if (currentRadius > 0.1) {
                        node.targetPos.multiplyScalar(nextRadius / currentRadius);
                    }
                }
                
                // Vanish if no activity for 2 hours
                if (inactivity > 7200) {
                    node.mesh.visible = false;
                } else {
                    node.mesh.visible = true;
                }
            }

            // Common wandering logic
            const wanderSpeed = 0.2;
            const wanderRadius = (node.type === 'worker' ? 200 : (node.type === 'bot' ? 400 : 600)) * this.config.wanderScale;
            
            const wanderX = Math.sin(time * wanderSpeed + offset) * wanderRadius;
            const wanderY = Math.cos(time * (wanderSpeed * 0.8) + offset) * wanderRadius;
            const wanderZ = Math.sin(time * (wanderSpeed * 1.2) + offset) * wanderRadius;

            node.mesh.position.x = node.targetPos.x + wanderX;
            node.mesh.position.y = node.targetPos.y + wanderY;
            node.mesh.position.z = node.targetPos.z + wanderZ;
            
            // Gentle swaying and pulsing
            node.mesh.material.rotation = Math.sin(time * 0.3 + offset) * 0.2;
            const baseScale = node.type === 'user' ? 700 : 900; // Increased 10x
            const pulse = 1 + Math.sin(time * 1.2 + offset) * 0.08;
            node.mesh.scale.set(baseScale * pulse, baseScale * pulse, 1);

            // Update Orbs Animation
            if (node.orbs.length > 0) {
                node.orbs.forEach(orb => {
                    orb.orbitData.angle += orb.orbitData.speed * 0.01;
                    const x = Math.cos(orb.orbitData.angle) * orb.orbitData.radius;
                    const y = Math.sin(orb.orbitData.angle) * orb.orbitData.radius;
                    
                    // Rotate around node position
                    orb.position.copy(node.mesh.position);
                    const relativePos = new THREE.Vector3(x, y, 0);
                    relativePos.applyAxisAngle(orb.orbitData.axis, orb.orbitData.angle);
                    orb.position.add(relativePos);
                });
            }

            // Update Group Connection Line (User -> Group)
            if (node.groupId && this.nodes.has(node.groupId)) {
                const group = this.nodes.get(node.groupId);
                if (!node.groupLine) {
                    const lineMat = new THREE.LineDashedMaterial({ 
                        color: 0xffd700, // Golden for group connections
                        transparent: true, 
                        opacity: 0.1,
                        dashSize: 40, // Scaled for 10x
                        gapSize: 20,
                        blending: THREE.AdditiveBlending
                    });
                    const lineGeom = new THREE.BufferGeometry().setFromPoints([
                        group.mesh.position.clone(),
                        node.mesh.position.clone()
                    ]);
                    node.groupLine = new THREE.Line(lineGeom, lineMat);
                    node.groupLine.computeLineDistances();
                    this.scene.add(node.groupLine);
                } else {
                    const positions = node.groupLine.geometry.attributes.position.array;
                    positions[0] = group.mesh.position.x;
                    positions[1] = group.mesh.position.y;
                    positions[2] = group.mesh.position.z;
                    positions[3] = node.mesh.position.x;
                    positions[4] = node.mesh.position.y;
                    positions[5] = node.mesh.position.z;
                    node.groupLine.geometry.attributes.position.needsUpdate = true;
                    node.groupLine.computeLineDistances();
                    node.groupLine.material.opacity = 0.15 + Math.sin(time * 3 + offset) * 0.05;
                    node.groupLine.visible = group.mesh.visible && node.mesh.visible;
                }
            } else if (node.groupLine) {
                this.scene.remove(node.groupLine);
                node.groupLine = null;
            }

            // Update Bot Connection Line (Group -> Bot)
            if (node.type === 'group' && node.botId && this.nodes.has(node.botId)) {
                const bot = this.nodes.get(node.botId);
                if (!node.botLine) {
                    const lineMat = new THREE.LineDashedMaterial({ 
                        color: 0x00d2ff, // Cyan for bot connections
                        transparent: true, 
                        opacity: 0.15,
                        dashSize: 60,
                        gapSize: 30,
                        blending: THREE.AdditiveBlending
                    });
                    const lineGeom = new THREE.BufferGeometry().setFromPoints([
                        bot.mesh.position.clone(),
                        node.mesh.position.clone()
                    ]);
                    node.botLine = new THREE.Line(lineGeom, lineMat);
                    node.botLine.computeLineDistances();
                    this.scene.add(node.botLine);
                } else {
                    const positions = node.botLine.geometry.attributes.position.array;
                    positions[0] = bot.mesh.position.x;
                    positions[1] = bot.mesh.position.y;
                    positions[2] = bot.mesh.position.z;
                    positions[3] = node.mesh.position.x;
                    positions[4] = node.mesh.position.y;
                    positions[5] = node.mesh.position.z;
                    node.botLine.geometry.attributes.position.needsUpdate = true;
                    node.botLine.computeLineDistances();
                    node.botLine.material.opacity = 0.2 + Math.sin(time * 2.5 + offset) * 0.08;
                    node.botLine.visible = bot.mesh.visible && node.mesh.visible;
                }
            } else if (node.botLine) {
                this.scene.remove(node.botLine);
                node.botLine = null;
            }

            // Update Nexus Line (Bot -> Nexus)
            if (node.type === 'bot') {
                const nexus = this.nodes.get('nexus');
                if (nexus) {
                    if (!node.nexusLine) {
                        const lineMat = new THREE.LineDashedMaterial({ 
                            color: 0x00ff41, // Green for core connections
                            transparent: true, 
                            opacity: 0.2,
                            dashSize: 100,
                            gapSize: 50,
                            blending: THREE.AdditiveBlending
                        });
                        const lineGeom = new THREE.BufferGeometry().setFromPoints([
                            new THREE.Vector3(0,0,0),
                            node.mesh.position.clone()
                        ]);
                        node.nexusLine = new THREE.Line(lineGeom, lineMat);
                        node.nexusLine.computeLineDistances();
                        this.scene.add(node.nexusLine);
                    } else {
                        const positions = node.nexusLine.geometry.attributes.position.array;
                        positions[0] = 0; positions[1] = 0; positions[2] = 0;
                        positions[3] = node.mesh.position.x;
                        positions[4] = node.mesh.position.y;
                        positions[5] = node.mesh.position.z;
                        node.nexusLine.geometry.attributes.position.needsUpdate = true;
                        node.nexusLine.computeLineDistances();
                        node.nexusLine.material.opacity = 0.25 + Math.sin(time * 2 + offset) * 0.1;
                        node.nexusLine.visible = node.mesh.visible;
                    }
                }
            }
        });

        // Update Particles and Hints
        const nowMs = Date.now();
        for (let i = this.particles.length - 1; i >= 0; i--) {
            const p = this.particles[i];
            const elapsed = nowMs - p.createdAt;
            
            // Update dynamic start/end positions from nodes
            if (p.startNode && p.startNode.mesh) p.startPos.copy(p.startNode.mesh.position);
            if (p.endNode && p.endNode.mesh) p.endPos.copy(p.endNode.mesh.position);

            // Update Line Mesh (Dashed Line) to connect moving nodes
            if (p.lineMesh) {
                if (p.lineMesh.isLine) {
                    // Update Line segments
                    const positions = p.lineMesh.geometry.attributes.position.array;
                    positions[0] = p.startPos.x;
                    positions[1] = p.startPos.y;
                    positions[2] = p.startPos.z;
                    positions[3] = p.endPos.x;
                    positions[4] = p.endPos.y;
                    positions[5] = p.endPos.z;
                    p.lineMesh.geometry.attributes.position.needsUpdate = true;
                    p.lineMesh.computeLineDistances(); // Update dash scaling for moving distance
                }
            }

            // 1. Update Particle Ball Position (Fast travel)
            if (p.progress < 1) {
                p.progress += p.speed;
                p.mesh.position.lerpVectors(p.startPos, p.endPos, p.progress);
                p.mesh.visible = true;
                if (p.lineMesh) p.lineMesh.visible = true;
            } else {
                p.mesh.visible = false;
                if (p.lineMesh) {
                    // Line fades out slowly after delivery
                    const lineLife = 1 - (elapsed - (1/p.speed * 16)) / 1000;
                    if (lineLife > 0) {
                        p.lineMesh.material.opacity = 0.6 * lineLife;
                    } else {
                        p.lineMesh.visible = false;
                    }
                }
            }

            // 2. Update Hint Sprite (Floating upwards)
            if (p.hintMesh) {
                p.hintMesh.position.y += 2;
                const life = 1 - (elapsed / p.ttl);
                p.hintMesh.material.opacity = 0.9 * life;
                if (life <= 0) {
                    this.scene.remove(p.hintMesh);
                    p.hintMesh = null;
                }
            }

            // 3. Cleanup
            if (elapsed > p.ttl) {
                this.scene.remove(p.mesh);
                if (p.lineMesh) this.scene.remove(p.lineMesh);
                if (p.hintMesh) this.scene.remove(p.hintMesh);
                this.particles.splice(i, 1);
            }
        }

        if (this.controls) this.controls.update();
        this.renderer.render(this.scene, this.camera);
    }

    onDoubleClick(event) {
        const rect = this.container.getBoundingClientRect();
        this.mouse.x = ((event.clientX - rect.left) / this.container.clientWidth) * 2 - 1;
        this.mouse.y = -((event.clientY - rect.top) / this.container.clientHeight) * 2 + 1;

        this.raycaster.setFromCamera(this.mouse, this.camera);
        const intersects = this.raycaster.intersectObjects(this.scene.children);

        if (intersects.length > 0) {
            const object = intersects[0].object;
            // Find which node this mesh belongs to
            let foundNode = null;
            this.nodes.forEach(node => {
                if (node.mesh === object) foundNode = node;
            });

            if (foundNode) {
                console.log("Clicked node:", foundNode);
                // Future: Show detail panel
            }
        }
    }

    cleanupNodes() {
        const now = Date.now();
        this.nodes.forEach((node, id) => {
            if (id === 'nexus') return;
            // Remove users/groups inactive for more than 1 hour (3600s)
            if (now - node.lastActive > 3600000) {
                this.scene.remove(node.mesh);
                if (node.groupLine) this.scene.remove(node.groupLine);
                if (node.botLine) this.scene.remove(node.botLine);
                if (node.nexusLine) this.scene.remove(node.nexusLine);
                node.orbs.forEach(o => this.scene.remove(o));
                this.nodes.delete(id);
            }
        });
    }

    clear() {
        this.nodes.forEach(node => {
            if (node.id === 'nexus') return;
            this.scene.remove(node.mesh);
            if (node.groupLine) this.scene.remove(node.groupLine);
            if (node.botLine) this.scene.remove(node.botLine);
            if (node.nexusLine) this.scene.remove(node.nexusLine);
            node.orbs.forEach(o => this.scene.remove(o));
        });
        const nexus = this.nodes.get('nexus');
        this.nodes.clear();
        if (nexus) this.nodes.set('nexus', nexus);
        
        this.particles.forEach(p => {
            this.scene.remove(p.mesh);
            if (p.lineMesh) this.scene.remove(p.lineMesh);
            if (p.hintMesh) this.scene.remove(p.hintMesh);
        });
        this.particles = [];
        this.totalMessages = 0;
        this.updateStatsUI();
    }
}

window.visualizer = null;
let lastSyncState = null;

export function initVisualizer() {
    if (!window.visualizer) {
        window.visualizer = new RoutingVisualizer();
        window.visualizer.getOrCreateNode('nexus', 'nexus');
        
        // Apply last sync state if available
        if (lastSyncState) {
            handleSyncState(lastSyncState);
        }
    }
}

export function handleRoutingEvent(data) {
    if (window.visualizer) {
        window.visualizer.addEvent(data);
    }
    // 同时在最近活动中显示消息
    // 1. bot_to_nexus: 机器人收到的消息 (User -> Bot -> Nexus)
    // 2. bot_to_user: 机器人发送的消息 (Nexus -> Bot -> User)
    if (data.msg_type === 'message' && (data.direction === 'bot_to_nexus' || data.direction === 'bot_to_user')) {
        // 如果内容为空，则不记录日志（减少噪音）
        if (!data.content || data.content.trim() === '') {
            return;
        }

        const logData = {
            post_type: data.direction === 'bot_to_user' ? 'message_sent' : 'message',
            user_id: data.user_id,
            group_id: data.group_id,
            message: data.content,
            sender: {
                nickname: data.user_name || data.user_id
            },
            platform: data.platform
        };
        addEventLog(logData);
    }
}

export function handleSyncState(data) {
    lastSyncState = data;
    
    // 1. Update Global State & UI (Always do this)
    if (data.bots) {
        // Trigger UI updates
        renderBots();
        updateGlobalBotSelectors();
    }

    // 2. Update Dashboard Metrics
    if (data.total_messages !== undefined) {
        const el = document.getElementById('metric-msgs-total');
        if (el) el.innerText = data.total_messages;
    }
    if (data.groups) {
        const el = document.getElementById('metric-groups-total');
        if (el) el.innerText = Object.keys(data.groups).length;
    }
    if (data.friends) {
        const el = document.getElementById('metric-users-total');
        if (el) el.innerText = Object.keys(data.friends).length;
    }

    // 3. Update Visualizer if active
    if (window.visualizer) {
        const t = translations[currentLang] || translations['zh-CN'] || {};
        
        // Sync bots
        if (data.bots) {
            data.bots.forEach(b => {
                let avatarUrl = null;
                if (b.platform && b.platform.toUpperCase() === 'QQ') {
                    avatarUrl = `https://q1.qlogo.cn/g?b=qq&nk=${b.self_id}&s=640`;
                    avatarUrl = `/api/proxy/avatar?url=${encodeURIComponent(avatarUrl)}`;
                }
                window.visualizer.getOrCreateNode(b.self_id, 'bot', b.nickname || b.self_id, avatarUrl, null, null, t.viz_bot || 'Bot');
            });
        }

        // Sync groups
        if (data.groups) {
            Object.values(data.groups).forEach(g => {
                window.visualizer.getOrCreateNode(g.group_id, 'group', g.group_name || `Group ${g.group_id}`, null, null, g.bot_id, t.viz_group || 'Group');
            });
        }
        
        // Sync friends
        if (data.friends) {
            Object.values(data.friends).forEach(f => {
                window.visualizer.getOrCreateNode(f.user_id, 'user', f.nickname || f.user_id, null, null, null, t.viz_user || 'User');
            });
        }
        
        // Sync members
        if (data.members) {
            Object.values(data.members).forEach(m => {
                window.visualizer.getOrCreateNode(m.user_id, 'user', m.nickname || m.user_id, null, m.group_id, null, t.viz_member || 'Member');
            });
        }
    }
}

export function clearVisualization() {
    if (window.visualizer) window.visualizer.clear();
}

export function toggleVisualizerFullScreen() {
    const container = document.getElementById('visualization-container');
    if (!container) return;

    if (!document.fullscreenElement) {
        container.requestFullscreen().catch(err => {
            console.error(`Error attempting to enable full-screen mode: ${err.message} (${err.name})`);
        });
    } else {
        document.exitFullscreen();
    }
}

// Handle full screen change
document.addEventListener('fullscreenchange', () => {
    const isFullScreen = !!document.fullscreenElement;
    const btn = document.getElementById('btn-fullscreen');
    const icon = document.getElementById('icon-fullscreen');
    const text = document.getElementById('text-fullscreen');

    if (btn && icon && text) {
        if (isFullScreen) {
            icon.className = 'bi bi-fullscreen-exit';
            text.setAttribute('data-i18n', 'exit_full_screen');
        } else {
            icon.className = 'bi bi-arrows-fullscreen';
            text.setAttribute('data-i18n', 'full_screen');
        }
        // Trigger re-translation
        setLanguage(currentLang);
    }

    if (window.visualizer) {
        setTimeout(() => window.visualizer.resize(), 100);
    }
});
