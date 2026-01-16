using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using System.Data;

namespace BotWorker.Domain.Repositories
{
    public interface IHandleQuestionRepository
    {
        Task<long> InsertAsync(HandleQuestion question, IDbTransaction? trans = null);
        // Add other methods if needed
    }
}
