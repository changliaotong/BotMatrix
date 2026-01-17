using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers;
using Microsoft.Extensions.Logging;
using System;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Utils;
using BotWorker.Domain.Entities;

namespace BotWorker.Modules.AI.Services
{
    public class AgentService : IAgentService
    {
        private readonly IAgentRepository _agentRepository;
        private readonly IUserRepository _userRepository;
        private readonly IUserService _userService;
        private readonly ILLMRepository _llmRepository;
        private readonly IAgentLogRepository _agentLogRepository;
        private readonly LLMApp _llmApp;
        private readonly IBotCmdService _botCmdService;
        private readonly IGroupSendMessageRepository _groupSendMessageRepository;
        private readonly IServiceProvider _serviceProvider;
        private readonly ILogger<AgentService> _logger;

        public AgentService(
            IAgentRepository agentRepository,
            IUserRepository userRepository,
            IUserService userService,
            ILLMRepository llmRepository,
            IAgentLogRepository agentLogRepository,
            LLMApp llmApp,
            IBotCmdService botCmdService,
            IGroupSendMessageRepository groupSendMessageRepository,
            IServiceProvider serviceProvider,
            ILogger<AgentService> logger)
        {
            _agentRepository = agentRepository;
            _userRepository = userRepository;
            _userService = userService;
            _llmRepository = llmRepository;
            _agentLogRepository = agentLogRepository;
            _llmApp = llmApp;
            _botCmdService = botCmdService;
            _groupSendMessageRepository = groupSendMessageRepository;
            _serviceProvider = serviceProvider;
            _logger = logger;
        }

        private static readonly long MinTokens = -300000;
        private static readonly long MaxTokensDay = 30000;
        private static readonly long MaxTokens = -1000000;

        private static readonly string[] ExitTips =
        {
            "如需退出，发送“结束”即可～",
            "输入“结束”可随时切换智能体哦。",
            "觉得聊够了吗？发送“结束”就可以退出啦。",
            "💡发送“结束”可以换个智能体继续聊。",
            "🤖小提示：发送“结束”即可退出当前智能体。"
        };

        private static readonly string[] ImpatientKeywords =
        {
            "闭嘴", "别说了", "够了", "烦", "滚", "走开", "别讲了", "安静",
            "你够了", "你闭嘴", "别再说了", "打住", "住口", "别来烦我",
            "不说了", "结束", "撤了", "拜拜", "再见", "退下", "不聊了", "不想说了",
            "歇了", "累了", "收工", "没兴趣了", "停", "停一下", "停下",
            "886", "88", "溜了", "闪了", "撤退", "撤回", "撤离", "bye", "byebye",
            "气死我了", "受够了", "头疼", "好烦", "懒得理", "莫名其妙", "没劲", "无聊",
            "你在说啥", "说什么呢", "说了半天啥也没说", "你在干嘛", "这啥玩意", "废话",
            "闭嘴吧", "够够的了", "你行你上", "你走吧", "我不想听了", "少来这套", "没完没了"
        };

        public async Task<bool> IsEnoughAsync(BotMessage botMsg)
        {
            if (botMsg.Group.IsOwnerPay)
                return await _userRepository.GetTokensAsync(botMsg.Group.RobotOwner) > MinTokens;
            else
            {
                var tokens = await _userRepository.GetTokensAsync(botMsg.UserId);
                return (tokens > MinTokens || await _userRepository.GetDayTokensGroupAsync(botMsg.GroupId, botMsg.UserId) > -MaxTokensDay) && tokens > MaxTokens;
            }
        }

        public async Task<bool> TryParseAgentCallAsync(BotMessage botMsg)
        {
            if (string.IsNullOrWhiteSpace(botMsg.Message)) return false;

            var match = botMsg.Message.Trim().RegexMatch(@"^[#＃](\S+)(?:\s+(.*))?$");
            if (!match.Success) return false;

            var agentName = match.Groups[1].Value.Trim();
            var cmdPara = match.Groups[2].Success ? match.Groups[2].Value.Trim() : "";

            var agent = await _agentRepository.GetByNameAsync(agentName);
            if (agent == null)
                return false;

            botMsg.CurrentAgent = agent;
            botMsg.IsCallAgent = true;
            botMsg.CmdPara = cmdPara;
            return true;
        }

        public async Task<string> ChangeAgentAsync(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;
            botMsg.CurrentAgent = await _agentRepository.GetByIdAsync(botMsg.User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : botMsg.User.AgentId) ?? new();
            var agentName = botMsg.CurrentAgent.Name == "早喵" ? "" : $"【{botMsg.CurrentAgent.Name}】";
            
            if (botMsg.CmdPara == "")
            {
                var names = await _agentRepository.GetNamesByTagAsync(1);
                return $"🤖 {agentName}可变身的智能体有:\n{names}";
            }

            var targetAgent = await _agentRepository.GetByNameAsync(botMsg.CmdPara);
            if (targetAgent != null)
            {
                botMsg.IsCallAgent = true;
                botMsg.CurrentAgent = targetAgent;
                return await _userRepository.SetValueAsync("AgentId", targetAgent.Id, botMsg.UserId) == -1
                    ? $"变身 失败，请稍后重试"
                    : $"🤖【{botMsg.CurrentAgent.Name}】{botMsg.CurrentAgent.Info}\n退出与智能体{botMsg.CurrentAgent.Name}对话请发送【结束】";
            }
            else
                return "您要切换的智能体不存在";
        }

