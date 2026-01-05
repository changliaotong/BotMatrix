import json
import sys
import logging
import asyncio
import re
from typing import Callable, Dict, Any, List, Optional, Awaitable, Union

# Configure logging to stderr
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    stream=sys.stderr
)
logger = logging.getLogger("BotMatrixSDK")

class Context:
    def __init__(self, event: Dict[str, Any], plugin: 'BotMatrixPlugin'):
        self.event = event
        self.plugin = plugin
        self.actions: List[Dict[str, Any]] = []
        self.args: List[str] = [] # For command arguments
        self.params: Dict[str, str] = {} # For regex named groups

    @property
    def sender(self) -> str:
        return self.event.get("payload", {}).get("from", "")

    @property
    def group_id(self) -> str:
        return self.event.get("payload", {}).get("group_id", "")

    @property
    def text(self) -> str:
        return self.event.get("payload", {}).get("text", "")

    def reply(self, text: str):
        self.call_action("send_message", text=text)

    def reply_image(self, url: str):
        self.call_action("send_image", url=url)

    async def ask(self, prompt: str, timeout: float = 30.0) -> Optional['Context']:
        """
        Powerful Interactive Feature: Send a prompt and wait for the user's next message.
        Uses Correlation ID for distributed tracking.
        """
        correlation_id = f"ask_{self.event.get('id')}_{int(asyncio.get_event_loop().time() * 1000)}"
        
        # Manually create the action to include correlation_id
        payload = self.event.get("payload", {})
        action = {
            "type": "send_message",
            "target": payload.get("from", ""),
            "target_id": payload.get("group_id", ""),
            "text": prompt,
            "correlation_id": correlation_id
        }
        self.actions.append(action)
        
        future = asyncio.get_running_loop().create_future()
        self.plugin._waiting_sessions[correlation_id] = future
        
        try:
            result_ctx = await asyncio.wait_for(future, timeout=timeout)
            return result_ctx
        except asyncio.TimeoutError:
            return None
        finally:
            self.plugin._waiting_sessions.pop(correlation_id, None)

    async def call_skill(self, target_id: str, skill_name: str, payload: Dict[str, Any]):
        """
        Call a skill exported by another plugin.
        """
        self.call_action("call_skill", 
            target_plugin=target_id,
            skill_name=skill_name,
            payload=payload
        )

    def call_action(self, action_type: str, **kwargs):
        # Permission check against plugin.json 'actions'
        if not self.plugin.has_permission(action_type):
            logger.error(f"Permission denied: Action '{action_type}' is not declared in plugin.json")
            return

        payload = self.event.get("payload", {})
        action = {
            "type": action_type,
            "target": payload.get("from", ""),
            "target_id": payload.get("group_id", ""),
            **kwargs
        }
        self.actions.append(action)

HandlerFunc = Callable[[Context], Awaitable[None]]
MiddlewareFunc = Callable[[HandlerFunc], HandlerFunc]

class BotMatrixPlugin:
    def __init__(self, config_path: str = "plugin.json"):
        self.handlers: Dict[str, HandlerFunc] = {}
        self.middlewares: List[MiddlewareFunc] = []
        self._lock = asyncio.Lock()
        self._waiting_sessions: Dict[str, asyncio.Future] = {}
        self.config = self._load_config(config_path)

    def _load_config(self, path: str) -> Dict[str, Any]:
        try:
            with open(path, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception as e:
            logger.warning(f"Could not load plugin config: {e}")
            return {}

    def has_permission(self, action: str) -> bool:
        if not self.config: return True # Legacy mode
        allowed_actions = self.config.get("actions", [])
        return action in allowed_actions

    def use(self, middleware: MiddlewareFunc):
        self.middlewares.append(middleware)

    def on(self, event_name: str):
        def decorator(func: HandlerFunc):
            self.handlers[event_name] = func
            return func
        return decorator

    def on_message(self):
        return self.on("on_message")

    def on_intent(self, intent_name: str):
        return self.on(f"intent_{intent_name}")

    def export_skill(self, name: str):
        return self.on(f"skill_{name}")

    def command(self, cmd: str, regex: bool = False):
        """
        Advanced Command Router: Supports simple prefix or Regex with arguments.
        """
        def decorator(func: HandlerFunc):
            async def wrapped_handler(ctx: Context):
                text = ctx.text
                if regex:
                    match = re.search(cmd, text)
                    if match:
                        ctx.params = match.groupdict()
                        ctx.args = list(match.groups())
                        await func(ctx)
                else:
                    if text.startswith(f"{cmd} ") or text == cmd:
                        ctx.args = text[len(cmd):].strip().split()
                        await func(ctx)
            
            self.on_message()(wrapped_handler)
            return func
        return decorator

    async def _handle_event(self, msg: Dict[str, Any]):
        event_id = msg.get("id")
        event_name = msg.get("name")
        correlation_id = msg.get("correlation_id")
        
        # 1. Check by CorrelationID first (The most reliable way in distributed systems)
        if correlation_id and correlation_id in self._waiting_sessions:
            future = self._waiting_sessions[correlation_id]
            if not future.done():
                future.set_result(Context(msg, self))
                self._send_response_sync(event_id, True, [], "")
                return

        # 2. Fallback to session key
        if event_name == "on_message":
            payload = msg.get("payload", {})
            session_key = f"{payload.get('group_id')}:{payload.get('from')}"
            if session_key in self._waiting_sessions:
                future = self._waiting_sessions[session_key]
                if not future.done():
                    future.set_result(Context(msg, self))
                    # Important: We consume this message, don't trigger other handlers
                    # Alternatively, you might want to trigger both. Here we consume.
                    self._send_response_sync(event_id, True, [], "")
                    return

        handler = self.handlers.get(event_name)
        if not handler:
            self._send_response_sync(event_id, True, [], "")
            return

        final_handler = handler
        for mw in reversed(self.middlewares):
            final_handler = mw(final_handler)

        ctx = Context(msg, self)
        try:
            await final_handler(ctx)
            await self._send_response(event_id, True, ctx.actions, "")
        except Exception as e:
            logger.error(f"Handler error for {event_name}: {e}", exc_info=True)
            await self._send_response(event_id, False, [], str(e))

    def _send_response_sync(self, event_id: str, ok: bool, actions: List[Dict[str, Any]], error_msg: str):
        """Immediate response for internal consumption."""
        response = {"id": event_id, "ok": ok, "actions": actions}
        if error_msg: response["error"] = error_msg
        sys.stdout.write(json.dumps(response) + "\n")
        sys.stdout.flush()

    async def _send_response(self, event_id: str, ok: bool, actions: List[Dict[str, Any]], error_msg: str):
        response = {"id": event_id, "ok": ok, "actions": actions}
        if error_msg: response["error"] = error_msg
        async with self._lock:
            sys.stdout.write(json.dumps(response) + "\n")
            sys.stdout.flush()

    async def run_async(self):
        logger.info("Plugin SDK 3.0 (Interactive) started")
        loop = asyncio.get_event_loop()
        reader = asyncio.StreamReader()
        protocol = asyncio.StreamReaderProtocol(reader)
        await loop.connect_read_pipe(lambda: protocol, sys.stdin)

        while True:
            line = await reader.readline()
            if not line: break
            try:
                msg = json.loads(line.decode())
                if msg.get("type") == "event":
                    asyncio.create_task(self._handle_event(msg))
            except Exception as e:
                logger.error(f"Loop error: {e}")

    def run(self):
        try:
            asyncio.run(self.run_async())
        except KeyboardInterrupt:
            pass
