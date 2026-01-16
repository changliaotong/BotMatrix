using System.Threading.Tasks;

namespace BotWorker.Domain.Repositories
{
    public interface IIDCRepository
    {
        Task<string?> GetAreaNameAsync(string areaCode);
        Task<string?> GetRandomBmAsync(string? dq = null);
    }
}
