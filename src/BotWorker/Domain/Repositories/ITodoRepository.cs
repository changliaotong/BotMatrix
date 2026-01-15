using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Tools;

namespace BotWorker.Domain.Repositories
{
    public interface ITodoRepository : IBaseRepository<Todo>
    {
        Task<Todo?> GetByNoAsync(long userId, int todoNo);
        Task<int> DeleteByNoAsync(long userId, int todoNo);
        Task<int> GetMaxNoAsync(long userId);
        Task<IEnumerable<Todo>> GetListAsync(long userId);
    }
}
