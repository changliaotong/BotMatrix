using System.Data;
using sz84.Agents.Entries;
using sz84.Agents.Providers;
using sz84.Agents.Providers.Txt2Img;
using sz84.Bots.Entries;
using sz84.Bots.Groups;
using sz84.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public static long MinTokens => -300000;
        public static long MaxTokensDay => 30000;
        public static long MaxTokens => -1000000;

        //算力是否充足
        public bool IsEnough()
        {
            if (Group.IsOwnerPay)            
                return UserInfo.GetTokens(Group.RobotOwner) >  MinTokens;            
            else
            {
                var tokens = UserInfo.GetTokens(UserId);
                return (tokens > MinTokens || UserInfo.GetDayTokensGroup(GroupId, UserId) > -MaxTokensDay) && tokens > MaxTokens;
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

            var agentGuid = Agent.GetWhere<Guid>("Guid", $"Name = {agentName.Quotes()} and private <> 2");
            if (agentGuid == Guid.Empty)
                return false;

            CurrentAgent = await Agent.LoadAsync(agentGuid);
            IsCallAgent = true;        
            CmdPara = cmdPara;
            return true;
        }

        //变身
        public async Task<string> ChangeAgentAsync()
        {
            IsCancelProxy = true;
            CurrentAgent = await Agent.LoadAsync(User.AgentId);            
            var agentName = CurrentAgent.Name == "早喵" ? "" : $"【{CurrentAgent.Name}】";
            if (CmdPara == "")            
                return $"🤖 {agentName}可变身的智能体有:\n{Agent.QueryWhere("Name", $"Id in (select AgentId from {AgentTags.FullName} WHERE TagId = 1)", "usedtimes desc", " {0}")}";
            
            var agentId = Agent.GetIdByName(CmdPara);
            if (agentId != 0)
            {
                IsCallAgent = true;               
                CurrentAgent = await Agent.LoadAsync(agentId);                
                return UserInfo.SetValue("AgentId", agentId, UserId) == -1 
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

            if (IsGroup && !Group.IsAI)
            {
                if (CmdName.In("AI"))
                    Answer = "AI功能已关闭";
                else
                    Reason += "[关闭AI]";
                return;
            }

            if (!User.IsAI)
            {
                if (IsAtMe || !IsGroup || IsPublic)                
                    Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[禁用AI]";
                return;
            }

            if (!IsEnough())
            {
                if (IsAtMe || !IsGroup || IsPublic)
                    Answer = $"你的算力已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[无算力]";
                return;
            }

            if (User.Credit <= 0)
            {
                if (IsAtMe || !IsGroup || IsPublic)
                    Answer = $"你的积分已用完。请每日签到获得或联系客服购买";
                else
                    Reason += "[负分]";
                return;
            }
            
            CurrentAgent = await Agent.LoadAsync(User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : User.AgentId);

            if (IsAgent && CmdPara == "结束")
            {               
                Answer = $"✅ 已结束与智能体【{CurrentAgent.Name}】的对话";
                UserInfo.SetValue("AgentId", AgentInfos.DefaultAgent.Id, UserId);
                return;
            }

            IsAI = true;            

            GetChatHistory();

            (ModelId, var providerName, var modelName) = LLMModel.GetModelInfo(CurrentAgent.ModelId);
            var provider = LLMApp._manager.GetProvider(providerName ?? "Doubao");
            if (provider != null)
            {
                AnswerAI = await provider.ExecuteAsync(History, modelName);
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
                Debug(Answer);
                return;
            }

            BatchInsertAgent();

            if (IsGuild && IsGroup && !User.IsAI)
            {
                var credit = TokensMinus;
                UserInfo.MinusCredit(SelfId, GroupId, GroupName, UserId, Name, credit, "使用AI");
            }        
        }

        public async Task GenerateImageAsync()
        {
            long tokens = UserInfo.GetTokens(UserId);
            if (tokens < 12000)
                Answer = $"生图功能每张图需消耗12000算力单位，您的算力({tokens})不足";
            else
            {
                if (IsPublic)
                {
                    var url = $"< a href =\"{Common.url}/ai?t={Token.GetToken(UserId)}";
                    Answer = $"以下地址直接<a href=\"{url}\">进入后台</a>使用文生图:\n{url}";
                }
                else if (CmdPara.IsNull())
                    Answer = $"命令格式：生图 + 提示词\n模型：豆包通用2.1-文生图\n消耗算力：12000单位/图";
                else
                {
                    TokensMinus = 12000;
                    _ = MinusTokensRes("使用生图模型 豆包");
                    Answer = await Doubao.GenerateImageDoubaoAsync(CmdPara);
                    Answer = Answer.IsNull() ? RetryMsg : Answer;
                    IsAI = !Answer.IsNull();

                    //保存到数据库
                    //if (bm.Answer != RetryMsg)
                    //_ = DalleImages.SaveImageAsync(bm, prompt, "", bm.Answer);
                }
            }
        }

        public string MinusTokensRes(string tokensInfo)
        {
            return UserInfo.MinusTokensRes(SelfId, GroupId, GroupName, Group.IsOwnerPay ? Group.RobotOwner : UserId, Name, TokensMinus, $"{tokensInfo} {(Group.IsOwnerPay ? $" 群主付(QQ:{UserId})" : "")}");
        }

        public string BatchInsertAgent()
        {
            OutputTokens = Answer.GetTokensCount();
            TokensMinus = (InputTokens * Agent.tokensTimes + OutputTokens * Agent.tokensTimesOutput) / 2;
            AgentLog.Append(this);
            return MinusTokensRes($"使用AI");
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

        public void GetChatHistory(int his = 3)
        {
            var systemPrompt = GetSystemPrompt();

            int context = IsAgent ? his :Group.ContextCount;            

            if (CurrentAgent.Guid.In(AgentInfos.PromptAgent.Guid, AgentInfos.InfoAgent.Guid)) context = 0;

            var questions = string.Empty;

            if (context > 0)
            {
                var query = $"SELECT TOP {context} Question, CASE WHEN IsAI = 1 THEN AnswerAI ELSE Message END AS Answer, UserName FROM {GroupSendMessage.FullName} " +
                            $"WHERE (AnswerId <> 0 or IsAI = 1) AND GroupId = {GroupId} {(Group.IsMultAI ? "" : $"AND UserId = {UserId}")} " +
                            $"AND ABS(DATEDIFF(HOUR, GETDATE(), InsertDate)) <= 24 ORDER BY Id DESC";
                DataSet ds = QueryDataset(query);

                if (ds != null)
                {
                    foreach (DataRow dr in ds.Tables[0].Rows)
                    {
                        var question = dr["question"].AsString().RemoveUserId(SelfId);
                        var re = BotCmd.GetRegexCmd();

                        if (question.IsMatch(re))
                            (_, question) = GetCmdPara(question, re);

                        questions += question + "\n";

                        var answer = dr["answer"].AsString().RegexReplace(@"\n积分：.*?累计：.*", "");
                        answer = answer.RegexReplace(@"^【\w*】", "");
                        long tokenCount = (question + answer).GetTokensCount();

                        if (InputTokens + tokenCount < Agent.tokensLimit - Agent.tokensOutputLimit)
                        {
                            History.AddAssistantMessage(answer);
                            History.AddUserMessage(question);
                            InputTokens += tokenCount + 4;
                        }
                        else break;
                    }
                    InputTokens += 2;
                }
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
}
