using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "core.oracle",
        Name = "矩阵先知系统",
        Version = "1.0.1",
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
                Commands = ["咨询", "问问", "oracle", "帮助", "help"],
                Description = "【咨询 问题】通过 AI 获取系统运行逻辑与操作指引"
            }, HandleCommandAsync);

            // 注册跨插件调用接口
            await robot.RegisterSkillAsync(new SkillCapability { Name = "oracle.query" }, async (ctx, args) => {
                if (args == null || args.Length == 0) return "❌ 错误：缺少咨询问题。";
                return await AskOracleAsync(ctx.UserId, args[0]);
            });

            // 延迟执行系统说明书索引，确保所有插件已加载
            _ = Task.Run(async () =>
            {
                try
                {
                    await Task.Delay(10000); // 等待 10 秒确保所有插件加载完毕
                    await IndexSystemManualAsync();
                }
                catch (Exception ex)
                {
                    _logger?.LogError(ex, "索引系统说明书失败");
                }
            });
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0)
            {
                return "👁️ 矩阵先知正注视着你。请描述您的疑问，例如：【咨询 如何提升位面？】\n\n您也可以直接输入【帮助】查看功能列表。";
            }

            string question = string.Join(" ", args);
            return await AskOracleAsync(ctx.UserId, question);
        }

        private async Task<string> IndexSystemManualAsync()
        {
            if (_robot == null) return string.Empty;

            var manual = new StringBuilder();
            manual.AppendLine("# 矩阵机器人系统说明书");
            manual.AppendLine("本机器人由 BotMatrix 驱动，集成 AI 与 RAG 增强。");
            manual.AppendLine();
            manual.AppendLine("## 核心功能清单：");

            foreach (var skill in _robot.Skills)
            {
                manual.AppendLine($"### 功能：{skill.Capability.Name}");
                manual.AppendLine($"- 指令：{string.Join(", ", skill.Capability.Commands)}");
                manual.AppendLine($"- 说明：{skill.Capability.Description}");
                manual.AppendLine();
            }

            // 同时将插件自身的 Metadata 也加入索引
            if (_robot is PluginManager pm)
            {
                // 这里可以通过反射获取所有插件的 BotPluginAttribute
                // 但为了简单，先用 Skills 里的信息
            }

            await _robot.Rag.IndexDocumentAsync(manual.ToString(), "system_manual");
            // _logger?.LogInformation("[Oracle] 系统说明书已完成 RAG 索引。");

            return "OK";
        }

        private async Task<string> AskOracleAsync(string userId, string question)
        {
            if (_robot == null) return "❌ 系统未就绪。";

            try
            {
                // 1. RAG 检索
                var chunks = await _robot.Rag.SearchAsync(question, 5);
                var context = string.Join("\n---\n", chunks.Select(c => c.Content));

                // 2. 构造 Prompt
                var prompt = $"你是一个专业的 AI 助手，名为“矩阵先知”。请根据以下提供的系统功能说明，回答用户关于机器人的提问。\n\n" +
                             $"【系统参考资料】\n{context}\n\n" +
                             $"【用户提问】\n{question}\n\n" +
                             $"请给出简洁明了、友好的回答。如果参考资料中没有相关信息，请告知用户并建议其联系管理员。";

                // 3. AI 生成
                return await _robot.AI.ChatAsync(prompt);
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "Oracle 咨询失败");
                return $"🔮 【先知预言】\n目前位面波纹过于剧烈，我暂时无法看清未来...\n错误原因：{ex.Message}";
            }
        }
    }
}