        public async Task GetAgentResAsync(BotMessage botMsg)
        {
            if (botMsg.IsGuild)
            {
                botMsg.Answer = "不支持此平台";
                return;
            }

            if (botMsg.IsRealProxy)
                botMsg.IsCancelProxy = true;

            if (!botMsg.IsNested && botMsg.IsGroup && !botMsg.Group.IsAI)
            {
                if (botMsg.CmdName.In("AI"))
                    botMsg.Answer = "AI功能已关闭";
                else
                    botMsg.Reason += "[关闭AI]";
                return;
            }

            if (!botMsg.IsNested && !botMsg.User.IsAI)
            {
                if (botMsg.IsAtMe || !botMsg.IsGroup || botMsg.IsPublic)
                    botMsg.Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    botMsg.Reason += "[禁用AI]";
                return;
            }

            if (!botMsg.IsNested && !await IsEnoughAsync(botMsg))
            {
                if (botMsg.IsAtMe || !botMsg.IsGroup || botMsg.IsPublic)
                    botMsg.Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    botMsg.Reason += "[无算力]";
                return;
            }

            if (!botMsg.IsNested && botMsg.User.Credit <= 0)
            {
                if (botMsg.IsAtMe || !botMsg.IsGroup || botMsg.IsPublic)
                    botMsg.Answer = $"你的积分已用完。请每日签到获得或联系客服购买";
                else
                    botMsg.Reason += "[负分]";
                return;
            }

            botMsg.CurrentAgent = await _agentRepository.GetByIdAsync(botMsg.User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : botMsg.User.AgentId) ?? new();

            if (botMsg.IsAgent && botMsg.CmdPara == "结束")
            {
                botMsg.Answer = $"✅ 已结束与智能体【{botMsg.CurrentAgent.Name}】的对话";
                await _userRepository.SetValueAsync("AgentId", AgentInfos.DefaultAgent.Id, botMsg.UserId);
                return;
            }

            botMsg.IsAI = true;

            await GetChatHistoryAsync(botMsg);

            var model = await _llmRepository.GetModelByIdAsync(botMsg.CurrentAgent.ModelId);
            var providerObj = model != null ? await _llmRepository.GetProviderByIdAsync(model.ProviderId) : null;

            botMsg.ModelId = model?.Id ?? 0;
            var providerName = providerObj?.Name ?? "Doubao";
            var modelName = model?.Name;

            var provider = _llmApp._manager.GetProvider(providerName);
            if (provider != null)
            {
                botMsg.AnswerAI = await provider.ExecuteAsync(botMsg.History, new ModelExecutionOptions { ModelId = modelName });
                botMsg.AnswerAI = botMsg.AnswerAI.Trim();

                bool ContainsImpatientWord(string input) =>
                    ImpatientKeywords.Any(k => input.Contains(k, StringComparison.OrdinalIgnoreCase));

                bool Chance(int percentage) => Random.Shared.Next(100) < percentage;

                bool ShouldAddExitTip(string userInput)
                {
                    if (ContainsImpatientWord(userInput))
                    {
                        return Chance(50);
                    }
                    return Chance(20);
                }

                botMsg.Answer = (botMsg.CurrentAgent.Name.IsNull() || !botMsg.IsAgent) && !botMsg.IsCallAgent
                    ? botMsg.AnswerAI
                    : !botMsg.IsCallAgent && ShouldAddExitTip(botMsg.Message)
                        ? $"【{botMsg.CurrentAgent.Name}】{botMsg.AnswerAI} {ExitTips[Random.Shared.Next(ExitTips.Length)]}"
                        : $"【{botMsg.CurrentAgent.Name}】{botMsg.AnswerAI}";
                if (botMsg.IsCallAgent)
                    botMsg.AnswerAI = $"【{botMsg.CurrentAgent.Name}】{botMsg.AnswerAI}";
            }
            else
            {
                botMsg.Answer = "模型提供者不存在";
                return;
            }
        }

        public async Task GetImageResAsync(BotMessage botMsg)
        {
            if (string.IsNullOrWhiteSpace(botMsg.CmdPara))
            {
                botMsg.Answer = "🎨 请输入图片描述，例如：画图 一个赛博朋克风格的猫";
                return;
            }

            if (!await IsEnoughAsync(botMsg))
            {
                botMsg.Answer = "❌ 您的算力不足，无法生成图片。";
                return;
            }

            var cost = 12000;
            var resAdd = await _userService.AddTokensTransAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, -cost, $"生成图片: {botMsg.CmdPara.Truncate(20)}");
            if (resAdd.Result == -1)
            {
                botMsg.Answer = "❌ 算力扣除失败，请重试。";
                return;
            }

            botMsg.Answer = "🎨 正在为您生成图片，请稍等...";
            await botMsg.SendMessageAsync();

