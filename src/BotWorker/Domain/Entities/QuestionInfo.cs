namespace BotWorker.Domain.Entities
{
    public class QuestionInfo : MetaDataGuid<QuestionInfo>
    {
        public override string TableName => "Question";
        public override string KeyField => "Id";

        public long QuestionId { get; set; }
        public string Question { get; set; } = string.Empty;
        public long UserId { get; set; }        
        public long GroupID { get; set; }
        public int CUsed { get; set; }
        public int Audit2 { get; set; }
        public bool IsSystem { get; set; }

        public static bool IsExists(string question)
        {
            return ExistsWhere($"question = {GetNew(question).Quotes()}");
        }

        // 新增问题
        public static long Append(long botUin, long groupId, long qq, string question) => AppendAsync(botUin, groupId, qq, question).GetAwaiter().GetResult();

        public static async Task<long> AppendAsync(long botUin, long groupId, long qq, string question)
        {
            question = GetNew(question);
            if (question.IsNull())
                return 0;
            else
            {
                long questionId = await GetQIdAsync(question);
                if (questionId == 0)
                {
                    if (question.Length < 200)
                    {
                        if (await InsertAsync([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("UserId", qq),
                            new Cov("question", question)
                            ]) == -1)
                            Logger.Error("添加问答问题失败");
                        else
                            questionId = await GetAutoIdAsync(FullName);
                    }
                    else
                        questionId = 0;
                }
                return questionId;
            }
        }

        // 使用次数+1
        public static async Task<int> PlusUsedTimesAsync(long questionId)
        {
            return await PlusAsync("CUsed", 1, questionId);
        }

        // 是否系统问题
        public static bool GetIsSystem(long QuestionId) => GetIsSystemAsync(QuestionId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsSystemAsync(long QuestionId)
        {
            return await GetBoolAsync("IsSystem", QuestionId);
        }

      
        /// 审核完成并升级为系统问题
        public static int Audit(long questionId, int audit2, int isSystem) => AuditAsync(questionId, audit2, isSystem).GetAwaiter().GetResult();

        public static async Task<int> AuditAsync(long questionId, int audit2, int isSystem)
        {            
            return await UpdateAsync($"Audit2 = {audit2}, Audit2Date = {SqlDateTime}, Audit2By = {BotInfo.SystemUid}, IsSystem = {isSystem}", questionId);
        }


        // 学习功能之问题是否存在
        public static long GetQId(string text) => GetQIdAsync(text).GetAwaiter().GetResult();

        public static async Task<long> GetQIdAsync(string text)
        {
            if (text.Length > 200)
                return 0;
            else
                return (await GetWhereAsync(Key, $"question = {text.Quotes()}", "Id")).AsLong();
        }

        // 去掉标点符号、表情 如果全是标点或全是表情则不去掉
        //public static string GetNew(string text)
        //{            
        //    text = text.RemoveWhiteSpaces();
        //    var res = text.RemoveBiaodian();
        //    if (res.IsNull()) return text;

        //    text = res;
        //    res = text.RemoveQqFace();
        //    return res.IsNull() ? text : res;
        //}

        public static string GetNew(string text)
        {
            if (string.IsNullOrWhiteSpace(text))
                return text;

            text = text.RemoveWhiteSpaces();

            // 精准匹配表情
            var faceRegex = @"\[Face\d{1,3}\.gif\]";
            var matches = text.Matches(faceRegex);

            // 替换成不含任何标点的占位符
            int faceIndex = 0;
            var faceDict = new Dictionary<string, string>(); // 占位符 -> 表情
            string temp = text.RegexReplace(faceRegex, m =>
            {
                string key = $"QQFACE{faceIndex}PLACEHOLDER";
                faceDict[key] = m.Value;
                faceIndex++;
                return key;
            });

            // 去标点（你已有的）
            temp = temp.RegexReplace(Regexs.BiaoDian, "");

            // 还原表情
            foreach (var kv in faceDict)
            {
                temp = temp.Replace(kv.Key, kv.Value);
            }

            // 判断是否全为表情或标点
            bool isAllRemoved = temp.RemoveQqFace().RemoveBiaodian().IsNull();
            return isAllRemoved ? text : temp;
        }

    }
}
