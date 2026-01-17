using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Plugins;
using Microsoft.Extensions.Logging;
using System.Text;
using System.Reflection;
using System.Text.Json;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.pet",
        Name = "宠物养成",
        Version = "2.1.0",
        Author = "Matrix",
        Description = "超越市面水平的宠物系统：性格差异、随机事件、打工冒险、道具背包、ASCII艺术与深度互动",
        Category = "Games"
    )]
    public class PetPlugin : IPlugin
    {
        private readonly ILogger<PetPlugin> _logger;
        private readonly PetService _service;
        private readonly IPetRepository _petRepo;
        private PetConfig? _config;

        public IModuleMetadata Metadata => typeof(PetPlugin).GetCustomAttribute<BotPluginAttribute>()!;

        public PetPlugin(ILogger<PetPlugin> logger, PetService service, IPetRepository petRepo)
        {
            _logger = logger;
            _service = service;
            _petRepo = petRepo;
        }

        public async Task StopAsync() => await Task.CompletedTask;

        public async Task InitAsync(IRobot robot)
        {
            // 1. 加载配置
            _config = await LoadConfigAsync();

            // 2. 注册指令
            robot.RegisterSkill(new SkillCapability("宠物系统", GetCommandAliases()), DispatchCommandAsync);

            // 3. 注册通用事件钩子：用户发言增加亲密度
            await robot.RegisterEventAsync("message", HandleUserMessageAsync);

            _logger?.LogInformation("{PluginName} v{Version} 已启动。", Metadata.Name, Metadata.Version);
        }

        private async Task HandleUserMessageAsync(IPluginContext ctx)
        {
            // 过滤掉指令消息
            if (GetCommandAliases().Any(a => ctx.RawMessage.StartsWith(a, StringComparison.OrdinalIgnoreCase)))
                return;

            var pet = await _petRepo.GetByUserIdAsync(ctx.UserId);
            if (pet == null) return;

            // 只有闲逛状态且精力充足才增加亲密度
            if (pet.CurrentState == PetState.Idle && pet.Energy > 10)
            {
                // 先更新时间状态
                await _service.UpdateStateByTimeAsync(pet);

                pet.Intimacy += 0.1 * _config!.IntimacyGainRate;
                pet.Experience += 0.5;
                await _petRepo.UpdateAsync(pet);
            }
        }

        private string[] GetCommandAliases()
        {
            return typeof(PetService)
                .GetMethods()
                .SelectMany(m => m.GetCustomAttributes<PetCommandAttribute>())
                .SelectMany(a => a.Aliases)
                .Concat(new[] { "宠物帮助", "pet" })
                .Distinct()
                .ToArray();
        }

        private async Task<string> DispatchCommandAsync(IPluginContext ctx, string[] args)
        {
            var rawCmd = ctx.RawMessage.Trim().Split(' ')[0];
            
            var method = typeof(PetService).GetMethods()
                .FirstOrDefault(m => m.GetCustomAttributes<PetCommandAttribute>()
                    .Any(a => a.Aliases.Contains(rawCmd, StringComparer.OrdinalIgnoreCase)));

            if (method == null) return GetHelpInfo();

            try
            {
                var task = method.Invoke(_service!, new object[] { ctx, args }) as Task<string>;
                return await (task ?? Task.FromResult("指令执行未返回结果"));
            }
            catch (TargetInvocationException ex)
            {
                _logger.LogError(ex.InnerException, "执行宠物指令 {Command} 时出错", rawCmd);
                return $"❌ 指令执行失败: {ex.InnerException?.Message}";
            }
        }

        private async Task<PetConfig> LoadConfigAsync()
        {
            var configDir = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "Plugins", "configs");
            if (!Directory.Exists(configDir)) Directory.CreateDirectory(configDir);

            var configFile = Path.Combine(configDir, "game.pet.json");
            if (!File.Exists(configFile))
            {
                var defaultConfig = new PetConfig();
                var json = JsonSerializer.Serialize(defaultConfig, new JsonSerializerOptions { WriteIndented = true });
                await File.WriteAllTextAsync(configFile, json);
                return defaultConfig;
            }

            try
            {
                var json = await File.ReadAllTextAsync(configFile);
                return JsonSerializer.Deserialize<PetConfig>(json) ?? new PetConfig();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "加载宠物系统配置失败，使用默认配置");
                return new PetConfig();
            }
        }

        private string GetHelpInfo()
        {
            var sb = new StringBuilder();
            sb.AppendLine("🐾 【宠物系统 - 工业级插件模板】");
            sb.AppendLine("----------------------------");
            
            var commands = typeof(PetService).GetMethods()
                .Select(m => new { 
                    Method = m, 
                    Attr = m.GetCustomAttribute<PetCommandAttribute>() 
                })
                .Where(x => x.Attr != null)
                .OrderBy(x => x.Attr!.Order);

            foreach (var cmd in commands)
            {
                sb.AppendLine($"{cmd.Attr!.Order}. {string.Join("/", cmd.Attr.Aliases)} - {cmd.Attr.Description}");
            }

            return sb.ToString();
        }
    }
}