            try
            {
                var aiService = _serviceProvider.GetRequiredService<IAIService>();
                if (aiService == null)
                {
                    botMsg.Answer = "❌ 错误：AI 服务不可用。";
                    return;
                }

                IPluginContext context = new PluginContext(
                    new BotMessageEvent(botMsg),
                    botMsg.Platform,
                    botMsg.SelfId.ToString(),
                    aiService,
                    _serviceProvider.GetRequiredService<II18nService>(),
                    _serviceProvider.GetRequiredService<ILogger<AgentService>>(),
                    botMsg.User,
                    botMsg.Group,
                    null,
                    botMsg.SelfInfo,
                    async msg => { botMsg.Answer = msg; await botMsg.SendMessageAsync(); },
                    async (title, artist, jumpUrl, coverUrl, audioUrl) => { await botMsg.SendMusicAsync(title, artist, jumpUrl, coverUrl, audioUrl); }
                );

                var res = await aiService.GenerateImageAsync(botMsg.CmdPara, context);
                botMsg.Answer = res;
            }
            catch (Exception ex)
            {
                botMsg.Answer = $"❌ 生图失败：{ex.Message}";
            }

            await BatchInsertAgentAsync(botMsg);

            if (botMsg.IsGuild && botMsg.IsGroup && !botMsg.User.IsAI)
            {
                var credit = botMsg.TokensMinus;
                await _userService.AddCreditTransAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, -credit, "使用AI");
            }
        }

        private async Task BatchInsertAgentAsync(BotMessage botMsg)
        {
            botMsg.OutputTokens = botMsg.Answer.GetTokensCount();
            botMsg.TokensMinus = (botMsg.InputTokens * botMsg.CurrentAgent.tokensTimes + botMsg.OutputTokens * botMsg.CurrentAgent.tokensTimesOutput) / 2;
            await _agentLogRepository.AppendAsync(botMsg);
            await _userService.AddTokensTransAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.Group.IsOwnerPay ? botMsg.Group.RobotOwner : botMsg.UserId, botMsg.Name, -botMsg.TokensMinus, $"使用AI {(botMsg.Group.IsOwnerPay ? $" 群主付(QQ:{botMsg.UserId})" : "")}");
        }

        private async Task GetChatHistoryAsync(BotMessage botMsg, int his = 3)
        {
            var systemPrompt = GetSystemPrompt(botMsg);
            int contextCount = botMsg.IsAgent ? his : botMsg.Group.ContextCount;

            if (botMsg.CurrentAgent.Guid.In(AgentInfos.PromptAgent.Guid, AgentInfos.InfoAgent.Guid)) contextCount = 0;

            if (contextCount > 0)
            {
                var historyItems = await _groupSendMessageRepository.GetChatHistoryAsync(botMsg.GroupId, botMsg.UserId, botMsg.Group.IsMultAI, contextCount);

                foreach (var item in historyItems)
                {
                    var question = item.Question.RemoveUserId(botMsg.SelfId);
                    var re = await _botCmdService.GetRegexCmdAsync();

                    if (question.IsMatch(re))
                    {
                        // Logic to get cmd para, usually on BotMessage but let's keep it simple for now or move it to a helper
                        // For now, assume question is clean enough
                    }

                    var answer = item.Answer.RegexReplace(@"\n积分：.*?累计：.*", "");
                    answer = answer.RegexReplace(@"^【\w*】", "");
                    long tokenCount = (question + answer).GetTokensCount();

                    if (botMsg.InputTokens + tokenCount < botMsg.CurrentAgent.tokensLimit - botMsg.CurrentAgent.tokensOutputLimit)
                    {
                        botMsg.History.AddAssistantMessage(answer);
                        botMsg.History.AddUserMessage(question);
                        botMsg.InputTokens += tokenCount + 4;
                    }
                    else break;
                }
                botMsg.InputTokens += 2;
            }

            systemPrompt += $"\n当前时间: {DateTime.Now:yyyy-MM-dd HH:mm:ss}";
            botMsg.InputTokens += systemPrompt.GetTokensCount();

            botMsg.History.AddSystemMessage(systemPrompt);
            botMsg.History = [.. botMsg.History.Reverse()];

            botMsg.History.AddUserMessage(botMsg.CurrentMessage.RemoveUserId(botMsg.SelfId));
            botMsg.InputTokens += botMsg.CurrentMessage.GetTokensCount();
        }

        private string GetSystemPrompt(BotMessage botMsg)
        {
            string systemPrompt;

            if (botMsg.IsCallAgent || botMsg.IsAgent)
                systemPrompt = botMsg.CurrentAgent.Prompt;
            else
            {
                systemPrompt = botMsg.IsGroup
                    ? botMsg.Group.SystemPrompt
                    : botMsg.User.SystemPrompt;

                if (systemPrompt.IsNull())
                    systemPrompt = botMsg.IsGroup
                        ? GroupInfo.GetValue("SystemPrompt", "你是一个由 sz84.com 开发的智能助手。")
                        : GroupInfo.GetValue("SystemPrompt", "你是一个由 sz84.com 开发的智能助手。");
            }

            return systemPrompt;
        }


    }
}
