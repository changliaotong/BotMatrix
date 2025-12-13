import os
import importlib
import sys
import traceback

class PluginManager:
    def __init__(self, plugin_dir="plugins"):
        self.plugin_dir = plugin_dir
        self.plugins = []
        self.load_plugins()

    def load_plugins(self):
        self.plugins = []
        if not os.path.exists(self.plugin_dir):
            os.makedirs(self.plugin_dir)
        
        # Ensure the parent directory is in path so we can import plugins.module
        if os.getcwd() not in sys.path:
            sys.path.append(os.getcwd())
        
        print(f"[PluginManager] Scanning {self.plugin_dir}...")
        for filename in os.listdir(self.plugin_dir):
            if filename.endswith(".py") and not filename.startswith("__"):
                module_name = filename[:-3]
                try:
                    # Import as package: plugins.demo
                    full_module_name = f"{self.plugin_dir}.{module_name}"
                    if full_module_name in sys.modules:
                        module = importlib.reload(sys.modules[full_module_name])
                    else:
                        module = importlib.import_module(full_module_name)
                    
                    if hasattr(module, "handle"):
                        self.plugins.append({
                            "name": module_name,
                            "module": module,
                            "handle": module.handle
                        })
                        print(f"[PluginManager] Loaded plugin: {module_name}")
                except Exception as e:
                    print(f"[PluginManager] Failed to load {module_name}: {e}")
                    traceback.print_exc()

    def process(self, context):
        """
        context: dict containing message info (content, sender, group, etc.)
        Returns: (reply_str, should_block)
        """
        final_reply = None
        should_block = False
        
        for plugin in self.plugins:
            try:
                res = plugin['handle'](context)
                if res:
                    if isinstance(res, str):
                        # Legacy support: string means reply
                        # We prioritize the first reply we get
                        if final_reply is None:
                            final_reply = res
                    elif isinstance(res, dict):
                        # New support: dict with 'reply' and 'block'
                        if "reply" in res and final_reply is None:
                            final_reply = res["reply"]
                        if res.get("block"):
                            should_block = True
                        if res.get("stop_chain"):
                            # Stop executing further plugins
                            break
            except Exception as e:
                print(f"[PluginManager] Error in {plugin['name']}: {e}")
        return final_reply, should_block
