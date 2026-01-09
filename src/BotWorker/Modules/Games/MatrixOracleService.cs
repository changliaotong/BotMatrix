using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "core.oracle",
        Name = "矩阵先知系统",
        Version = "1.0.0",
        Author = "BotMatrix AI",
        Description = "基于矩阵知识库的 AI 引导员，能够通过自然语言解答您关于系统的任何疑问。",
        Category = "Core"
    )]
    public class MatrixOracleService : IPlugin
    {
        private readonly ILogger<MatrixOracleService>? _logger;
        private IRobot? _robot;

        public MatrixOracleService() { }
        public MatrixOracleService(ILogger<MatrixOracleService> logger)
        {
            _logger = logger;
        }

        public List<Intent> Intents => [
            new() { Name = "先知咨询", Keywords = ["咨询", "问问", "oracle", "help"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;

            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "矩阵先知",
                Commands = ["咨询", "问问", "oracle"],
                Description = "【咨询 问题】通过 AI 获取系统运行逻辑与操作指引"
            }, HandleCommandAsync);

            // 注册跨插件调用接口
            await robot.RegisterSkillAsync(new SkillCapability { Name = "oracle.query" }, async (ctx, args) => {
                if (args == null || args.Length == 0) return "❌ 错误：缺少咨询问题。";
                return await AskOracleAsync(ctx.UserId, args[0]);
            });
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0)
            {
                return "👁️ 矩阵先知正注视着你。请描述您的疑问，例如：【咨询 如何提升位面？】";
            }

            string question = string.Join(" ", args);
            return await AskOracleAsync(ctx.UserId, question);
        }

        private async Task<string> AskOracleAsync(string userId, string question)
        {
            // TODO: 接入向量数据库检索与 LLM 生成逻辑
            // 目前先返回一个基于当前进度的占位回复
            
            _logger?.LogInformation($"[Oracle] 用户 {userId} 提问: {question}");

            return $"🔮 【先知预言】\n关于“{question}”的逻辑正在同步至向量矩阵...\n\n目前我已掌握：\n- 位面进化法则 (Evolution)\n- 积分金融准则 (Points)\n- 资源中心权限 (Market)\n\n请耐心等待 AI 逻辑核心完全启动。";
        }
    }
}
