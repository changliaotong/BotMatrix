console.log('matrix.js loading...');

// Global error handler for easier debugging
window.onerror = function(msg, url, lineNo, columnNo, error) {
    const errorMsg = `[Global Error] ${msg} at ${url}:${lineNo}:${columnNo}`;
    console.error(errorMsg, error);
    // Only alert for non-extension errors if possible, but for now alert all to help user
    if (url && !url.startsWith('chrome-extension')) {
        // alert(errorMsg);
    }
    return false;
};

window.onunhandledrejection = function(event) {
    console.error('[Unhandled Rejection]', event.reason);
    // alert(`[Promise Error] ${event.reason}`);
};

if (typeof Vue === 'undefined') {
    const errorTitle = (window.t && window.t('vue_not_loaded_error')) || 'Vue is NOT defined! Check index.html script tags.';
    const errorMsg = (window.t && window.t('vue_not_loaded_msg')) || 'Error: Vue.js not loaded. Please check your internet connection or script paths.';
    console.error(errorTitle);
    document.body.innerHTML = `<div style="color:red;padding:20px;">${errorMsg}</div>`;
}
const { createApp, ref, computed, onMounted, onUnmounted, watch, nextTick, toRaw } = Vue;

const app = createApp({
    setup() {
        console.log('Vue setup() starting...');
        
        // Helper for safe localStorage access
        const safeStorage = {
            getItem: (key) => {
                try {
                    return localStorage.getItem(key);
                } catch (e) {
                    console.warn('localStorage access failed:', e);
                    return null;
                }
            },
            setItem: (key, value) => {
                try {
                    localStorage.setItem(key, value);
                } catch (e) {
                    console.warn('localStorage write failed:', e);
                }
            }
        };

        const lang = ref(safeStorage.getItem('language') || 'zh-CN');
        const isDark = ref(safeStorage.getItem('theme') !== 'light'); // Default to dark
        const shieldActive = ref(window.__shield_active || false);
        
        // Auth state
        const token = safeStorage.getItem('wxbot_token');
        
        // Basic token validation: must exist and look like a JWT (3 parts)
        const isValidToken = (t) => {
            console.log('Checking token validity:', t ? (t.substring(0, 10) + '...') : 'null');
            if (!t || t === 'undefined' || t === 'null') {
                console.log('Token is null or undefined string');
                return false;
            }
            // Simple JWT check
            const parts = t.split('.');
            console.log('Token parts count:', parts.length);
            return parts.length === 3;
        };

        const isLoggedIn = ref(isValidToken(token));
        if (isLoggedIn.value) {
            window.authToken = token;
        }
        console.log('Auth state:', { 
            isLoggedIn: isLoggedIn.value, 
            hasToken: !!token,
            tokenType: typeof token,
            tokenValue: token ? (token.substring(0, 10) + '...') : 'none'
        });
        const loginLoading = ref(false);
        const loginError = ref('');
        const loginData = ref({
            username: '',
            password: ''
        });

        const safeCreateIcons = () => {
            if (typeof lucide !== 'undefined') {
                nextTick(() => {
                    try {
                        lucide.createIcons();
                    } catch (e) {
                        console.warn('Lucide icons error:', e);
                    }
                });
            }
        };

        const handleLogin = async () => {
            if (!loginData.value.username || !loginData.value.password) {
                loginError.value = t('alert_enter_user_pass') || 'Please enter username and password';
                return;
            }

            loginLoading.value = true;
            loginError.value = '';

            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(loginData.value)
                });

                const data = await response.json();

                if (data.success && data.token) {
                    safeStorage.setItem('wxbot_token', data.token);
                    window.authToken = data.token; // Ensure global token is set for other modules
                    if (data.role) safeStorage.setItem('wxbot_role', data.role);
                    isLoggedIn.value = true;
                    // Reset login data
                    loginData.value.username = '';
                    loginData.value.password = '';
                    // Fetch data for the main app
                    nextTick(() => {
                        fetchAllData();
                        fetchUserInfo();
                        initWebSocket();
                        safeCreateIcons();
                    });
                } else {
                    loginError.value = data.message || t('login_failed') || 'Login failed';
                }
            } catch (err) {
                console.error('Login error:', err);
                loginError.value = t('network_error') || 'Network error or server unavailable';
            } finally {
                loginLoading.value = false;
            }
        };

        const t = (key) => {
            if (!key) return '';
            const dict = translations[lang.value] || translations['zh-CN'] || {};
            return dict[key] || key;
        };
        window.t = t;

        const toggleLang = () => {
            lang.value = lang.value === 'zh-CN' ? 'en' : 'zh-CN';
            safeStorage.setItem('language', lang.value);
            document.documentElement.lang = lang.value;
            safeCreateIcons();
        };

        const toggleTheme = () => {
            isDark.value = !isDark.value;
            safeStorage.setItem('theme', isDark.value ? 'dark' : 'light');
            updateThemeClass();
            safeCreateIcons();
            // Update visualizer theme if active
            if (window.visualizer) {
                window.visualizer.setTheme(isDark.value);
            }
        };

        const updateThemeClass = () => {
            if (isDark.value) {
                document.documentElement.classList.add('dark');
                document.documentElement.setAttribute('data-theme', 'dark');
            } else {
                document.documentElement.classList.remove('dark');
                document.documentElement.setAttribute('data-theme', 'light');
            }
        };

        // Routing Visualizer Class
        const selectedNodeDetails = ref(null);
        const showingNodeDetails = ref(false);

        class RoutingVisualizer {
            constructor(containerId) {
                this.container = document.getElementById(containerId);
                if (!this.container) return;
                
                this.scene = new THREE.Scene();
                this.camera = new THREE.PerspectiveCamera(75, this.container.clientWidth / this.container.clientHeight, 0.1, 200000);
                this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
                this.renderer.setSize(this.container.clientWidth, this.container.clientHeight);
                this.renderer.setPixelRatio(window.devicePixelRatio);
                this.container.appendChild(this.renderer.domElement);

                this.controls = new THREE.OrbitControls(this.camera, this.renderer.domElement);
                this.controls.enableDamping = true;
                this.controls.dampingFactor = 0.05;
                this.camera.position.z = 15000;

                this.nodes = new Map();
                this.links = new Map();
                this.particles = [];
                this.labels = [];
                this.running = true;
                this.theme = document.documentElement.classList.contains('dark') ? 'dark' : 'light';
                
                // Optimization: Cache geometries and textures
                this.geometries = {
                    particle: new THREE.SphereGeometry(60, 8, 8)
                };
                this.textureCache = new Map();
                 this.materialCache = new Map();

                 this.raycaster = new THREE.Raycaster();
                this.mouse = new THREE.Vector2();
                this.hoveredNode = null;
                this.tooltip = this.createTooltip();

                this.init();
                this.animate();

                this.container.addEventListener('mousemove', (e) => this.onMouseMove(e));
                this.container.addEventListener('click', (e) => this.onClick(e));
                window.addEventListener('resize', () => this.onResize());
            }

            createTooltip() {
                const div = document.createElement('div');
                div.className = 'fixed pointer-events-none bg-black/80 backdrop-blur-md border border-white/20 text-white p-3 rounded-xl text-xs z-[100] hidden opacity-0 transition-opacity duration-200';
                div.style.fontFamily = 'JetBrains Mono, monospace';
                document.body.appendChild(div);
                return div;
            }

            onMouseMove(event) {
                const rect = this.container.getBoundingClientRect();
                this.mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
                this.mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
                
                this.updateTooltip(event.clientX, event.clientY);
            }

            updateTooltip(x, y) {
                if (this.hoveredNode) {
                    const data = this.hoveredNode.userData;
                    this.tooltip.innerHTML = `
                        <div class="flex flex-col gap-1">
                            <div class="font-bold text-yellow-400 uppercase tracking-tighter">${data.type}</div>
                            <div class="text-lg font-mono">${data.label}</div>
                            <div class="text-[10px] opacity-50 font-mono">${data.id}</div>
                        </div>
                    `;
                    this.tooltip.style.left = `${x + 20}px`;
                    this.tooltip.style.top = `${y + 20}px`;
                    this.tooltip.classList.remove('hidden');
                    setTimeout(() => this.tooltip.classList.add('opacity-100'), 10);
                } else {
                    this.tooltip.classList.remove('opacity-100');
                    setTimeout(() => this.tooltip.classList.add('hidden'), 200);
                }
            }

            onClick(event) {
                if (this.hoveredNode) {
                    this.pulseNode(this.hoveredNode);
                    
                    const nodeMesh = this.hoveredNode;
                    const nodeId = nodeMesh.userData.id;
                    const nodeType = nodeMesh.userData.type;
                    
                    let details = { id: nodeId, type: nodeType };
                    
                    if (nodeType === 'bot') {
                        const bot = bots.value.find(b => b.self_id == nodeId);
                        if (bot) {
                            details = { 
                                ...details, 
                                name: bot.nickname || bot.self_id,
                                status: bot.is_alive ? 'Online' : 'Offline',
                                last_seen: bot.last_seen_time,
                                messages: bot.msg_count,
                                ip: bot.remote_addr
                            };
                        }
                    } else if (nodeType === 'worker') {
                        const worker = workers.value.find(w => w.id == nodeId);
                        if (worker) {
                            details = {
                                ...details,
                                name: worker.id,
                                status: 'Active',
                                type: worker.type,
                                ip: worker.remote_addr
                            };
                        }
                    } else if (nodeType === 'group') {
                        const group = groups.value.find(g => (g.group_id || g.id) == nodeId);
                        if (group) {
                            details = {
                                ...details,
                                name: group.group_name || group.id,
                                status: 'Active Group',
                                member_count: group.member_count || 0,
                                bot_id: group.bot_id
                            };
                        }
                    } else if (nodeType === 'user') {
                        const friend = friends.value.find(f => (f.user_id || f.id) == nodeId);
                        if (friend) {
                            details = {
                                ...details,
                                name: friend.nickname || friend.id,
                                status: 'Active User',
                                platform: friend.platform || 'QQ',
                                bot_id: friend.bot_id
                            };
                        }
                    } else if (nodeType === 'nexus') {
                        details = {
                            ...details,
                            name: 'BotNexus Central',
                            status: 'Core System',
                            uptime: stats.value.uptime,
                            total_bots: stats.value.total_bots,
                            active_workers: stats.value.active_workers
                        };
                    }
                    
                    selectedNodeDetails.value = details;
                    showingNodeDetails.value = true;
                    
                    safeCreateIcons();

                    // Center camera on node
                    const targetPos = nodeMesh.position.clone();
                    new TWEEN.Tween(this.controls.target)
                        .to({ x: targetPos.x, y: targetPos.y, z: targetPos.z }, 500)
                        .easing(TWEEN.Easing.Quadratic.Out)
                        .start();
                }
            }

            pulseNode(node) {
                const originalScale = node.userData.id === 'nexus' ? 2200 : 1800;
                const targetScale = originalScale * 1.5;
                let start = null;
                const duration = 600;
                
                const step = (time) => {
                    if (!start) start = time;
                    const progress = (time - start) / duration;
                    if (progress < 1) {
                        const s = originalScale + (targetScale - originalScale) * Math.sin(progress * Math.PI);
                        node.scale.set(s, s, 1);
                        requestAnimationFrame(step);
                    } else {
                        node.scale.set(originalScale, originalScale, 1);
                    }
                };
                requestAnimationFrame(step);
            }

            onResize() {
                if (!this.container) return;
                this.camera.aspect = this.container.clientWidth / this.container.clientHeight;
                this.camera.updateProjectionMatrix();
                this.renderer.setSize(this.container.clientWidth, this.container.clientHeight);
            }

            init() {
                this.scene.background = null;
                const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
                this.scene.add(ambientLight);
                
                const starsGeometry = new THREE.BufferGeometry();
                const starsVertices = [];
                for (let i = 0; i < 5000; i++) {
                    starsVertices.push((Math.random() - 0.5) * 100000, (Math.random() - 0.5) * 100000, (Math.random() - 0.5) * 100000);
                }
                starsGeometry.setAttribute('position', new THREE.Float32BufferAttribute(starsVertices, 3));
                this.stars = new THREE.Points(starsGeometry, new THREE.PointsMaterial({ 
                    color: this.theme === 'dark' ? 0x3b82f6 : 0x33ccff, 
                    size: 4,
                    transparent: true,
                    opacity: 0.6
                }));
                this.scene.add(this.stars);

                this.getOrCreateNode('nexus', 'nexus', 'NEXUS');

                // Load existing bots and workers
                if (bots.value) {
                    bots.value.forEach(bot => {
                        this.getOrCreateNode(bot.self_id, 'bot', bot.nickname || bot.self_id);
                    });
                }
                if (workers.value) {
                    workers.value.forEach(worker => {
                        this.getOrCreateNode(worker.id, 'worker', worker.id);
                    });
                }
                if (groups.value) {
                    groups.value.forEach(group => {
                        this.getOrCreateNode(group.group_id || group.id, 'group', group.group_name || group.id);
                    });
                }
                if (friends.value) {
                    friends.value.forEach(friend => {
                        this.getOrCreateNode(friend.user_id || friend.id, 'user', friend.nickname || friend.id);
                    });
                }

                // Load cached links
                this.loadLinksFromCache();
            }

            saveLinksToCache() {
                const linksToSave = [];
                this.links.forEach((link, id) => {
                    linksToSave.push({
                        id,
                        source: { id: link.source.id, type: link.source.type, label: link.source.label },
                        target: { id: link.target.id, type: link.target.type, label: link.target.label }
                    });
                });
                localStorage.setItem('viz_links_cache', JSON.stringify(linksToSave.slice(-500))); // Limit to last 500 links
            }

            loadLinksFromCache() {
                try {
                    const cached = localStorage.getItem('viz_links_cache');
                    if (cached) {
                        const linksToLoad = JSON.parse(cached);
                        linksToLoad.forEach(linkData => {
                            const source = this.getOrCreateNode(linkData.source.id, linkData.source.type, linkData.source.label);
                            const target = this.getOrCreateNode(linkData.target.id, linkData.target.type, linkData.target.label);
                            if (source && target) {
                                this.createLink(source, target, false); // false = don't save during batch load
                            }
                        });
                    }
                } catch (e) {
                    console.error('Failed to load links from cache:', e);
                }
            }

            createLink(source, target, save = true, color = 0x3b82f6, opacity = 0.1) {
                const linkId = [source.id, target.id].sort().join('-');
                if (this.links.has(linkId)) return this.links.get(linkId);

                const material = new THREE.LineBasicMaterial({
                    color: color,
                    transparent: true,
                    opacity: opacity,
                    depthWrite: false
                });

                const geometry = new THREE.BufferGeometry().setFromPoints([
                    source.mesh.position,
                    target.mesh.position
                ]);

                const line = new THREE.Line(geometry, material);
                this.scene.add(line);
                const link = { line, source, target, baseOpacity: opacity };
                this.links.set(linkId, link);
                
                if (save) {
                    this.saveLinksToCache();
                }
                return link;
            }

            getAvatarUrl(id, type) {
                if (!id || id === 'nexus') return null;
                
                const numericId = parseInt(id);
                const isLargeId = !isNaN(numericId) && numericId > 980000000000;

                let url = '';
                if (type === 'group') {
                    url = `https://p.qlogo.cn/gh/${id}/${id}/100`;
                } else if (type === 'user') {
                    url = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
                } else if (type === 'bot') {
                    if (isLargeId) {
                        return 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';
                    }
                    url = `https://q1.qlogo.cn/g?b=qq&nk=${id}&s=100`;
                } else {
                    return null;
                }
                
                return `/api/proxy/avatar?url=${encodeURIComponent(url)}`;
            }

            getOrCreateNode(id, type, label) {
                if (!id) return null;
                if (this.nodes.has(id)) return this.nodes.get(id);
                
                const cacheKey = `${type}_${id}_${label}`;
                let texture = this.textureCache.get(cacheKey);

                const canvas = document.createElement('canvas');
                canvas.width = 512; canvas.height = 512;
                const ctx = canvas.getContext('2d');
                
                const drawNode = (img = null) => {
                    ctx.clearRect(0, 0, 512, 512);
                    
                    // Draw glow
                    const gradient = ctx.createRadialGradient(256, 256, 0, 256, 256, 240);
                    let color = '#3b82f6';
                    if (type === 'worker') color = '#33ccff';
                    if (type === 'bot') color = '#ff3366';
                    if (type === 'group') color = '#a855f7';
                    if (type === 'user') color = '#10b981';
                    if (type === 'nexus') color = '#facc15';
                    
                    gradient.addColorStop(0, color);
                    gradient.addColorStop(0.3, color + '66');
                    gradient.addColorStop(0.7, color + '22');
                    gradient.addColorStop(1, 'transparent');
                    
                    ctx.fillStyle = gradient;
                    ctx.beginPath(); ctx.arc(256, 256, 240, 0, Math.PI * 2); ctx.fill();
                    
                    if (img) {
                        // Draw avatar circle
                        ctx.save();
                        ctx.beginPath();
                        ctx.arc(256, 256, 180, 0, Math.PI * 2);
                        ctx.clip();
                        ctx.drawImage(img, 256 - 180, 256 - 180, 360, 360);
                        ctx.restore();
                    }
                    
                    // Draw outer ring
                    ctx.strokeStyle = color;
                    ctx.lineWidth = 15;
                    ctx.setLineDash([20, 10]);
                    ctx.beginPath(); ctx.arc(256, 256, 200, 0, Math.PI * 2); ctx.stroke();
                    
                    // Text (only if no image or for nexus)
                    if (!img || type === 'nexus') {
                        ctx.shadowColor = color;
                        ctx.shadowBlur = 30;
                        ctx.fillStyle = '#fff'; 
                        ctx.font = 'bold 90px JetBrains Mono'; // Much larger for distant visibility
                        ctx.textAlign = 'center';
                        ctx.fillText(label, 256, 285);
                    } else {
                        // Label below avatar
                        ctx.shadowColor = 'black';
                        ctx.shadowBlur = 15;
                        ctx.fillStyle = '#fff'; 
                        ctx.font = 'bold 70px JetBrains Mono'; // Significantly increased from 40px
                        ctx.textAlign = 'center';
                        
                        // Slightly move up to avoid being too close to the edge
                        const displayLabel = label.length > 15 ? label.substring(0, 13) + '..' : label;
                        ctx.fillText(displayLabel, 256, 470);
                    }
                    
                    if (texture) texture.needsUpdate = true;
                };

                drawNode(); // Draw placeholder/base first

                if (!texture) {
                    texture = new THREE.CanvasTexture(canvas);
                    this.textureCache.set(cacheKey, texture);
                    
                    // Load avatar if applicable
                    const avatarUrl = this.getAvatarUrl(id, type);
                    if (avatarUrl) {
                        const img = new Image();
                        img.crossOrigin = "Anonymous";
                        img.onload = () => drawNode(img);
                        img.src = avatarUrl;
                    }
                }

                const material = new THREE.SpriteMaterial({ map: texture, transparent: true });
                const sprite = new THREE.Sprite(material);
                
                let pos;
                if (id === 'nexus') {
                    pos = new THREE.Vector3(0, 0, 0);
                } else if (type === 'worker') {
                    const angle = Math.random() * Math.PI * 2;
                    const dist = 3000 + Math.random() * 1500;
                    pos = new THREE.Vector3(Math.cos(angle) * dist, Math.sin(angle) * dist, (Math.random() - 0.5) * 1000);
                } else if (type === 'bot') {
                    const angle = Math.random() * Math.PI * 2;
                    const dist = 6000 + Math.random() * 2500;
                    pos = new THREE.Vector3(Math.cos(angle) * dist, Math.sin(angle) * dist, (Math.random() - 0.5) * 2000);
                } else if (type === 'group') {
                    const angle = Math.random() * Math.PI * 2;
                    const dist = 10000 + Math.random() * 3000;
                    pos = new THREE.Vector3(Math.cos(angle) * dist, Math.sin(angle) * dist, (Math.random() - 0.5) * 3000);
                } else { // user
                    const angle = Math.random() * Math.PI * 2;
                    const dist = 14000 + Math.random() * 4000;
                    pos = new THREE.Vector3(Math.cos(angle) * dist, Math.sin(angle) * dist, (Math.random() - 0.5) * 4000);
                }
                
                sprite.position.copy(pos);
                sprite.scale.set(1800, 1800, 1); // Increased from 1000 for better distant visibility
                sprite.userData = { id, type, label };
                
                this.scene.add(sprite);
                const node = { id, type, label, mesh: sprite, targetPos: pos, pulse: 0 };
                this.nodes.set(id, node);
                return node;
            }

            handleRoutingEvent(event) {
                const source = this.getOrCreateNode(event.source || 'nexus', event.source_type || 'bot', event.source_label || event.source || 'BOT');
                const target = this.getOrCreateNode(event.target || 'nexus', event.target_type || 'worker', event.target_label || event.target || 'WORKER');
                
                if (source && target) {
                    // Determine link color and opacity based on types
                    let linkColor = 0x3b82f6; // Default blue
                    let linkOpacity = 0.1;

                    if (event.source_type === 'group' || event.target_type === 'group') {
                        linkColor = 0x60a5fa; // Lighter blue for group links
                        linkOpacity = 0.15;
                    }

                    // Create persistent link if it doesn't exist
                    const mainLink = this.createLink(source, target, true, linkColor, linkOpacity);
                    if (mainLink) {
                        mainLink.line.material.opacity = 0.8; // Flash on transmission
                    }

                    this.createParticle(source, target, event.color || (event.msg_type === 'request' ? '#ff3366' : '#33ccff'));
                    source.pulse = 1.0;
                    target.pulse = 0.5;
                    
                    if (event.content) {
                        this.createFloatingLabel(event.content, source.mesh.position);
                    }
                    
                    if (event.total_messages) {
                        stats.value.total_msgs = event.total_messages;
                    }
                }
            }

            handleSyncState(state) {
                if (state.bots) {
                    state.bots.forEach(bot => {
                        this.getOrCreateNode(bot.self_id, 'bot', bot.nickname || bot.self_id);
                    });
                }
                if (state.workers) {
                    state.workers.forEach(worker => {
                        if (worker.status === 'Online' || worker.status === 'Active') {
                            this.getOrCreateNode(worker.id, 'worker', worker.id);
                        }
                    });
                }
                if (state.groups) {
                    state.groups.forEach(group => {
                        this.getOrCreateNode(group.group_id || group.id, 'group', group.group_name || group.id);
                    });
                }
                if (state.friends) {
                    state.friends.forEach(friend => {
                        this.getOrCreateNode(friend.user_id || friend.id, 'user', friend.nickname || friend.id);
                    });
                }
                if (state.nodes) {
                    state.nodes.forEach(n => this.getOrCreateNode(n.id, n.type, n.label));
                }
                
                // Cleanup removed nodes
                const activeIds = new Set(['nexus', 
                    ...(state.bots?.map(b => b.self_id) || []), 
                    ...(state.workers?.filter(w => w.status === 'Online' || w.status === 'Active').map(w => w.id) || []),
                    ...(state.groups?.map(g => g.group_id || g.id) || []),
                    ...(state.friends?.map(f => f.user_id || f.id) || []),
                    ...(state.nodes?.map(n => n.id) || []),
                    ...(groups.value?.map(g => g.group_id || g.id) || []),
                    ...(friends.value?.map(f => f.user_id || f.id) || [])
                ]);

                this.nodes.forEach((node, id) => {
                    if (!activeIds.has(id)) {
                        this.scene.remove(node.mesh);
                        this.nodes.delete(id);
                        
                        // Cleanup associated links
                        this.links.forEach((link, linkId) => {
                            if (link.source.id === id || link.target.id === id) {
                                this.scene.remove(link.line);
                                this.links.delete(linkId);
                            }
                        });

                        if (selectedNodeDetails.value && selectedNodeDetails.value.id === id) {
                            showingNodeDetails.value = false;
                            selectedNodeDetails.value = null;
                        }
                    }
                });

                // Update cache after cleanup
                this.saveLinksToCache();
            }

            createFloatingLabel(text, position) {
                const canvas = document.createElement('canvas');
                canvas.width = 512; canvas.height = 128;
                const ctx = canvas.getContext('2d');
                
                ctx.fillStyle = 'rgba(0,0,0,0.6)';
                ctx.roundRect(0, 0, 512, 80, 20);
                ctx.fill();
                
                ctx.strokeStyle = '#3b82f6';
                ctx.lineWidth = 2;
                ctx.stroke();

                ctx.fillStyle = '#fff';
                ctx.font = '32px JetBrains Mono';
                ctx.textAlign = 'center';
                const displayMsg = text.length > 25 ? text.substring(0, 22) + '...' : text;
                ctx.fillText(displayMsg, 256, 50);

                const texture = new THREE.CanvasTexture(canvas);
                const material = new THREE.SpriteMaterial({ map: texture, transparent: true, opacity: 0 });
                const sprite = new THREE.Sprite(material);
                sprite.position.copy(position);
                sprite.position.y += 400;
                sprite.scale.set(1000, 250, 1);
                
                this.scene.add(sprite);
                this.labels.push({
                    mesh: sprite,
                    life: 1.0,
                    velocity: new THREE.Vector3(0, 2, 0)
                });
            }

            createParticle(start, end, color = 0x3b82f6) {
                let mat = this.materialCache.get(color);
                if (!mat) {
                    mat = new THREE.MeshBasicMaterial({ color: color, transparent: true, opacity: 0.9 });
                    this.materialCache.set(color, mat);
                }
                
                const mesh = new THREE.Mesh(this.geometries.particle, mat);
                mesh.position.copy(start.mesh.position);
                
                // Only add light for important/rare events to save performance
                if (this.particles.length < 50) {
                    const light = new THREE.PointLight(color, 1, 500);
                    mesh.add(light);
                }
                
                this.scene.add(mesh);
                
                this.particles.push({
                    mesh, 
                    start: start.mesh.position.clone(), 
                    end: end.mesh.position.clone(), 
                    progress: 0,
                    speed: 0.005 + Math.random() * 0.01
                });
            }

            animate() {
                if (!this.running) return;
                requestAnimationFrame(() => this.animate());
                
                const time = Date.now() * 0.001;
                
                // Raycasting for hover effects
                this.raycaster.setFromCamera(this.mouse, this.camera);
                const nodeSprites = Array.from(this.nodes.values()).map(n => n.mesh);
                const intersects = this.raycaster.intersectObjects(nodeSprites);

                if (intersects.length > 0) {
                    if (this.hoveredNode !== intersects[0].object) {
                        this.hoveredNode = intersects[0].object;
                    }
                } else {
                    this.hoveredNode = null;
                }

                // Update Particles
                for (let i = this.particles.length - 1; i >= 0; i--) {
                    const p = this.particles[i];
                    p.progress += p.speed;
                    p.mesh.position.lerpVectors(p.start, p.end, p.progress);
                    
                    // Arc trajectory
                    const arcHeight = 500;
                    p.mesh.position.y += Math.sin(p.progress * Math.PI) * arcHeight;
                    
                    if (p.progress >= 1) {
                        this.scene.remove(p.mesh);
                        this.particles.splice(i, 1);
                    }
                }

                // Update Labels
                for (let i = this.labels.length - 1; i >= 0; i--) {
                    const l = this.labels[i];
                    l.life -= 0.005;
                    l.mesh.position.add(l.velocity);
                    l.mesh.material.opacity = Math.min(l.life * 2, 1);
                    
                    if (l.life <= 0) {
                        this.scene.remove(l.mesh);
                        this.labels.splice(i, 1);
                    }
                }

                // Update Nodes
                this.nodes.forEach(node => {
                    const baseScale = node.id === 'nexus' ? 2200 : 1800; // Match new larger scales
                    const isHovered = this.hoveredNode === node.mesh;
                    const hoverScale = isHovered ? 1.2 : 1.0;
                    const pulseScale = (1 + Math.sin(time * 2) * 0.05 + (node.pulse || 0) * 0.3) * hoverScale;
                    node.mesh.scale.set(baseScale * pulseScale, baseScale * pulseScale, 1);
                    
                    if (node.pulse > 0) node.pulse -= 0.02;
                    
                    if (node.id !== 'nexus') {
                        node.mesh.position.y += Math.sin(time + node.mesh.position.x) * 1.5;
                        // node.mesh.material.rotation += 0.002; // Removed to keep avatar orientation fixed
                    }

                    if (isHovered) {
                        node.mesh.userData = { id: node.id, type: node.type, label: node.label };
                    }
                });

                // Update Links (follow floating nodes and handle dynamic opacity)
                this.links.forEach(link => {
                    const positions = link.line.geometry.attributes.position.array;
                    positions[0] = link.source.mesh.position.x;
                    positions[1] = link.source.mesh.position.y;
                    positions[2] = link.source.mesh.position.z;
                    positions[3] = link.target.mesh.position.x;
                    positions[4] = link.target.mesh.position.y;
                    positions[5] = link.target.mesh.position.z;
                    link.line.geometry.attributes.position.needsUpdate = true;

                    // Opacity decay
                    const baseOpacity = link.baseOpacity || 0.1;
                    if (link.line.material.opacity > baseOpacity) {
                        link.line.material.opacity -= 0.005; // Fade out slowly
                        if (link.line.material.opacity < baseOpacity) {
                            link.line.material.opacity = baseOpacity;
                        }
                    }
                });

                if (this.stars) {
                    this.stars.rotation.y += 0.0003;
                    this.stars.rotation.z += 0.0001;
                }
                
                this.controls.update();
                this.renderer.render(this.scene, this.camera);
            }

            setTheme(isDark) {
                this.theme = isDark ? 'dark' : 'light';
                if (this.stars) {
                    this.stars.material.color.set(isDark ? 0x3b82f6 : 0x33ccff);
                }
            }

            resize() {
                if (!this.container) return;
                this.camera.aspect = this.container.clientWidth / this.container.clientHeight;
                this.camera.updateProjectionMatrix();
                this.renderer.setSize(this.container.clientWidth, this.container.clientHeight);
            }
        }

        const currentTime = ref('');
        const loading = ref(false);
        const searchQuery = ref('');
        const filterTab = ref('all');
        const activeTab = ref(localStorage.getItem('activeTab') || 'dashboard');
        const showMobileMenu = ref(window.innerWidth >= 1024);
        const isSidebarCollapsed = ref(localStorage.getItem('sidebarCollapsed') === 'true');
        const bots = ref([]);
        const workers = ref([]);
        const groups = ref([]);
        const friends = ref([]);
        const groupMembers = ref([]);
        const currentGroup = ref(null);
        const systemLogs = ref([]);
        const logFilter = ref('');
        const filteredLogs = computed(() => {
            if (!logFilter.value || logFilter.value === 'all') return systemLogs.value;
            const filter = logFilter.value.toLowerCase();
            return systemLogs.value.filter(log => {
                const text = (log.msg || log.message || JSON.stringify(log)).toLowerCase();
                const level = (log.level || log.type || 'info').toLowerCase();
                return text.includes(filter) || level.includes(filter);
            });
        });

        const downloadLogs = () => {
            if (systemLogs.value.length === 0) return;
            const logText = systemLogs.value.map(log => `[${log.time || 'SYSTEM'}] ${log.level || log.type || 'INFO'}: ${log.msg || log.message || JSON.stringify(log)}`).join('\n');
            const blob = new Blob([logText], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `system_logs_${new Date().getTime()}.log`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        };

        const memberSearchQuery = ref('');
        const memberSortBy = ref('role');
        const memberSortAsc = ref(false);
        const routingRules = ref([]);
        const dockerContainers = ref([]);
        const dockerLogs = ref('');
        const dockerLogFilter = ref('');
        const showingDockerLogs = ref(false);
        const currentDockerContainer = ref(null);
        const loadingContainers = ref(new Set());

        const filteredDockerLogs = computed(() => {
            if (!dockerLogFilter.value) return dockerLogs.value;
            const filter = dockerLogFilter.value.toLowerCase();
            return dockerLogs.value.split('\n')
                .filter(line => line.toLowerCase().includes(filter))
                .join('\n');
        });

        const downloadDockerLogs = () => {
            if (!dockerLogs.value || !currentDockerContainer.value) return;
            const blob = new Blob([dockerLogs.value], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `docker_${currentDockerContainer.value.name}_logs_${new Date().getTime()}.log`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        };
        const systemUsers = ref([]);
        const showingUserModal = ref(false);
        const userModalData = ref({
            action: 'create', // 'create' or 'edit'
            username: '',
            password: '',
            is_admin: false
        });
        const showAddBotModal = ref(false);
        const newBotData = ref({
            self_id: '',
            nickname: '',
            platform: 'QQ'
        });
        const debugResponse = ref('');
        const showingMassSend = ref(false);
        const massSendType = ref('group'); // 'group' or 'friend'
        const massSendTargets = ref([]);
        const massSendMessage = ref('');
        const massSendSelectedTargets = ref([]);
        const massSendStatus = ref({ current: 0, total: 0, running: false });
        const backendConfig = ref({
            ws_port: ':3001',
            webui_port: ':5000',
            redis_addr: 'localhost:6379',
            redis_pwd: '',
            jwt_secret: '',
            default_admin_password: '',
            stats_file: 'stats.json',
            pg_host: 'localhost',
            pg_port: 5432,
            pg_user: 'postgres',
            pg_password: '',
            pg_dbname: 'botmatrix',
            pg_sslmode: 'disable',
            enable_skill: true,
            log_level: 'INFO',
            auto_reply: false
        });
        const userInfo = ref({});
        const selectedBotId = ref('');
        const stats = ref({
            total_msgs: 0,
            total_bots: 0,
            active_workers: 0,
            uptime: '0d 0h 0m',
            cpu_usage: '0%',
            memory_usage: '0%',
            memory_used_mb: 0,
            msg_per_sec: 0,
            sent_per_sec: 0,
            goroutines: 0,
            disk_usage: '0%',
            os_platform: '',
            os_arch: '',
            cpu_model: ''
        });

        const menuGroups = computed(() => [
            {
                title: 'overview',
                items: [
                    { id: 'dashboard', icon: 'layout-dashboard' },
                    { id: 'bots', icon: 'bot' },
                    { id: 'monitor', icon: 'activity' },
                    { id: 'visualization', icon: 'zap' }
                ]
            },
            {
                title: 'management',
                items: [
                    { id: 'groups', icon: 'users' },
                    { id: 'friends', icon: 'user-plus' },
                    { id: 'system_logs', icon: 'file-text' }
                ]
            },
            {
                title: 'system',
                items: [
                    { id: 'docker', icon: 'container' },
                    { id: 'users', icon: 'shield-check' },
                    { id: 'settings', icon: 'settings' }
                ]
            }
        ]);

        const statsCards = computed(() => [
            { label: 'total_bots', value: stats.value.total_bots, icon: 'bot', colorClass: 'bg-blue-500', textColor: 'text-blue-500' },
            { label: 'active_workers', value: stats.value.active_workers, icon: 'cpu', colorClass: 'bg-purple-500', textColor: 'text-purple-500' },
            { label: 'messages_today', value: stats.value.total_msgs, icon: 'message-square', colorClass: 'bg-green-500', textColor: 'text-green-500' },
            { label: 'current_time', value: currentTime.value, icon: 'clock', colorClass: 'bg-matrix', textColor: 'text-matrix' }
        ]);

        const uptimeDisplay = computed(() => {
            const parts = stats.value.uptime.split(' ');
            if (parts.length >= 2) {
                return { value: parts[0] + parts[1], unit: parts.slice(2).join(' ') || 'UPTIME' };
            }
            return { value: stats.value.uptime, unit: 'UPTIME' };
        });

        const recentLogs = computed(() => {
            return systemLogs.value.filter(log => log).slice(0, 10).map(log => ({
                time: log.time || new Date().toISOString(),
                message: log.msg || log.message || JSON.stringify(log)
            }));
        });

        const singleMsgModal = ref({
            show: false,
            target: null,
            type: 'private', // private, group
            content: '',
            sending: false
        });

        // WebSocket State
        const wsStatus = ref('offline'); // offline, connecting, online, error
        const wsConnected = ref(false);
        let wsSubscriber = null;
        let wsReconnectAttempts = 0;
        const MAX_WS_RECONNECT_ATTEMPTS = 5;
        const WS_RECONNECT_DELAY = 5000;

        const navItems = computed(() => [
            { id: 'overview', label: t('overview'), icon: 'layout-dashboard' },
            { id: 'bots', label: t('bot_matrix'), icon: 'bot' },
            { id: 'monitor', label: t('monitor_events'), icon: 'activity' },
            { id: 'visualization', label: t('visualization'), icon: 'zap' },
            { id: 'groups', label: t('groups'), icon: 'users' },
            { id: 'friends', label: t('friends'), icon: 'user-plus' },
            { id: 'system_logs', label: t('system_logs'), icon: 'file-text' },
            { id: 'docker', label: t('docker_mgmt'), icon: 'container' },
            { id: 'users', label: t('user_mgmt'), icon: 'shield-check' },
            { id: 'settings', label: t('settings'), icon: 'settings' }
        ]);

        const activeNavItem = computed(() => navItems.value.find(item => item.id === activeTab.value));

        const pulseBars = ref(Array.from({ length: 48 }, () => Math.floor(Math.random() * 60) + 20));

        // Core Stats mapping
        const coreStats = computed(() => [
            { label: t('total_bots'), value: stats.value.total_bots, icon: 'bot', color: 'blue', trend: '+5%' },
            { label: t('active_workers'), value: stats.value.active_workers, icon: 'cpu', color: 'purple', trend: 'STABLE' },
            { label: t('messages_today'), value: stats.value.total_msgs, icon: 'message-square', color: 'green', trend: '+18%' },
            { label: t('system_uptime'), value: stats.value.uptime, icon: 'clock', color: 'orange', trend: 'UP' }
        ]);

        const filteredBots = computed(() => {
            let result = bots.value;
            
            if (filterTab.value === 'active') {
                result = result.filter(b => b.is_alive);
            } else if (filterTab.value === 'offline') {
                result = result.filter(b => !b.is_alive);
            }

            if (searchQuery.value) {
                const query = searchQuery.value.toLowerCase();
                result = result.filter(b => 
                    (b.nickname && b.nickname.toLowerCase().includes(query)) || 
                    b.self_id.toString().includes(query)
                );
            }

            return result;
        });

        // WebSocket Integration
        const initWebSocket = () => {
            if (wsSubscriber && wsSubscriber.readyState === WebSocket.CONNECTING) return;
            
            if (wsReconnectAttempts >= MAX_WS_RECONNECT_ATTEMPTS) {
                console.error('WebSocket reconnection limit reached');
                wsStatus.value = 'error';
                return;
            }

            const token = localStorage.getItem('wxbot_token');
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsPort = window.location.port ? `:${window.location.port}` : '';
            let wsUrl = `${protocol}//${window.location.hostname}${wsPort}/ws/subscriber?role=subscriber`;
            if (token) wsUrl += `&token=${encodeURIComponent(token)}`;

            try {
                wsStatus.value = 'connecting';
                wsSubscriber = new WebSocket(wsUrl);

                wsSubscriber.onopen = () => {
                    wsReconnectAttempts = 0;
                    wsStatus.value = 'online';
                    wsConnected.value = true;
                    console.log('WebSocket connected');
                };

                wsSubscriber.onclose = () => {
                    wsConnected.value = false;
                    wsStatus.value = 'offline';
                    wsSubscriber = null;
                    wsReconnectAttempts++;
                    setTimeout(initWebSocket, WS_RECONNECT_DELAY);
                };

                wsSubscriber.onerror = (err) => {
                    console.error('WebSocket error:', err);
                    wsStatus.value = 'error';
                };

                wsSubscriber.onmessage = (evt) => {
                    try {
                        const data = JSON.parse(evt.data);
                        
                        // Pass to global event logger if available
                        if (window.addEventLog) {
                            window.addEventLog(data);
                        }

                        // Real-time log updates
                        if (data.post_type === 'log' && data.data) {
                            const logEntry = {
                                ...data.data,
                                type: (data.data.level || data.data.type || 'info').toLowerCase()
                            };
                            systemLogs.value.unshift(logEntry);
                            if (systemLogs.value.length > 100) systemLogs.value.pop();
                        }
                        // Handle visualization events
                        if (data.type === 'routing_event' || data.type === 'sync_state') {
                            if (window.visualizer) {
                                if (data.type === 'routing_event') {
                                    window.visualizer.handleRoutingEvent(data);
                                    if (data.total_messages) stats.value.total_msgs = data.total_messages;
                                } else {
                                    window.visualizer.handleSyncState(data);
                                }
                            }
                        }
                        // Handle docker events
                        if (data.type === 'docker_event') {
                            if (activeTab.value === 'docker') {
                                fetchDockerContainers();
                            }
                            if (showingDockerLogs.value && currentDockerContainer.value && currentDockerContainer.value.id === data.container_id) {
                                if (data.status === 'deleted') {
                                    showingDockerLogs.value = false;
                                    alert(t('container_deleted') || 'Container was deleted');
                                } else {
                                    fetchDockerLogs(currentDockerContainer.value);
                                }
                            }
                        }

                        // Handle worker updates
                        if (data.type === 'worker_update' && data.data) {
                            const updatedWorker = data.data;
                            const index = workers.value.findIndex(w => w.id === updatedWorker.id);
                            if (index !== -1) {
                                // Update existing worker
                                workers.value[index] = { ...workers.value[index], ...updatedWorker };
                            } else {
                                // Add new worker if not exists
                                workers.value.push(updatedWorker);
                            }
                            // Also update visualization if it exists
                            if (window.visualizer) {
                                window.visualizer.getOrCreateNode(updatedWorker.id, 'worker', updatedWorker.id);
                            }
                        }

                        // Refresh bots on lifecycle events
                        if (data.post_type === 'meta_event' && data.meta_event_type === 'lifecycle') {
                            fetchAllData();
                        }
                    } catch (e) {
                        console.error('WS parse error:', e);
                    }
                };
            } catch (e) {
                console.error('WS init failed:', e);
                wsStatus.value = 'error';
            }
        };

        // Global API fetch helper
        const apiFetch = async (url, options = {}) => {
            const token = localStorage.getItem('wxbot_token');
            if (!token) {
                isLoggedIn.value = false;
                return null;
            }

            const defaultOptions = {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            };

            try {
                const res = await fetch(url, { ...defaultOptions, ...options });
                if (res.status === 401) {
                    localStorage.removeItem('wxbot_token');
                    isLoggedIn.value = false;
                    return null;
                }
                if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
                return await res.json();
            } catch (err) {
                console.error(`API Fetch error (${url}):`, err);
                return null;
            }
        };

        const fetchFriends = async (botId, refresh = false) => {
            loading.value = true;
            try {
                const url = botId ? `/api/contacts?bot_id=${botId}${refresh ? '&refresh=true' : ''}` : '/api/contacts';
                const data = await apiFetch(url);
                if (data) {
                    const friendList = Array.isArray(data) ? data : (data.friends || []);
                    friends.value = friendList.filter(item => item.type === 'friend' || item.type === 'contact' || item.type === 'private' || item.user_id);
                    console.log(`Fetched ${friends.value.length} friends`);
                }
            } finally {
                loading.value = false;
            }
        };

        const fetchRoutingRules = async () => {
            const data = await apiFetch('/api/admin/routing');
            if (data) {
                const rulesMap = data.rules || {};
                routingRules.value = Object.entries(rulesMap).map(([pattern, target]) => ({ pattern, target }));
            }
        };

        const saveRoutingRule = async (rule) => {
            const token = localStorage.getItem('wxbot_token');
            try {
                const res = await fetch('/api/admin/routing', {
                    method: 'POST',
                    headers: { 
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(rule)
                });
                if (res.ok) {
                    fetchRoutingRules();
                    return true;
                }
            } catch (err) {
                console.error('Failed to save routing rule:', err);
            }
            return false;
        };

        const deleteRoutingRule = async (ruleName) => {
            const token = localStorage.getItem('wxbot_token');
            try {
                const res = await fetch(`/api/admin/routing?name=${encodeURIComponent(ruleName)}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (res.ok) {
                    fetchRoutingRules();
                    return true;
                }
            } catch (err) {
                console.error('Failed to delete routing rule:', err);
            }
            return false;
        };

        const fetchLogs = async () => {
            loading.value = true;
            try {
                const data = await apiFetch(`/api/admin/logs?level=${logFilter.value}`);
                if (data) {
                    const logList = data.logs || (data.data && data.data.logs) || [];
                    systemLogs.value = logList.filter(log => log).map(log => ({
                        ...log,
                        type: (log.level || log.type || 'info').toLowerCase()
                    }));
                }
            } finally {
                loading.value = false;
            }
        };

        const clearLogs = async () => {
            if (!confirm('Clear all system logs?')) return;
            loading.value = true;
            try {
                const data = await apiFetch('/api/admin/logs/clear', { method: 'POST' });
                if (data && data.success) {
                    systemLogs.value = [];
                }
            } finally {
                loading.value = false;
            }
        };

        const fetchDockerContainers = async () => {
            const data = await apiFetch('/api/docker/list');
            if (data) {
                dockerContainers.value = data.containers || [];
            }
        };

        const manageDockerContainer = async (id, action) => {
            if (action === 'delete') {
                if (!confirm(t('confirm_delete_container') || 'Are you sure you want to delete this container?')) {
                    return;
                }
            }
            loadingContainers.value.add(id);
            try {
                const data = await apiFetch('/api/admin/docker/action', { 
                    method: 'POST',
                    body: JSON.stringify({ container_id: id, action: action })
                });
                if (data && (data.success || data.status === 'ok' || data.status === 'success')) {
                    // Success is usually handled by WebSocket broadcast, but we refresh anyway
                    fetchDockerContainers();
                } else {
                    alert(data?.message || t('action_failed') || 'Action failed');
                }
            } catch (err) {
                alert(t('action_failed') || 'Action failed');
            } finally {
                loadingContainers.value.delete(id);
            }
        };

        const addDockerBot = async () => {
            const confirmed = confirm(t('confirm_add_bot') || 'Deploy a new Bot container?');
            if (!confirmed) return;
            
            const data = await apiFetch('/api/admin/docker/add-bot', { method: 'POST' });
            if (data && (data.status === 'ok' || data.status === 'success')) {
                alert(data.message || 'Bot deployed successfully');
                fetchDockerContainers();
            } else {
                alert(data?.message || 'Failed to deploy Bot');
            }
        };

        const addDockerWorker = async () => {
            const confirmed = confirm(t('confirm_add_worker') || 'Deploy a new Worker container?');
            if (!confirmed) return;

            const data = await apiFetch('/api/admin/docker/add-worker', { method: 'POST' });
            if (data && (data.status === 'ok' || data.status === 'success')) {
                alert(data.message || 'Worker deployed successfully');
                fetchDockerContainers();
            } else {
                alert(data?.message || 'Failed to deploy Worker');
            }
        };

        const fetchDockerLogs = async (container) => {
            currentDockerContainer.value = container;
            dockerLogs.value = 'Loading logs...';
            showingDockerLogs.value = true;
            
            const data = await apiFetch(`/api/admin/docker/logs?id=${container.id}`);
            if (data && data.logs) {
                dockerLogs.value = data.logs;
            } else {
                dockerLogs.value = 'No logs found or error fetching logs.';
            }
        };

        const fetchSystemUsers = async () => {
            loading.value = true;
            try {
                const data = await apiFetch('/api/admin/users');
                if (data) {
                    systemUsers.value = data.users || [];
                }
            } finally {
                loading.value = false;
            }
        };

        const manageSystemUser = async (user, action) => {
            if (action === 'create') {
                userModalData.value = {
                    action: 'create',
                    username: '',
                    password: '',
                    is_admin: false
                };
                showingUserModal.value = true;
                return;
            } else if (action === 'delete') {
                if (!confirm(t('confirm_delete_user') || `Are you sure you want to delete user ${user.username}?`)) return;
                const data = await apiFetch('/api/admin/users', {
                    method: 'POST',
                    body: JSON.stringify({ action: 'delete', username: user.username })
                });
                if (data && data.success) {
                    alert(t('action_success') || 'Operation successful');
                    fetchSystemUsers();
                }
                return;
            } else if (action === 'reset_password') {
                const newPassword = prompt(t('prompt_new_password') || 'Enter new password:');
                if (!newPassword) return;
                const data = await apiFetch('/api/admin/users', {
                    method: 'POST',
                    body: JSON.stringify({ action: 'reset_password', username: user.username, password: newPassword })
                });
                if (data && data.success) {
                    alert(t('action_success') || 'Operation successful');
                }
                return;
            } else if (action === 'toggle_active') {
                const data = await apiFetch('/api/admin/users', {
                    method: 'POST',
                    body: JSON.stringify({ action: 'toggle_active', username: user.username })
                });
                if (data && data.success) {
                    fetchSystemUsers();
                }
                return;
            }
        };

        const submitUserModal = async () => {
            if (!userModalData.value.username) {
                alert(t('error_username_required') || 'Username is required');
                return;
            }
            if (userModalData.value.action === 'create' && !userModalData.value.password) {
                alert(t('error_password_required') || 'Password is required');
                return;
            }
            if (userModalData.value.action === 'create' && userModalData.value.password.length < 6) {
                alert(t('error_password_too_short') || 'Password must be at least 6 characters');
                return;
            }

            const data = await apiFetch('/api/admin/users', {
                method: 'POST',
                body: JSON.stringify(userModalData.value)
            });

            if (data && data.success) {
                alert(t('action_success') || 'Operation successful');
                showingUserModal.value = false;
                fetchSystemUsers();
            } else {
                alert(data?.message || t('action_failed') || 'Operation failed');
            }
        };

        const openMassSend = (type) => {
            massSendType.value = type;
            massSendTargets.value = type === 'group' ? groups.value : friends.value;
            massSendSelectedTargets.value = [];
            massSendMessage.value = '';
            showingMassSend.value = true;
            massSendStatus.value = { current: 0, total: 0, running: false };
        };

        const toggleMassSendTarget = (id) => {
            const index = massSendSelectedTargets.value.indexOf(id);
            if (index > -1) {
                massSendSelectedTargets.value.splice(index, 1);
            } else {
                massSendSelectedTargets.value.push(id);
            }
        };

        const selectAllMassSendTargets = () => {
            if (massSendSelectedTargets.value.length === massSendTargets.value.length && massSendTargets.value.length > 0) {
                massSendSelectedTargets.value = [];
            } else {
                massSendSelectedTargets.value = massSendTargets.value.map(t => t.group_id || t.guild_id || t.user_id);
            }
        };

        const startMassSend = async () => {
            if (!massSendMessage.value || massSendSelectedTargets.value.length === 0) return;
            if (!selectedBotId.value) {
                alert(t('select_bot_hint'));
                return;
            }
            
            massSendStatus.value = { 
                current: 0, 
                total: massSendSelectedTargets.value.length, 
                running: true 
            };

            for (const targetId of massSendSelectedTargets.value) {
                if (!massSendStatus.value.running) break;

                try {
                    const params = massSendType.value === 'group' 
                        ? { group_id: targetId, message: massSendMessage.value }
                        : { user_id: targetId, message: massSendMessage.value };

                    await apiFetch('/api/bot/action', {
                        method: 'POST',
                        body: JSON.stringify({
                            bot_id: selectedBotId.value,
                            action: massSendType.value === 'group' ? 'send_group_msg' : 'send_private_msg',
                            params
                        })
                    });
                } catch (e) {
                    console.error(`Failed to send mass message to ${targetId}:`, e);
                }

                massSendStatus.value.current++;
                // Add a small delay between messages to avoid rate limiting
                await new Promise(r => setTimeout(r, 600));
            }

            if (massSendStatus.value.running) {
                alert(`${t('action_success')}: Sent ${massSendStatus.value.current}/${massSendStatus.value.total}`);
            }
            massSendStatus.value.running = false;
            showingMassSend.value = false;
        };

        const fetchBackendConfig = async () => {
            const data = await apiFetch('/api/admin/config');
            if (data && data.config) {
                backendConfig.value = { ...backendConfig.value, ...data.config };
            }
        };

        const saveBackendConfig = async () => {
            console.log('SAVE BUTTON CLICKED - Starting save process'); 
            try {
                console.log('Current backendConfig state:', JSON.stringify(backendConfig.value));
                
                // Ensure numeric types for fields that require them in backend
                if (backendConfig.value.pg_port !== undefined && backendConfig.value.pg_port !== '') {
                    const oldPort = backendConfig.value.pg_port;
                    backendConfig.value.pg_port = parseInt(backendConfig.value.pg_port) || 0;
                    console.log(`Converted pg_port from ${oldPort} to ${backendConfig.value.pg_port}`);
                }
                
                const payload = JSON.stringify(backendConfig.value);
                console.log('Sending payload to /api/admin/config:', payload);
                
                const data = await apiFetch('/api/admin/config', {
                    method: 'POST',
                    body: payload
                });
                
                console.log('API Response received:', data);
                
                if (data && (data.success || data.status === 'ok')) {
                    alert(t('action_success') || 'Configuration saved successfully');
                } else {
                    const errorMsg = data && data.message ? data.message : (data ? 'Action failed' : 'Network or server error');
                    alert((t('action_failed') || 'Failed to save configuration') + ': ' + errorMsg);
                }
            } catch (err) {
                console.error('Save config fatal error:', err);
                alert((t('error') || 'Error') + ': ' + err.message);
            }
        };

        const updatePassword = async (oldPwd, newPwd) => {
            await apiFetch('/api/user/password', {
                method: 'POST',
                body: JSON.stringify({ old: oldPwd, new: newPwd })
            });
        };

        const manageBot = async (botId, action) => {
            const data = await apiFetch(`/api/bot/manage?bot_id=${botId}&action=${action}`, { method: 'POST' });
            if (data && data.success) {
                fetchAllData();
            }
        };

        const fetchGroups = async (botId, refresh = false) => {
            loading.value = true;
            try {
                const url = botId ? `/api/contacts?bot_id=${botId}${refresh ? '&refresh=true' : ''}` : '/api/contacts';
                const data = await apiFetch(url);
                if (data) {
                    const groupList = Array.isArray(data) ? data : (data.groups || []);
                    groups.value = groupList.filter(item => item.type === 'group' || item.type === 'guild' || item.group_id);
                    console.log(`Fetched ${groups.value.length} groups`);
                }
            } finally {
                loading.value = false;
            }
        };

        const fetchGroupMembers = async (groupId) => {
            if (!selectedBotId.value || !groupId) return;
            
            // Try different actions for different platforms
            let action = 'get_group_member_list';
            const group = groups.value.find(g => (g.group_id || g.guild_id) === groupId);
            if (group && group.type === 'guild') {
                action = 'get_guild_member_list';
            }

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: action,
                    params: { group_id: groupId, guild_id: groupId }
                })
            });

            if (data && data.success) {
                let members = [];
                const res = data.data || {};
                if (Array.isArray(res)) members = res;
                else if (Array.isArray(res.members)) members = res.members;
                else if (Array.isArray(res.list)) members = res.list;
                
                groupMembers.value = members.map(m => ({
                    ...m,
                    user_id: m.user_id || m.id || m.uid,
                    nickname: m.nickname || m.name || 'Unknown'
                }));
                currentGroup.value = group;
            }
        };

        const filteredMembers = computed(() => {
            let result = groupMembers.value;
            if (memberSearchQuery.value) {
                const query = memberSearchQuery.value.toLowerCase();
                result = result.filter(m => 
                    (m.nickname && m.nickname.toLowerCase().includes(query)) ||
                    (m.card && m.card.toLowerCase().includes(query)) ||
                    m.user_id.toString().includes(query)
                );
            }

            return [...result].sort((a, b) => {
                let res = 0;
                if (memberSortBy.value === 'role') {
                    const roles = { owner: 3, admin: 2, member: 1 };
                    res = (roles[a.role] || 0) - (roles[b.role] || 0);
                } else if (memberSortBy.value === 'nickname') {
                    res = (a.card || a.nickname).localeCompare(b.card || b.nickname);
                } else if (memberSortBy.value === 'user_id') {
                    res = a.user_id.toString().localeCompare(b.user_id.toString());
                }
                return memberSortAsc.value ? res : -res;
            });
        });

        const sortMembers = (field) => {
            if (memberSortBy.value === field) {
                memberSortAsc.value = !memberSortAsc.value;
            } else {
                memberSortBy.value = field;
                memberSortAsc.value = true;
            }
        };

        const banMember = async (groupId, userId) => {
            const durationStr = prompt(t('prompt_ban_duration') || 'Enter duration in minutes (0 to unban):', '30');
            if (durationStr === null) return;
            const duration = parseInt(durationStr);
            if (isNaN(duration)) return;

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: 'set_group_ban',
                    params: {
                        group_id: groupId,
                        user_id: userId,
                        duration: duration * 60
                    }
                })
            });

            if (data && data.success) {
                alert(t('action_success') || 'Operation successful');
            }
        };

        const kickMember = async (groupId, userId) => {
            if (!confirm(t('confirm_kick') || 'Are you sure you want to kick this member?')) return;

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: 'set_group_kick',
                    params: {
                        group_id: groupId,
                        user_id: userId
                    }
                })
            });

            if (data && data.success) {
                fetchGroupMembers(groupId);
            }
        };

        const setMemberCard = async (groupId, userId, currentCard) => {
            const newCard = prompt(t('prompt_new_card') || 'Enter new card name:', currentCard);
            if (newCard === null) return;

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: 'set_group_card',
                    params: {
                        group_id: groupId,
                        user_id: userId,
                        card: newCard
                    }
                })
            });

            if (data && data.success) {
                fetchGroupMembers(groupId);
            }
        };

        const leaveGroup = async (groupId) => {
            if (!confirm(t('confirm_leave_group') || `Are you sure you want to leave group ${groupId}?`)) return;

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: 'set_group_leave',
                    params: { group_id: groupId }
                })
            });

            if (data && data.success) {
                alert(t('action_success') || 'Operation successful');
                fetchGroups(selectedBotId.value, true);
                currentGroup.value = null;
                groupMembers.value = [];
            }
        };

        const checkGroupMember = async () => {
            if (!currentGroup.value) return;
            const userId = prompt(t('prompt_check_member') || 'Enter user ID to check:');
            if (!userId) return;

            const data = await apiFetch(`/api/bot/action`, {
                method: 'POST',
                body: JSON.stringify({
                    bot_id: selectedBotId.value,
                    action: 'get_group_member_info',
                    params: {
                        group_id: currentGroup.value.group_id || currentGroup.value.guild_id,
                        user_id: userId,
                        no_cache: true
                    }
                })
            });

            if (data && data.success && data.data) {
                const m = data.data;
                alert(`${t('member_found') || 'Member found'}: ${m.card || m.nickname} (${m.user_id})`);
            } else {
                alert(t('member_not_found') || 'Member not found');
            }
        };

        const showSingleMsgModal = (target, type) => {
            singleMsgModal.value.target = target;
            singleMsgModal.value.type = type;
            singleMsgModal.value.show = true;
            singleMsgModal.value.content = '';
        };

        const executeSingleSend = async () => {
            if (!selectedBotId.value || !singleMsgModal.value.target) return;
            
            singleMsgModal.value.sending = true;
            try {
                const targetId = singleMsgModal.value.target.group_id || singleMsgModal.value.target.user_id;
                await apiFetch('/api/bot/action', {
                    method: 'POST',
                    body: JSON.stringify({
                        bot_id: selectedBotId.value,
                        action: singleMsgModal.value.type === 'group' ? 'send_group_msg' : 'send_private_msg',
                        params: {
                            [singleMsgModal.value.type === 'group' ? 'group_id' : 'user_id']: targetId,
                            message: singleMsgModal.value.content
                        }
                    })
                });
                singleMsgModal.value.show = false;
            } catch (e) {
                console.error('Failed to send message:', e);
            }
            singleMsgModal.value.sending = false;
        };

        const resetCamera = () => {
            if (window.visualizer) {
                window.visualizer.camera.position.set(0, 0, 2000);
                window.visualizer.controls.target.set(0, 0, 0);
                window.visualizer.controls.update();
            }
        };

        const fetchUserInfo = async () => {
            const data = await apiFetch('/api/me');
            if (data && data.success) {
                userInfo.value = data.user || {};
            }
        };

        const handleAvatarError = (e) => {
            e.target.src = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';
        };

        const submitAddBot = async () => {
            alert('Manual bot addition is not implemented. Bots should connect automatically or via Docker.');
            showAddBotModal.value = false;
        };

        const fetchAllData = async () => {
            if (!isLoggedIn.value) return;
            loading.value = true;
            try {
                const botsData = await apiFetch('/api/bots');
                if (botsData) {
                    bots.value = botsData.bots || [];
                    if (!selectedBotId.value && bots.value.length > 0) {
                        selectedBotId.value = bots.value[0].self_id;
                    }
                }

                const statsData = await apiFetch('/api/stats');
                if (statsData) {
                    // 更新 Vue 响应式数据
                    stats.value.total_msgs = statsData.message_count || 0;
                    stats.value.total_bots = statsData.bot_count_total || (bots.value ? bots.value.length : 0);
                    stats.value.active_workers = statsData.worker_count || 0;
                    
                    if (statsData.cpu_usage !== undefined) {
                        stats.value.cpu_usage_raw = statsData.cpu_usage;
                        stats.value.cpu_usage = `${parseFloat(statsData.cpu_usage).toFixed(1)}%`;
                    }
                    if (statsData.memory_used_percent !== undefined) {
                        stats.value.memory_usage = `${parseFloat(statsData.memory_used_percent).toFixed(1)}%`;
                    }
                    if (statsData.memory_used !== undefined) {
                        stats.value.memory_used = statsData.memory_used;
                        stats.value.memory_used_mb = Math.round(statsData.memory_used / 1024 / 1024);
                    }
                    if (statsData.goroutines !== undefined) {
                        stats.value.goroutines = statsData.goroutines;
                    }

                    // 保存趋势数据到 stats.value，以便图表重新初始化时使用
                    stats.value.cpu_trend = statsData.cpu_trend || [];
                    stats.value.mem_trend = statsData.mem_trend || [];
                    stats.value.msg_trend = statsData.msg_trend || [];
                    stats.value.sent_trend = statsData.sent_trend || [];
                    stats.value.recv_trend = statsData.recv_trend || [];
                    
                    // 保存其他指标到 stats.value
                    stats.value.memory_total = statsData.memory_total;
                    stats.value.memory_alloc = statsData.memory_alloc;
                    stats.value.bot_count = statsData.bot_count;
                    stats.value.bot_count_offline = statsData.bot_count_offline;
                    stats.value.bot_count_total = statsData.bot_count_total;
                    stats.value.worker_count = statsData.worker_count;
                    stats.value.active_groups = statsData.active_groups;
                    stats.value.active_groups_today = statsData.active_groups_today;
                    stats.value.active_users = statsData.active_users;
                    stats.value.active_users_today = statsData.active_users_today;
                    stats.value.sent_message_count = statsData.sent_message_count;
                    stats.value.message_count = statsData.message_count;

                    // 更新趋势相关
                    if (statsData.msg_trend && statsData.msg_trend.length > 0) {
                        const lastMsg = statsData.msg_trend[statsData.msg_trend.length - 1];
                        stats.value.msg_per_sec = lastMsg / 5.0;
                    }
                    if (statsData.sent_trend && statsData.sent_trend.length > 0) {
                        const lastSent = statsData.sent_trend[statsData.sent_trend.length - 1];
                        stats.value.sent_per_sec = lastSent / 5.0;
                    }

                    // 更新系统信息
                    if (statsData.os_platform) stats.value.os_platform = statsData.os_platform;
                    if (statsData.os_arch) stats.value.os_arch = statsData.os_arch;
                    if (statsData.cpu_model) stats.value.cpu_model = statsData.cpu_model;
                    if (statsData.disk_usage) stats.value.disk_usage = statsData.disk_usage;
                    
                    if (statsData.uptime) {
                        stats.value.uptime = statsData.uptime;
                    } else if (statsData.start_time) {
                        const uptimeSeconds = Math.floor(Date.now() / 1000 - statsData.start_time);
                        const d = Math.floor(uptimeSeconds / 86400);
                        const h = Math.floor((uptimeSeconds % 86400) / 3600);
                        const m = Math.floor((uptimeSeconds % 3600) / 60);
                        if (d > 0) {
                            stats.value.uptime = `${d}d ${h}h ${m}m`;
                        } else if (h > 0) {
                            stats.value.uptime = `${h}h ${m}m`;
                        } else {
                            stats.value.uptime = `${m}m`;
                        }
                    }

                    // 重要：更新首页图表 (stats.js 中的函数)
                    if (window.updateStats) {
                        window.updateStats(statsData);
                    }
                    if (window.updateChatStats && activeTab.value === 'dashboard') {
                        window.updateChatStats();
                    }
                }

                const workersData = await apiFetch('/api/workers');
                if (workersData) {
                    workers.value = workersData.workers || [];
                }

                // Load contacts for visualization and tabs
                if (selectedBotId.value) {
                    if (groups.value.length === 0) fetchGroups(selectedBotId.value);
                    if (friends.value.length === 0) fetchFriends(selectedBotId.value);
                }

                if (activeTab.value === 'system_logs' && systemLogs.value.length === 0) fetchLogs();
                if (activeTab.value === 'docker' && dockerContainers.value.length === 0) fetchDockerContainers();
                if (activeTab.value === 'users' && systemUsers.value.length === 0) fetchSystemUsers();

                pulseBars.value = pulseBars.value.map(() => Math.floor(Math.random() * 80) + 20);
            } catch (err) {
                console.error('Failed to fetch data:', err);
            } finally {
                loading.value = false;
            }
        };

        const getAvatar = (bot) => {
            let url = 'https://cdn.staticfile.org/bootstrap-icons/1.8.1/icons/robot.svg';
            if (bot.platform === 'QQ' || bot.platform === 'qq') {
                url = `https://q1.qlogo.cn/g?b=qq&nk=${bot.self_id}&s=100`;
            }
            return `/api/proxy/avatar?url=${encodeURIComponent(url)}`;
        };

        const goBack = () => {
            window.location.href = 'index.html';
        };

        const logout = () => {
            localStorage.removeItem('wxbot_token');
            localStorage.removeItem('wxbot_role');
            if (wsSubscriber) {
                try {
                    wsSubscriber.onclose = null; // Prevent reconnect
                    wsSubscriber.close();
                } catch (e) {}
                wsSubscriber = null;
            }
            isLoggedIn.value = false;
            window.location.href = 'index.html';
        };

        // Matrix Rain Effect
        let matrixInterval;
        const initMatrix = () => {
            const canvas = document.getElementById('matrixCanvas');
            if (!canvas) return;
            const ctx = canvas.getContext('2d');
            
            const resize = () => {
                canvas.width = window.innerWidth;
                canvas.height = window.innerHeight;
            };
            window.addEventListener('resize', resize);
            resize();

            const characters = "0123456789ABCDEFHIJKLMNOPQRSTUVWXYZ@#$%^&*()";
            const fontSize = 14;
            const columns = canvas.width / fontSize;
            const drops = Array(Math.floor(columns)).fill(1);

            const draw = () => {
                ctx.fillStyle = "rgba(10, 10, 10, 0.05)";
                ctx.fillRect(0, 0, canvas.width, canvas.height);
                ctx.fillStyle = "#00ff41";
                ctx.font = fontSize + "px monospace";

                for (let i = 0; i < drops.length; i++) {
                    const text = characters.charAt(Math.floor(Math.random() * characters.length));
                    ctx.fillText(text, i * fontSize, drops[i] * fontSize);
                    if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
                        drops[i] = 0;
                    }
                    drops[i]++;
                }
            };
            matrixInterval = setInterval(draw, 33);
        };

        const updateTime = () => {
            const now = new Date();
            currentTime.value = now.toLocaleTimeString(lang.value === 'zh-CN' ? 'zh-CN' : 'en-US', {
                hour12: false,
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
        };
        updateTime();

        watch(groups, (newGroups) => {
            if (window.visualizer && newGroups) {
                newGroups.forEach(group => {
                    window.visualizer.getOrCreateNode(group.group_id || group.id, 'group', group.group_name || group.id);
                });
            }
        });

        watch(friends, (newFriends) => {
            if (window.visualizer && newFriends) {
                newFriends.forEach(friend => {
                    window.visualizer.getOrCreateNode(friend.user_id || friend.id, 'user', friend.nickname || friend.id);
                });
            }
        });

        onMounted(() => {
            console.log('Vue instance mounted!');
            updateThemeClass();
            updateTime();
            
            if (isLoggedIn.value) {
                fetchAllData();
                fetchUserInfo();
                initWebSocket();
                
                // Initialize legacy stats with a more robust retry mechanism
                const initLegacyStats = () => {
            if (window.initCharts) {
                console.log('Initializing legacy charts: window.initCharts is available');
                window.initCharts();
                if (window.updateStats) {
                    console.log('Updating stats: window.updateStats is available');
                    window.updateStats(toRaw(stats.value));
                }
                if (window.updateChatStats) window.updateChatStats();
                return true;
            }
            console.log('window.initCharts not yet available, will retry...');
            return false;
        };

                if (!initLegacyStats()) {
                    const retryInterval = setInterval(() => {
                        if (initLegacyStats()) {
                            clearInterval(retryInterval);
                        }
                    }, 500);
                    // Stop retrying after 5 seconds
                    setTimeout(() => clearInterval(retryInterval), 5000);
                }
            } else {
                // If not logged in, we still might want to fetch some public info or just wait for login
                fetchAllData(); 
            }
            
            const timeInterval = setInterval(updateTime, 1000);

            const dataInterval = setInterval(() => {
                if (isLoggedIn.value) {
                    fetchAllData();
                }
            }, 5000);

            initMatrix();
            
            watch(showMobileMenu, (newVal) => {
                if (newVal) {
                    safeCreateIcons();
                }
            });
            
            watch(activeTab, (newTab) => {
                localStorage.setItem('activeTab', newTab);
                safeCreateIcons();
                if (!isLoggedIn.value) return;
                
                // Re-initialize charts when switching back to dashboard
                if (newTab === 'dashboard') {
                    setTimeout(() => {
                        if (window.initCharts) window.initCharts();
                        if (window.updateStats) window.updateStats(toRaw(stats.value));
                    }, 100);
                }

                if (newTab === 'groups') {
                    if (!selectedBotId.value && bots.value.length > 0) {
                        selectedBotId.value = bots.value[0].self_id;
                    } else if (selectedBotId.value) {
                        fetchGroups(selectedBotId.value);
                    }
                }
                
                if (newTab === 'friends') {
                    if (!selectedBotId.value && bots.value.length > 0) {
                        selectedBotId.value = bots.value[0].self_id;
                    } else if (selectedBotId.value) {
                        fetchFriends(selectedBotId.value);
                    }
                }

                if (newTab === 'system_logs') {
                    fetchLogs();
                }

                if (newTab === 'docker') {
                    fetchDockerContainers();
                }

                if (newTab === 'users') {
                    fetchSystemUsers();
                }

                if (newTab === 'settings') {
                    fetchRoutingRules();
                    fetchDockerContainers();
                    fetchBackendConfig();
                    fetchUserInfo();
                }

                if (newTab === 'visualization') {
                    // Fetch all contacts from all bots for visualization
                    fetchGroups();
                    fetchFriends();
                    
                    nextTick(() => {
                        if (!window.visualizer) {
                            window.visualizer = new RoutingVisualizer('visualizerContainer');
                        }
                    });
                }

                safeCreateIcons();
            });

            watch(selectedBotId, (newId) => {
                if (!isLoggedIn.value) return;
                if (activeTab.value === 'groups') {
                    fetchGroups(newId);
                }
                if (activeTab.value === 'friends') {
                    fetchFriends(newId);
                }
            });

            safeCreateIcons();

            onUnmounted(() => {
                clearInterval(timeInterval);
                clearInterval(dataInterval);
                clearInterval(matrixInterval);
                if (wsSubscriber) {
                    wsSubscriber.onclose = null;
                    wsSubscriber.close();
                    wsSubscriber = null;
                }
            });
        });

        const toggleSidebar = () => {
            isSidebarCollapsed.value = !isSidebarCollapsed.value;
            localStorage.setItem('sidebarCollapsed', isSidebarCollapsed.value);
        };

        return {
            toggleSidebar,
            lang,
            isDark,
            shieldActive,
            isLoggedIn,
            loginLoading,
            loginError,
            loginData,
            handleLogin,
            t,
            toggleLang,
            toggleTheme,
            currentTime,
            loading,
            searchQuery,
            filterTab,
            activeTab,
                showMobileMenu,
                navItems,
                selectedNodeDetails,
                showingNodeDetails,
                activeNavItem,
            bots,
            workers,
            groups,
            friends,
            systemLogs,
            isSidebarCollapsed,
            showAddBotModal,
            newBotData,
            submitAddBot,
            debugResponse,
            menuGroups,
            statsCards,
            uptimeDisplay,
            recentLogs,
            stats,
            clearLogs,
            routingRules,
            dockerContainers,
                dockerLogs,
                dockerLogFilter,
                filteredDockerLogs,
                downloadDockerLogs,
                showingDockerLogs,
                currentDockerContainer,
                loadingContainers,
                fetchDockerLogs,
                addDockerBot,
                addDockerWorker,
                manageDockerContainer,
                systemUsers,
                showingUserModal,
                userModalData,
                submitUserModal,
                manageSystemUser,
                showingMassSend,
                massSendType,
                massSendTargets,
                massSendMessage,
                massSendSelectedTargets,
                massSendStatus,
                openMassSend,
                toggleMassSendTarget,
                selectAllMassSendTargets,
                startMassSend,
                backendConfig,
            userInfo,
            selectedBotId,
            filteredBots,
            coreStats,
            pulseBars,
            fetchGroups,
            fetchFriends,
            fetchLogs,
            fetchRoutingRules,
            saveRoutingRule,
            deleteRoutingRule,
            fetchDockerContainers,
            fetchUserInfo,
            manageDockerContainer,
            fetchBackendConfig,
            saveBackendConfig,
            fetchSystemUsers,
            updatePassword,
            manageBot,
            getAvatar,
            handleAvatarError,
            goBack,
            logout,
            wsStatus,
            wsConnected,
            singleMsgModal,
            fetchAllData,
            showSingleMsgModal,
            executeSingleSend,
            resetCamera,
            groupMembers,
            currentGroup,
            memberSearchQuery,
            memberSortBy,
            memberSortAsc,
            fetchGroupMembers,
            filteredMembers,
            sortMembers,
            banMember,
            kickMember,
            setMemberCard,
            leaveGroup,
            checkGroupMember
        };
    }
});

app.config.errorHandler = (err, vm, info) => {
    console.error('Vue Error:', err, info);
};

app.mount('#app');

console.log('matrix.js execution finished, Vue mounted to #app');
window.__matrix_loaded = true;
