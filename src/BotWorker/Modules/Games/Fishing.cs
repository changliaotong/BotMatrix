using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models.BotMessages;
using System.Threading.Tasks;
using System;
using System.Collections.Generic;
using System.Linq;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.fishing.v2",
        Name = "新版钓鱼王",
        Version = "2.0.0",
        Author = "Matrix",
        Description = "深度钓鱼模拟：多场景探索、装备强化、鱼种图鉴、实时交易",
        Category = "Games"
    )]
    public class FishingPlugin : IPlugin
    {
        public List<Intent> Intents => [
            new() { Name = "钓鱼", Keywords = ["钓鱼", "钓鱼状态"] },
            new() { Name = "抛竿", Keywords = ["抛竿"] },
            new() { Name = "收竿", Keywords = ["收竿"] },
            new() { Name = "鱼篓", Keywords = ["鱼篓"] },
            new() { Name = "卖鱼", Keywords = ["卖鱼"] },
            new() { Name = "钓鱼商店", Keywords = ["钓鱼商店"] },
            new() { Name = "升级鱼竿", Keywords = ["升级鱼竿"] }
        ];

        private readonly IFishingService _fishingService;
        private readonly ILogger<FishingPlugin> _logger;

        public FishingPlugin(
            IFishingService fishingService,
            ILogger<FishingPlugin> logger)
        {
            _fishingService = fishingService;
            _logger = logger;
        }

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "新版钓鱼",
                Commands = ["钓鱼", "抛竿", "收竿", "鱼篓", "卖鱼", "钓鱼商店", "升级鱼竿", "钓鱼状态"],
                Description = "【钓鱼】查看当前状态；【抛竿】开始钓鱼；【收竿】看看收获；【鱼篓】查看战利品；【卖鱼】换取金币"
            }, HandleFishingAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task<string> HandleFishingAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return await _fishingService.HandleFishingAsync(userId, ctx.User?.Name ?? "钓鱼佬", cmd);
        }
    }
}
