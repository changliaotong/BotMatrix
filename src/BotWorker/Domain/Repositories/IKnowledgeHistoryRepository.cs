using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IKnowledgeHistoryRepository : IBaseRepository<KnowledgeHistory>
    {
        Task<int> AddAsync(string question, string targetQuestion, long targetQuestionId, float similarity, long answerId, string answer);
    }
}
