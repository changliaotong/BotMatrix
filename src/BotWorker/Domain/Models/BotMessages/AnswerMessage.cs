using System.Text.RegularExpressions;
using BotWorker.Common.Utily;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        private const int MaxDepth = 5;

        public async Task<string> GetQaAnswerAsync(string question)
        {
            if (string.IsNullOrWhiteSpace(question)) return string.Empty;

            var oldCmdPara = CmdPara;
            var oldAnswer = Answer;
            var oldAnswerId = AnswerId;
            var oldIsCmd = IsCmd;
            var oldIsCancelProxy = IsCancelProxy;

            try
            {
                CmdPara = QuestionInfoService.GetNew(question);
                long qid = await QuestionInfoService.GetIdByQuestionAsync(CmdPara);

                if (qid == 0) return string.Empty;

                await QuestionInfoService.IncrementUsedTimesAsync(qid);

                int cloud = !IsGroup || IsGuild ? 5 : !User.IsShutup ? Group.IsCloudAnswer : 0;
                long ansId = 0;

                if (await QuestionInfoService.IsSystemAsync(qid))
                {
                    ansId = await GetDefaultAnswerAtAsync(qid);
                }
                else
                {
                    ansId = await GetGroupAnswerAsync(GroupId, qid);
                    if (ansId == 0 && cloud >= 2)
                    {
                        ansId = await GetDefaultAnswerAsync(qid);
                        if (ansId != 0) ansId = await GetDefaultAnswerAtAsync(qid);

                        if (ansId == 0 && cloud >= 3)
                        {
                            ansId = await GetAllAnswerAuditAsync(qid);
                            if (ansId == 0) ansId = await GetAllAnswerNotAuditAsync(qid);
                            if (ansId == 0 && cloud >= 4) ansId = await GetAllAnswerAsync(qid);
                        }
                    }
                }

                if (ansId == 0) return string.Empty;

                AnswerId = ansId;
                Answer = await AnswerRepository.GetValueAsync<string>("answer", ansId);
                await ResolveAnswerRefsAsync();
                
                string result = Answer;
                if (result.Equals("#none", StringComparison.CurrentCultureIgnoreCase))
                    result = string.Empty;
                else
                    result = result.Replace("??", "  ");

                return result;
            }
            finally
            {
                CmdPara = oldCmdPara;
                Answer = oldAnswer;
                AnswerId = oldAnswerId;
                IsCmd = oldIsCmd;
                IsCancelProxy = oldIsCancelProxy;
            }
        }

        public async Task GetAnswerAsync()
        {
            if (CurrentMessage.StartsWith("[自动回复]"))
            {
                Reason += "[自动回复]";
                return;
            }

            if (IsRefresh)
            {
                Reason += "[刷屏]";
                return;
            }

            if (User.CreditTotal < -5000)
            {
                Reason += "[负分]";
                return;
            }

            int cloud = !IsGroup || IsGuild ? 5 : !User.IsShutup ? Group.IsCloudAnswer : 0;

            if (User.IsAI && Group.IsAI && cloud != 0)
            {
                AgentId = User.AgentId == 0 ? AgentInfos.DefaultAgent.Id : User.AgentId;
                
                if (IsAgent)
                {
                    if (!IsWeb) 
                        await GetAgentResAsync();
                    return;
                }
            }

            if (!IsDup && !IsMusic)            
                CmdPara = Message;            

            if (IsReply)
                CmdPara = CmdPara.RemoveAt();

            CmdPara = CmdPara.RemoveQqImage();
            CmdPara = CmdPara.RemoveUserId(SelfId);
            var newPara = CmdPara.RemoveUserIds();
            if (!newPara.IsNull())
                CmdPara = newPara;

            if (CmdPara == "图片") CmdPara = "图片系统";

            if (CmdPara.IsNull())
            {
                if (IsAtMe)
                {
                    Answer = IsAgent || IsCallAgent
                        ? await AgentRepository.GetValueAsync("Info", AgentId)
                        : "我在~";
                    IsCancelProxy = true;
                    return;
                }

                if (SelfId.In(2195307828))
                    return;

                if ((IsImage || IsFlashImage) && Group.IsReplyImage && cloud >= 3)
                {
                    CmdPara = "图片";
                    //RecallAfterMs = 15000;
                }
            }            

            CmdPara = QuestionInfoService.GetNew(CmdPara);
            long qid = IsAtOthers ? 0 : await QuestionInfoService.GetIdByQuestionAsync(CmdPara);                     

            if (qid == 0)
            {
                if ((!IsGroup || IsPublic || IsAtMe) && CmdPara.Length < 30)                
                    await QuestionInfoService.AddQuestionAsync(SelfId, RealGroupId, UserId, CmdPara);

                if (IsAtOthers && !IsAtMe)
                {
                    Reason += "[艾特他人]";
                    return;
                }
            }

            if (qid != 0)
            {
                await QuestionInfoService.IncrementUsedTimesAsync(qid);

                if (await QuestionInfoService.IsSystemAsync(qid))
                    {
                        IsCmd = true;
                        AnswerId = await GetDefaultAnswerAtAsync(qid);
                    }
                else if (!IsAgent && cloud != 6)
                {
                    if (cloud == 0)
                    {
                        Reason += "[闭嘴模式]";
                        return;
                    }

                    if (IsGroup && IsAtMe && User.IsShutup)
                    {
                        Answer = "请先发：关闭 闭嘴模式";
                        return;
                    }

                    var userCount = GroupSendMessage.UserCount(GroupId);

                    if (IsAtMe || (!IsAtOthers && userCount == 1))
                    {
                        cloud = cloud switch
                        {
                            1 => 1,
                            2 => 4,
                            3 => 4,
                            4 => 4,                            
                            5 => 5,
                            6 => 6,
                            _ => 5
                        };
                    }

                    if (IsGuild)
                    {
                        Answer = NoAnswer;
                        return;
                    }

                    AnswerId = await GetGroupAnswerAsync(GroupId, qid);
                    var length = Group.IsVoiceReply ? 4 : 0;

                    if (AnswerId == 0 && cloud >= 2)
                    {
                        AnswerId = await GetDefaultAnswerAsync(qid, length);
                        if (AnswerId != 0)
                            AnswerId = await GetDefaultAnswerAtAsync(qid, length);

                        if (AnswerId == 0 && cloud >= 3)
                        {
                            AnswerId = await GetAllAnswerAuditAsync(qid, length);
                            if (AnswerId == 0)
                                AnswerId = await GetAllAnswerNotAuditAsync(qid, length);

                            if (AnswerId == 0 && cloud >= 4)
                                AnswerId = await GetAllAnswerAsync(qid, length);
                        }
                    }

                    if (AnswerId == 0)
                    {
                        if (User.IsShutup)
                            Reason += "[闭嘴]";     
                        
                        Reason += Group.IsCloudAnswer switch
                        {
                            1 => "[本群模式]",
                            2 => "[官方模式]",
                            3 => "[话痨模式]",
                            4 => "[终极模式]",
                            _ => "",
                        };
                    }
                }

                Answer = await AnswerRepository.GetValueAsync<string>("answer", AnswerId);

                // 递归引用 例如：{{客服QQ}}
                await ResolveAnswerRefsAsync();
            }

            if (Answer.Equals("#none", StringComparison.CurrentCultureIgnoreCase))
                Answer = string.Empty;
            else 
            {
                Answer = Answer.Replace("??", "  "); //Emoji 表情
            }

            if (AnswerId != 0 && !IsGuild)
                IsCancelProxy = true;

            if (!IsDup) await UpdateCountUsedAsync();
        }

        public async Task<long> GetGroupAnswerAsync(long group, long question, int length = 0)
        {
            //本群 及 系统级答案（audit2=3）
            return group == BotInfo.GroupIdDef
                ? await GetDefaultAnswerAsync(question, length)
                : await AnswerRepository.GetGroupAnswerIdAsync(group, question, length);
        }

        // 官方 
        public async Task<long> GetDefaultAnswerAsync(long question, int length = 0) =>
            await AnswerRepository.GetDefaultAnswerIdAsync(question, GroupId, length);

        // 话痨(官方群+审核升级到默认群内容) audit2 >= 1 (1,2,3) 
        public async Task<long> GetDefaultAnswerAtAsync(long question, int length = 0)
        {
            return await AnswerRepository.GetDefaultAnswerAtIdAsync(question, length);
        }

        //终极
        public async Task<long> GetAllAnswerAuditAsync(long question, int length = 0) =>
            await AnswerRepository.GetAllAnswerAuditIdAsync(question, length);

        //终极+1 
        public async Task<long> GetAllAnswerNotAuditAsync(long question, int length = 0) =>
            await AnswerRepository.GetAllAnswerNotAuditIdAsync(question, length);

        public async Task<long> GetAllAnswerAsync(long question, int length = 0) =>
            await AnswerRepository.GetAllAnswerIdAsync(question, length);

        public async Task ResolveAnswerRefsAsync(int depth = 0)
        {
            if (depth > MaxDepth || string.IsNullOrWhiteSpace(Answer))
                return;

            string result = Answer; // 用于构建最终的替换后内容

            if (Answer.IsMatch(Regexs.QuestionRef))
            {
                foreach (Match match in Answer.Matches(Regexs.QuestionRef))
                {
                    var refQuestion = match.Groups["question"].Value;
                    if (string.IsNullOrWhiteSpace(refQuestion))
                        continue;

                    string placeholder = $"{{{{{refQuestion}}}}}";

                    var oldCmdPara = CmdPara;
                    var oldIsDup = IsDup;
                    CmdPara = refQuestion;                    
                    IsDup = true;

                    var answerBackup = Answer;

                    await GetAnswerAsync();

                    string resolved = Answer;

                    await AnswerRepository.IncrementUsedTimesAsync(AnswerId);

                    CmdPara = oldCmdPara;
                    Answer = answerBackup;
                    IsDup = oldIsDup;

                    result = result.Replace(placeholder, resolved);
                }

                Answer = result;

                if (Answer.IsMatch(Regexs.QuestionRef))
                    await ResolveAnswerRefsAsync(depth + 1);
            }
        }

        public async Task GetNewQuestionIdAsync()
        {
            if (KbService == null) return;

            var qaService = new QueryAnswerService(KbService);
            (NewQuestionId, Similarity, NewQuestion) = await qaService.GetTargetQuestionAsync(CmdPara);
        }

        // 新增答案 (异步事务重构版)
        public async Task<string> AppendAnswerAsync(string que, string ans)
        {
            string res = SetupPrivate(teachRight: true);
            if (res != "")
                return res;

            long creditValue = await GetCreditAsync();
            if (creditValue < 0)
                return $"您已负分（{creditValue}），不能教我说话了";

            ans = ans.Replace("｛", "{").Replace("｝", "}");

            if (ans == "")
                return "答案不能为空（图片无效）";

            long questionId = await QuestionInfoService.AddQuestionAsync(SelfId, RealGroupId, UserId, que);
            if (questionId == 0)
                return "问题不能为空（图片无效）";

            string refInfo = "";
            if (!IsSuperAdmin)
            {
                if (await QuestionInfoService.IsSystemAsync(questionId) || await QuestionInfoRepository.GetValueAsync<bool>("IsLock", questionId))
                    return AnswerExists;

                if (que.Length > 30)
                    return "问题不能超过30字";

                if (ans.Length > 300)
                    return "答案不能超过300字";
            }
            else
            {
                if (ans.StartsWith('#'))
                {
                    ans = ans[1..];
                    var refId = await QuestionInfoService.GetIdByQuestionAsync(ans);
                    var countAnswer = await QuestionInfoRepository.GetValueAsync<int>("CAnswer", refId);

                    if (countAnswer > 0)
                        await QuestionInfoRepository.SetValueAsync("audit2", 1, refId);

                    refInfo = $"{ans} 答案数：{countAnswer}/{await QuestionInfoRepository.GetValueAsync<int>("CAnswerAll", refId)}";
                    ans = $"{{{{{ans}}}}}";

                    if (await AnswerRepository.CountAsync($"QuestionId = {questionId} AND answer = @ans", new { ans }) > 0)
                        return $"{AnswerExists}\n{refInfo}";
                }
            }

            if (await AnswerRepository.ExistsAsync(questionId, ans, GroupId))
                return AnswerExists;

            (int audit, int audit2, int minus, res) = await GetAuditAsync(questionId, que, ans);

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取并锁定积分
                creditValue = await UserRepository.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < 0)
                {
                    await wrapper.RollbackAsync();
                    return $"您已负分（{creditValue}），不能教我说话了";
                }

                // 2. 添加答案记录
                await AnswerRepository.AppendAsync(SelfId, RealGroupId, UserId, GroupId, questionId, que, ans, audit, -minus, audit2, "", wrapper.Transaction);

                // 3. 通用加积分函数 (含日志记录)
                var addRes = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -minus, minus < 0 ? "教学加分" : "教学扣分", wrapper.Transaction);
                if (addRes.Result == -1) throw new Exception("更新积分失败");
                creditValue = addRes.CreditValue;

                await wrapper.CommitAsync();

                // 4. 同步缓存
                await UserRepository.SyncCreditCacheAsync(SelfId, GroupId, UserId, creditValue);

                if (!IsGroup)
                    res += $"\n默认群：{GroupId}";

                await QuestionInfoRepository.UpdateAsync($"CAnswer = CAnswer + 1, CAnswerAll = CAnswerAll + 1", questionId);

                return $"{res}\n💎 积分：{-minus}, 累计：{creditValue}\n{refInfo}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[AppendAnswer Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 判断用户提交问题 审核广告/脏话/说话语气等
        public async Task<(int, int, int, string)> GetAuditAsync(long questionId, string textQuestion, string textAnswer)
        {
            int audit = -1;
            int audit2 = 0;
            int minus = 10;
            string msg = $"{textQuestion} {textAnswer}";

            if (!User.IsTeach || msg.IsMatch(Regexs.DirtyWords))
            {
                audit = -4;
                audit2 = -4;
                minus = 100;
            }
            else if (textAnswer.HaveUserId() || textAnswer.IsMatch(Regexs.AdWords) || textAnswer.ContainsURL())
            {
                audit = -3;
                audit2 = -3;
                minus = 50;
            }
            else if (textQuestion.HaveUserId())
                audit2 = -3;

            if ((IsRobotOwner() || IsWhiteList()) && audit2 != -4)
            {
                audit = 1;
                if (audit != -1)
                    audit2 = -3;
            }

            if (IsSuperAdmin)
            {
                audit = 1;
                audit2 = 2;
            }

            if (audit2 > -3 && await AnswerRepository.ExistsAsync(SelfId, questionId, textAnswer))
            {
                audit2 = -3;
            }

            int c_answer = await QuestionInfoRepository.GetValueAsync<int>("CAnswer", questionId);
            int c_answer_all = await QuestionInfoRepository.GetValueAsync<int>("CAnswerAll", questionId);
            int c_used = await QuestionInfoRepository.GetValueAsync<int>("CUsed", questionId);

            if (audit2 > -3 && c_used > 1 && c_answer <= 2 && c_answer_all < 20 && textQuestion.Length < 10 && textAnswer.Length < 20 && textQuestion != textAnswer)
            {
                minus = -10;
            }

            string res = audit < -1 || audit2 == -4 ? "✅ 教学成功，请等待群主审核" : "✅ 教学成功，谢谢您！";
            return (audit, audit2, minus, res);
        }


        //更新答案使用次数
        public async Task UpdateCountUsedAsync()
        {
            if (AnswerId == 0) return;

            var lastId = await UserRepository.GetValueAsync<long>("AnswerId", UserId);

            if (AnswerId != lastId)
            {
                await AnswerRepository.IncrementUsedTimesAsync(AnswerId);

                var timeDiff = await UserRepository.ExecuteScalarAsync<int>($"SELECT ABS({SqlDateDiff("MINUTE", SqlDateTime, "AnswerDate")}) FROM {UserRepository.TableName} WHERE Id = @id", new { id = UserId });
                if (timeDiff <= 5)
                    await AnswerRepository.IncrementValueAsync("GoonTimes", 1, lastId);

                await UserRepository.UpdateAsync($"AnswerId = {AnswerId}, AnswerDate = {SqlDateTime}", UserId);
            }

            lastId = await GroupRepository.GetValueAsync<long>("LastAnswerId", GroupId);

            if (AnswerId != lastId)
            {
                await AnswerRepository.IncrementValueAsync("UsedTimesGroup", 1, AnswerId);

                var timeDiffGroup = await GroupRepository.ExecuteScalarAsync<int>($"SELECT ABS({SqlDateDiff("MINUTE", SqlDateTime, "LastAnswerDate")}) FROM {GroupRepository.TableName} WHERE Id = @id", new { id = GroupId });
                if (timeDiffGroup <= 5)
                    await AnswerRepository.IncrementValueAsync("GoonTimesGroup", 1, lastId);

                await GroupRepository.UpdateAsync($"LastAnswerId = {AnswerId}, LastAnswer = {Answer.Quotes()}, LastAnswerDate = {SqlDateTime}", GroupId);
            }
        }

        public async Task GetAnswerAsync(long answerId = 0)
        {
            if (answerId != 0) AnswerId = answerId;
            Answer = await AnswerRepository.GetValueAsync<string>("answer", AnswerId);
            await UpdateCountUsedAsync();
        }

        //笑话

        public async Task<string> GetJokeResAsync()
        {
            return await GetQaAnswerAsync("笑话");
        }

        //故事
        public async Task GetStoryAsync()
        {
            AnswerId = await AnswerRepository.GetStoryIdAsync();
            await GetAnswerAsync();
        }

        //鬼故事
        public async Task GetGhostStoryAsync()
        {
            AnswerId = await AnswerRepository.GetGhostStoryIdAsync();
            await GetAnswerAsync();
            await ResolveAnswerRefsAsync();
            if (!string.IsNullOrEmpty(Answer))
            {
                Answer = $"✅ 鬼故事\n{Answer}" + MinusCreditRes(10, "鬼故事扣分");
            }
        }

        // 对联
        public async Task GetCoupletsAsync()
        {
            AnswerId = await AnswerRepository.GetCoupletsIdAsync();
            await GetAnswerAsync();
            await ResolveAnswerRefsAsync();
            if (!string.IsNullOrEmpty(Answer))
            {
                Answer = $"✅ 对联\n{Answer}" + MinusCreditRes(10, "对联扣分");
            }
        }

        /// 抽签
        public async Task GetChouqianAsync()
        {
            AnswerId = await AnswerRepository.GetChouqianIdAsync();
            await GetAnswerAsync();
            await ResolveAnswerRefsAsync();
            if (!string.IsNullOrEmpty(Answer))
            {
                Answer = $"✅ {Answer}\n✨ 古签藏玄意，早喵见真机。\n发送【解签】为你精准解读";
            }
        }

        /// 解签
        public async Task GetJieqianAsync()
        {
            var answerId = await AnswerRepository.GetJieqianAnswerIdAsync(GroupId, UserId);
            if (answerId != 0)
            {
                AnswerId = await AnswerRepository.GetAnswerIdByParentAsync(answerId);
                await GetAnswerAsync();
                await ResolveAnswerRefsAsync();
                if (!string.IsNullOrEmpty(Answer))
                {
                    Answer = Answer.StripMarkdown();
                }
            }
        }

        public const long group_dati = 453174086; //客户群 红楼梦

        // 答题
        public async Task<BotMessage> GetDatiAsync(BotMessage bm)
        {
            if (bm.CmdName == "答案")
            {
                long answerId = await UserRepository.GetValueAsync<long>("AnswerId", bm.UserId);
                string question = await AnswerRepository.GetValueAsync<string>("question", answerId);
                if (question.IsMatch(Regexs.Dati.Replace("$", "\\d*答案")))
                {
                    bm.AnswerId = answerId;
                    await bm.GetAnswerAsync();
                }
                else
                {
                    if (question.IsMatch(Regexs.Dati.Replace("$", "\\d*")))
                    {
                        bm.Message = question + bm.CmdName;
                        bm.IsDup = true;
                        await bm.GetAnswerAsync();
                    }
                    else
                        await bm.GetAnswerAsync();
                }
            }
            else
            {
                long answerId = await AnswerRepository.GetDatiIdAsync(bm.CmdName);
                bm.AnswerId = answerId;
                await bm.GetAnswerAsync();
                bm.Answer = $"{bm.Answer} ——查答案发送【{await AnswerRepository.GetValueAsync<string>("question", answerId)}答案】或【答案】";
            }
            return bm;
        }
}
