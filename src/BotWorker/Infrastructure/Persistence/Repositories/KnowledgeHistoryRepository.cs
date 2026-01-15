using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class KnowledgeHistoryRepository : BaseRepository<KnowledgeHistory>, IKnowledgeHistoryRepository
    {
        public KnowledgeHistoryRepository(string? connectionString = null) : base("KnowledgeHistory", connectionString)
        {
        }

        public async Task<int> AddAsync(string question, string targetQuestion, long targetQuestionId, float similarity, long answerId, string answer)
        {
            var history = new KnowledgeHistory
            {
                Question = question,
                TargetQuestion = targetQuestion,
                TargetQuestionId = targetQuestionId,
                Similarity = similarity,
                AnswerId = answerId,
                Answer = answer,
                InsertDate = DateTime.Now
            };
            return await AddAsync(history);
        }
    }
}
