using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("Question")]
    public class QuestionInfo
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string Question { get; set; } = string.Empty;
        public int CUsed { get; set; }
        public int Audit2 { get; set; }
        public DateTime? Audit2Date { get; set; }
        public long Audit2By { get; set; }
        public bool IsSystem { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;

        // Static wrappers for compatibility
        public static bool IsExists(string question)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return repo.ExistsByQuestionAsync(question).GetAwaiter().GetResult();
        }

        public static long Append(long botUin, long groupId, long qq, string question) 
            => AppendAsync(botUin, groupId, qq, question).GetAwaiter().GetResult();

        public static async Task<long> AppendAsync(long botUin, long groupId, long qq, string question)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return await repo.AddQuestionAsync(botUin, groupId, qq, question);
        }

        public static async Task<int> PlusUsedTimesAsync(long questionId)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return await repo.IncrementUsedTimesAsync(questionId);
        }

        public static bool GetIsSystem(long questionId) 
            => GetIsSystemAsync(questionId).GetAwaiter().GetResult();

        public static async Task<bool> GetIsSystemAsync(long questionId)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return await repo.IsSystemAsync(questionId);
        }

        public static int Audit(long questionId, int audit2, int isSystem) 
            => AuditAsync(questionId, audit2, isSystem).GetAwaiter().GetResult();

        public static async Task<int> AuditAsync(long questionId, int audit2, int isSystem)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return await repo.AuditAsync(questionId, audit2, isSystem);
        }

        public static long GetQId(string text) 
            => GetQIdAsync(text).GetAwaiter().GetResult();

        public static async Task<long> GetQIdAsync(string text)
        {
            var repo = GlobalConfig.ServiceProvider.GetService<IQuestionInfoRepository>();
            return await repo.GetIdByQuestionAsync(text);
        }

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
