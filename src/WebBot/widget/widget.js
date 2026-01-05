(function() {
    // WebBot Widget - Vanilla JS
    const script = document.currentScript;
    const appKey = script.getAttribute('data-app-key') || 'default';
    const serverAddr = script.getAttribute('data-server') || (window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host;
    const themeColor = script.getAttribute('data-theme') || '#00FF9D';
    const title = script.getAttribute('data-title') || 'Web Assistant';

    // Create UI Elements
    const container = document.createElement('div');
    container.id = 'webbot-widget-container';
    container.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        z-index: 9999;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    `;

    const button = document.createElement('div');
    button.style.cssText = `
        width: 60px;
        height: 60px;
        border-radius: 30px;
        background: ${themeColor};
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    `;
    button.innerHTML = `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>`;

    const chatWindow = document.createElement('div');
    chatWindow.style.cssText = `
        position: absolute;
        bottom: 80px;
        right: 0;
        width: 350px;
        height: 500px;
        background: white;
        border-radius: 20px;
        box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        display: none;
        flex-direction: column;
        overflow: hidden;
        transition: all 0.3s ease;
        opacity: 0;
        transform: translateY(20px);
    `;

    chatWindow.innerHTML = `
        <div style="padding: 20px; background: ${themeColor}; color: white; display: flex; justify-content: space-between; align-items: center;">
            <div style="font-weight: bold;">${title}</div>
            <div id="webbot-close" style="cursor: pointer; opacity: 0.8;">✕</div>
        </div>
        <div id="webbot-messages" style="flex: 1; overflow-y: auto; padding: 15px; background: #f8f9fa; display: flex; flex-direction: column; gap: 10px;"></div>
        <div style="padding: 15px; border-top: 1px solid #eee; display: flex; gap: 10px; background: white;">
            <input id="webbot-input" type="text" placeholder="Type a message..." style="flex: 1; border: 1px solid #ddd; padding: 8px 12px; border-radius: 20px; outline: none;">
            <button id="webbot-send" style="background: ${themeColor}; color: white; border: none; padding: 8px 15px; border-radius: 20px; cursor: pointer; font-weight: bold;">Send</button>
        </div>
    `;

    container.appendChild(chatWindow);
    container.appendChild(button);
    document.body.appendChild(container);

    // State
    let isOpen = false;
    let ws = null;
    let userId = localStorage.getItem('webbot_user_id');

    // Toggle Chat
    button.onclick = () => {
        isOpen = !isOpen;
        if (isOpen) {
            chatWindow.style.display = 'flex';
            setTimeout(() => {
                chatWindow.style.opacity = '1';
                chatWindow.style.transform = 'translateY(0)';
            }, 10);
            connect();
        } else {
            chatWindow.style.opacity = '0';
            chatWindow.style.transform = 'translateY(20px)';
            setTimeout(() => chatWindow.style.display = 'none', 300);
        }
    };

    document.getElementById('webbot-close').onclick = button.onclick;

    function connect() {
        if (ws && ws.readyState === WebSocket.OPEN) return;

        // 如果 serverAddr 不包含协议，自动补全
        let wsUrl = serverAddr;
        if (!wsUrl.startsWith('ws://') && !wsUrl.startsWith('wss://')) {
            const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
            wsUrl = protocol + wsUrl;
        }
        
        wsUrl += `/ws/widget?app_key=${appKey}&user_id=${userId || ''}`;
        ws = new WebSocket(wsUrl);

        ws.onmessage = (e) => {
            const msg = JSON.parse(e.data);
            if (msg.type === 'init') {
                userId = msg.data.user_id;
                localStorage.setItem('webbot_user_id', userId);
                if (msg.data.welcome) addMessage('bot', msg.data.welcome);
            } else if (msg.type === 'text') {
                addMessage(msg.from === 'bot' ? 'bot' : 'other', msg.content, msg.from);
            }
        };

        ws.onclose = () => {
            console.log('WebBot connection closed');
            setTimeout(connect, 3000);
        };
    }

    function addMessage(type, content, nickname) {
        const msgDiv = document.createElement('div');
        const isSelf = type === 'self';
        const isBot = type === 'bot';
        
        msgDiv.style.cssText = `
            max-width: 80%;
            padding: 10px 14px;
            border-radius: 15px;
            font-size: 14px;
            line-height: 1.4;
            align-self: ${isSelf ? 'flex-end' : 'flex-start'};
            background: ${isSelf ? themeColor : (isBot ? '#fff' : '#e9ecef')};
            color: ${isSelf ? 'white' : '#333'};
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            position: relative;
        `;

        if (!isSelf && nickname) {
            const nameDiv = document.createElement('div');
            nameDiv.style.cssText = 'font-size: 10px; color: #888; margin-bottom: 2px;';
            nameDiv.innerText = nickname;
            msgDiv.prepend(nameDiv);
        }

        const textDiv = document.createElement('div');
        textDiv.innerText = content;
        msgDiv.appendChild(textDiv);

        const messages = document.getElementById('webbot-messages');
        messages.appendChild(msgDiv);
        messages.scrollTop = messages.scrollHeight;
    }

    const input = document.getElementById('webbot-input');
    const sendBtn = document.getElementById('webbot-send');

    const sendMessage = () => {
        const text = input.value.trim();
        if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;

        ws.send(JSON.stringify({ type: 'text', content: text }));
        addMessage('self', text);
        input.value = '';
    };

    sendBtn.onclick = sendMessage;
    input.onkeypress = (e) => { if (e.key === 'Enter') sendMessage(); };

})();
