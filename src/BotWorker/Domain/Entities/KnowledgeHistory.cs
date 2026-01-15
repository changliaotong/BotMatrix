using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("KnowledgeHistory")]
    public class KnowledgeHistory
    {
        private static IKnowledgeHistoryRepository Repo => GlobalConfig.ServiceProvider!.GetRequiredService<IKnowledgeHistoryRepository>();

        [Key]
        public long Id { get; set; }
        public string Question { get; set; } = string.Empty;
        public string TargetQuestion { get; set; } = string.Empty;
        public long TargetQuestionId { get; set; }
        public float Similarity { get; set; }
        public long AnswerId { get; set; }
        public string Answer { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static int AddKnowledgeHistroy(string question, string targetQuestion, long targetQuestionId, float similarity, long answerId, string answer)
        {
            return Repo.AddAsync(question, targetQuestion, targetQuestionId, similarity, answerId, answer).GetAwaiter().GetResult();
        }
    }
}
