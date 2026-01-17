using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("knowledge_history")]
    public class KnowledgeHistory
    {
        [Key]
        public long Id { get; set; }
        public string Question { get; set; } = string.Empty;
        public string TargetQuestion { get; set; } = string.Empty;
        public long TargetQuestionId { get; set; }
        public float Similarity { get; set; }
        public long AnswerId { get; set; }
        public string Answer { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static async Task<int> AddKnowledgeHistroyAsync(BotMessage bm, string question, string targetQuestion, long targetQuestionId, float similarity, long answerId, string answer)
        {
            return await bm.KnowledgeHistoryRepository.AddAsync(question, targetQuestion, targetQuestionId, similarity, answerId, answer);
        }
    }
}
