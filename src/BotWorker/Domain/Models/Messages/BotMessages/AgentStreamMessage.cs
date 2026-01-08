using sz84.Agents.Entries;
using sz84.Agents.Plugins;
using sz84.Agents.Providers;
using sz84.Bots.Groups;
using sz84.Bots.Users;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 接收客户端的问题并处理
        public async Task StartStreamChatAsync(CancellationToken cts)
        {
            RealGroupId = GroupId;   
            CurrentMessage = Message;
            
            if (!IsAgent && !IsCallAgent)
            {
                Group.Id = AgentInfos.DefaultAgent.GroupId;  
                await HandleEventAsync();                
                if (!Answer.IsNull())
                {
                    await SendMessageAsync();
                    return;
                }

                IsSend = true;
            }

            if (!IsCallAgent)
                CurrentAgent = await Agent.LoadAsync(AgentId);

            CmdPara = Message;

            if (IsAgent && CmdPara == "结束")
            {
                Answer = $"✅ 已结束与智能体【{CurrentAgent.Name}】的对话";
                UserInfo.SetValue("AgentId", AgentInfos.DefaultAgent.Id, UserId);
                await SendMessage();
                return;
            }

            // 2. 算力检测
            if (!IsEnough())
            {
                Answer = $"您的算力已用完。请每日签到获取算力或联系客服购买。客服QQ:{{客服QQ}}。"; 
                await SendMessage();
                return;
            }

            // 3. 加载聊天历史
            GetChatHistory(HistoryMessageCount);            

            IsAI = true; 

            try
            {
                (ModelId, var providerName, var modelId) = LLMModel.GetModelInfo(CurrentAgent.ModelId);
                var provider = LLMApp._manager.GetProvider(providerName ?? "Doubao");
                if (provider == null)
                {
                    Answer = "模型提供者不存在";
                    await SendMessage();
                    ErrorMessage(Answer);
                    return;
                }

                // 4. 发送流开始事件                            
                await StreamBegin(cts);

                if (IsCallAgent || (!IsWeb && IsAgent))
                {
                    var nickName = IsCallAgent ? $"【{CurrentAgent.Name}】" : AgentName.IsNull() ? "" : $"【{AgentName}】";                    
                    Answer += nickName;
                    await Stream(nickName, cts);
                }
                                
                if (Group.IsUseKnowledgebase && KbService != null)
                {
                    var pluginKnowledge = new KnowledgeBasePlugin(KbService, GroupId);
                    var plugins = new[] { pluginKnowledge };

                    await provider.StreamExecuteAsync(History, modelId, async (data, isStreaming, token) =>
                    {
                        token.ThrowIfCancellationRequested();
                        await Stream(data, token);
                        AnswerAI += data;
                        Answer += data;
                    }, plugins, this, cts);
                }
                else
                {
                    await provider.StreamExecuteAsync(History, modelId, async (data, isStreaming, token) =>
                    {
                        token.ThrowIfCancellationRequested();
                        await Stream(data, token);
                        AnswerAI += data;
                        Answer += data;
                    }, cts);
                }
            }
            catch (OperationCanceledException)
            {
                await Stream("[已取消]", cts);
            }
            finally
            {
                await StreamEnd(cts);
            }

            // 6. 保存数据
            BatchInsertAgent();

            // 👇 停止计时并记录耗时
            CurrentStopwatch?.Stop();
            CostTime = CurrentStopwatch is null ? 0 : CurrentStopwatch.Elapsed.TotalSeconds;

            GroupSendMessage.Append(this);
        }

        private async Task SendMessage()
        {
            if (Answer.IsNull()) return;         
            await GetFriendlyResAsync();
            GroupSendMessage.Append(this);
            if (ReplyMessageAsync == null) return;
            await ReplyMessageAsync();
        }

        private async Task StreamBegin(CancellationToken cts = default)
        {
            if (ReplyStreamBeginMessageAsync == null) return;
            await ReplyStreamBeginMessageAsync(cts);
        }

        private async Task Stream(string data, CancellationToken cts = default)
        {
            if (ReplyStreamMessageAsync == null) return;
            await ReplyStreamMessageAsync(data, cts);
        }

        private async Task StreamEnd(CancellationToken cts = default)
        {
            if (ReplyStreamEndMessageAsync == null) return;
            await ReplyStreamEndMessageAsync(cts);
        }
    }
}
