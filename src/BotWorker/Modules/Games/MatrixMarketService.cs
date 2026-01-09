using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using BotWorker.Infrastructure.Utils.Schema;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    public class UserModuleAccess : MetaDataGuid<UserModuleAccess>
    {
        public override string TableName => "UserModuleAccess";
        public override string KeyField => "Id";

        public string UserId { get; set; } = string.Empty;
        public string ModuleId { get; set; } = string.Empty;
        public DateTime UnlockTime { get; set; }
        public int Level { get; set; } = 1;
    }

    public class MarketModule
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string Category { get; set; } = string.Empty;
        public long UnlockCost { get; set; }
        public int RequiredLevel { get; set; }
        public string Icon { get; set; } = "📁";
    }

    [BotPlugin(
        Id = "matrix_market",
        Name = "矩阵市场系统",
        Version = "1.0.0",
        Author = "BotMatrix Core",
        Description = "管理全服功能系统的开启与资源调度，将插件封装为用户可感知的系统模块。",
        Category = "Core"
    )]
    public class MatrixMarketService : IPlugin
    {
        private readonly ILogger<MatrixMarketService>? _logger;
        private IRobot? _robot;

        private readonly List<MarketModule> _modules = new()
        {
            new MarketModule { Id = "game.pet.v2", Name = "生命模拟系统", Description = "跨位面的生命形式模拟，可孵化并培养您的电子宠物。", Category = "Life", UnlockCost = 1000, RequiredLevel = 1, Icon = "🐾" },
            new MarketModule { Id = "game.marriage.v2", Name = "协议共鸣系统", Description = "建立深度逻辑链接，与其他实体达成共鸣契约。", Category = "Social", UnlockCost = 5000, RequiredLevel = 5, Icon = "💍" },
            new MarketModule { Id = "game.fishing.v2", Name = "位面垂钓系统", Description = "从虚空裂缝中打捞失落的数据残片。", Category = "Game", UnlockCost = 500, RequiredLevel = 1, Icon = "🎣" },
            new MarketModule { Id = "game.music", Name = "音频流转系统", Description = "解析并重构矩阵中的波形数据，享受跨时空的听觉盛宴。", Category = "Media", UnlockCost = 2000, RequiredLevel = 3, Icon = "🎵" },
            new MarketModule { Id = "core.oracle", Name = "矩阵先知系统", Description = "接入 AI 逻辑核心，通过自然语言实时解答系统疑问。", Category = "Core", UnlockCost = 10000, RequiredLevel = 10, Icon = "🔮" },
            new MarketModule { Id = "core.digital_staff", Name = "数字员工系统", Description = "组建自动化团队，雇佣 AI 员工为您自动开发系统或赚取积分。", Category = "Core", UnlockCost = 50000, RequiredLevel = 15, Icon = "💼" }
        };

        public MatrixMarketService() { }
        public MatrixMarketService(ILogger<MatrixMarketService> logger)
        {
            _logger = logger;
        }

        public List<Intent> Intents => [
            new() { Name = "资源中心", Keywords = ["资源中心", "市场", "market", "shop"] },
            new() { Name = "系统激活", Keywords = ["激活", "开启", "unlock"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();

            var capability = new SkillCapability
            {
                Name = "资源中心",
                Commands = ["资源中心", "市场", "market", "激活", "开启"],
                Description = "【资源中心】查看可用的系统模块；【激活 模块名】开启新系统"
            };

            await robot.RegisterSkillAsync(capability, HandleCommandAsync);
            
            // 额外注册 matrix_market ID，方便 MenuService 内部调用
            await robot.RegisterSkillAsync(new SkillCapability { Name = "matrix_market" }, async (ctx, args) => {
                return await GetMarketDisplayAsync(ctx.UserId);
            });
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                var checkTable = await UserModuleAccess.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserModuleAccess'");
                if (checkTable == 0)
                {
                    await UserModuleAccess.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<UserModuleAccess>());
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "MatrixMarketService 数据库初始化失败");
            }
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0].TrimStart('!', '！', '/', ' ');

            if (cmd == "资源中心" || cmd == "市场" || cmd == "market")
            {
                return await GetMarketDisplayAsync(ctx.UserId);
            }

            if ((cmd == "激活" || cmd == "开启" || cmd == "unlock") && args.Length > 0)
            {
                return await UnlockModuleAsync(ctx, args[0]);
            }

            return "💡 请输入【资源中心】查看可用系统，或【激活 系统名】进行开启。";
        }

        private async Task<string> GetMarketDisplayAsync(string userId)
        {
            var userAccess = await UserModuleAccess.QueryWhere("UserId = @p1", UserModuleAccess.SqlParams(("@p1", userId)));
            var unlockedIds = userAccess.Select(a => a.ModuleId).ToHashSet();

            var sb = new System.Text.StringBuilder();
            sb.AppendLine("🌌 --- 矩阵资源中心 (Matrix Resource Center) ---");
            sb.AppendLine("这里展示了您可以接入的逻辑系统。");
            sb.AppendLine();

            foreach (var category in _modules.GroupBy(m => m.Category))
            {
                sb.AppendLine($"【{category.Key} 类系统】");
                foreach (var m in category)
                {
                    bool isUnlocked = unlockedIds.Contains(m.Id);
                    string status = isUnlocked ? "✅ 已接入" : $"🔒 需 {m.UnlockCost} 积分 / Lv.{m.RequiredLevel}";
                    sb.AppendLine($"{m.Icon} {m.Name}：{m.Description}");
                    sb.AppendLine($"   > 状态: {status}");
                }
                sb.AppendLine();
            }

            sb.AppendLine("💡 使用【激活 系统名】来接入新的逻辑模块。");
            return sb.ToString();
        }

        private async Task<string> UnlockModuleAsync(IPluginContext ctx, string moduleName)
        {
            var userId = ctx.UserId;
            var module = _modules.FirstOrDefault(m => m.Name == moduleName || m.Id.Equals(moduleName, StringComparison.OrdinalIgnoreCase));
            if (module == null) return $"❌ 错误：在矩阵记录中未找到名为“{moduleName}”的系统。";

            // 检查是否已激活
            var existing = await UserModuleAccess.QueryWhere("UserId = @p1 AND ModuleId = @p2", UserModuleAccess.SqlParams(("@p1", userId), ("@p2", module.Id)));
            if (existing.Any()) return $"✨ 系统提示：“{module.Name}”已处于激活状态，无需重复接入。";

            // 检查等级 (调用 EvolutionService)
            // 这里我们通过数据库直接查，解耦插件调用
            var levelData = await UserLevel.GetByUserIdAsync(userId);
            var currentLevel = levelData?.Level ?? 1;
            if (currentLevel < module.RequiredLevel)
            { 
                return $"🚫 接入权限不足：您的进化等级为 Lv.{currentLevel}，而接入“{module.Name}”需要达到 Lv.{module.RequiredLevel}。";
            }

            // 检查积分 (调用 PointsService 进行扣费)
            // 我们通过 EventNexus 发布扣费请求，或者直接通过 robot.CallSkillAsync
            if (_robot != null)
            {
                var result = await _robot.CallSkillAsync("points.transfer", ctx, new[] { userId, "SYSTEM_REVENUE", module.UnlockCost.ToString(), $"激活系统模块: {module.Name}" });

                if (result?.ToString()?.Contains("成功") == true)
                {
                    // 记录激活
                    var access = new UserModuleAccess
                    {
                        UserId = userId,
                        ModuleId = module.Id,
                        UnlockTime = DateTime.Now,
                        Level = 1
                    };
                    await access.InsertAsync();

                    // 发布审计事件
                    await _robot.Events.PublishAsync(new SystemAuditEvent
                    {
                        Level = "Success",
                        Source = "MatrixMarket",
                        Message = $"用户 {userId} 成功激活了 {module.Name} 系统。",
                        TargetUser = userId
                    });

                    return $"🎊 恭喜！您已成功接入“{module.Name}”。系统逻辑正在同步中...";
                }
                else
                {
                    return $"⚠️ 接入失败：您的积分储备不足（需要 {module.UnlockCost} 积分）。";
                }
            }

            return "❌ 市场系统暂时无法连接到核心逻辑，请稍后再试。";
        }
    }
}
