using System.Text.RegularExpressions;
using BotWorker.Common.Utily;

namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        private const int MaxDepth = 5;

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
                        ? Agent.GetValue("Info", AgentId)
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

            CmdPara = QuestionInfo.GetNew(CmdPara);
            long qid = IsAtOthers ? 0 : QuestionInfo.GetQId(CmdPara);                     

            if (qid == 0)
            {
                if ((!IsGroup || IsPublic || IsAtMe) && CmdPara.Length < 30)                
                    QuestionInfo.Append(SelfId, RealGroupId, UserId, CmdPara);

                if (IsAtOthers && !IsAtMe)
                {
                    Reason += "[艾特他人]";
                    return;
                }
            }

            if (qid != 0)
            {
                QuestionInfo.PlusUsedTimes(qid);

                if (QuestionInfo.GetIsSystem(qid))
                {
                    IsCmd = true;
                    AnswerId = GetDefaultAnswerAt(qid);
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

                    AnswerId = GetGroupAnswer(GroupId, qid);
                    var length = Group.IsVoiceReply ? 4 : 0;

                    if (AnswerId == 0 && cloud >= 2)
                    {
                        AnswerId = GetDefaultAnswer(qid, length);
                        if (AnswerId != 0)
                            AnswerId = GetDefaultAnswerAt(qid, length);

                        if (AnswerId == 0 && cloud >= 3)
                        {
                            AnswerId = GetAllAnswerAudit(qid, length);
                            if (AnswerId == 0)
                                AnswerId = GetAllAnswerNotAudit(qid, length);

                            if (AnswerId == 0 && cloud >= 4)
                                AnswerId = GetAllAnswer(qid, length);
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

                Answer = AnswerInfo.GetValue("answer", AnswerId);

                // 递归引用 例如：{{客服QQ}}
                await ResolveAnswerRefsAsync();
            }

            if (Answer.Equals("#none", StringComparison.CurrentCultureIgnoreCase))
                Answer = string.Empty;
            else 
            {
                //@机器人或私聊机器人没有答案的使用ai
                if (!CmdPara.IsNull() && Answer.IsNull() && !IsDup && !IsMusic)
                {
                    if ((IsAgent || IsCallAgent || IsAtMe || IsGuild || !IsGroup || IsPublic || (cloud >= 5 && !IsAtOthers)) && !IsWeb)
                    {
                        await GetAgentResAsync();                              
                    }                    
                }
                else
                    Answer = Answer.Replace("??", "  "); //Emoji 表情
            }

            if (AnswerId != 0 && !IsGuild)
                IsCancelProxy = true;

            if (!IsDup) UpdateCountUsed();
        }

        public long GetGroupAnswer(long group, long question, int length = 0)
        {
            //本群 及 系统级答案（audit2=3）
            return group == BotInfo.GroupIdDef
                ? GetDefaultAnswer(question)
                : AnswerInfo.GetWhere("Id", $"QuestionId = {question} AND ABS(audit) = 1 AND ((RobotId = {group} AND audit2 <> -4) OR audit2 = 3) {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}", "NEWID()").AsLong();
        }

        // 官方 
        public long GetDefaultAnswer(long question, int length = 0) =>
            AnswerInfo.GetWhere("Id", $"QuestionId = {question} AND ABS(audit) = 1 AND RobotId = {GroupId} AND audit2 >= 0 {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}", "NEWID()").AsLong();

        // 话痨(官方群+审核升级到默认群内容) audit2 >= 1 (1,2,3) 
        public static long GetDefaultAnswerAt(long question, int length = 0)
        {
            var sql = $"QuestionId = {question} AND ABS(audit) = 1 ";
            sql += $"AND (Id IN (SELECT TOP 20 ID FROM {AnswerInfo.FullName} WHERE QuestionId = {question} AND Audit2 >= 1 ORDER BY ((ISNULL(GoonTimes,0) + 1)/(ISNULL(UsedTimes,0) + 1)) DESC)";
            sql += $"OR Id IN (SELECT TOP 10 Id FROM {AnswerInfo.FullName} WHERE QuestionId = {question} AND Audit2 >= 1 AND UsedTimes < 100 ORDER BY UsedTimes DESC)) {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}";
            var res = AnswerInfo.GetWhere("Id", sql, "NEWID()").AsLong();
            return res;
        }

        //终极
        public static long GetAllAnswerAudit(long question, int length = 0) =>
            AnswerInfo.GetWhere("Id", $"QuestionId = {question} AND ABS(audit) = 1 AND audit2 >= 0 {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}", "NEWID()").AsLong();

        //终极+1 
        public static long GetAllAnswerNotAudit(long question, int length = 0) =>
            AnswerInfo.GetWhere("Id", $"QuestionId = {question} AND ABS(audit) = 1 AND audit2 >= -1 {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}", "NEWID()").AsLong();

        public static long GetAllAnswer(long question, int length = 0) =>
            AnswerInfo.GetWhere("Id", $"QuestionId = {question} AND ABS(audit) = 1 AND audit2 >= -2 {(length > 0 ? $" AND LEN(answer) >= {length}" : "")}", "NEWID()").AsLong();

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

                    AnswerInfo.CountUsedPlus(AnswerId);

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

            long creditValue = GetCredit();
            if (creditValue < 0)
                return $"您已负分（{creditValue}），不能教我说话了";

            ans = ans.Replace("｛", "{").Replace("｝", "}");

            if (ans == "")
                return "答案不能为空（图片无效）";

            long questionId = QuestionInfo.Append(SelfId, RealGroupId, UserId, que);
            if (questionId == 0)
                return "问题不能为空（图片无效）";

            string refInfo = "";
            if (!IsSuperAdmin)
            {
                if (QuestionInfo.GetIsSystem(questionId) || QuestionInfo.GetBool("IsLock", questionId))
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
                    var refId = QuestionInfo.GetQId(ans);
                    var countAnswer = QuestionInfo.GetInt("CAnswer", refId);

                    if (countAnswer > 0)
                        QuestionInfo.SetValue("audit2", 1, refId);

                    refInfo = $"{ans} 答案数：{countAnswer}/{QuestionInfo.GetInt("CAnswerAll", refId)}";
                    ans = $"{{{{{ans}}}}}";

                    if (AnswerInfo.ExistsAandB("QuestionId", questionId, "answer", ans))
                        return $"{AnswerExists}\n{refInfo}";
                }
            }

            if (AnswerInfo.Exists(questionId, ans, GroupId))
                return AnswerExists;

            (int audit, int audit2, int minus, res) = GetAudit(questionId, que, ans);

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 添加答案记录
                var (sql1, paras1) = AnswerInfo.SqlAppend(SelfId, RealGroupId, UserId, GroupId, questionId, que, ans, audit, -minus, audit2, "");
                await ExecAsync(sql1, trans, paras1);

                // 2. 通用加积分函数 (含日志记录)
                var addRes = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -minus, minus < 0 ? "教学加分" : "教学扣分", trans);
                if (addRes.Result == -1) throw new Exception("更新积分失败");
                creditValue = addRes.CreditValue;

                await trans.CommitAsync();

                if (!IsGroup)
                    res += $"\n默认群：{GroupId}";

                QuestionInfo.Update($"CAnswer = CAnswer + 1, CAnswerAll = CAnswerAll + 1", questionId);

                return $"{res}\n💎 积分：{-minus}, 累计：{creditValue}\n{refInfo}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[AppendAnswer Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 新增答案
        public string AppendAnswer(string que, string ans) => AppendAnswerAsync(que, ans).GetAwaiter().GetResult();

        // 判断用户提交问题 审核广告/脏话/说话语气等

        public (int, int, int, string) GetAudit(long questionId, string textQuestion, string textAnswer)
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

            if (audit2 > -3 &&  AnswerInfo.Exists(SelfId, questionId, textAnswer))
            {
                audit2 = -3;
            }

            int c_answer = QuestionInfo.GetInt("CAnswer", questionId);
            int c_answer_all = QuestionInfo.GetInt("CAnswerAll", questionId);
            int c_used = QuestionInfo.GetInt("CUsed", questionId);

            if (audit2 > -3 && c_used > 1 && c_answer <= 2 && c_answer_all < 20 && textQuestion.Length < 10 && textAnswer.Length < 20 && textQuestion != textAnswer)
            {
                minus = -10;
            }

            string res = audit < -1 || audit2 == -4 ? "✅ 教学成功，请等待群主审核" : "✅ 教学成功，谢谢您！";
            return (audit, audit2, minus, res);
        }


        //更新答案使用次数
        public void UpdateCountUsed()
        {
            if (AnswerId == 0) return;

            var lastId = UserInfo.GetLong("AnswerId", UserId);

            if (AnswerId != lastId)
            {
                AnswerInfo.CountUsedPlus(AnswerId);

                if (UserInfo.GetInt("ABS(DATEDIFF(MINUTE, GETDATE(), AnswerDate))", UserId) <= 5)
                    AnswerInfo.Plus("GoonTimes", 1, lastId);

                UserInfo.Update($"AnswerId = {AnswerId}, AnswerDate = GETDATE()", UserId);
            }

            lastId = GroupInfo.GetLong("LastAnswerId", UserId);

            if (AnswerId != lastId)
            {
                AnswerInfo.Plus("UsedTimesGroup", 1, AnswerId);

                if (GroupInfo.GetInt("ABS(DATEDIFF(MINUTE, GETDATE(), LastAnswerDate))", GroupId) <= 5)
                    AnswerInfo.Plus("GoonTimesGroup", 1, lastId);

                GroupInfo.Update($"LastAnswerId = {AnswerId}, LastAnswer = {Answer.Quotes()}, LastAnswerDate = GETDATE()", GroupId);
            }
        }

        public void GetAnswer()
        {
            Answer = AnswerInfo.GetValue("answer", AnswerId);
            UpdateCountUsed();
        }

        //笑话
        public string GetJokeRes()
        {
            AnswerId = AnswerInfo.GetWhere("Id", $"QuestionId = 2303 AND ABS(audit) = 1 AND audit2 >= 0", "NEWID()").AsLong();
            GetAnswer();
            return Answer;
        }

        //故事
        public void GetStory()
        {
            AnswerId = AnswerInfo.GetWhere("Id", $"QuestionId IN (50701, 545) AND LEN(answer) > 40 AND ABS(audit) = 1 AND audit2 >= 0 ", "NEWID()").AsLong();
            GetAnswer();
        }

        //鬼故事
        public void GetGhostStory()
        {
            AnswerId = AnswerInfo.GetWhere("Id",
                $"QuestionId IN (SELECT Id FROM {QuestionInfo.FullName} WHERE question like '鬼故事%') " +
                $"AND LEN(answer) > 40 AND ABS(audit) = 1 AND audit2 > -3", "NEWID()").AsLong();
            GetAnswer();
            Answer = $"✅ 鬼故事\n{Answer}" + MinusCreditRes(10, "鬼故事扣分");
        }

        // 对联
        public void GetCouplets()
        {
            AnswerId = AnswerInfo.GetWhere("Id", $"QuestionId IN (SELECT Id FROM {QuestionInfo.FullName} WHERE question LIKE '%对联%') " +
                                   $"AND LEN(answer) > 12 AND ABS(audit) = 1 AND audit2 > -3 ", "NEWID()").AsLong();
            GetAnswer();
            Answer = $"✅ 对联\n{Answer}" + MinusCreditRes(10, "对联扣分");           
        }

        /// 抽签
        public void GetChouqian()
        {
            var sql = $"SELECT TOP 1 Id FROM {AnswerInfo.FullName} WHERE RobotId = 286946883 and QuestionId = 225781 AND AUDIT2 > 0" +
                      $"ORDER BY NEWID()";
            AnswerId = QueryScalar<long>(sql);
            GetAnswer();
            Answer = $"✅ {Answer}\n✨ 古签藏玄意，早喵见真机。\n发送【解签】为你精准解读";
        }

        /// 解签
        public void GetJieqian()
        {
            var sql = $"SELECT TOP 1 AnswerId FROM {GroupSendMessage.FullName} " +
                      $"WHERE GroupId = {GroupId} AND UserId = {UserId} " +
                      $"AND AnswerId IN (SELECT Id FROM {AnswerInfo.FullName} WHERE RobotId = 286946883 and QuestionId = 225781)" +
                      $"ORDER BY Id DESC";
            var answerId = QueryScalar<long>(sql);
            if (answerId != 0)
            {
                AnswerId = AnswerInfo.GetWhere("Id", $"parentanswer = {answerId}").AsLong();
                GetAnswer();
                Answer = Answer.StripMarkdown();
            }
        }

        public const long group_dati = 453174086; //客户群 红楼梦

        // 答题
        public async Task<BotMessage> GetDatiAsync(BotMessage bm)
        {
            if (bm.CmdName == "答案")
            {
                long answerId = UserInfo.GetLong("AnswerId", bm.UserId);
                string question = AnswerInfo.GetValue("question", answerId);
                if (question.IsMatch(Regexs.Dati.Replace("$", "\\d*答案")))
                {
                    bm.AnswerId = answerId;
                    bm.GetAnswer();
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
                long answerId = AnswerInfo.GetWhere("Id", $"RobotId = {group_dati} AND question LIKE '%{bm.CmdName}%' AND question NOT LIKE '%答案%' AND ABS(audit) = 1 AND audit2 <> -4", "NEWID()").AsLong();
                bm.AnswerId = answerId;
                bm.GetAnswer();
                bm.Answer = $"{bm.Answer} ——查答案发送【{GetValue("question", answerId)}答案】或【答案】";
            }
            return bm;
        }
}
