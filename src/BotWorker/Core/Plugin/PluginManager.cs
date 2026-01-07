using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Core.OneBot;

namespace BotWorker.Core.Plugin
{
    public class PluginManager : IRobot
    {
        private readonly List<Skill> _skills = new();
        private readonly List<IPlugin> _plugins = new();

        public IReadOnlyList<Skill> Skills => _skills;
        public IReadOnlyList<IPlugin> Plugins => _plugins;

        public Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler)
        {
            _skills.Add(new Skill { Capability = capability, Handler = handler });
            return Task.CompletedTask;
        }

        public async Task LoadPluginAsync(IPlugin plugin)
        {
            _plugins.Add(plugin);
            await plugin.InitAsync(this);
        }

        public async Task<string> HandleMessageAsync(EventBase ev)
        {
            var message = ev.RawMessage.Trim();
            if (string.IsNullOrEmpty(message)) return string.Empty;

            foreach (var skill in _skills)
            {
                foreach (var cmd in skill.Capability.Commands)
                {
                    if (message.StartsWith(cmd, StringComparison.OrdinalIgnoreCase))
                    {
                        var args = message.Substring(cmd.Length).Trim().Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                        var ctx = new PluginContext(ev, "onebot", ev.SelfId.ToString())
                        {
                            RawMessage = message
                        };
                        return await skill.Handler(ctx, args);
                    }
                }
            }

            return string.Empty;
        }
    }
}
