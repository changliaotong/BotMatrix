const { createApp, ref, computed, onMounted, onUnmounted, watch, nextTick } = Vue;

const translations = {
    'en-US': {
        dashboard: 'Dashboard',
        bot_matrix: 'Bot Matrix',
        groups: 'Groups',
        friends: 'Friends',
        system_logs: 'System Logs',
        settings: 'Settings',
        overview: 'Overview',
        total_bots: 'Total Bots',
        active_workers: 'Active Workers',
        messages_today: 'Messages Today',
        system_uptime: 'System Uptime',
        target_node: 'Target Node',
        refresh_sector: 'Refresh Sector',
        members: 'MEMBERS',
        manage_node: 'Manage Node',
        no_nodes: 'No neural nodes found in this sector',
        no_clusters: 'No active clusters found in this node',
        deployment_in_progress: 'MODULE DEPLOYMENT IN PROGRESS...',
        back: 'Back',
        logout: 'Logout',
        search_placeholder: 'Search nodes...',
        status_active: 'Active',
        status_offline: 'Offline',
        all_sectors: 'All Sectors',
        active_sectors: 'Active Sectors',
        offline_sectors: 'Offline Sectors',
        msgs: 'MSGS',
        id: 'ID',
        security_logs: 'View Security Logs',
        protocol_time: 'Node Protocol Time',
        pulse_map: 'Neural Pulse Map',
        realtime_stream: 'Real-time Stream',
        node_events: 'Node Events',
        secure_link: 'established secure link.',
        visualization: 'Visualization',
        neural_nodes: 'NEURAL NODES',
        active_routing: 'ACTIVE ROUTING',
        sync_status: 'SYNC STATUS',
        reset_camera: 'Reset View',
        mass_send: 'Mass Send',
        select_bot: 'Select Bot',
        select_bot_hint: 'Please select a bot to view data',
        mass_send_groups: 'Mass Send Groups',
        mass_send_friends: 'Mass Send Friends',
        message_content: 'Message Content',
        select_targets: 'Select Targets',
        select_all: 'Select All',
        deselect_all: 'Deselect All',
        cancel: 'Cancel',
        execute: 'Execute',
        sending: 'Sending...',
        send_message: 'Send Message',
        enter_message: 'Enter message content...',
        send: 'Send',
        // New Translations
        routing_rules: 'Routing Rules',
        docker_mgmt: 'Docker Management',
        user_mgmt: 'User Management',
        backend_config: 'Backend Config',
        change_password: 'Change Password',
        add_rule: 'Add Rule',
        rule_name: 'Rule Name',
        pattern: 'Pattern',
        target_worker: 'Target Worker',
        actions: 'Actions',
        status: 'Status',
        container_id: 'Container ID',
        image: 'Image',
        cpu_usage: 'CPU Usage',
        mem_usage: 'Memory Usage',
        save_changes: 'Save Changes',
        reset: 'Reset',
        admin_privileges: 'Admin Privileges Required',
        confirm_leave_group: 'Are you sure you want to leave this group?',
        prompt_check_member: 'Enter user ID to check:',
        member_found: 'Member found',
        member_not_found: 'Member not found',
        prompt_username: 'Enter username:',
        prompt_password: 'Enter password:',
        confirm_admin: 'Is administrator?',
        confirm_delete_user: 'Are you sure you want to delete this user?',
        confirm_delete_container: 'Are you sure you want to delete this container?',
        prompt_new_password: 'Enter new password:',
        docker_containers: 'Docker Containers',
        username: 'Username',
        role: 'Role',
        add_user: 'Add User',
        reset_pwd: 'Reset PWD',
        confirm_kick: 'Are you sure you want to kick this member?',
        prompt_ban_duration: 'Enter duration in minutes (0 to unban):',
        prompt_new_card: 'Enter new card name:',
        search_members: 'Search members...',
        sort_role: 'By Role',
        sort_nickname: 'By Nickname',
        sort_user_id: 'By ID'
    },
    'zh-CN': {
        dashboard: '仪表盘',
        bot_matrix: '云端矩阵',
        groups: '群组管理',
        friends: '好友管理',
        system_logs: '系统日志',
        settings: '系统设置',
        overview: '运行概览',
        total_bots: '在线机器人',
        active_workers: '处理节点',
        messages_today: '今日消息',
        system_uptime: '运行时间',
        target_node: '目标节点',
        refresh_sector: '刷新扇区',
        members: '成员',
        manage_node: '管理节点',
        no_nodes: '未发现神经节点',
        no_clusters: '未发现活跃簇',
        deployment_in_progress: '模块部署中...',
        back: '返回',
        logout: '退出',
        search_placeholder: '搜索节点...',
        status_active: '活跃',
        status_offline: '离线',
        all_sectors: '全部扇区',
        active_sectors: '活跃扇区',
        offline_sectors: '离线扇区',
        msgs: '消息',
        id: '标识',
        security_logs: '查看安全日志',
        protocol_time: '节点协议时间',
        pulse_map: '神经脉冲图谱',
        realtime_stream: '实时数据流',
        node_events: '节点事件',
        secure_link: '建立安全连接',
        visualization: '神经可视化',
        neural_nodes: '神经节点',
        active_routing: '活跃路由',
        sync_status: '同步状态',
        reset_camera: '重置视角',
        mass_send: '批量发送',
        select_bot: '选择机器人',
        select_bot_hint: '请选择机器人以查看数据',
        mass_send_groups: '批量发送群组',
        mass_send_friends: '批量发送好友',
        message_content: '消息内容',
        select_targets: '选择目标',
        select_all: '全选',
        deselect_all: '取消全选',
        cancel: '取消',
        execute: '执行发送',
        sending: '发送中...',
        send_message: '发送消息',
        enter_message: '输入消息内容...',
        send: '发送',
        // 新增翻译
        routing_rules: '路由规则',
        docker_mgmt: 'Docker 管理',
        user_mgmt: '用户管理',
        backend_config: '后端配置',
        change_password: '修改密码',
        add_rule: '添加规则',
        rule_name: '规则名称',
        pattern: '匹配模式',
        target_worker: '目标节点',
        actions: '操作',
        status: '状态',
        container_id: '容器 ID',
        image: '镜像',
        cpu_usage: 'CPU 使用率',
        mem_usage: '内存使用率',
        save_changes: '保存更改',
        reset: '重置',
        admin_privileges: '需要管理员权限',
        confirm_leave_group: '确定要退出该群聊吗？',
        prompt_check_member: '请输入要查询的用户ID:',
        member_found: '已找到成员',
        member_not_found: '未找到该成员',
        prompt_username: '请输入用户名:',
        prompt_password: '请输入密码:',
        confirm_admin: '是否设为管理员？',
        confirm_delete_user: '确定要删除该用户吗？',
        prompt_new_password: '请输入新密码:',
        docker_containers: 'Docker 容器',
        username: '用户名',
        role: '角色',
        add_user: '添加用户',
        reset_pwd: '重置密码',
        confirm_kick: '确定要移出群聊吗？',
        prompt_ban_duration: '请输入禁言时长（分钟，0为解除）:',
        prompt_new_card: '请输入新的名片内容:',
        search_members: '搜索成员...',
        sort_role: '按角色',
        sort_nickname: '按昵称',
        sort_user_id: '按账号',
        docker_logs: '容器日志',
        confirm_add_bot: '部署新的机器人容器？',
        confirm_add_worker: '部署新的 Worker 容器？',
        add_bot: '部署机器人',
        add_worker: '部署 Worker',
        filter_logs: '过滤日志...',
        download: '下载',
        node_details: '节点详情',
        last_seen: '最后在线',
        action_success: '操作成功',
        action_failed: '操作失败',
        container_deleted: '容器已删除',
        error_username_required: '请输入用户名',
        error_password_required: '请输入密码',
        error_password_too_short: '密码长度至少为 6 位',
        confirm_clear_logs: '确定要清除所有日志吗？'
    },
    'en': {
        dashboard: 'Dashboard',
        bot_matrix: 'Cloud Matrix',
        groups: 'Groups',
        friends: 'Friends',
        system_logs: 'Logs',
        settings: 'Settings',
        overview: 'Overview',
        total_bots: 'Online Bots',
        active_workers: 'Active Workers',
        messages_today: 'Messages Today',
        system_uptime: 'Uptime',
        target_node: 'Target Node',
        refresh_sector: 'Refresh Sector',
        members: 'Members',
        manage_node: 'Manage Node',
        no_nodes: 'No neural nodes discovered',
        no_clusters: 'No active clusters',
        deployment_in_progress: 'Module deploying...',
        back: 'Back',
        logout: 'Logout',
        search_placeholder: 'Search nodes...',
        status_active: 'Active',
        status_offline: 'Offline',
        all_sectors: 'All Sectors',
        active_sectors: 'Active Sectors',
        offline_sectors: 'Offline Sectors',
        msgs: 'Msgs',
        id: 'ID',
        security_logs: 'Security Logs',
        protocol_time: 'Protocol Time',
        pulse_map: 'Pulse Map',
        realtime_stream: 'Realtime Stream',
        node_events: 'Node Events',
        secure_link: 'Secure Link',
        visualization: 'Visualization',
        neural_nodes: 'Neural Nodes',
        confirm_add_bot: 'Deploy new Bot container?',
        confirm_add_worker: 'Deploy new Worker container?',
        add_bot: 'Add Bot',
        add_worker: 'Add Worker',
        active_routing: 'Active Routing',
        sync_status: 'Sync Status',
        reset_camera: 'Reset Camera',
        mass_send: 'Mass Send',
        select_bot: 'Select Bot',
        select_bot_hint: 'Please select a bot to view data',
        mass_send_groups: 'Mass Send Groups',
        mass_send_friends: 'Mass Send Friends',
        message_content: 'Message',
        select_targets: 'Select Targets',
        select_all: 'Select All',
        deselect_all: 'Deselect All',
        cancel: 'Cancel',
        execute: 'Execute',
        sending: 'Sending...',
        send_message: 'Send Message',
        enter_message: 'Enter message...',
        send: 'Send',
        routing_rules: 'Routing Rules',
        docker_mgmt: 'Docker Mgmt',
        user_mgmt: 'User Mgmt',
        backend_config: 'Backend Config',
        change_password: 'Change Password',
        add_rule: 'Add Rule',
        rule_name: 'Rule Name',
        pattern: 'Pattern',
        target_worker: 'Target Worker',
        actions: 'Actions',
        status: 'Status',
        container_id: 'Container ID',
        image: 'Image',
        cpu_usage: 'CPU',
        mem_usage: 'Memory',
        save_changes: 'Save',
        reset: 'Reset',
        admin_privileges: 'Admin Privileges',
        confirm_leave_group: 'Are you sure you want to leave this group?',
        prompt_check_member: 'Enter user ID:',
        member_found: 'Member found',
        member_not_found: 'Member not found',
        prompt_username: 'Enter username:',
        prompt_password: 'Enter password:',
        confirm_admin: 'Set as administrator?',
        confirm_delete_user: 'Are you sure you want to delete this user?',
        prompt_new_password: 'Enter new password:',
        docker_containers: 'Docker Containers',
        username: 'Username',
        role: 'Role',
        add_user: 'Add User',
        reset_pwd: 'Reset Password',
        confirm_kick: 'Are you sure you want to kick this member?',
        prompt_ban_duration: 'Enter ban duration (minutes, 0 to unban):',
        prompt_new_card: 'Enter new nickname:',
        search_members: 'Search members...',
        sort_role: 'By Role',
        sort_nickname: 'By Nickname',
        sort_user_id: 'By ID',
        docker_logs: 'Container Logs',
        filter_logs: 'Filter logs...',
        download: 'Download',
        node_details: 'Node Details',
        last_seen: 'Last Seen',
        action_success: 'Success',
        action_failed: 'Operation failed',
        container_deleted: 'Container was deleted',
        error_username_required: 'Username is required',
        error_password_required: 'Password is required',
        error_password_too_short: 'Password must be at least 6 characters',
        confirm_clear_logs: 'Clear all logs?'
    }
};

