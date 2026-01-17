using System.Data;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public static long MinTokens => -300000;
        public static long MaxTokensDay => 30000;
        public static long MaxTokens => -1000000;

        //增加算力
        public (int Result, long TokensValue) AddTokens(long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
            => AddTokensAsync(tokensAdd, tokensInfo, trans).GetAwaiter().GetResult();

        public async Task<(int Result, long TokensValue)> AddTokensAsync(long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
        {
            if (trans != null)
            {
                var res = await UserService.AddTokensAsync(SelfId, GroupId, GroupName, UserId, Name, tokensAdd, tokensInfo, trans);
                return (res.Result, res.TokensValue);
            }
            else
            {
                var res = await UserService.AddTokensTransAsync(SelfId, GroupId, GroupName, UserId, Name, tokensAdd, tokensInfo);
                return (res.Result, res.TokensValue);
            }
        }

        //减少算力
        public (int Result, long TokensValue) MinusTokens(long tokensMinus, string tokensInfo, IDbTransaction? trans = null)
            => MinusTokensAsync(tokensMinus, tokensInfo, trans).GetAwaiter().GetResult();

        public async Task<(int Result, long TokensValue)> MinusTokensAsync(long tokensMinus, string tokensInfo, IDbTransaction? trans = null)
        {
            return await AddTokensAsync(-tokensMinus, tokensInfo, trans);
        }

        //算力是否充足
        public async Task<bool> IsEnoughAsync()
        {
            if (Group.IsOwnerPay)            
                return await UserRepository.GetTokensAsync(Group.RobotOwner) >  MinTokens;            
            else
            {
                var tokens = await UserRepository.GetTokensAsync(UserId);
                return (tokens > MinTokens || await UserRepository.GetDayTokensGroupAsync(GroupId, UserId) > -MaxTokensDay) && tokens > MaxTokens;
            }
        }

        //#智能体 快捷对话
        public async Task<bool> TryParseAgentCall()
        { 
            if (string.IsNullOrWhiteSpace(Message)) return false;

            var match = Message.Trim().RegexMatch(@"^[#＃](\S+)(?:\s+(.*))?$");
            if (!match.Success) return false;

            var agentName = match.Groups[1].Value.Trim();
            var cmdPara = match.Groups[2].Success ? match.Groups[2].Value.Trim() : "";

            var agent = await AgentRepository.GetByNameAsync(agentName);
            if (agent == null)
                return false;

            CurrentAgent = agent;
            IsCallAgent = true;        
            CmdPara = cmdPara;
            return true;
        }

        //变身
        public async Task<string> ChangeAgentAsync()
        {
            IsCancelProxy = true;
            CurrentAgent = await AgentRepository.GetByIdAsync(User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : User.AgentId) ?? new();            
            var agentName = CurrentAgent.Name == "早喵" ? "" : $"【{CurrentAgent.Name}】";
            if (CmdPara == "")            
            {
                var names = await AgentRepository.GetNamesByTagAsync(1);
                return $"🤖 {agentName}可变身的智能体有:\n{names}";
            }
            
            var targetAgent = await AgentRepository.GetByNameAsync(CmdPara);
            if (targetAgent != null)
            {
                IsCallAgent = true;               
                CurrentAgent = targetAgent;                
                return await UserRepository.SetValueAsync("AgentId", targetAgent.Id, UserId) == -1 
                    ? $"变身{RetryMsg}" 
                    : $"🤖【{CurrentAgent.Name}】{CurrentAgent.Info}\n退出与智能体{CurrentAgent.Name}对话请发送【结束】";
            }
            else
                return "您要切换的智能体不存在";
        }

        static readonly string[] ExitTips =
        [
            "如需退出，发送“结束”即可～",
            "输入“结束”可随时切换智能体哦。",
            "觉得聊够了吗？发送“结束”就可以退出啦。",
            "💡发送“结束”可以换个智能体继续聊。",
            "🤖小提示：发送“结束”即可退出当前智能体。"
        ];

        static readonly string[] ImpatientKeywords =
        [
            // 明确表达厌烦
            "闭嘴", "别说了", "够了", "烦", "滚", "走开", "别讲了", "安静",
            "你够了", "你闭嘴", "别再说了", "打住", "住口", "别来烦我",
    
            // 想要结束
            "不说了", "结束", "撤了", "拜拜", "再见", "退下", "不聊了", "不想说了",
            "歇了", "累了", "收工", "没兴趣了", "停", "停一下", "停下",

            // 网络用语/缩写
            "886", "88", "溜了", "闪了", "撤退", "撤回", "撤离", "bye", "byebye",

            // 含情绪的词汇
            "气死我了", "受够了", "头疼", "好烦", "懒得理", "莫名其妙", "没劲", "无聊",

            // 质疑类（视语境而定是否判断为不耐烦）
            "你在说啥", "说什么呢", "说了半天啥也没说", "你在干嘛", "这啥玩意", "废话",

            // 高强度的拒绝
            "闭嘴吧", "够够的了", "你行你上", "你走吧", "我不想听了", "少来这套", "没完没了",
        ];

        // AI 智能体

        public async Task GetAgentResAsync()
        {
            if (IsGuild)
            {
                Answer = NoAnswer;
                return;
            } 
            
            if (IsRealProxy)
                IsCancelProxy = true;   

            if (!IsNested && IsGroup && !Group.IsAI)
            {
                if (CmdName.In("AI"))
                    Answer = "AI功能已关闭";
                else
                    Reason += "[关闭AI]";
                return;
            }

            if (!IsNested && !User.IsAI)
            {
                if (IsAtMe || !IsGroup || IsPublic)                
                    Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[禁用AI]";
                return;
            }

            if (!IsNested && !await IsEnoughAsync())
            {
                if (IsAtMe || !IsGroup || IsPublic)
                    Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[无算力]";
                return;
            }

            if (!IsNested && User.Credit <= 0)
            {
                if (IsAtMe || !IsGroup || IsPublic)
                    Answer = $"你的积分已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[负分]";
                return;
            }
            
            CurrentAgent = await AgentRepository.GetByIdAsync(User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : User.AgentId) ?? new();

            if (IsAgent && CmdPara == "结束")
            {               
                Answer = $"✅ 已结束与智能体【{CurrentAgent.Name}】的对话";
                await UserRepository.SetValueAsync("AgentId", AgentInfos.DefaultAgent.Id, UserId);
                return;
            }

            IsAI = true;            

            await GetChatHistoryAsync();

            var model = await LLMRepository.GetModelByIdAsync(CurrentAgent.ModelId);
            var providerObj = model != null ? await LLMRepository.GetProviderByIdAsync(model.ProviderId) : null;
            
            ModelId = model?.Id ?? 0;
            var providerName = providerObj?.Name ?? "Doubao";
            var modelName = model?.Name;

            var provider = LLMApp?._manager.GetProvider(providerName);
            if (provider != null)
            {
                AnswerAI = await provider.ExecuteAsync(History, new ModelExecutionOptions { ModelId = modelName });
                AnswerAI = AnswerAI.Trim();

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
          
                Answer = (CurrentAgent.Name.IsNull() || !IsAgent) && !IsCallAgent
                    ? AnswerAI
                    : !IsCallAgent && ShouldAddExitTip(Message)
                        ? $"【{CurrentAgent.Name}】{AnswerAI} {ExitTips[Random.Shared.Next(ExitTips.Length)]}"
                        : $"【{CurrentAgent.Name}】{AnswerAI}";
                if (IsCallAgent)
                    AnswerAI = $"【{CurrentAgent.Name}】{AnswerAI}";
            }
            else
            {
                Answer = "模型提供者不存在";
                return;
            }
        }

        // 生图功能实现
        public async Task GetImageResAsync()
        {
            if (string.IsNullOrWhiteSpace(CmdPara))
            {
                Answer = "🎨 请输入图片描述，例如：画图 一个赛博朋克风格的猫";
                return;
            }

            if (!await IsEnoughAsync())
            {
                Answer = "❌ 您的算力不足，无法生成图片。";
                return;
            }

            // 扣除算力 (生图消耗 12000 算力)
            var cost = 12000;
            var (result, tokens) = await MinusTokensAsync(cost, $"生成图片: {CmdPara.Truncate(20)}");
            if (result == -1)
            {
                Answer = "❌ 算力扣除失败，请重试。";
                return;
            }

            Answer = "🎨 正在为您生成图片，请稍等...";
            await SendMessageAsync(); // 先发送提示消息

            try
            {
                var aiService = PluginManager?.AI;
                if (aiService == null)
                {
                    Answer = "❌ 错误：AI 服务不可用。";
                    return;
                }

                // 尝试构建插件上下文
                IPluginContext? context = null;
                if (PluginManager != null)
                {
                    context = new PluginContext(
                        new BotMessageEvent(this),
                        this.Platform,
                        this.SelfId.ToString(),
                        PluginManager.AI,
                         PluginManager.I18n,
                         PluginManager.Logger,
                         this.User,
                        this.Group,
                        null,
                        this.SelfInfo,
                        async msg => { this.Answer = msg; await this.SendMessageAsync(); },
                        async (title, artist, jumpUrl, coverUrl, audioUrl) => { await this.SendMusicAsync(title, artist, jumpUrl, coverUrl, audioUrl); }
                    );
                }

                var res = await aiService.GenerateImageAsync(CmdPara, context);
                
                Answer = res;
            }
            catch (Exception ex)
            {
                Answer = $"❌ 生图失败：{ex.Message}";
            }

            await BatchInsertAgentAsync();

            if (IsGuild && IsGroup && !User.IsAI)
            {
                var credit = TokensMinus;
                await UserService.AddCreditTransAsync(SelfId, GroupId, GroupName, UserId, Name, -credit, "使用AI");
            }
        }

        public async Task<string> MinusTokensResAsync(string tokensInfo)
        {
            var res = await UserService.AddTokensTransAsync(SelfId, GroupId, GroupName, Group.IsOwnerPay ? Group.RobotOwner : UserId, Name, -TokensMinus, $"{tokensInfo} {(Group.IsOwnerPay ? $" 群主付(QQ:{UserId})" : "")}");
            return res.Result == -1 ? "" : "";
        }

        public async Task<string> BatchInsertAgentAsync()
        {
            OutputTokens = Answer.GetTokensCount();
            TokensMinus = (InputTokens * CurrentAgent.tokensTimes + OutputTokens * CurrentAgent.tokensTimesOutput) / 2;
            await AgentLog.AppendAsync(this);
            return await MinusTokensResAsync($"使用AI");
        }

        public string GetSystemPrompt()
        {
            string systemPrompt;

            if (IsCallAgent || IsAgent)
                systemPrompt = CurrentAgent.Prompt;            
            else
            {
                systemPrompt = IsGroup
                    ? Group.SystemPrompt
                    : User.SystemPrompt;

                if (systemPrompt.IsNull())
                    systemPrompt = IsGroup
                        ? GroupInfo.GetValue("SystemPrompt", SystemPromptGroup)
                        : GroupInfo.GetValue("SystemPrompt", C2CMessageGroupId);

                //if (CurrentGroup.IsUseKnowledgebase)
                //{
                //    systemPrompt = $"{systemPrompt}\n如果用户的问题可能涉及本群的知识库内容，请调用函数查询后再回答。";
                //}
            }

            return systemPrompt;
        }

        public async Task GetChatHistoryAsync(int his = 3)
        {
            var systemPrompt = GetSystemPrompt();

            int context = IsAgent ? his :Group.ContextCount;            

            if (CurrentAgent.Guid.In(AgentInfos.PromptAgent.Guid, AgentInfos.InfoAgent.Guid)) context = 0;

            var questions = string.Empty;

            if (context > 0)
            {
                var historyItems = await GroupSendMessageRepository.GetChatHistoryAsync(GroupId, UserId, Group.IsMultAI, context);

                foreach (var item in historyItems)
                {
                    var question = item.Question.RemoveUserId(SelfId);
                    var re = await BotCmdService.GetRegexCmdAsync();

                    if (question.IsMatch(re))
                        (_, question) = await GetCmdParaAsync(question, re);

                    questions += question + "\n";

                    var answer = item.Answer.RegexReplace(@"\n积分：.*?累计：.*", "");
                    answer = answer.RegexReplace(@"^【\w*】", "");
                    long tokenCount = (question + answer).GetTokensCount();

                    if (InputTokens + tokenCount < CurrentAgent.tokensLimit - CurrentAgent.tokensOutputLimit)
                    {
                        History.AddAssistantMessage(answer);
                        History.AddUserMessage(question);
                        InputTokens += tokenCount + 4;
                    }
                    else break;
                }
                InputTokens += 2;
            }

            systemPrompt += $"\n当前时间: {GetTimeStamp()}";
            InputTokens += systemPrompt.GetTokensCount();

            //InfoMessage(systemPrompt);

            // 当前 history 顺序为：A → B → C
            History.AddSystemMessage(systemPrompt);
            // 顺序变为：A → B → C → systemPrompt

            // 如果你想倒序成：systemPrompt ← C ← B ← A
            History = [.. History.Reverse()];

            History.AddUserMessage(CurrentMessage.RemoveUserId(SelfId));
            InputTokens += CurrentMessage.GetTokensCount();
        }
}
