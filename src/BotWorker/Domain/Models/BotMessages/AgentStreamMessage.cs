using BotWorker.Modules.AI.Plugins;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Domain.Models.BotMessages;

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
                CurrentAgent = await Agent.LoadAsync(AgentId) ?? new();

            CmdPara = Message;

            if (IsAgent && CmdPara == "结束")
            {
                Answer = $"✅ 已结束与智能体【{CurrentAgent.Name}】的对话";
                UserInfo.SetValue("AgentId", AgentInfos.DefaultAgent.Id, UserId);
                await SendMessageAsync();
                return;
            }

            // 2. 算力检测
            if (!IsEnough())
            {
                Answer = $"您的算力已用完。请每日签到获取算力或联系客服购买。客服QQ:{BotInfo.CrmUin}。"; 
                await SendMessageAsync();
                return;
            }

            // 3. 加载聊天历史
            GetChatHistory(HistoryMessageCount);            

            // --- RAG 预检索优化 ---
            if (Group.IsUseKnowledgebase && KbService != null)
            {
                var knowledge = await KbService.BuildPrompt(CurrentMessage, RealGroupId);
                if (!string.IsNullOrEmpty(knowledge))
                {
                    History.AddSystemMessage(knowledge);
                }
            }

            IsAI = true; 

            try
            {
                (ModelId, var providerName, var modelId) = LLMModel.GetModelInfo(CurrentAgent.ModelId);
                var provider = LLMApp?._manager.GetProvider(providerName ?? "Doubao");
                if (provider == null)
                {
                    Answer = "模型提供者不存在";
                    await SendMessageAsync();
                    Logger.Error(Answer);
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
                    var plugins = new Microsoft.SemanticKernel.KernelPlugin[] { pluginKnowledge };

                    var options = new ModelExecutionOptions 
                    { 
                        ModelId = modelId, 
                        Plugins = plugins,
                        CancellationToken = cts 
                    };

                    await foreach (var data in provider.StreamExecuteAsync(History, options).WithCancellation(cts))
                    {
                        await Stream(data, cts);
                        AnswerAI += data;
                        Answer += data;
                    }
                }
                else
                {
                    var options = new ModelExecutionOptions 
                    { 
                        ModelId = modelId, 
                        CancellationToken = cts 
                    };

                    await foreach (var data in provider.StreamExecuteAsync(History, options).WithCancellation(cts))
                    {
                        await Stream(data, cts);
                        AnswerAI += data;
                        Answer += data;
                    }
                }
                await StreamEnd(cts);
                await SendMessageAsync();
            }
            catch (OperationCanceledException)
            {
                await Stream("[已取消]", cts);
                await StreamEnd(cts);
                await SendMessageAsync();
            }
            catch (Exception ex)
            {                
                Logger.Error($"智能体聊天异常: {ex.Message}");
                await Stream($"\n⚠️ 出错了: {ex.Message}", cts);
                await StreamEnd(cts);
                await SendMessageAsync();
            }

            // 6. 保存数据
            BatchInsertAgent();

            // 👇 停止计时并记录耗时
            CurrentStopwatch?.Stop();
            CostTime = CurrentStopwatch is null ? 0 : CurrentStopwatch.Elapsed.TotalSeconds;

            GroupSendMessage.Append(this);
        }

        private async Task SendMessageAsync()
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