createApp({
    setup() {
        const lang = ref(localStorage.getItem('language') || 'zh-CN');
        const isDark = ref(localStorage.getItem('theme') !== 'light'); // Default to dark
        const shieldActive = ref(window.__shield_active || false);

        const t = (key) => {
            if (!translations[lang.value]) {
                return translations['zh-CN'][key] || key;
            }
            return translations[lang.value][key] || key;
        };

        const toggleLang = () => {
            lang.value = lang.value === 'zh-CN' ? 'en' : 'zh-CN';
            localStorage.setItem('language', lang.value);
            document.documentElement.lang = lang.value;
            nextTick(() => {
                lucide.createIcons();
            });
        };

        const toggleTheme = () => {
            isDark.value = !isDark.value;
            localStorage.setItem('theme', isDark.value ? 'dark' : 'light');
            updateThemeClass();
            nextTick(() => {
                lucide.createIcons();
            });
            // Update visualizer theme if active
            if (window.visualizer) {
                window.visualizer.setTheme(isDark.value);
            }
        };

        const updateThemeClass = () => {
            if (isDark.value) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
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
                this.camera.position.z = 8000;

                this.nodes = new Map();
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
                    
                    nextTick(() => {
                        lucide.createIcons();
                    });

                    // Center camera on node
                    const targetPos = nodeMesh.position.clone();
                    new TWEEN.Tween(this.controls.target)
                        .to({ x: targetPos.x, y: targetPos.y, z: targetPos.z }, 500)
                        .easing(TWEEN.Easing.Quadratic.Out)
                        .start();
                }
            }

            pulseNode(node) {
                const originalScale = 1000;
                const targetScale = 1500;
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
                    color: this.theme === 'dark' ? 0x00ff41 : 0x33ccff, 
                    size: 4,
                    transparent: true,
                    opacity: 0.6
                }));
                this.scene.add(this.stars);

                this.getOrCreateNode('nexus', 'nexus', 'NEXUS');
            }

            getOrCreateNode(id, type, label) {
                if (this.nodes.has(id)) return this.nodes.get(id);
                
                const cacheKey = `${type}_${label}`;
                let texture = this.textureCache.get(cacheKey);

                if (!texture) {
                    const canvas = document.createElement('canvas');
                    canvas.width = 512; canvas.height = 512;
                    const ctx = canvas.getContext('2d');
                    
                    // Draw glow
                    const gradient = ctx.createRadialGradient(256, 256, 0, 256, 256, 240);
                    let color = '#00ff41';
                    if (type === 'worker') color = '#33ccff';
                    if (type === 'bot') color = '#ff3366';
                    if (type === 'nexus') color = '#facc15';
                    
                    gradient.addColorStop(0, color);
                    gradient.addColorStop(0.3, color + '66');
                    gradient.addColorStop(0.7, color + '22');
                    gradient.addColorStop(1, 'transparent');
                    
                    ctx.fillStyle = gradient;
                    ctx.beginPath(); ctx.arc(256, 256, 240, 0, Math.PI * 2); ctx.fill();
                    
                    // Draw circle
                    ctx.strokeStyle = color;
                    ctx.lineWidth = 15;
                    ctx.setLineDash([20, 10]);
                    ctx.beginPath(); ctx.arc(256, 256, 200, 0, Math.PI * 2); ctx.stroke();
                    
                    // Text
                    ctx.shadowColor = color;
                    ctx.shadowBlur = 20;
                    ctx.fillStyle = '#fff'; ctx.font = 'bold 60px JetBrains Mono'; ctx.textAlign = 'center';
                    ctx.fillText(label, 256, 275);

                    texture = new THREE.CanvasTexture(canvas);
                    this.textureCache.set(cacheKey, texture);
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
                } else {
                    const angle = Math.random() * Math.PI * 2;
                    const dist = 6000 + Math.random() * 2500;
                    pos = new THREE.Vector3(Math.cos(angle) * dist, Math.sin(angle) * dist, (Math.random() - 0.5) * 2000);
                }
                
                sprite.position.copy(pos);
                sprite.scale.set(1000, 1000, 1);
                
                this.scene.add(sprite);
                const node = { id, type, label, mesh: sprite, targetPos: pos, pulse: 0 };
                this.nodes.set(id, node);
                return node;
            }

            handleRoutingEvent(event) {
                const source = this.getOrCreateNode(event.source || 'nexus', event.source_type || 'bot', event.source_label || event.source || 'BOT');
                const target = this.getOrCreateNode(event.target || 'nexus', event.target_type || 'worker', event.target_label || event.target || 'WORKER');
                
                if (source && target) {
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
                        this.getOrCreateNode(worker.id, 'worker', worker.id);
                    });
                }
                if (state.nodes) {
                    state.nodes.forEach(n => this.getOrCreateNode(n.id, n.type, n.label));
                }
                
                // Cleanup removed nodes
                const activeIds = new Set(['nexus', 
                    ...(state.bots?.map(b => b.self_id) || []), 
                    ...(state.workers?.map(w => w.id) || []),
                    ...(state.nodes?.map(n => n.id) || [])
                ]);
                this.nodes.forEach((node, id) => {
                    if (!activeIds.has(id)) {
                        this.scene.remove(node.mesh);
                        this.nodes.delete(id);
                        
                        if (selectedNodeDetails.value && selectedNodeDetails.value.id === id) {
                            showingNodeDetails.value = false;
                            selectedNodeDetails.value = null;
                        }
                    }
                });
            }

            createFloatingLabel(text, position) {
                const canvas = document.createElement('canvas');
                canvas.width = 512; canvas.height = 128;
                const ctx = canvas.getContext('2d');
                
                ctx.fillStyle = 'rgba(0,0,0,0.6)';
                ctx.roundRect(0, 0, 512, 80, 20);
                ctx.fill();
                
                ctx.strokeStyle = '#00ff41';
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

            createParticle(start, end, color = 0x00ff41) {
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
                    const baseScale = node.id === 'nexus' ? 1200 : 1000;
                    const isHovered = this.hoveredNode === node.mesh;
                    const hoverScale = isHovered ? 1.2 : 1.0;
                    const pulseScale = (1 + Math.sin(time * 2) * 0.05 + (node.pulse || 0) * 0.3) * hoverScale;
                    node.mesh.scale.set(baseScale * pulseScale, baseScale * pulseScale, 1);
                    
                    if (node.pulse > 0) node.pulse -= 0.02;
                    
                    if (node.id !== 'nexus') {
                        node.mesh.position.y += Math.sin(time + node.mesh.position.x) * 1.5;
                        node.mesh.material.rotation += 0.002;
                    }

                    if (isHovered) {
                        node.mesh.userData = { id: node.id, type: node.type, label: node.label };
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
                    this.stars.material.color.set(isDark ? 0x00ff41 : 0x33ccff);
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
        const searchQuery = ref('');
        const filterTab = ref('all');
        const activeTab = ref('overview');
        const showMobileMenu = ref(false);
        const bots = ref([]);
        const workers = ref([]);
        const groups = ref([]);
        const friends = ref([]);
        const groupMembers = ref([]);
        const currentGroup = ref(null);
        const logs = ref([]);
        const logFilter = ref('');
        const filteredLogs = computed(() => {
            if (!logFilter.value) return logs.value;
            const filter = logFilter.value.toLowerCase();
            return logs.value.filter(log => {
                const text = (log.msg || log.message || JSON.stringify(log)).toLowerCase();
                return text.includes(filter) || (log.level && log.level.toLowerCase().includes(filter));
            });
        });

        const downloadLogs = () => {
            if (logs.value.length === 0) return;
            const logText = logs.value.map(log => `[${log.time || 'SYSTEM'}] ${log.level || 'INFO'}: ${log.msg || log.message || JSON.stringify(log)}`).join('\n');
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
        const showingMassSend = ref(false);
        const massSendType = ref('group'); // 'group' or 'friend'
        const massSendTargets = ref([]);
        const massSendMessage = ref('');
        const massSendSelectedTargets = ref([]);
        const massSendStatus = ref({ current: 0, total: 0, running: false });
        const backendConfig = ref({});
        const userInfo = ref({});
        const selectedBotId = ref('');
        const stats = ref({
            total_msgs: 0,
            total_bots: 0,
            active_workers: 0,
            uptime: '0d 0h 0m'
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
            { id: 'visualization', label: t('visualization'), icon: 'zap' },
            { id: 'groups', label: t('groups'), icon: 'users' },
            { id: 'friends', label: t('friends'), icon: 'user-plus' },
            { id: 'logs', label: t('system_logs'), icon: 'file-text' },
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

        const updateTime = () => {
            const now = new Date();
            currentTime.value = now.toLocaleTimeString('en-US', { hour12: false });
        };

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
                        // Real-time log updates
                        if (data.post_type === 'log') {
                            logs.value.unshift(data.data);
                            if (logs.value.length > 100) logs.value.pop();
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
                window.location.href = 'index.html';
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
                    window.location.href = 'index.html';
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
            if (!botId) return;
            const data = await apiFetch(`/api/contacts?bot_id=${botId}${refresh ? '&refresh=true' : ''}`);
            if (data) {
                friends.value = Array.isArray(data) ? data.filter(item => item.type === 'friend' || item.type === 'contact' || item.type === 'private') : [];
                console.log(`Fetched ${friends.value.length} friends for bot ${botId}`);
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
            const data = await apiFetch('/api/logs');
            if (data) {
                logs.value = data.logs || [];
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
            const data = await apiFetch('/api/admin/users');
            if (data) {
                systemUsers.value = data.users || [];
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
            if (data) {
                backendConfig.value = data.config || {};
            }
        };

        const saveBackendConfig = async () => {
            const data = await apiFetch('/api/admin/config', {
                method: 'POST',
                body: JSON.stringify(backendConfig.value)
            });
            if (data && data.success) {
                // Config saved
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
            if (!botId) return;
            const data = await apiFetch(`/api/contacts?bot_id=${botId}${refresh ? '&refresh=true' : ''}`);
            if (data) {
                groups.value = Array.isArray(data) ? data.filter(item => item.type === 'group' || item.type === 'guild') : [];
                console.log(`Fetched ${groups.value.length} groups for bot ${botId}`);
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

        const fetchAllData = async () => {
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
                    stats.value.total_msgs = statsData.message_count || 0;
                    stats.value.total_bots = statsData.bot_count_total || bots.value.length;
                    stats.value.active_workers = statsData.worker_count || 0;
                    
                    if (statsData.uptime) {
                        const match = statsData.uptime.match(/(\d+h)?(\d+m)?/);
                        if (match && (match[1] || match[2])) {
                            stats.value.uptime = `${match[1] || ''} ${match[2] || ''}`.trim() || '0h 0m';
                        } else {
                            stats.value.uptime = '< 1m';
                        }
                    }
                }

                const workersData = await apiFetch('/api/workers');
                if (workersData) {
                    workers.value = workersData.workers || [];
                }

                fetchLogs();
                pulseBars.value = pulseBars.value.map(() => Math.floor(Math.random() * 80) + 20);
            } catch (err) {
                console.error('Failed to fetch data:', err);
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

        onMounted(() => {
            updateThemeClass();
            updateTime();
            setInterval(updateTime, 1000);
            
            fetchAllData();
            fetchUserInfo();
            const dataInterval = setInterval(fetchAllData, 5000);

            initMatrix();
            
            watch(showMobileMenu, (newVal) => {
                if (newVal) {
                    nextTick(() => {
                        lucide.createIcons();
                    });
                }
            });

            watch(activeTab, (newTab) => {
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

                if (newTab === 'logs') {
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
                    nextTick(() => {
                        if (!window.visualizer) {
                            window.visualizer = new RoutingVisualizer('visualizerContainer');
                        }
                    });
                }

                nextTick(() => {
                    lucide.createIcons();
                });
            });

            watch(selectedBotId, (newId) => {
                if (activeTab.value === 'groups') {
                    fetchGroups(newId);
                }
                if (activeTab.value === 'friends') {
                    fetchFriends(newId);
                }
            });

            lucide.createIcons();

            onUnmounted(() => {
                clearInterval(dataInterval);
                clearInterval(matrixInterval);
            });
        });

        return {
            lang,
            isDark,
            shieldActive,
            t,
            toggleLang,
            toggleTheme,
            currentTime,
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
            logs,
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
            stats,
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
            checkGroupMember,
            manageSystemUser
        };
    }
}).mount('#app');
