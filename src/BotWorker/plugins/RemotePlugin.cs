using System;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Plugins
{
    /// <summary>
    /// 远程插件适配器：支持通过 HTTP 调用外部语言（Python, Go, JS 等）编写的插件
    /// </summary>
    public class RemotePlugin : IPlugin
    {
        private readonly string _endpoint;
        private readonly IModuleMetadata _metadata;
        private readonly HttpClient _httpClient;

        public IModuleMetadata Metadata => _metadata;

        public RemotePlugin(IModuleMetadata metadata, string endpoint)
        {
            _metadata = metadata;
            _endpoint = endpoint;
            _httpClient = new HttpClient { Timeout = TimeSpan.FromSeconds(10) };
        }

        public async Task InitAsync(IRobot robot)
        {
            // 1. 尝试从远程获取该插件定义的技能
            try
            {
                var capabilities = await _httpClient.GetFromJsonAsync<SkillCapability[]>(_endpoint + "/capabilities");
                if (capabilities != null)
                {
                    foreach (var cap in capabilities)
                    {
                        await robot.RegisterSkillAsync(cap, async (ctx, args) =>
                        {
                            return await CallRemoteHandlerAsync(ctx, cap.Name, args);
                        });
                    }
                }

                // 2. 注册通用事件监听
                await robot.RegisterEventAsync("message", async (ctx) =>
                {
                    await ForwardEventAsync(ctx);
                });
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[RemotePlugin] Failed to connect to {_metadata.Name} at {_endpoint}: {ex.Message}");
            }
        }

        private async Task<string> CallRemoteHandlerAsync(IPluginContext ctx, string skillName, string[] args)
        {
            try
            {
                var payload = new
                {
                    Skill = skillName,
                    Args = args,
                    Context = new
                    {
                        ctx.UserId,
                        ctx.GroupId,
                        ctx.Platform,
                        ctx.RawMessage,
                        ctx.IsMessage,
                        ctx.EventType,
                        ctx.BotId,
                        // 包含基础实体信息
                        User = ctx.User != null ? new { ctx.User.Id, Nickname = ctx.User.Name, Level = ctx.User.LvValue } : null,
                        Group = ctx.Group != null ? new { ctx.Group.Id, ctx.Group.GroupName } : null
                    }
                };

                var response = await _httpClient.PostAsJsonAsync(_endpoint + "/execute", payload);
                if (response.IsSuccessStatusCode)
                {
                    var result = await response.Content.ReadAsStringAsync();
                    return result;
                }
                return $"远程插件响应失败: {response.StatusCode}";
            }
            catch (Exception ex)
            {
                return $"调用远程插件异常: {ex.Message}";
            }
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task ForwardEventAsync(IPluginContext ctx)
        {
            try
            {
                await _httpClient.PostAsJsonAsync(_endpoint + "/event", new
                {
                    Type = ctx.EventType,
                    ctx.RawMessage,
                    ctx.UserId,
                    ctx.GroupId
                });
            }
            catch
            {
                // 忽略转发失败
            }
        }
    }

    public class RemotePluginMetadata : IModuleMetadata
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = "RemotePlugin";
        public string Version { get; set; } = "1.0.0";
        public string Author { get; set; } = "Unknown";
        public string Description { get; set; } = "A remote plugin connected via HTTP";
        public string Category { get; set; } = "General";
        public string[] Permissions { get; set; } = Array.Empty<string>();
        public string[] Dependencies { get; set; } = Array.Empty<string>();
        public bool IsEssential { get; set; } = false;

        public List<Intent> Intents { get; set; } = new();
        public List<UIComponent> UI { get; set; } = new();
        public string[] Events { get; set; } = Array.Empty<string>();
    }
}